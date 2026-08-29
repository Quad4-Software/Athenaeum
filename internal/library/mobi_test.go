package library

import (
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHTMLEscape(t *testing.T) {
	got := htmlEscape(`a&b<c>d"e`)
	want := `a&amp;b&lt;c&gt;d&quot;e`
	if got != want {
		t.Fatalf("got %q", got)
	}
}

func TestMobiTextToHTML(t *testing.T) {
	html := mobiTextToHTML("Para one\n\nPara <two>", 65001)
	if !strings.Contains(html, "<p>Para one</p>") {
		t.Fatalf("html=%s", html)
	}
	if !strings.Contains(html, "Para &lt;two&gt;") {
		t.Fatalf("escape missing: %s", html)
	}
	single := mobiTextToHTML("line1\nline2", 1252)
	if !strings.Contains(single, "<p>line1</p>") || !strings.Contains(single, "<p>line2</p>") {
		t.Fatalf("single=%s", single)
	}
	badUTF := mobiTextToHTML("ok\xffmore", 65001)
	if !strings.Contains(badUTF, "<p>") {
		t.Fatalf("bad utf=%s", badUTF)
	}
}

func TestPalmDOCDecompress(t *testing.T) {
	lit := palmDOCDecompress([]byte{3, 'a', 'b', 'c', 'x'})
	if string(lit) != "abcx" {
		t.Fatalf("literal=%q", lit)
	}
	null := palmDOCDecompress([]byte{0, 'z'})
	if string(null) != "\x00z" {
		t.Fatalf("null=%q", null)
	}
	empty := palmDOCDecompress(nil)
	if len(empty) != 0 {
		t.Fatalf("empty=%v", empty)
	}
	raw := decompressPalm([]byte("hi"), 1)
	if string(raw) != "hi" {
		t.Fatalf("comp1=%q", raw)
	}
	raw2 := decompressPalm([]byte{3, 'a', 'b', 'c'}, 2)
	if string(raw2) != "abc" {
		t.Fatalf("comp2=%q", raw2)
	}
	raw3 := decompressPalm([]byte("x"), 99)
	if string(raw3) != "x" {
		t.Fatalf("default=%q", raw3)
	}
	if decompressPalm(nil, 2) != nil {
		t.Fatal("empty in")
	}
}

func TestParseEXTH(t *testing.T) {
	var buf []byte
	buf = append(buf, []byte("EXTH")...)
	buf = append(buf, 0, 0, 0, 0)
	count := make([]byte, 4)
	binary.BigEndian.PutUint32(count, 4)
	buf = append(buf, count...)

	add := func(typ uint32, val string) {
		h := make([]byte, 8)
		binary.BigEndian.PutUint32(h[0:4], typ)
		binary.BigEndian.PutUint32(h[4:8], uint32(len(val)))
		buf = append(buf, h...)
		buf = append(buf, val...)
	}
	add(100, "Author Name")
	add(503, "Book Title")
	add(103, "A description")
	add(524, "en")

	var meta mobiMeta
	parseEXTH(buf, &meta)
	if meta.Author != "Author Name" || meta.Title != "Book Title" {
		t.Fatalf("meta=%+v", meta)
	}
	if meta.Description != "A description" || meta.Language != "en" {
		t.Fatalf("meta=%+v", meta)
	}
	parseEXTH([]byte("nope"), &meta)
	parseEXTH(nil, &mobiMeta{})
}

func TestWriteMinimalEPUBNoSections(t *testing.T) {
	src := filepath.Join(t.TempDir(), "empty.mobi")
	if err := os.WriteFile(src, []byte("not palm"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := writeMinimalEPUB(src, filepath.Join(t.TempDir(), "out.epub"))
	if err != io.EOF {
		t.Fatalf("err=%v", err)
	}
}

func TestConvertToEPUBWithoutCalibre(t *testing.T) {
	if calibreAvailable() {
		t.Skip("calibre installed")
	}
	src := filepath.Join(t.TempDir(), "empty.mobi")
	if err := os.WriteFile(src, []byte("not palm"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := ConvertToEPUB(src, filepath.Join(t.TempDir(), "out.epub"))
	if err != io.EOF {
		t.Fatalf("err=%v", err)
	}
}

func TestMobiSectionsEmpty(t *testing.T) {
	src := filepath.Join(t.TempDir(), "x.kfx")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if secs := MobiSections(src); secs != nil {
		t.Fatalf("secs=%v", secs)
	}
}
