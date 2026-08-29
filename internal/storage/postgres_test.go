package storage

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"athenaeum/internal/models"
)

// TestPostgresSmoke runs against TEST_DATABASE_URL when set (otherwise skipped).
func TestPostgresSmoke(t *testing.T) {
	url := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	store, err := OpenWith(OpenOptions{Driver: DriverPostgres, URL: url})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer store.Close()
	if store.Driver() != DriverPostgres {
		t.Fatalf("driver = %s", store.Driver())
	}

	ctx := context.Background()
	id, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: 1,
		Title:     "Dune",
		Author:    "Frank Herbert",
		Series:    "Dune",
		Format:    "epub",
		RelPath:   "dune.epub",
		AbsPath:   "/library/dune.epub",
	}, time.Now().Unix())
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected id, got %d", id)
	}
	if err := store.ReplaceBookContent(ctx, id, []string{"The spice must flow across Arrakis"}); err != nil {
		t.Fatalf("content: %v", err)
	}

	page, err := store.ListBooks(ctx, models.BookQuery{Search: "dune", Limit: 10})
	if err != nil {
		t.Fatalf("search metadata: %v", err)
	}
	if page.Total < 1 {
		t.Fatalf("expected metadata FTS hit, total=%d", page.Total)
	}
	contentIDs, err := store.SearchBookContentIDs(ctx, "spice")
	if err != nil {
		t.Fatalf("content search: %v", err)
	}
	if len(contentIDs) < 1 {
		t.Fatal("expected content FTS hit")
	}
}
