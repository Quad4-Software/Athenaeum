package library

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

type stubBibStore struct {
	books   map[int64]models.Book
	updates []models.BookUpdate
}

func (s *stubBibStore) FindBookByDOI(_ context.Context, doi string, opts storage.CitationLookupOpts) (models.Book, error) {
	for _, b := range s.books {
		if b.DOI == doi && libraryAllowed(b.LibraryID, opts) {
			return b, nil
		}
	}
	return models.Book{}, storage.ErrNotFound
}

func (s *stubBibStore) FindBookByArxivID(_ context.Context, id string, opts storage.CitationLookupOpts) (models.Book, error) {
	for _, b := range s.books {
		if b.ArxivID == id && libraryAllowed(b.LibraryID, opts) {
			return b, nil
		}
	}
	return models.Book{}, storage.ErrNotFound
}

func (s *stubBibStore) FindBookByPubmedID(_ context.Context, id string, opts storage.CitationLookupOpts) (models.Book, error) {
	for _, b := range s.books {
		if b.PubmedID == id && libraryAllowed(b.LibraryID, opts) {
			return b, nil
		}
	}
	return models.Book{}, storage.ErrNotFound
}

func (s *stubBibStore) FindBookByTitleAuthor(_ context.Context, title, author string, opts storage.CitationLookupOpts) (models.Book, error) {
	for _, b := range s.books {
		if b.Title == title && b.Author == author && libraryAllowed(b.LibraryID, opts) {
			return b, nil
		}
	}
	return models.Book{}, storage.ErrNotFound
}

func (s *stubBibStore) UpdateBookMetadata(_ context.Context, id int64, u models.BookUpdate) (models.Book, error) {
	b := s.books[id]
	b.Title = u.Title
	b.Author = u.Author
	b.DOI = u.DOI
	b.Journal = u.Journal
	s.books[id] = b
	s.updates = append(s.updates, u)
	return b, nil
}

func libraryAllowed(libraryID int64, opts storage.CitationLookupOpts) bool {
	if len(opts.LibraryIDs) == 0 {
		return true
	}
	return slices.Contains(opts.LibraryIDs, libraryID)
}

func TestImportBibTeXRespectsLibraryScope(t *testing.T) {
	store := &stubBibStore{books: map[int64]models.Book{
		1: {ID: 1, LibraryID: 1, Title: "Paper A", Author: "A", DOI: "10.1000/a"},
		2: {ID: 2, LibraryID: 2, Title: "Paper B", Author: "B", DOI: "10.1000/b"},
	}}
	raw := `
@article{a, title={Paper A}, author={A}, doi={10.1000/a}, journal={J}}
@article{b, title={Paper B}, author={B}, doi={10.1000/b}, journal={J}}
`
	res, err := ImportBibTeX(context.Background(), store, raw, BibImportOptions{LibraryIDs: []int64{1}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Updated != 1 || res.Matched != 1 || res.Unmatched != 1 {
		t.Fatalf("result=%+v", res)
	}
	if store.books[1].Journal != "J" {
		t.Fatalf("lib1 not updated: %+v", store.books[1])
	}
	if store.books[2].Journal != "" {
		t.Fatalf("lib2 should be untouched: %+v", store.books[2])
	}
}

func TestImportBibTeXDoesNotOverwriteTitle(t *testing.T) {
	store := &stubBibStore{books: map[int64]models.Book{
		1: {ID: 1, LibraryID: 1, Title: "Good PDF Title", Author: "Author", DOI: "10.1000/x"},
	}}
	raw := `@article{x, title={Different Bib Title}, author={Author}, doi={10.1000/x}, journal={Nature}}`
	res, err := ImportBibTeX(context.Background(), store, raw, BibImportOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Updated != 1 {
		t.Fatalf("%+v", res)
	}
	if store.books[1].Title != "Good PDF Title" {
		t.Fatalf("title overwritten: %q", store.books[1].Title)
	}
	if store.books[1].Journal != "Nature" {
		t.Fatalf("journal=%q", store.books[1].Journal)
	}
}

func TestDirHasSingleBookFile(t *testing.T) {
	dir := t.TempDir()
	if err := writeTemp(dir, "a.pdf", "x"); err != nil {
		t.Fatal(err)
	}
	if !dirHasSingleBookFile(dir) {
		t.Fatal("expected single")
	}
	if err := writeTemp(dir, "b.pdf", "y"); err != nil {
		t.Fatal(err)
	}
	if dirHasSingleBookFile(dir) {
		t.Fatal("expected multi")
	}
}

func writeTemp(dir, name, body string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600)
}
