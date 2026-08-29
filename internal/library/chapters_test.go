package library

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func writeTestID3WithChapters(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "chapters.mp3")

	ch1Body := []byte("ch1\x00")
	start := make([]byte, 8)
	binary.BigEndian.PutUint32(start[0:4], 0)
	binary.BigEndian.PutUint32(start[4:8], 60000)
	ch1Body = append(ch1Body, start...)
	ch1Body = append(ch1Body, make([]byte, 8)...)
	tit2Body := []byte{0x00, 'O', 'n', 'e'}
	tit2 := append([]byte("TIT2"), syncSafe(len(tit2Body))...)
	tit2 = append(tit2, 0x00, 0x00)
	tit2 = append(tit2, tit2Body...)
	ch1Body = append(ch1Body, tit2...)
	ch1 := frameBytes("CHAP", ch1Body)

	ch2Body := []byte("ch2\x00")
	binary.BigEndian.PutUint32(start[0:4], 125000)
	ch2Body = append(ch2Body, start...)
	ch2Body = append(ch2Body, make([]byte, 8)...)
	tit2bBody := []byte{0x00, 'T', 'w', 'o'}
	tit2b := append([]byte("TIT2"), syncSafe(len(tit2bBody))...)
	tit2b = append(tit2b, 0x00, 0x00)
	tit2b = append(tit2b, tit2bBody...)
	ch2Body = append(ch2Body, tit2b...)
	ch2 := frameBytes("CHAP", ch2Body)

	body := append(ch1, ch2...)
	tagSize := syncSafe(len(body))

	var out []byte
	out = append(out, []byte("ID3")...)
	out = append(out, 0x04, 0x00, 0x00)
	out = append(out, tagSize...)
	out = append(out, body...)
	out = append(out, 0x00)

	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func buildAtom(name string, data []byte) []byte {
	size := 8 + len(data)
	out := make([]byte, size)
	binary.BigEndian.PutUint32(out[0:4], uint32(size))
	copy(out[4:8], []byte(name))
	copy(out[8:], data)
	return out
}

func buildCHPLAtom(titles []string, startsMS []uint64) []byte {
	var data []byte
	data = append(data, 0, 0, 0, 0)
	for i, title := range titles {
		data = append(data, byte(len(title)))
		data = append(data, []byte(title)...)
		ts := make([]byte, 8)
		binary.BigEndian.PutUint64(ts, startsMS[i])
		data = append(data, ts...)
	}
	return buildAtom("chpl", data)
}

func writeTestM4BWithChapters(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "chapters.m4b")
	chpl := buildCHPLAtom([]string{"Intro", "Part 2"}, []uint64{0, 90000})
	moov := buildAtom("moov", chpl)
	ftyp := buildAtom("ftyp", []byte("M4A \x00\x00\x00\x00M4A mp42"))
	file := append(ftyp, moov...)
	if err := os.WriteFile(path, file, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractCHAPTitleDirect(t *testing.T) {
	rest := make([]byte, 8)
	tit2Body := []byte{0x00, 'O', 'n', 'e'}
	tit2 := append([]byte("TIT2"), syncSafe(len(tit2Body))...)
	tit2 = append(tit2, 0x00, 0x00)
	tit2 = append(tit2, tit2Body...)
	rest = append(rest, tit2...)
	title := extractCHAPTitle(rest)
	if title != "One" {
		t.Fatalf("title = %q rest=% x", title, rest)
	}
}

func TestParseCHAPFrameDirect(t *testing.T) {
	ch1Body := []byte("ch1\x00")
	start := make([]byte, 8)
	binary.BigEndian.PutUint32(start[0:4], 0)
	binary.BigEndian.PutUint32(start[4:8], 60000)
	ch1Body = append(ch1Body, start...)
	ch1Body = append(ch1Body, make([]byte, 8)...)
	tit2Body := []byte{0x00, 'O', 'n', 'e'}
	tit2 := append([]byte("TIT2"), syncSafe(len(tit2Body))...)
	tit2 = append(tit2, 0x00, 0x00)
	tit2 = append(tit2, tit2Body...)
	ch1Body = append(ch1Body, tit2...)
	m, ok := parseCHAPFrame(ch1Body)
	if !ok {
		t.Fatal("parseCHAPFrame failed")
	}
	if m.title != "One" {
		t.Fatalf("title = %q", m.title)
	}
}

func TestChapterItoa(t *testing.T) {
	cases := map[int]string{0: "0", 1: "1", 10: "10", 99: "99", 1234: "1234"}
	for n, want := range cases {
		if got := chapterItoa(n); got != want {
			t.Errorf("chapterItoa(%d)=%q want %q", n, got, want)
		}
	}
	marks := toModelChapters([]chapterMark{{title: " ", startMS: 0}, {title: "", startMS: 1000}})
	if len(marks) != 2 || marks[0].Title != "Chapter 1" || marks[1].Title != "Chapter 2" {
		t.Fatalf("marks=%+v", marks)
	}
}

func TestParseID3Chapters(t *testing.T) {
	path := writeTestID3WithChapters(t)
	ch := parseID3Chapters(path)
	if len(ch) != 2 {
		t.Fatalf("chapters = %d, want 2", len(ch))
	}
	if ch[0].Title != "One" || ch[0].StartSec != 0 {
		t.Errorf("ch[0] = %+v", ch[0])
	}
	if ch[1].Title != "Two" || ch[1].StartSec != 125 {
		t.Errorf("ch[1] = %+v", ch[1])
	}
}

func TestParseMP4Chapters(t *testing.T) {
	path := writeTestM4BWithChapters(t)
	ch := parseMP4Chapters(path)
	if len(ch) != 2 {
		t.Fatalf("chapters = %d, want 2", len(ch))
	}
	if ch[0].Title != "Intro" {
		t.Errorf("ch[0].Title = %q", ch[0].Title)
	}
	if ch[1].StartSec != 90 {
		t.Errorf("ch[1].StartSec = %v, want 90", ch[1].StartSec)
	}
}
