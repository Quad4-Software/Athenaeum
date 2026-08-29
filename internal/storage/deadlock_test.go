package storage

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"athenaeum/internal/models"
)

// Regression: ListCollections must not query counts while the list cursor is open.
func TestListCollectionsConcurrentNoDeadlock(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	for i := range 20 {
		if _, err := s.UpsertBook(ctx, &models.Book{
			Title: "T", Author: "Gibson", Format: models.FormatEPUB, RelPath: fmt.Sprintf("b%d.epub", i),
		}, time.Now().Unix()); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.EnsureAutoCollections(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateSmartCollection(ctx, models.AnonymousUserID, "PDFs", "", models.SmartQuery{Format: models.FormatPDF}); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() {
				for range 20 {
					if _, err := s.ListCollections(ctx, models.AnonymousUserID); err != nil {
						t.Error(err)
						return
					}
				}
			})
		}
		wg.Wait()
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("ListCollections deadlocked under concurrency")
	}
}

func TestEnsureAutoCollectionsConcurrentNoDeadlock(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	for i := range 30 {
		if _, err := s.UpsertBook(ctx, &models.Book{
			Title: "T", Author: "Author" + fmt.Sprint(i%10), Format: models.FormatEPUB,
			RelPath: fmt.Sprintf("x%d.epub", i),
		}, time.Now().Unix()); err != nil {
			t.Fatal(err)
		}
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		var wg sync.WaitGroup
		for range 4 {
			wg.Go(func() {
				if err := s.EnsureAutoCollections(ctx); err != nil {
					t.Error(err)
				}
			})
		}
		wg.Wait()
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("EnsureAutoCollections deadlocked under concurrency")
	}
}
