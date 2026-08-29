package library

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func TestCleanDisplayText(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"  Hello   World  ", "Hello World"},
		{"Coffee consumption and\x00migraine", "Coffee consumption and migraine"},
		{"Line\u200bone", "Line one"},
		{"Bad\uFFFDtext", "Bad text"},
	}
	for _, tc := range tests {
		got := CleanDisplayText(tc.in)
		if got != tc.want {
			t.Errorf("CleanDisplayText(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestIsGarbledText(t *testing.T) {
	if IsGarbledText("Masters of Doom") {
		t.Fatal("expected readable title")
	}
	if IsGarbledText("RAND RR2742") {
		t.Fatal("expected readable title")
	}
	if !IsGarbledText("1\uFFFD\uFFFD\uFFFD\uFFFDbS\\b~") {
		t.Fatal("expected garbled title")
	}
	if !IsGarbledText("0\uFFFD5\uFFFD1\uFFFD\uFFFD\uFFFD1v\uFFFD") {
		t.Fatal("expected garbled title")
	}
}

func TestCleanBookTitleFallsBackToFilename(t *testing.T) {
	got := CleanBookTitle("1\uFFFD\uFFFD\uFFFD\uFFFDbS\\b~", "/books/Masters of Doom.pdf")
	if got != "Masters of Doom" {
		t.Fatalf("title = %q", got)
	}
}

func TestCleanStoredBookText(t *testing.T) {
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
		Title:     "1\uFFFD\uFFFD\uFFFD\uFFFDbS\\b~",
		Author:    "???",
		Format:    models.FormatPDF,
		RelPath:   "Masters of Doom.pdf",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}

	updated, err := CleanStoredBookText(ctx, store)
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
	if book.Title != "Masters of Doom" {
		t.Fatalf("title=%q", book.Title)
	}
}
