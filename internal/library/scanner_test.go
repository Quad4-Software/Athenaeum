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

func TestScannerIndexesLibrary(t *testing.T) {
	libDir := t.TempDir()
	coverDir := filepath.Join(t.TempDir(), "covers")

	// Place a valid EPUB and a PDF into the library.
	epubSrc := writeTestEPUB(t)
	copyFile(t, epubSrc, filepath.Join(libDir, "tales.epub"))
	if err := os.WriteFile(filepath.Join(libDir, "manual_v2.pdf"), []byte("%PDF-1.4 fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A non-book file should be ignored.
	if err := os.WriteFile(filepath.Join(libDir, "notes.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatal(err)
	}

	store, err := storage.Open(filepath.Join(t.TempDir(), "s.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.EnsureDefaultLibrary(context.Background(), libDir); err != nil {
		t.Fatal(err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	sc := New(store, coverDir, t.TempDir(), log, 2)

	if err := sc.Scan(context.Background()); err != nil {
		t.Fatalf("scan: %v", err)
	}

	stats, err := store.Stats(context.Background(), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalBooks != 2 {
		t.Fatalf("indexed %d books, want 2", stats.TotalBooks)
	}
	if stats.EPUBCount != 1 || stats.PDFCount != 1 {
		t.Fatalf("counts epub=%d pdf=%d, want 1/1", stats.EPUBCount, stats.PDFCount)
	}

	// The EPUB cover should have been extracted to the cover cache.
	page, _ := store.ListBooks(context.Background(), models.BookQuery{Format: models.FormatEPUB})
	if len(page.Items) != 1 {
		t.Fatalf("expected one epub, got %d", len(page.Items))
	}
	coverPath := CoverPath(coverDir, page.Items[0].ID)
	if _, err := os.Stat(coverPath); err != nil {
		t.Errorf("expected cover at %s: %v", coverPath, err)
	}

	// The PDF title should derive from its filename.
	pdfs, _ := store.ListBooks(context.Background(), models.BookQuery{Format: models.FormatPDF})
	if len(pdfs.Items) != 1 || pdfs.Items[0].Title != "manual v2" {
		t.Errorf("pdf title = %q, want %q", pdfs.Items[0].Title, "manual v2")
	}
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	in, err := os.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		t.Fatal(err)
	}
}
