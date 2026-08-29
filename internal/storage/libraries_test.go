package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/models"
)

func TestLibrariesCRUD(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	if err := s.EnsureDefaultLibrary(ctx, t.TempDir()); err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	mount := filepath.Join(dir, "audiobooks")
	if err := os.MkdirAll(mount, 0o755); err != nil {
		t.Fatal(err)
	}

	lib, err := s.CreateLibrary(ctx, "Audiobooks", mount)
	if err != nil {
		t.Fatal(err)
	}
	if lib.Name != "Audiobooks" || lib.MountPath != mount {
		t.Fatalf("lib=%+v", lib)
	}

	libs, err := s.ListLibraries(ctx)
	if err != nil || len(libs) < 2 {
		t.Fatalf("list=%d err=%v", len(libs), err)
	}

	if err := s.ReorderLibraries(ctx, []int64{lib.ID, 1}); err != nil {
		t.Fatal(err)
	}

	if err := s.DeleteLibrary(ctx, lib.ID); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteDefaultLibrary(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	root := t.TempDir()
	if err := s.EnsureDefaultLibrary(ctx, root); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpsertBook(ctx, &models.Book{LibraryID: 1, Title: "A", Format: "epub", RelPath: "a.epub"}, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteLibrary(ctx, 1); err != nil {
		t.Fatal(err)
	}
	libs, err := s.ListLibraries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(libs) != 0 {
		t.Fatalf("libs=%d want 0", len(libs))
	}
}

func TestPrunePathsScoped(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	if err := s.EnsureDefaultLibrary(ctx, t.TempDir()); err != nil {
		t.Fatal(err)
	}

	for _, rp := range []string{"a.epub", "b.epub"} {
		if _, err := s.UpsertBook(ctx, &models.Book{LibraryID: 1, Title: rp, Format: "epub", RelPath: rp}, 1); err != nil {
			t.Fatal(err)
		}
	}
	dir := t.TempDir()
	mount := filepath.Join(dir, "other")
	if err := os.MkdirAll(mount, 0o755); err != nil {
		t.Fatal(err)
	}
	other, err := s.CreateLibrary(ctx, "Other", mount)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpsertBook(ctx, &models.Book{LibraryID: other.ID, Title: "x", Format: "epub", RelPath: "x.epub"}, 1); err != nil {
		t.Fatal(err)
	}

	removed, err := s.PrunePaths(ctx, 1, map[string]struct{}{"a.epub": {}})
	if err != nil || removed != 1 {
		t.Fatalf("removed=%d err=%v", removed, err)
	}
	stats, _ := s.Stats(ctx, 0, 0)
	if stats.TotalBooks != 2 {
		t.Fatalf("total=%d want 2", stats.TotalBooks)
	}
}
