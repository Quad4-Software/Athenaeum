package library

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func buildMinimalPalmDB(titleText string) []byte {
	const numRecords = 2
	rec0 := make([]byte, 88)
	copy(rec0[0:4], "MOBI")
	binary.BigEndian.PutUint32(rec0[4:8], 88)
	binary.BigEndian.PutUint32(rec0[28:32], 65001)
	rec0[45] = 1
	binary.BigEndian.PutUint32(rec0[64:68], 1)
	binary.BigEndian.PutUint32(rec0[68:72], 1)

	text := []byte(titleText)
	recListOff := palmDBHeaderLen
	rec0Off := uint32(recListOff + numRecords*8)
	rec1Off := rec0Off + uint32(len(rec0))

	hdr := make([]byte, palmDBHeaderLen)
	copy(hdr[0:], "MiniMobi")
	copy(hdr[60:], "BOOKMOBI")
	binary.BigEndian.PutUint16(hdr[76:78], numRecords)

	out := make([]byte, 0, int(rec1Off)+len(text))
	out = append(out, hdr...)
	e0 := make([]byte, 8)
	binary.BigEndian.PutUint32(e0[0:4], rec0Off)
	out = append(out, e0...)
	e1 := make([]byte, 8)
	binary.BigEndian.PutUint32(e1[0:4], rec1Off)
	out = append(out, e1...)
	out = append(out, rec0...)
	out = append(out, text...)
	return out
}

func TestParseMobiRecordsAndCache(t *testing.T) {
	mobiMetaMu.Lock()
	mobiMetaByPath = make(map[string]mobiCacheEntry)
	mobiMetaOrder = nil
	mobiMetaMu.Unlock()

	dir := t.TempDir()
	path := filepath.Join(dir, "sample.mobi")
	data := buildMinimalPalmDB("Hello mobi\n\nSecond paragraph")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	meta := parseMobiFamily(path)
	if !meta.Readable {
		t.Fatalf("meta=%+v", meta)
	}
	if len(meta.Sections) != 1 || !strings.Contains(meta.Sections[0].HTML, "Hello mobi") {
		t.Fatalf("sections=%+v", meta.Sections)
	}

	meta2 := parseMobiFamily(path)
	if !meta2.Readable || len(meta2.Sections) != 1 {
		t.Fatalf("cached meta=%+v", meta2)
	}

	direct := parseMobiRecords(data)
	if !direct.Readable {
		t.Fatalf("direct=%+v", direct)
	}

	records := []palmRecord{{off: 0, size: 4}, {off: 4, size: 100}}
	if slicePalmRecord(data, records, -1) != nil {
		t.Fatal("bad idx")
	}
	if slicePalmRecord(data, records, 99) != nil {
		t.Fatal("oob idx")
	}
	if slicePalmRecord([]byte("abcd"), []palmRecord{{off: 100, size: 1}}, 0) != nil {
		t.Fatal("off past end")
	}

	extraDir := filepath.Join(dir, "extra")
	if err := os.MkdirAll(extraDir, 0o750); err != nil {
		t.Fatal(err)
	}
	for i := range maxMobiCacheEntries + 2 {
		p := filepath.Join(extraDir, "book-"+string(rune('a'+i))+".mobi")
		if err := os.WriteFile(p, buildMinimalPalmDB("cache body"), 0o644); err != nil {
			t.Fatal(err)
		}
		_ = parseMobiFamily(p)
	}
	mobiMetaMu.RLock()
	n := len(mobiMetaByPath)
	mobiMetaMu.RUnlock()
	if n > maxMobiCacheEntries {
		t.Fatalf("cache size=%d", n)
	}

	secs := MobiSections(path)
	if len(secs) != 1 {
		t.Fatalf("MobiSections=%v", secs)
	}

	emptyPath := filepath.Join(dir, "tiny.mobi")
	if err := os.WriteFile(emptyPath, []byte("short"), 0o644); err != nil {
		t.Fatal(err)
	}
	if parseMobiFamily(emptyPath).Readable {
		t.Fatal("expected empty for short file")
	}
	badNum := make([]byte, palmDBHeaderLen)
	copy(badNum[60:], "BOOKMOBI")
	binary.BigEndian.PutUint16(badNum[76:78], 0)
	if parseMobiRecords(badNum).Readable {
		t.Fatal("zero records")
	}
	noMobi := make([]byte, palmDBHeaderLen+16)
	copy(noMobi[60:], "BOOKMOBI")
	binary.BigEndian.PutUint16(noMobi[76:78], 1)
	binary.BigEndian.PutUint32(noMobi[palmDBHeaderLen:palmDBHeaderLen+4], uint32(palmDBHeaderLen+8))
	copy(noMobi[palmDBHeaderLen+8:], "XXXX")
	if parseMobiRecords(noMobi).Readable {
		t.Fatal("expected non-MOBI header reject")
	}
}
