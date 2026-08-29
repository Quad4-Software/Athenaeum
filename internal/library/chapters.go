package library

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"os"
	"sort"
	"strings"

	"athenaeum/internal/models"
	"github.com/dhowden/tag"
)

type chapterMark struct {
	title   string
	startMS uint32
}

// parseAudioChapters extracts chapter markers from M4B/MP4 or ID3-tagged audio.
func parseAudioChapters(filePath string) []models.Chapter {
	if ch := parseMP4Chapters(filePath); len(ch) > 0 {
		return ch
	}
	return parseID3Chapters(filePath)
}

func toModelChapters(marks []chapterMark) []models.Chapter {
	if len(marks) == 0 {
		return nil
	}
	sort.Slice(marks, func(i, j int) bool {
		if marks[i].startMS == marks[j].startMS {
			return marks[i].title < marks[j].title
		}
		return marks[i].startMS < marks[j].startMS
	})
	out := make([]models.Chapter, len(marks))
	for i, m := range marks {
		title := strings.TrimSpace(m.title)
		if title == "" {
			title = "Chapter " + chapterItoa(i+1)
		}
		out[i] = models.Chapter{
			Index:    i,
			Title:    title,
			StartSec: float64(m.startMS) / 1000,
		}
	}
	return out
}

func chapterItoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func parseID3Chapters(filePath string) []models.Chapter {
	data, err := os.ReadFile(filePath) // #nosec G304
	if err != nil || len(data) < 10 || string(data[:3]) != "ID3" {
		return nil
	}
	size := syncSafeInt(data[6:10])
	if 10+size > len(data) {
		return nil
	}
	body := data[10 : 10+size]
	var marks []chapterMark
	off := 0
	for off+10 <= len(body) {
		id := string(body[off : off+4])
		fsize := syncSafeInt(body[off+4 : off+8])
		off += 10
		if fsize < 0 || off+fsize > len(body) {
			break
		}
		payload := body[off : off+fsize]
		off += fsize
		if id != "CHAP" {
			continue
		}
		if m, ok := parseCHAPFrame(payload); ok {
			marks = append(marks, m)
		}
	}
	return toModelChapters(marks)
}

func parseCHAPFrame(payload []byte) (chapterMark, bool) {
	if len(payload) < 5 {
		return chapterMark{}, false
	}
	n := bytes.IndexByte(payload, 0)
	if n < 0 {
		return chapterMark{}, false
	}
	if n+1+8 > len(payload) {
		return chapterMark{}, false
	}
	startMS := binary.BigEndian.Uint32(payload[n+1:])
	title := extractCHAPTitle(payload[n+1+8:])
	return chapterMark{title: title, startMS: startMS}, true
}

func extractCHAPTitle(rest []byte) string {
	if len(rest) < 8 {
		return ""
	}
	sub := rest[8:]
	for len(sub) >= 10 {
		id := string(sub[0:4])
		fsize := syncSafeInt(sub[4:8])
		sub = sub[10:]
		if fsize < 0 || len(sub) < fsize {
			break
		}
		payload := sub[:fsize]
		sub = sub[fsize:]
		if id == "TIT2" && len(payload) > 1 {
			return strings.TrimSpace(string(payload[1:]))
		}
	}
	return ""
}

func syncSafeInt(b []byte) int {
	if len(b) < 4 {
		return 0
	}
	return int(b[0]&0x7f)<<21 | int(b[1]&0x7f)<<14 | int(b[2]&0x7f)<<7 | int(b[3]&0x7f)
}

func parseMP4Chapters(filePath string) []models.Chapter {
	f, err := os.Open(filePath) // #nosec G304
	if err != nil {
		return nil
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil || info.Size() < 16 {
		return nil
	}
	ra := io.NewSectionReader(f, 0, info.Size())
	var chpl []byte
	walkAtoms(ra, info.Size(), func(name string, data []byte) {
		if name == "chpl" && len(data) > 0 {
			chpl = append([]byte(nil), data...)
		}
	})
	if len(chpl) == 0 {
		return nil
	}
	return toModelChapters(parseCHPLAtom(chpl))
}

func walkAtoms(r io.ReaderAt, size int64, fn func(name string, data []byte)) {
	var off int64
	for off+8 <= size {
		hdr := make([]byte, 8)
		if _, err := r.ReadAt(hdr, off); err != nil {
			return
		}
		atomSize := int64(binary.BigEndian.Uint32(hdr[0:4]))
		name := string(hdr[4:8])
		if atomSize < 8 {
			return
		}
		headerSize := int64(8)
		if atomSize == 1 && off+16 <= size {
			ext := make([]byte, 8)
			if _, err := r.ReadAt(ext, off+8); err != nil {
				return
			}
			extSize := binary.BigEndian.Uint64(ext)
			if extSize > math.MaxInt64 {
				return
			}
			atomSize = int64(extSize)
			headerSize = 16
		}
		end := off + atomSize
		if end > size {
			return
		}
		dataOff := off + headerSize
		switch name {
		case "moov", "trak", "mdia", "minf", "stbl", "udta", "meta":
			walkAtoms(io.NewSectionReader(r, dataOff, end-dataOff), end-dataOff, fn)
		default:
			if name == "chpl" {
				buf := make([]byte, end-dataOff)
				if _, err := r.ReadAt(buf, dataOff); err == nil {
					fn(name, buf)
				}
			}
		}
		off = end
	}
}

func parseCHPLAtom(data []byte) []chapterMark {
	if len(data) < 5 {
		return nil
	}
	off := 4
	var out []chapterMark
	for off < len(data) {
		n := int(data[off])
		off++
		if n <= 0 || off+n+8 > len(data) {
			break
		}
		title := string(data[off : off+n])
		off += n
		startMS := binary.BigEndian.Uint64(data[off : off+8])
		off += 8
		var markMS uint32
		if startMS > math.MaxUint32 {
			markMS = math.MaxUint32
		} else {
			markMS = uint32(startMS)
		}
		out = append(out, chapterMark{title: title, startMS: markMS})
	}
	return out
}

func parseTagIdentifiers(m tag.Metadata) (isbn, asin string) {
	raw := m.Raw()
	for k, v := range raw {
		kl := strings.ToLower(k)
		val := tagString(v)
		switch kl {
		case "isbn":
			isbn = val
		case "asin", "audible_asin", "audible asin":
			asin = val
		}
		if isbn == "" && strings.Contains(kl, "isbn") {
			isbn = val
		}
		if asin == "" && strings.Contains(kl, "asin") {
			asin = val
		}
	}
	return strings.TrimSpace(isbn), strings.TrimSpace(asin)
}

func tagString(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case []string:
		if len(t) > 0 {
			return strings.TrimSpace(t[0])
		}
	}
	return ""
}
