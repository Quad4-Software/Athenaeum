package storage

import (
	"context"
	"testing"

	"athenaeum/internal/models"
)

func TestTagsCRUD(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	id, err := s.UpsertBook(ctx, &models.Book{Title: "Dune", Format: models.FormatEPUB, RelPath: "dune.epub"}, 1)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.CreateTag(ctx, "sci-fi"); err != nil {
		t.Fatal(err)
	}
	// Creating the same tag again (different case) should not duplicate it.
	if _, err := s.CreateTag(ctx, "Sci-Fi"); err != nil {
		t.Fatal(err)
	}
	tags, err := s.ListTags(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag, got %d: %+v", len(tags), tags)
	}

	names, err := s.SetBookTags(ctx, id, []string{"sci-fi", "favorite", "favorite"})
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 tags on book, got %+v", names)
	}

	names, err = s.AddBookTag(ctx, id, "classic")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 3 {
		t.Fatalf("expected 3 tags after add, got %+v", names)
	}

	all, err := s.ListTags(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var classicID int64
	for _, tag := range all {
		if tag.Name == "classic" {
			classicID = tag.ID
		}
	}
	if classicID == 0 {
		t.Fatalf("classic tag not found in %+v", all)
	}

	if err := s.RemoveBookTag(ctx, id, classicID); err != nil {
		t.Fatal(err)
	}
	names, err = s.ListBookTags(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 tags after remove, got %+v", names)
	}

	if err := s.RemoveBookTag(ctx, id, classicID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound removing again, got %v", err)
	}

	batch, err := s.ListBookTagsBatch(ctx, []int64{id})
	if err != nil {
		t.Fatal(err)
	}
	if len(batch[id]) != 2 {
		t.Fatalf("expected batch to return 2 tags, got %+v", batch)
	}
}

func TestBookTagsFilter(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	id1, _ := s.UpsertBook(ctx, &models.Book{Title: "A", Format: models.FormatEPUB, RelPath: "a.epub"}, 1)
	id2, _ := s.UpsertBook(ctx, &models.Book{Title: "B", Format: models.FormatEPUB, RelPath: "b.epub"}, 1)

	if _, err := s.SetBookTags(ctx, id1, []string{"fiction"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetBookTags(ctx, id2, []string{"nonfiction"}); err != nil {
		t.Fatal(err)
	}

	page, err := s.ListBooks(ctx, models.BookQuery{Tag: "fiction"})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != id1 {
		t.Fatalf("unexpected page for tag filter: %+v", page)
	}
}

func TestRatingsCRUD(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	id, err := s.UpsertBook(ctx, &models.Book{Title: "Dune", Format: models.FormatEPUB, RelPath: "dune.epub"}, 1)
	if err != nil {
		t.Fatal(err)
	}

	r, err := s.GetRating(ctx, 1, id)
	if err != nil {
		t.Fatal(err)
	}
	if r.Rating != 0 {
		t.Fatalf("expected no rating yet, got %+v", r)
	}

	r, err = s.SetRating(ctx, 1, id, 5)
	if err != nil {
		t.Fatal(err)
	}
	if r.Rating != 5 {
		t.Fatalf("expected rating 5, got %+v", r)
	}

	r, err = s.GetRating(ctx, 1, id)
	if err != nil {
		t.Fatal(err)
	}
	if r.Rating != 5 {
		t.Fatalf("expected persisted rating 5, got %+v", r)
	}

	if _, err := s.SetRating(ctx, 1, id, 6); err == nil {
		t.Fatal("expected error for out-of-range rating")
	}

	if _, err := s.SetRating(ctx, 1, id, 0); err != nil {
		t.Fatal(err)
	}
	r, err = s.GetRating(ctx, 1, id)
	if err != nil {
		t.Fatal(err)
	}
	if r.Rating != 0 {
		t.Fatalf("expected rating cleared, got %+v", r)
	}

	if _, err := s.SetRating(ctx, 2, id, 3); err != nil {
		t.Fatal(err)
	}
	batch, err := s.RatingsBatch(ctx, 2, []int64{id})
	if err != nil {
		t.Fatal(err)
	}
	if batch[id] != 3 {
		t.Fatalf("expected batch rating 3, got %+v", batch)
	}
}

func TestReaderPrefsRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	prefs, err := s.GetReaderPrefs(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(prefs.Prefs) != 0 {
		t.Fatalf("expected empty prefs, got %+v", prefs)
	}

	saved, err := s.SaveReaderPrefs(ctx, 1, map[string]any{"theme": "sepia", "fontPct": float64(120)})
	if err != nil {
		t.Fatal(err)
	}
	if saved.Prefs["theme"] != "sepia" {
		t.Fatalf("unexpected saved prefs: %+v", saved)
	}

	loaded, err := s.GetReaderPrefs(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Prefs["theme"] != "sepia" || loaded.Prefs["fontPct"] != float64(120) {
		t.Fatalf("unexpected loaded prefs: %+v", loaded)
	}

	other, err := s.GetReaderPrefs(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(other.Prefs) != 0 {
		t.Fatalf("expected empty prefs for other user, got %+v", other)
	}
}
