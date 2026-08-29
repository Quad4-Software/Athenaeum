package library

import (
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/models"
)

func TestSidecarMetadataJSON(t *testing.T) {
	dir := t.TempDir()
	book := filepath.Join(dir, "audiobook.m4b")
	if err := os.WriteFile(book, []byte{1}, 0o644); err != nil {
		t.Fatal(err)
	}
	metaPath := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(metaPath, []byte(`{
		"title": "Sidecar Title",
		"author": "Sidecar Author",
		"isbn": "9780143127741",
		"asin": "B00TEST123"
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	side := sidecarMetadata(book)
	if side.Title != "Sidecar Title" || side.Author != "Sidecar Author" {
		t.Fatalf("side = %+v", side)
	}
	if side.ISBN != "9780143127741" || side.ASIN != "B00TEST123" {
		t.Fatalf("identifiers = %+v", side)
	}

	bookModel := &models.Book{}
	isbn, asin := applySidecarMeta(bookModel, side)
	if bookModel.Title != "Sidecar Title" {
		t.Errorf("title = %q", bookModel.Title)
	}
	if isbn != "9780143127741" || asin != "B00TEST123" {
		t.Errorf("isbn=%q asin=%q", isbn, asin)
	}
}

func TestSidecarMetadataOPF(t *testing.T) {
	dir := t.TempDir()
	book := filepath.Join(dir, "book.mp3")
	if err := os.WriteFile(book, []byte{1}, 0o644); err != nil {
		t.Fatal(err)
	}
	opf := `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>OPF Title</dc:title>
    <dc:creator>OPF Author</dc:creator>
    <dc:language>en</dc:language>
    <dc:description>A test description</dc:description>
    <meta property="belongs-to-collection">Test Series</meta>
    <meta property="group-position">2</meta>
  </metadata>
</package>`
	if err := os.WriteFile(filepath.Join(dir, "book.opf"), []byte(opf), 0o644); err != nil {
		t.Fatal(err)
	}

	side := sidecarMetadata(book)
	if side.Title != "OPF Title" || side.Author != "OPF Author" {
		t.Fatalf("side = %+v", side)
	}
	if side.Series != "Test Series" || side.SeriesIndex != 2 {
		t.Fatalf("series = %+v", side)
	}
}

func TestApplyLookupMetaOnlyFillsGaps(t *testing.T) {
	book := &models.Book{Title: "Existing", Author: ""}
	applyLookupMeta(book, sidecarFields{
		Title:       "Lookup Title",
		Author:      "Lookup Author",
		Description: "Desc",
	})
	if book.Title != "Existing" {
		t.Errorf("title = %q", book.Title)
	}
	if book.Author != "Lookup Author" || book.Description != "Desc" {
		t.Errorf("book = %+v", book)
	}
}
