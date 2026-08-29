package libfs

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFSRoundTrip(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sub := filepath.Join(root, "books")
	if err := os.MkdirAll(sub, 0o750); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(sub, "a.epub")
	if err := os.WriteFile(src, []byte("epub-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}

	fs, err := Open(Config{Backend: BackendLocal, Path: root})
	if err != nil {
		t.Fatal(err)
	}
	if fs.Backend() != BackendLocal {
		t.Fatalf("backend %q", fs.Backend())
	}

	ctx := context.Background()
	info, err := fs.Stat(ctx, "books/a.epub")
	if err != nil {
		t.Fatal(err)
	}
	if info.Size != 10 {
		t.Fatalf("size %d", info.Size)
	}

	var seen []string
	if err := fs.Walk(ctx, func(fi FileInfo) error {
		seen = append(seen, fi.RelPath)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 1 || seen[0] != "books/a.epub" {
		t.Fatalf("walk %#v", seen)
	}

	rc, err := fs.Open(ctx, "books/a.epub")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "epub-bytes" {
		t.Fatalf("got %q", data)
	}

	if err := fs.Write(ctx, "books/b.epub", bytes.NewReader([]byte("new")), 3); err != nil {
		t.Fatal(err)
	}
	if err := fs.Remove(ctx, "books/a.epub"); err != nil {
		t.Fatal(err)
	}
	if _, err := fs.Stat(ctx, "books/a.epub"); err != ErrNotExist {
		t.Fatalf("expected ErrNotExist, got %v", err)
	}
}

func TestLocalFSRejectsEscape(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	fs, err := Open(Config{Backend: BackendLocal, Path: root})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.Stat(context.Background(), "../outside"); err == nil {
		t.Fatal("expected escape error")
	}
}

func TestMountLabel(t *testing.T) {
	t.Parallel()
	if got := MountLabel("b", ""); got != "s3://b" {
		t.Fatalf("got %q", got)
	}
	if got := MountLabel("b", "p/"); got != "s3://b/p" {
		t.Fatalf("got %q", got)
	}
}

func TestMaterialize(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "x.pdf"), []byte("pdf"), 0o600); err != nil {
		t.Fatal(err)
	}
	fs, err := Open(Config{Backend: BackendLocal, Path: root})
	if err != nil {
		t.Fatal(err)
	}
	tmp, err := Materialize(context.Background(), fs, "x.pdf", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp)
	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "pdf" {
		t.Fatalf("got %q", data)
	}
}
