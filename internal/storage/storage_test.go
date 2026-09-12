package storage

import (
	"context"
	"path/filepath"
	"testing"

	"athenaeum/internal/models"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestUpsertAndGet(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	book := &models.Book{
		Title:    "Dune",
		Author:   "Frank Herbert",
		Format:   models.FormatEPUB,
		RelPath:  "herbert/dune.epub",
		AbsPath:  "/library/herbert/dune.epub",
		HasCover: true,
	}
	id, err := s.UpsertBook(ctx, book, 1000)
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	got, err := s.GetBook(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "Dune" || got.Author != "Frank Herbert" || !got.HasCover {
		t.Errorf("unexpected book: %+v", got)
	}

	// Upsert again on the same path should update, not duplicate.
	book.Title = "Dune (Revised)"
	if _, err := s.UpsertBook(ctx, book, 2000); err != nil {
		t.Fatalf("re-upsert: %v", err)
	}
	stats, err := s.Stats(ctx, 0, 0)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats.TotalBooks != 1 {
		t.Errorf("total = %d, want 1", stats.TotalBooks)
	}
	if stats.EPUBCount != 1 {
		t.Errorf("epub count = %d, want 1", stats.EPUBCount)
	}
}

func TestUpdateBookMetadataPreservesOnRescan(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	book := &models.Book{
		Title:   "Original",
		Author:  "Author",
		Format:  models.FormatPDF,
		RelPath: "book.pdf",
		AbsPath: "/library/book.pdf",
	}
	id, err := s.UpsertBook(ctx, book, 1000)
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	updated, err := s.UpdateBookMetadata(ctx, id, models.BookUpdate{
		Title:  "Edited Title",
		Author: "Edited Author",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !updated.MetaEdited {
		t.Fatal("expected meta_edited")
	}

	scan := &models.Book{
		LibraryID:   1,
		Title:       "Scanned Title",
		Author:      "Scanned Author",
		Format:      models.FormatPDF,
		RelPath:     "book.pdf",
		AbsPath:     "/library/book.pdf",
		MetaEdited:  true,
		CoverEdited: false,
	}
	if _, err := s.UpsertBook(ctx, scan, 2000); err != nil {
		t.Fatalf("re-upsert: %v", err)
	}
	got, err := s.GetBook(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "Edited Title" || got.Author != "Edited Author" {
		t.Errorf("metadata overwritten: %+v", got)
	}
}

func TestListBooksSearchAndPaging(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	seed := []models.Book{
		{Title: "Go in Action", Author: "Kennedy", Format: models.FormatEPUB, RelPath: "1.epub"},
		{Title: "The Go Programming Language", Author: "Donovan", Format: models.FormatPDF, RelPath: "2.pdf"},
		{Title: "Clean Code", Author: "Martin", Format: models.FormatPDF, RelPath: "3.pdf"},
	}
	for i := range seed {
		if _, err := s.UpsertBook(ctx, &seed[i], int64(i)); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	page, err := s.ListBooks(ctx, models.BookQuery{Search: "go"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if page.Total != 2 {
		t.Errorf("search total = %d, want 2", page.Total)
	}

	pdfs, err := s.ListBooks(ctx, models.BookQuery{Format: models.FormatPDF})
	if err != nil {
		t.Fatalf("list pdf: %v", err)
	}
	if pdfs.Total != 2 {
		t.Errorf("pdf total = %d, want 2", pdfs.Total)
	}

	limited, err := s.ListBooks(ctx, models.BookQuery{Limit: 1, Offset: 0})
	if err != nil {
		t.Fatalf("list limited: %v", err)
	}
	if len(limited.Items) != 1 || limited.Total != 3 {
		t.Errorf("paging: got %d items, total %d", len(limited.Items), limited.Total)
	}
}

func TestListBooksStartsWith(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	seed := []models.Book{
		{Title: "Dune", Format: models.FormatEPUB, RelPath: "dune.epub"},
		{Title: "dune Messiah", Format: models.FormatEPUB, RelPath: "dm.epub"},
		{Title: "1984", Format: models.FormatEPUB, RelPath: "1984.epub"},
		{Title: "!Syntax Notes", Format: models.FormatEPUB, RelPath: "syn.epub"},
		{Title: "Foundation", Format: models.FormatEPUB, RelPath: "found.epub"},
	}
	for i := range seed {
		if _, err := s.UpsertBook(ctx, &seed[i], int64(i)); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	titles := func(q models.BookQuery) []string {
		page, err := s.ListBooks(ctx, q)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		var out []string
		for _, b := range page.Items {
			out = append(out, b.Title)
		}
		return out
	}
	sameTitles := func(got, want []string) bool {
		if len(got) != len(want) {
			return false
		}
		seen := map[string]int{}
		for _, title := range want {
			seen[title]++
		}
		for _, title := range got {
			seen[title]--
			if seen[title] < 0 {
				return false
			}
		}
		return true
	}

	for _, letter := range []string{"d", "D"} {
		got := titles(models.BookQuery{StartsWith: letter})
		if !sameTitles(got, []string{"Dune", "dune Messiah"}) {
			t.Errorf("letter %q got %v", letter, got)
		}
	}
	for _, bucket := range []string{"#", "0"} {
		got := titles(models.BookQuery{StartsWith: bucket})
		if !sameTitles(got, []string{"1984", "!Syntax Notes"}) {
			t.Errorf("bucket %q got %v", bucket, got)
		}
	}
	// Anything that is not one letter or the non-alpha bucket disables the filter.
	for _, invalid := range []string{"", "zz", "5", "!"} {
		got := titles(models.BookQuery{StartsWith: invalid})
		if len(got) != len(seed) {
			t.Errorf("letter %q got %d titles, want unfiltered %d", invalid, len(got), len(seed))
		}
	}
}

func TestProgressRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	book := &models.Book{Title: "X", Format: models.FormatPDF, RelPath: "x.pdf"}
	id, err := s.UpsertBook(ctx, book, 1)
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// No progress yet returns zero value without error.
	p, err := s.GetProgress(ctx, models.AnonymousUserID, id)
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}
	if p.Percent != 0 {
		t.Errorf("expected 0 percent, got %v", p.Percent)
	}

	if err := s.SaveProgress(ctx, models.AnonymousUserID, models.Progress{BookID: id, Location: "42", Percent: 0.5}); err != nil {
		t.Fatalf("save progress: %v", err)
	}
	p, err = s.GetProgress(ctx, models.AnonymousUserID, id)
	if err != nil {
		t.Fatalf("get progress 2: %v", err)
	}
	if p.Location != "42" || p.Percent != 0.5 {
		t.Errorf("unexpected progress: %+v", p)
	}
}

func TestPrunePaths(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	for _, rp := range []string{"a.epub", "b.epub", "c.pdf"} {
		if _, err := s.UpsertBook(ctx, &models.Book{Title: rp, Format: models.FormatEPUB, RelPath: rp}, 1); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	removed, err := s.PrunePaths(ctx, 1, map[string]struct{}{"a.epub": {}})
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if removed != 2 {
		t.Errorf("removed = %d, want 2", removed)
	}
	stats, _ := s.Stats(ctx, 0, 0)
	if stats.TotalBooks != 1 {
		t.Errorf("total after prune = %d, want 1", stats.TotalBooks)
	}
}

func TestGetBookNotFound(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.GetBook(context.Background(), 999); err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
