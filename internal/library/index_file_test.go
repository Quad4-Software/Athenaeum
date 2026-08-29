package library

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/storage"
)

func TestStringsTrimJoin(t *testing.T) {
	got := stringsTrimJoin("s3://bucket/", "path/to/file.epub")
	if got != "s3://bucket/path/to/file.epub" {
		t.Fatalf("got=%q", got)
	}
	got = stringsTrimJoin("s3://bucket", "a.epub")
	if got != "s3://bucket/a.epub" {
		t.Fatalf("got=%q", got)
	}
}

func TestIndexFile(t *testing.T) {
	libDir := t.TempDir()
	coverDir := filepath.Join(t.TempDir(), "covers")
	epubSrc := writeTestEPUB(t)
	copyFile(t, epubSrc, filepath.Join(libDir, "one.epub"))

	store, err := storage.Open(filepath.Join(t.TempDir(), "s.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.EnsureDefaultLibrary(ctx, libDir); err != nil {
		t.Fatal(err)
	}
	libs, err := store.ListLibraries(ctx)
	if err != nil || len(libs) == 0 {
		t.Fatalf("libs=%v err=%v", libs, err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	sc := New(store, coverDir, t.TempDir(), log, 1)
	id, err := sc.IndexFile(ctx, libs[0].ID, "one.epub")
	if err != nil {
		t.Fatal(err)
	}
	if id <= 0 {
		t.Fatalf("id=%d", id)
	}
	book, err := store.GetBook(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if book.RelPath != "one.epub" {
		t.Fatalf("rel=%q", book.RelPath)
	}

	if _, err := sc.IndexFile(ctx, libs[0].ID, "missing.epub"); err == nil {
		t.Fatal("expected missing error")
	}
	if err := os.MkdirAll(filepath.Join(libDir, "subdir"), 0o750); err != nil {
		t.Fatal(err)
	}
	if _, err := sc.IndexFile(ctx, libs[0].ID, "subdir"); err == nil {
		t.Fatal("expected dir error")
	}
}
