package library

import (
	"os"
	"path/filepath"
	"testing"
)

func frameBytes(id string, data []byte) []byte {
	size := syncSafe(len(data))
	f := append([]byte(id), size...)
	f = append(f, 0x00, 0x00)
	f = append(f, data...)
	return f
}

func syncSafe(n int) []byte {
	return []byte{
		byte(n>>21) & 0x7f,
		byte(n>>14) & 0x7f,
		byte(n>>7) & 0x7f,
		byte(n) & 0x7f,
	}
}

func writeTestID3(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tagged.mp3")

	jpeg := make([]byte, minCoverBytes)
	jpeg[0], jpeg[1], jpeg[2] = 0xff, 0xd8, 0xff
	jpeg[len(jpeg)-2], jpeg[len(jpeg)-1] = 0xff, 0xd9

	title := frameBytes("TIT2", []byte{0x00, 'T', 'a', 'g', 'g', 'e', 'd', ' ', 'T', 'i', 't', 'l', 'e'})
	artist := frameBytes("TPE1", []byte{0x00, 'T', 'a', 'g', 'g', 'e', 'd', ' ', 'A', 'r', 't', 'i', 's', 't'})
	album := frameBytes("TALB", []byte{0x00, 'T', 'a', 'g', 'g', 'e', 'd', ' ', 'A', 'l', 'b', 'u', 'm'})
	apicBody := append([]byte{0x00}, []byte("image/jpeg\x00")...)
	apicBody = append(apicBody, 0x03, 0x00)
	apicBody = append(apicBody, jpeg...)
	art := frameBytes("APIC", apicBody)

	body := append(title, artist...)
	body = append(body, album...)
	body = append(body, art...)
	tagSize := syncSafe(len(body))

	var out []byte
	out = append(out, []byte("ID3")...)
	out = append(out, 0x04, 0x00, 0x00)
	out = append(out, tagSize...)
	out = append(out, body...)
	out = append(out, 0x00, 0x00)

	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseAudioID3(t *testing.T) {
	path := writeTestID3(t)
	meta := parseAudio(path)
	if meta.Title != "Tagged Title" {
		t.Errorf("title = %q", meta.Title)
	}
	if meta.Author != "Tagged Artist" {
		t.Errorf("author = %q", meta.Author)
	}
	if len(meta.CoverData) == 0 {
		t.Error("expected embedded cover")
	}
}

func TestParseAudioFallback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plain_book.mp3")
	if err := os.WriteFile(path, []byte{0, 1, 2}, 0o644); err != nil {
		t.Fatal(err)
	}
	meta := parseAudio(path)
	if meta.Title != "plain book" {
		t.Errorf("title = %q", meta.Title)
	}
}
