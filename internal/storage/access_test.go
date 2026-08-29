package storage

import (
	"context"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

func TestUserLibraryAccess(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	userID, err := s.CreateUser(ctx, "reader", hash, false)
	if err != nil {
		t.Fatal(err)
	}
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}

	lib2, err := s.CreateLibrary(ctx, "Other", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	acc, err := s.AccessibleLibraries(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if acc.Restricted {
		t.Fatal("expected unrestricted by default")
	}

	if err := s.SetUserLibraries(ctx, user.ID, []int64{lib2.ID}); err != nil {
		t.Fatal(err)
	}
	acc, err = s.AccessibleLibraries(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if !acc.Restricted || len(acc.LibraryIDs) != 1 || acc.LibraryIDs[0] != lib2.ID {
		t.Fatalf("access = %+v", acc)
	}
	ok, err := s.UserCanAccessLibrary(ctx, user, lib2.ID)
	if err != nil || !ok {
		t.Fatalf("lib2 access = %v %v", ok, err)
	}
	ok, err = s.UserCanAccessLibrary(ctx, user, 1)
	if err != nil || ok {
		t.Fatalf("default lib access = %v %v", ok, err)
	}

	admin := models.User{ID: user.ID, IsAdmin: true}
	acc, err = s.AccessibleLibraries(ctx, admin)
	if err != nil {
		t.Fatal(err)
	}
	if acc.Restricted {
		t.Fatal("admin should be unrestricted")
	}
}

func TestContentHashDuplicate(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	b1 := &models.Book{
		LibraryID: 1,
		Title:     "First",
		Format:    models.FormatEPUB,
		RelPath:   "a.epub",
		AbsPath:   "/tmp/a.epub",
		FileSize:  10,
	}
	id1, err := s.UpsertBook(ctx, b1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ApplyContentHash(ctx, id1, "abc123"); err != nil {
		t.Fatal(err)
	}

	b2 := &models.Book{
		LibraryID: 1,
		Title:     "Second",
		Format:    models.FormatEPUB,
		RelPath:   "b.epub",
		AbsPath:   "/tmp/b.epub",
		FileSize:  10,
	}
	id2, err := s.UpsertBook(ctx, b2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ApplyContentHash(ctx, id2, "abc123"); err != nil {
		t.Fatal(err)
	}
	book, err := s.GetBook(ctx, id2)
	if err != nil {
		t.Fatal(err)
	}
	if book.DuplicateOf != id1 {
		t.Fatalf("duplicateOf = %d, want %d", book.DuplicateOf, id1)
	}
}
