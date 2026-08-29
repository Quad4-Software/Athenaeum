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

func TestMetadataMatcherStartDoesNotPanic(t *testing.T) {
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
	if _, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Sample",
		Format:    models.FormatEPUB,
		RelPath:   "sample.epub",
	}, 1); err != nil {
		t.Fatal(err)
	}

	m := NewMetadataMatcher(store, filepath.Join(dir, "covers"), slog.Default())
	if !m.Start(ctx, MetadataAutoMatchRequest{ApplyCover: true}) {
		t.Fatal("expected metadata match to start")
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !m.Running() {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if m.Running() {
		t.Fatal("metadata match still running after timeout")
	}

	st := m.Status()
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
