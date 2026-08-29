package storage

import (
	"context"
	"testing"

	"athenaeum/internal/models"
)

func TestBooksDeleteCoverAndFileState(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	libDir := t.TempDir()
	lib, err := s.CreateLibrary(ctx, "Books Gap", libDir)
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Gap Book",
		Author:    "A",
		Format:    models.FormatEPUB,
		RelPath:   "gap.epub",
		FileSize:  12,
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.GetBookByPath(ctx, lib.ID, "gap.epub")
	if err != nil || got.ID != id {
		t.Fatalf("by path id=%d err=%v", got.ID, err)
	}
	mtime, size, ok, err := s.FileState(ctx, lib.ID, "gap.epub")
	if err != nil || !ok || size != 12 {
		t.Fatalf("file state mtime=%d size=%d ok=%v err=%v", mtime, size, ok, err)
	}
	if err := s.SetBookCover(ctx, id, true); err != nil {
		t.Fatal(err)
	}
	book, err := s.GetBook(ctx, id)
	if err != nil || !book.HasCover {
		t.Fatalf("hasCover=%v err=%v", book.HasCover, err)
	}
	if err := s.SetBookHasCover(ctx, id, false); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteBook(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetBook(ctx, id); err == nil {
		t.Fatal("expected deleted")
	}
}

func TestMigratePostgresEnsureFTSNoop(t *testing.T) {
	s := &Store{driver: DriverPostgres}
	if err := s.ensureFTS(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestListBooksForMetadata(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	lib, err := s.CreateLibrary(ctx, "Meta Lib", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Meta",
		Format:    models.FormatEPUB,
		RelPath:   "m.epub",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	books, err := s.ListBooksForMetadata(ctx, lib.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(books) == 0 {
		t.Fatal("expected books by library")
	}
	byIDs, err := s.ListBooksForMetadata(ctx, 0, []int64{id})
	if err != nil || len(byIDs) != 1 {
		t.Fatalf("by ids=%d err=%v", len(byIDs), err)
	}
}
