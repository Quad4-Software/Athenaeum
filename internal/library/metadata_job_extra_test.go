package library

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func TestWriteCoverBytes(t *testing.T) {
	dir := t.TempDir()
	data := []byte("cover-bytes")
	if err := writeCoverBytes(dir, 42, data); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(CoverPath(dir, 42))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Fatalf("got=%q", got)
	}
}

func TestCleanStoredSeriesNames(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	store, err := storage.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	libDir := filepath.Join(dir, "lib")
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}
	lib, err := store.CreateLibrary(ctx, "Main", libDir)
	if err != nil {
		t.Fatal(err)
	}
	id, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Book",
		Series:    "[www.ebookism.net] the_expanse",
		Format:    models.FormatEPUB,
		RelPath:   "book.epub",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}

	updated, err := CleanStoredSeriesNames(ctx, store)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 1 {
		t.Fatalf("updated=%d", updated)
	}
	book, err := store.GetBook(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if book.Series != "the expanse" {
		t.Fatalf("series=%q", book.Series)
	}

	updated, err = CleanStoredSeriesNames(ctx, store)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 0 {
		t.Fatalf("second pass updated=%d", updated)
	}
}
