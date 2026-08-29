package demo_test

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"athenaeum/internal/demo"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func TestSeed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	libraryDir := filepath.Join(dir, "library")
	dataDir := filepath.Join(dir, "data")
	coverDir := filepath.Join(dataDir, "covers")
	if err := os.MkdirAll(libraryDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		t.Fatal(err)
	}

	store, err := storage.Open(filepath.Join(dataDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	if err := store.EnsureDefaultLibrary(ctx, libraryDir); err != nil {
		t.Fatal(err)
	}
	if err := demo.Seed(ctx, store, libraryDir, coverDir, nil); err != nil {
		t.Fatal(err)
	}
	if err := demo.Seed(ctx, store, libraryDir, coverDir, nil); err != nil {
		t.Fatal(err)
	}

	page, err := store.ListBooks(ctx, models.BookQuery{Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	want := len(demo.Catalog())
	if int(page.Total) < want {
		t.Fatalf("got %d books, want at least %d", page.Total, want)
	}

	var withCover int
	for _, b := range page.Items {
		if !b.HasCover {
			continue
		}
		withCover++
		p := filepath.Join(coverDir, strconv.FormatInt(b.ID, 10)+".img")
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing cover for %s: %v", b.Title, err)
		}
	}
	if withCover < want {
		t.Fatalf("covers: got %d want %d", withCover, want)
	}
}

func TestEncodeCoverBuffer(t *testing.T) {
	t.Parallel()
	b, err := demo.EncodeCoverBuffer("The Ember Protocol", "Mira Kade")
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 100 {
		t.Fatalf("cover too small: %d bytes", len(b))
	}
	if b[0] != 0x89 || b[1] != 'P' {
		t.Fatalf("expected PNG header, got %x", b[:4])
	}
}

func TestWriteCoverPNG(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "covers", "book.png")
	if err := demo.WriteCoverPNG(path, "Title", "Author"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 100 || data[0] != 0x89 {
		t.Fatalf("bad png: %d bytes", len(data))
	}
}
