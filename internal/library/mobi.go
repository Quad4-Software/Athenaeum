package library

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type mobiMeta struct {
	Title       string
	Author      string
	Language    string
	Description string
	CoverData   []byte
	Sections    []mobiSection
	Readable    bool
}

type mobiSection struct {
	Title string
	HTML  string
}

type palmRecord struct {
	off  uint32
	size uint32
}

type mobiCacheEntry struct {
	modTime time.Time
	size    int64
	meta    mobiMeta
}

const (
	palmDBHeaderLen     = 78
	maxMobiCacheEntries = 8
)

var (
	mobiMetaMu     sync.RWMutex
	mobiMetaByPath = make(map[string]mobiCacheEntry)
	mobiMetaOrder  []string
)

func parseMobiFamily(path string) mobiMeta {
	if strings.ToLower(filepath.Ext(path)) == ".kfx" {
		return mobiMeta{Readable: false}
	}
	fi, err := os.Stat(path)
	if err != nil {
		return mobiMeta{}
	}
	modTime, size := fi.ModTime(), fi.Size()

	mobiMetaMu.RLock()
	if entry, ok := mobiMetaByPath[path]; ok && entry.modTime.Equal(modTime) && entry.size == size {
		meta := entry.meta
		mobiMetaMu.RUnlock()
		return meta
	}
	mobiMetaMu.RUnlock()

	data, err := os.ReadFile(path) // #nosec G304
	if err != nil || len(data) < palmDBHeaderLen {
		return mobiMeta{}
	}
	if bytes.Index(data, []byte("BOOKMOBI")) < 0 && bytes.Index(data, []byte("TEXtREAd")) < 0 {
		return mobiMeta{}
	}
	meta := parseMobiRecords(data)

	mobiMetaMu.Lock()
	mobiMetaByPath[path] = mobiCacheEntry{modTime: modTime, size: size, meta: meta}
	touchMobiMetaOrder(path)
	evictMobiMetaLocked()
	mobiMetaMu.Unlock()
	return meta
}

func touchMobiMetaOrder(path string) {
	for i, p := range mobiMetaOrder {
		if p == path {
			mobiMetaOrder = append(mobiMetaOrder[:i], mobiMetaOrder[i+1:]...)
			break
		}
	}
	mobiMetaOrder = append(mobiMetaOrder, path)
}

func evictMobiMetaLocked() {
	for len(mobiMetaOrder) > maxMobiCacheEntries {
		oldest := mobiMetaOrder[0]
		mobiMetaOrder = mobiMetaOrder[1:]
		delete(mobiMetaByPath, oldest)
	}
}

func parseMobiRecords(data []byte) mobiMeta {
	numRecords := int(binary.BigEndian.Uint16(data[76:78]))
	if numRecords < 1 || numRecords > 50000 {
		return mobiMeta{}
	}
	recordListOff := palmDBHeaderLen
	if len(data) < recordListOff+numRecords*8 {
		return mobiMeta{}
	}
	records := make([]palmRecord, numRecords)
	for i := range numRecords {
		base := recordListOff + i*8
		records[i].off = binary.BigEndian.Uint32(data[base : base+4])
		if i+1 < numRecords {
			next := binary.BigEndian.Uint32(data[base+8 : base+12])
			if next > records[i].off {
				records[i].size = next - records[i].off
			}
		} else {
			remain := min(max(int64(len(data))-int64(records[i].off), 0), int64(^uint32(0)))
			records[i].size = uint32(remain) // #nosec G115 -- remain clamped to max uint32
		}
	}
	rec0 := slicePalmRecord(data, records, 0)
	if len(rec0) < 16 || string(rec0[0:4]) != "MOBI" {
		return mobiMeta{}
	}
	meta := mobiMeta{Readable: true}
	headerLen := int(binary.BigEndian.Uint32(rec0[4:8]))
	if len(rec0) >= 84 {
		exthFlag := binary.BigEndian.Uint32(rec0[80:84])
		if exthFlag&0x40 != 0 && len(rec0) > headerLen+4 && string(rec0[headerLen:headerLen+4]) == "EXTH" {
			parseEXTH(rec0[headerLen:], &meta)
		}
	}
	encoding := uint32(1252)
	if len(rec0) >= 32 {
		encoding = binary.BigEndian.Uint32(rec0[28:32])
	}
	firstText := 1
	lastText := numRecords - 1
	if len(rec0) >= 72 {
		firstText = int(binary.BigEndian.Uint32(rec0[64:68]))
		lastText = int(binary.BigEndian.Uint32(rec0[68:72]))
	}
	if firstText < 1 {
		firstText = 1
	}
	if lastText >= numRecords {
		lastText = numRecords - 1
	}
	comp := byte(2)
	if len(rec0) >= 48 {
		comp = rec0[45]
	}
	var chunks []string
	for i := firstText; i <= lastText && i < numRecords; i++ {
		raw := slicePalmRecord(data, records, i)
		chunks = append(chunks, string(decompressPalm(raw, comp)))
	}
	html := mobiTextToHTML(strings.Join(chunks, ""), encoding)
	if html != "" {
		meta.Sections = []mobiSection{{Title: "Content", HTML: html}}
	}
	return meta
}

