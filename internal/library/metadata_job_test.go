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
	RegisterMetadataProvider(MetadataProviderDef{
		Info: models.MetadataProvider{ID: "testlocal", Label: "Test Local"},
		Search: func(ctx context.Context, s *metadataSearcher, in MetadataSearchInput) []models.MetadataMatch {
			return nil
		},
	})
	metadataRegistryMu.Lock()
	prev := metadataRegistry
	metadataRegistry = []MetadataProviderDef{metadataRegistry[len(metadataRegistry)-1]}
	metadataRegistryMu.Unlock()
	t.Cleanup(func() {
		metadataRegistryMu.Lock()
		metadataRegistry = prev
		metadataRegistryMu.Unlock()
	})

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
	if !m.Start(ctx, MetadataAutoMatchRequest{ApplyCover: false}) {
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

// Guarantee: a metadata match job restricted to library A never touches
// books selected from library B, even when their ids are given explicitly.
func TestMetadataMatcherRespectsAllowedLibraries(t *testing.T) {
	RegisterMetadataProvider(MetadataProviderDef{
		Info: models.MetadataProvider{ID: "testlocal", Label: "Test Local"},
		Search: func(ctx context.Context, s *metadataSearcher, in MetadataSearchInput) []models.MetadataMatch {
			return []models.MetadataMatch{{Title: in.Title, Author: "REWRITTEN"}}
		},
	})
	metadataRegistryMu.Lock()
	prev := metadataRegistry
	metadataRegistry = []MetadataProviderDef{metadataRegistry[len(metadataRegistry)-1]}
	metadataRegistryMu.Unlock()
	t.Cleanup(func() {
		metadataRegistryMu.Lock()
		metadataRegistry = prev
		metadataRegistryMu.Unlock()
	})

	ctx := context.Background()
	dir := t.TempDir()
	store, err := storage.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	dirA := filepath.Join(dir, "libA")
	dirB := filepath.Join(dir, "libB")
	for _, d := range []string{dirA, dirB} {
		if err := os.MkdirAll(d, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	libA, err := store.CreateLibrary(ctx, "A", dirA)
	if err != nil {
		t.Fatal(err)
	}
	libB, err := store.CreateLibrary(ctx, "B", dirB)
	if err != nil {
		t.Fatal(err)
	}
	bookA, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: libA.ID, Title: "Mine", Format: models.FormatEPUB, RelPath: "a.epub",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	bookB, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: libB.ID, Title: "NotMine", Format: models.FormatEPUB, RelPath: "b.epub",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}

	m := NewMetadataMatcher(store, filepath.Join(dir, "covers"), slog.Default())
	started := m.Start(ctx, MetadataAutoMatchRequest{
		BookIDs:           []int64{bookA, bookB},
		AllowedLibraryIDs: []int64{libA.ID},
	})
	if !started {
		t.Fatal("expected metadata match to start")
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && m.Running() {
		time.Sleep(20 * time.Millisecond)
	}
	if m.Running() {
		t.Fatal("metadata match still running after timeout")
	}

	st := m.Status()
	if st.Total != 1 {
		t.Fatalf("total=%d want 1 (only the allowed book)", st.Total)
	}
	got, err := store.GetBook(ctx, bookB)
	if err != nil {
		t.Fatal(err)
	}
	if got.Author == "REWRITTEN" {
		t.Fatal("restricted book metadata rewritten by match job")
	}
}
