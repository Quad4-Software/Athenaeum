package library

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func TestVerifyIntegrityMissingFile(t *testing.T) {
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
	libID := lib.ID

	existing := filepath.Join(libDir, "present.epub")
	if err := os.WriteFile(existing, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: libID,
		Title:     "Present",
		Format:    models.FormatEPUB,
		RelPath:   "present.epub",
	}, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: libID,
		Title:     "Gone",
		Format:    models.FormatEPUB,
		RelPath:   "missing.epub",
	}, 1); err != nil {
		t.Fatal(err)
	}

	coverDir := filepath.Join(dir, "covers")
	if err := os.MkdirAll(coverDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(coverDir, "999.img"), []byte("orphan"), 0o600); err != nil {
		t.Fatal(err)
	}

	report, err := VerifyIntegrity(ctx, store, coverDir)
	if err != nil {
		t.Fatal(err)
	}
	if report.TotalBooks != 2 {
		t.Fatalf("total=%d", report.TotalBooks)
	}
	if report.MissingCount != 1 {
		t.Fatalf("missing=%d", report.MissingCount)
	}
	if report.OrphanCovers != 1 {
		t.Fatalf("orphan covers=%d", report.OrphanCovers)
	}
}

func TestPruneMissingBooks(t *testing.T) {
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
	libID := lib.ID
	if _, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: libID,
		Title:     "Gone",
		Format:    models.FormatPDF,
		RelPath:   "gone.pdf",
	}, 1); err != nil {
		t.Fatal(err)
	}

	removed, err := PruneMissingBooks(ctx, store)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatalf("removed=%d", removed)
	}
	page, err := store.ListBooks(ctx, models.BookQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 0 {
		t.Fatalf("total=%d", page.Total)
	}
}

func TestCleanupOrphanCovers(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	store, err := storage.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	coverDir := filepath.Join(dir, "covers")
	if err := os.MkdirAll(coverDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(coverDir, "42.img"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	removed, err := CleanupOrphanCovers(ctx, store, coverDir)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatalf("removed=%d", removed)
	}
}

func TestExtractCoverFromFileSidecar(t *testing.T) {
	dir := t.TempDir()
	book := filepath.Join(dir, "novel.epub")
	if err := os.WriteFile(book, []byte("zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	cover := filepath.Join(dir, "cover.jpg")
	jpeg := makeTestJPEG(512)
	if err := os.WriteFile(cover, jpeg, 0o644); err != nil {
		t.Fatal(err)
	}
	got := ExtractCoverFromFile(book, models.FormatEPUB)
	if len(got) != len(jpeg) {
		t.Fatalf("cover len=%d want %d", len(got), len(jpeg))
	}
}

func TestStartRegenerateCoversDoesNotPanic(t *testing.T) {
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
	bookPath := filepath.Join(libDir, "sample.epub")
	if err := os.WriteFile(bookPath, []byte("zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Sample",
		Format:    models.FormatEPUB,
		RelPath:   "sample.epub",
	}, 1); err != nil {
		t.Fatal(err)
	}

	coverDir := filepath.Join(dir, "covers")
	if err := os.MkdirAll(coverDir, 0o750); err != nil {
		t.Fatal(err)
	}

	m := NewMaintenance(store, coverDir, slog.Default())
	if !m.StartRegenerateCovers(ctx, 0) {
		t.Fatal("expected regenerate to start")
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !m.Running() {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if m.Running() {
		t.Fatal("regenerate still running after timeout")
	}

	st := m.Status()
	if st.Task != "regenerate_covers" {
		t.Fatalf("task=%q", st.Task)
	}
	if st.Total != 1 {
		t.Fatalf("total=%d", st.Total)
	}
	if st.Done != 1 {
		t.Fatalf("done=%d", st.Done)
	}
	if st.FinishedAt == nil {
		t.Fatal("expected finishedAt")
	}
}