func slicePalmRecord(data []byte, records []palmRecord, idx int) []byte {
	if idx < 0 || idx >= len(records) {
		return nil
	}
	off := int(records[idx].off)
	size := int(records[idx].size)
	if off < 0 || off >= len(data) {
		return nil
	}
	end := min(off+size, len(data))
	return data[off:end]
}

func parseEXTH(exth []byte, meta *mobiMeta) {
	if len(exth) < 12 || string(exth[0:4]) != "EXTH" {
		return
	}
	count := int(binary.BigEndian.Uint32(exth[8:12]))
	off := 12
	for i := 0; i < count && off+8 <= len(exth); i++ {
		typ := int(binary.BigEndian.Uint32(exth[off : off+4]))
		ln := int(binary.BigEndian.Uint32(exth[off+4 : off+8]))
		off += 8
		if ln < 0 || off+ln > len(exth) {
			break
		}
		val := strings.TrimSpace(string(exth[off : off+ln]))
		off += ln
		switch typ {
		case 100: // author
			if meta.Author == "" {
				meta.Author = val
			}
		case 503: // title
			if meta.Title == "" {
				meta.Title = val
			}
		case 103: // description
			if meta.Description == "" {
				meta.Description = val
			}
		case 524: // language
			if meta.Language == "" {
				meta.Language = val
			}
		}
	}
}

func decompressPalm(in []byte, comp byte) []byte {
	if len(in) == 0 {
		return nil
	}
	switch comp {
	case 1:
		return append([]byte(nil), in...)
	case 2:
		return palmDOCDecompress(in)
	default:
		return append([]byte(nil), in...)
	}
}

func palmDOCDecompress(in []byte) []byte {
	out := make([]byte, 0, len(in)*2)
	i := 0
	for i < len(in) {
		c := in[i]
		i++
		if c >= 1 && c <= 8 {
			if i+int(c) > len(in) {
				break
			}
			out = append(out, in[i:i+int(c)]...)
			i += int(c)
			continue
		}
		if c >= 0x80 && c <= 0xbf {
			if i >= len(in) {
				break
			}
			d := in[i]
			i++
			dist := (((int(c) << 8) | int(d)) & 0x3fff) + 3
			n := int(c>>6)&0x03 + 3
			if dist > len(out) {
				break
			}
			start := len(out) - dist
			for j := range n {
				out = append(out, out[start+j])
			}
			continue
		}
		if c == 0 {
			out = append(out, 0)
			continue
		}
		out = append(out, c)
	}
	return out
}

func mobiTextToHTML(text string, encoding uint32) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	if encoding == 65001 {
		if !utf8.ValidString(text) {
			text = strings.ToValidUTF8(text, "")
		}
	}
	paras := strings.Split(text, "\n\n")
	if len(paras) == 1 {
		paras = strings.Split(text, "\n")
	}
	var b strings.Builder
	b.WriteString(`<div class="mobi-content">`)
	for _, p := range paras {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		b.WriteString("<p>")
		b.WriteString(htmlEscape(p))
		b.WriteString("</p>")
	}
	b.WriteString("</div>")
	return b.String()
}

func htmlEscape(s string) string {
	repl := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return repl.Replace(s)
}

// MobiSections returns readable HTML sections for API responses.
func MobiSections(path string) []mobiSection {
	return parseMobiFamily(path).Sections
}

// ConvertToEPUB writes an EPUB using calibre when available, else a minimal zip EPUB.
func ConvertToEPUB(srcPath, destPath string) error {
	if err := convertWithCalibre(srcPath, destPath, "epub"); err == nil {
		return nil
	}
	return writeMinimalEPUB(srcPath, destPath)
}

func writeMinimalEPUB(srcPath, destPath string) error {
	meta := parseMobiFamily(srcPath)
	if len(meta.Sections) == 0 {
		return io.EOF
	}
	title := meta.Title
	if title == "" {
		title = filepath.Base(srcPath)
	}
	author := meta.Author
	body := meta.Sections[0].HTML
	return writeSimpleEPUB(destPath, title, author, body)
}
