package storage

import (
	"context"
	"testing"
	"time"

	"athenaeum/internal/models"
)

func TestListBooksSmartCollectionFilter(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	if _, err := s.UpsertBook(ctx, &models.Book{Title: "E", Author: "A", Format: models.FormatEPUB, RelPath: "e.epub"}, time.Now().Unix()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpsertBook(ctx, &models.Book{Title: "P", Author: "A", Format: models.FormatPDF, RelPath: "p.pdf"}, time.Now().Unix()); err != nil {
		t.Fatal(err)
	}

	smart, err := s.CreateSmartCollection(ctx, models.AnonymousUserID, "Only PDF", "", models.SmartQuery{Format: models.FormatPDF})
	if err != nil {
		t.Fatal(err)
	}

	page, err := s.ListBooks(ctx, models.BookQuery{CollectionID: smart.ID, UserID: models.AnonymousUserID})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].Format != models.FormatPDF {
		t.Fatalf("smart filter page=%+v", page)
	}
}

func TestListBooksAuthorFilter(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	if _, err := s.UpsertBook(ctx, &models.Book{Title: "1", Author: "Gibson", Format: models.FormatEPUB, RelPath: "1.epub"}, time.Now().Unix()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpsertBook(ctx, &models.Book{Title: "2", Author: "Other", Format: models.FormatEPUB, RelPath: "2.epub"}, time.Now().Unix()); err != nil {
		t.Fatal(err)
	}

	page, err := s.ListBooks(ctx, models.BookQuery{Author: "Gibson"})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 {
		t.Fatalf("author filter total=%d", page.Total)
	}
}
