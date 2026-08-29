package library

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func TestParseCBRInvalid(t *testing.T) {
	meta := parseCBR(filepath.Join(t.TempDir(), "missing.cbr"))
	if meta.PageCount != 0 {
		t.Fatalf("missing=%+v", meta)
	}
	bad := filepath.Join(t.TempDir(), "bad.cbr")
	if err := os.WriteFile(bad, []byte("not-rar"), 0o644); err != nil {
		t.Fatal(err)
	}
	meta = parseCBR(bad)
	if meta.PageCount != 0 {
		t.Fatalf("bad=%+v", meta)
	}
	_, _, err := openCBRPage(bad, "1.jpg")
	if err == nil {
		t.Fatal("expected openCBRPage error")
	}
	_, _, err = openCBRPage(filepath.Join(t.TempDir(), "nope.cbr"), "1.jpg")
	if err == nil {
		t.Fatal("expected missing open error")
	}
	if parseComic(bad).PageCount != 0 {
		t.Fatal("parseComic cbr")
	}
}

func TestFetchCoverImageEmpty(t *testing.T) {
	if _, err := FetchCoverImage(context.Background(), "  "); err == nil {
		t.Fatal("expected empty url error")
	}
}

func TestPDFInfoFromFile(t *testing.T) {
	pdf := []byte("%PDF-1.4\n3 0 obj<</Title (Hex Path)/Author (A)>>endobj\n%%EOF")
	path := filepath.Join(t.TempDir(), "x.pdf")
	if err := os.WriteFile(path, pdf, 0o644); err != nil {
		t.Fatal(err)
	}
	meta := pdfInfoFromFile(path, pdfMeta{Title: "fallback"})
	if meta.Title != "Hex Path" {
		t.Fatalf("title=%q", meta.Title)
	}
	meta = pdfInfoFromFile(filepath.Join(t.TempDir(), "missing.pdf"), pdfMeta{Title: "fallback"})
	if meta.Title != "fallback" {
		t.Fatalf("fallback=%q", meta.Title)
	}
}

func TestExtractCoverFromFileEPUB(t *testing.T) {
	path := writeTestEPUB(t)
	data := ExtractCoverFromFile(path, models.FormatEPUB)
	if string(data) != "PNGDATA" {
		t.Fatalf("cover=%q", data)
	}
	if ExtractCoverFromFile(path, models.FormatMOBI) != nil {
		t.Fatal("mobi empty")
	}
	if ExtractCoverFromFile(filepath.Join(t.TempDir(), "x.cbr"), models.FormatCBR) != nil {
		t.Fatal("cbr empty")
	}
}

func TestApplyFilenameMeta(t *testing.T) {
	book := &models.Book{}
	applyFilenameMeta(book, filenameMeta{
		Title: "T", Author: "A", Series: "Foo__Bar", SeriesIndex: 2,
	})
	if book.Title != "T" || book.Author != "A" || book.Series != "Foo Bar" || book.SeriesIndex != 2 {
		t.Fatalf("book=%+v", book)
	}
}

func TestCleanAuthorNameKeepGood(t *testing.T) {
	if CleanAuthorName("Real Author", "x.epub") != "Real Author" {
		t.Fatal("keep good author")
	}
	if CleanAuthorName("1\uFFFD\uFFFD\uFFFD\uFFFDx", "Ada Lovelace - Notes.pdf") == "1\uFFFD\uFFFD\uFFFD\uFFFDx" {
		t.Fatal("expected garbled author cleanup")
	}
}

func TestCoverHrefAndMime(t *testing.T) {
	pkg := opfPackage{}
	pkg.Manifest.Items = []struct {
		ID         string `xml:"id,attr"`
		Href       string `xml:"href,attr"`
		MediaType  string `xml:"media-type,attr"`
		Properties string `xml:"properties,attr"`
	}{
		{ID: "c1", Href: "cover.jpg", MediaType: "image/jpeg"},
		{ID: "other", Href: "x.png", MediaType: "image/png"},
	}
	if coverHref(pkg, "c1") != "cover.jpg" {
		t.Fatal("cover id")
	}
	pkg.Manifest.Items[0].Properties = "cover-image"
	if coverHref(pkg, "") != "cover.jpg" {
		t.Fatal("cover-image prop")
	}
	if mimeForItem(pkg, "x.png") != "image/png" {
		t.Fatal("mime")
	}
	if mimeForItem(pkg, "missing") != "image/jpeg" {
		t.Fatal("mime default")
	}
}

func TestParseFilenameMetaSeries(t *testing.T) {
	m := parseFilenameMeta("[2020] Expanse Book 2 - Leviathan Wakes.epub")
	if m.SeriesIndex != 2 {
		t.Fatalf("meta=%+v", m)
	}
	m2 := parseFilenameMeta("Ada Lovelace - Notes on Babbage.pdf")
	if m2.Author == "" && m2.Title == "" {
		t.Fatalf("meta=%+v", m2)
	}
}

func TestFetchCoverImageInvalidURL(t *testing.T) {
	if _, err := FetchCoverImage(context.Background(), "http://127.0.0.1/x"); err == nil {
		t.Fatal("expected blocked private url")
	}
}

func TestPalmDOCBackref(t *testing.T) {
	in := []byte{3, 'a', 'b', 'c', 0x80, 0x00}
	out := palmDOCDecompress(in)
	if len(out) < 3 || string(out[:3]) != "abc" {
		t.Fatalf("out=%q", out)
	}
	// Oversize length byte and truncated backref must not panic.
	truncated := palmDOCDecompress([]byte{5, 'a'})
	if truncated == nil {
		t.Fatal("oversize length should return a slice")
	}
	incomplete := palmDOCDecompress([]byte{3, 'a', 'b', 'c', 0x80})
	if incomplete == nil {
		t.Fatal("incomplete backref should return a slice")
	}
}

func TestMergeAudiobookFoldersNoop(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.Open(filepath.Join(dir, "s.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	libDir := filepath.Join(dir, "lib")
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	lib, err := store.CreateLibrary(ctx, "Main", libDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Lonely",
		Format:    models.FormatMP3,
		RelPath:   "lonely.mp3",
	}, 1); err != nil {
		t.Fatal(err)
	}
	sc := New(store, filepath.Join(dir, "covers"), dir, slog.New(slog.NewTextHandler(io.Discard, nil)), 1)
	if err := sc.MergeAudiobookFolders(ctx, lib.ID); err != nil {
		t.Fatal(err)
	}
	page, err := store.ListBooks(ctx, models.BookQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || page.Items[0].Format != models.FormatMP3 {
		t.Fatalf("single-track folder should stay mp3: %+v", page)
	}
}
