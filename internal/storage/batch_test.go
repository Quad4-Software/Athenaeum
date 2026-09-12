package storage

import (
	"context"
	"fmt"
	"testing"

	"athenaeum/internal/models"
)

func TestChunkIDs(t *testing.T) {
	cases := []struct {
		name string
		n    int
		size int
		want []int
	}{
		{"empty", 0, 3, nil},
		{"under chunk size", 2, 3, []int{2}},
		{"exact chunk size", 3, 3, []int{3}},
		{"over chunk size", 7, 3, []int{3, 3, 1}},
		{"exact multiple", 6, 3, []int{3, 3}},
		{"zero size", 4, 0, nil},
	}
	for _, tc := range cases {
		ids := make([]int64, tc.n)
		for i := range ids {
			ids[i] = int64(i)
		}
		chunks := chunkIDs(ids, tc.size)
		if len(chunks) != len(tc.want) {
			t.Fatalf("%s: got %d chunks, want %d", tc.name, len(chunks), len(tc.want))
		}
		var next int64
		for i, ch := range chunks {
			if len(ch) != tc.want[i] {
				t.Fatalf("%s: chunk %d len=%d want %d", tc.name, i, len(ch), tc.want[i])
			}
			for _, id := range ch {
				if id != next {
					t.Fatalf("%s: chunks broke ordering, got id %d want %d", tc.name, id, next)
				}
				next++
			}
		}
	}
}

func TestListBookTagsBatchOverChunkSize(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	n := inListChunkSize + 1
	ids := make([]int64, 0, n)
	for i := 0; i < n; i++ {
		rel := fmt.Sprintf("b%04d.epub", i)
		id, err := s.UpsertBook(ctx, &models.Book{Title: rel, Format: models.FormatEPUB, RelPath: rel}, 1)
		if err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
		ids = append(ids, id)
	}
	// Tag one book per chunk so merging across queries is observable.
	if _, err := s.SetBookTags(ctx, ids[0], []string{"first"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetBookTags(ctx, ids[n-1], []string{"last"}); err != nil {
		t.Fatal(err)
	}

	batch, err := s.ListBookTagsBatch(ctx, ids)
	if err != nil {
		t.Fatal(err)
	}
	if len(batch) != n {
		t.Fatalf("batch size=%d want %d", len(batch), n)
	}
	if got := batch[ids[0]]; len(got) != 1 || got[0] != "first" {
		t.Fatalf("first-chunk book tags=%v", got)
	}
	if got := batch[ids[n-1]]; len(got) != 1 || got[0] != "last" {
		t.Fatalf("last-chunk book tags=%v", got)
	}
}
