package library

import (
	"os"
	"testing"
)

func TestParseFilenameMeta(t *testing.T) {
	tests := []struct {
		path   string
		title  string
		author string
	}{
		{"/books/Andy Weir - The Martian.pdf", "The Martian", "Andy Weir"},
		{"/books/The_Martian.epub", "The Martian", ""},
		{"/books/[2020] Neuromancer - Gibson.epub", "Neuromancer", "Gibson"},
	}
	for _, tc := range tests {
		meta := parseFilenameMeta(tc.path)
		if meta.Title != tc.title {
			t.Errorf("parseFilenameMeta(%q) title = %q, want %q", tc.path, meta.Title, tc.title)
		}
		if meta.Author != tc.author {
			t.Errorf("parseFilenameMeta(%q) author = %q, want %q", tc.path, meta.Author, tc.author)
		}
	}
}

func TestSidecarCover(t *testing.T) {
	dir := t.TempDir()
	pdfPath := dir + "/story.pdf"
	if err := writeTestFile(pdfPath, []byte("%PDF-1.4")); err != nil {
		t.Fatal(err)
	}
	coverPath := dir + "/cover.jpg"
	jpeg := make([]byte, 3000)
	jpeg[0], jpeg[1], jpeg[2] = 0xff, 0xd8, 0xff
	jpeg[len(jpeg)-2], jpeg[len(jpeg)-1] = 0xff, 0xd9
	if err := writeTestFile(coverPath, jpeg); err != nil {
		t.Fatal(err)
	}
	got := sidecarCover(pdfPath)
	if len(got) != len(jpeg) {
		t.Fatalf("sidecar len=%d want=%d", len(got), len(jpeg))
	}
}

func writeTestFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}
