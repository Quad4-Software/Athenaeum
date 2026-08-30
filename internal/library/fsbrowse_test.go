package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathAllowedDotfileNotTraversal(t *testing.T) {
	// Filenames that contain ".." as a prefix of the name (not a parent segment)
	// must stay allowed under the root.
	if !pathAllowed("/lib/ety/..rarc", []string{"/lib"}) {
		t.Fatal("expected /lib/ety/..rarc under /lib")
	}
}

func TestBrowseDirsRoots(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "books")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := BrowseDirs([]string{root}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Entries) != 1 || res.Entries[0].Path != root {
		t.Fatalf("entries = %+v", res.Entries)
	}
}

func TestBrowseDirsChildren(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "beta"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := BrowseDirs([]string{root}, root)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Entries) != 2 {
		t.Fatalf("entries = %d", len(res.Entries))
	}
}

func TestBrowseDirsDenied(t *testing.T) {
	root := t.TempDir()
	other := t.TempDir()
	if _, err := BrowseDirs([]string{root}, other); err == nil {
		t.Fatal("expected error for path outside roots")
	}
}
