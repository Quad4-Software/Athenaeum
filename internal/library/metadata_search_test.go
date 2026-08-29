package library

import (
	"testing"

	"athenaeum/internal/models"
)

func TestBestMetadataMatch(t *testing.T) {
	book := models.Book{Title: "The Martian", Author: "Andy Weir"}
	matches := []models.MetadataMatch{
		{Source: "openlibrary", Title: "Other Book", Author: "Someone"},
		{Source: "google", Title: "The Martian", Author: "Andy Weir", ISBN: "978123"},
	}
	best, ok := BestMetadataMatch(book, matches)
	if !ok {
		t.Fatal("expected a match")
	}
	if best.Source != "google" {
		t.Fatalf("got source %q, want google", best.Source)
	}
}
