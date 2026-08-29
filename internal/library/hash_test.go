package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileSHA256(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bin")
	data := []byte("hello duplicate detection")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	h1, err := FileSHA256(path)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := FileSHA256(path)
	if err != nil {
		t.Fatal(err)
	}
	if h1 == "" || h1 != h2 {
		t.Fatalf("hash = %q, want stable non-empty", h1)
	}
}
