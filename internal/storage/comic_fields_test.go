package storage

import (
	"context"
	"testing"

	"athenaeum/internal/models"
)

func TestComicFieldsRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	book := &models.Book{
		Title:            "Hero Tales 7",
		Format:           models.FormatCBZ,
		RelPath:          "hero.cbz",
		Publisher:        "Indie Press",
		AgeRating:        "Mature 17+",
		ReadingDirection: models.ReadingDirectionRTL,
	}
	id, err := s.UpsertBook(ctx, book, 1000)
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := s.GetBook(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Publisher != "Indie Press" || got.AgeRating != "Mature 17+" || got.ReadingDirection != "rtl" {
		t.Errorf("comic fields = %q/%q/%q", got.Publisher, got.AgeRating, got.ReadingDirection)
	}

	page, err := s.ListBooks(ctx, models.BookQuery{Format: models.FormatCBZ})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("list items = %d, want 1", len(page.Items))
	}
	listed := page.Items[0]
	if listed.Publisher != "Indie Press" || listed.ReadingDirection != "rtl" {
		t.Errorf("listed comic fields = %q/%q", listed.Publisher, listed.ReadingDirection)
	}

	metas, err := s.ListBooksForMetadata(ctx, 0, []int64{id})
	if err != nil {
		t.Fatalf("metadata list: %v", err)
	}
	if len(metas) != 1 || metas[0].Publisher != "Indie Press" {
		t.Fatalf("metadata list comic fields = %+v", metas)
	}
}

func TestComicFieldsRespectMetaEdited(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	book := &models.Book{
		Title:            "Comic",
		Format:           models.FormatCBZ,
		RelPath:          "c.cbz",
		Publisher:        "Scan Pub",
		ReadingDirection: "rtl",
	}
	id, err := s.UpsertBook(ctx, book, 1000)
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// A plain rescan (no user edit) still refreshes the comic fields.
	rescan := &models.Book{
		Title:            "Comic",
		Format:           models.FormatCBZ,
		RelPath:          "c.cbz",
		Publisher:        "Scan Pub 2",
		AgeRating:        "Teen",
		ReadingDirection: "ltr",
	}
	if _, err := s.UpsertBook(ctx, rescan, 2000); err != nil {
		t.Fatalf("rescan upsert: %v", err)
	}
	got, err := s.GetBook(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Publisher != "Scan Pub 2" || got.AgeRating != "Teen" || got.ReadingDirection != "ltr" {
		t.Fatalf("rescan did not update comic fields: %+v", got)
	}

	// Once the user edits metadata, rescan values must not clobber.
	if _, err := s.UpdateBookMetadata(ctx, id, models.BookUpdate{
		Title:            "My Comic",
		Publisher:        "User Pub",
		AgeRating:        "Adult",
		ReadingDirection: "rtl",
	}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := s.UpsertBook(ctx, rescan, 3000); err != nil {
		t.Fatalf("second rescan: %v", err)
	}
	got, err = s.GetBook(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Publisher != "User Pub" || got.AgeRating != "Adult" || got.ReadingDirection != "rtl" {
		t.Errorf("edited comic fields clobbered: %q/%q/%q", got.Publisher, got.AgeRating, got.ReadingDirection)
	}
}
