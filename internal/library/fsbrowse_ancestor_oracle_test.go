package library

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// PROVED_FSBROWSE_ANCESTOR_JAIL
// Guarantee: BrowseDirs refuses paths that are ancestors of roots
// (siblings outside the mount must not be listable by walking Parent).

func TestBrowseDirsRejectsAncestorOfRoot(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "mount")
	secret := filepath.Join(base, "secret")
	if err := os.MkdirAll(filepath.Join(root, "books"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(secret, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := BrowseDirs([]string{root}, base); err == nil {
		t.Fatal("expected ancestor of root to be denied")
	} else {
		fmt.Println("PROVED_FSBROWSE_ANCESTOR_JAIL: ancestor denied:", err)
	}

	if _, err := BrowseDirs([]string{root}, secret); err == nil {
		t.Fatal("expected sibling of root to be denied")
	}
}

func TestBrowseDirsAllowsUnderRoot(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "mount")
	child := filepath.Join(root, "books")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := BrowseDirs([]string{root}, child)
	if err != nil {
		t.Fatal(err)
	}
	if res.Path != child {
		t.Fatalf("path=%q", res.Path)
	}
	if res.Parent != root {
		t.Fatalf("parent=%q want %q", res.Parent, root)
	}
}

func TestBrowseDirsRootHasNoParentOutsideJail(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "mount")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := BrowseDirs([]string{root}, root)
	if err != nil {
		t.Fatal(err)
	}
	if res.Parent != "" {
		t.Fatalf("parent should be empty at root, got %q", res.Parent)
	}
}
