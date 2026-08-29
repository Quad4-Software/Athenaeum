package library

import (
	"context"
	"strings"

	"athenaeum/internal/models"
)

func enrichBookMetadata(ctx context.Context, book *models.Book, isbn, asin string) {
	if !needsMetadataLookup(book) && isbn == "" && asin == "" {
		return
	}
	lookup := newMetadataLookup()
	fields := lookup.enrich(ctx, book.Title, book.Author, isbn, asin)
	applyLookupMeta(book, fields)
}

func needsMetadataLookup(book *models.Book) bool {
	return strings.TrimSpace(book.Title) == "" ||
		strings.TrimSpace(book.Author) == "" ||
		strings.TrimSpace(book.Description) == ""
}

func mergeIdentifiers(a, b string) string {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a != "" {
		return a
	}
	return b
}
