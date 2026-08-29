package storage

import (
	"context"
	"testing"
	"time"

	"athenaeum/internal/models"
)

func TestSmartCollectionAndAuto(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	if _, err := s.UpsertBook(ctx, &models.Book{
		Title: "A", Author: "Gibson", Format: models.FormatEPUB, RelPath: "a.epub",
	}, time.Now().Unix()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpsertBook(ctx, &models.Book{
		Title: "B", Author: "Gibson", Format: models.FormatEPUB, RelPath: "b.epub",
	}, time.Now().Unix()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpsertBook(ctx, &models.Book{
		Title: "C", Author: "Other", Format: models.FormatPDF, RelPath: "c.pdf",
	}, time.Now().Unix()); err != nil {
		t.Fatal(err)
	}

	if err := s.EnsureAutoCollections(ctx); err != nil {
		t.Fatal(err)
	}
	cols, err := s.ListCollections(ctx, models.AnonymousUserID)
	if err != nil {
		t.Fatal(err)
	}
	var recent, byGibson *models.Collection
	for i := range cols {
		switch cols[i].Name {
		case "Recently Added":
			recent = &cols[i]
		case "By Gibson":
			byGibson = &cols[i]
		}
	}
	if recent == nil || recent.Kind != models.CollectionAuto {
		t.Fatalf("missing recent auto collection: %+v", cols)
	}
	if byGibson == nil || byGibson.BookCount != 2 {
		t.Fatalf("by gibson = %+v", byGibson)
	}

	smart, err := s.CreateSmartCollection(ctx, models.AnonymousUserID, "PDFs", "", models.SmartQuery{Format: models.FormatPDF})
	if err != nil {
		t.Fatal(err)
	}
	page, err := s.ListBooks(ctx, models.BookQuery{CollectionID: smart.ID, UserID: models.AnonymousUserID})
	if err != nil || page.Total != 1 {
		t.Fatalf("smart list total=%d err=%v", page.Total, err)
	}

	if err := s.AddToCollection(ctx, models.AnonymousUserID, smart.ID, 1); err == nil {
		t.Fatal("expected error adding to smart collection")
	}
}

func TestApplySmartQueryAddedDays(t *testing.T) {
	q := ApplySmartQuery(models.BookQuery{}, models.SmartQuery{AddedDays: 7})
	if q.AddedAfter == 0 {
		t.Fatal("expected AddedAfter")
	}
	if q.AddedAfter > time.Now().Unix() {
		t.Fatal("AddedAfter in future")
	}
}
