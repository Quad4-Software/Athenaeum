package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalibreMissingPaths(t *testing.T) {
	if calibreAvailable() {
		t.Skip("calibre installed")
	}
	src := filepath.Join(t.TempDir(), "book.mobi")
	if err := os.WriteFile(src, []byte("not-mobi"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ConvertBook(src, t.TempDir(), "pdf")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "conversion requires calibre for this target format" {
		t.Fatalf("err=%v", err)
	}
	_, err = ConvertBook(src, t.TempDir(), "epub")
	if err == nil {
		t.Fatal("expected epub convert error without readable mobi")
	}
	err = calibreConvert("x.mobi", filepath.Join(t.TempDir(), "o.epub"), "epub")
	if err == nil || err.Error() != "calibre not installed" {
		t.Fatalf("err=%v", err)
	}
}
