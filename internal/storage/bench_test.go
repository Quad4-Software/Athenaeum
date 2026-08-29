package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"athenaeum/internal/models"
)

func seedBenchBooks(b *testing.B, s *Store, n int) {
	b.Helper()
	ctx := context.Background()
	authors := []string{"Gibson", "Stephenson", "Le Guin", "Asimov", "Clarke"}
	formats := []string{models.FormatEPUB, models.FormatPDF, models.FormatMP3}
	for i := range n {
		book := &models.Book{
			Title:   "Book Title With Searchable Words",
			Author:  authors[i%len(authors)],
			Series:  "Series Alpha",
			Format:  formats[i%len(formats)],
			RelPath: fmt.Sprintf("dir/book-%d.epub", i),
		}
		if _, err := s.UpsertBook(ctx, book, time.Now().Unix()); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListBooksRecent(b *testing.B) {
	s := newBenchStore(b)
	seedBenchBooks(b, s, 500)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.ListBooks(ctx, models.BookQuery{Limit: 60}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListBooksFTS(b *testing.B) {
	s := newBenchStore(b)
	seedBenchBooks(b, s, 500)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.ListBooks(ctx, models.BookQuery{Search: "searchable", Limit: 60}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListCollectionsWithAuto(b *testing.B) {
	s := newBenchStore(b)
	seedBenchBooks(b, s, 200)
	ctx := context.Background()
	if err := s.EnsureAutoCollections(ctx); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.ListCollections(ctx, models.AnonymousUserID); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEnsureAutoCollections(b *testing.B) {
	s := newBenchStore(b)
	seedBenchBooks(b, s, 200)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := s.EnsureAutoCollections(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func newBenchStore(b *testing.B) *Store {
	b.Helper()
	s, err := Open(b.TempDir() + "/bench.db")
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = s.Close() })
	return s
}
