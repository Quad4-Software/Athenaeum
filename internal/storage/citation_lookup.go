package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"athenaeum/internal/models"
)

// CitationLookupOpts scopes scholarly book lookups.
type CitationLookupOpts struct {
	LibraryIDs []int64 // empty means unrestricted
}

func citationLibraryClause(opts CitationLookupOpts) (clause string, args []any) {
	if len(opts.LibraryIDs) == 0 {
		return "", nil
	}
	placeholders := strings.Repeat("?,", len(opts.LibraryIDs))
	placeholders = placeholders[:len(placeholders)-1]
	args = make([]any, len(opts.LibraryIDs))
	for i, id := range opts.LibraryIDs {
		args[i] = id
	}
	return " AND library_id IN (" + placeholders + ")", args
}

// FindBookByDOI returns the first book with the given DOI.
func (s *Store) FindBookByDOI(ctx context.Context, doi string, opts CitationLookupOpts) (models.Book, error) {
	doi = strings.TrimSpace(doi)
	if doi == "" {
		return models.Book{}, ErrNotFound
	}
	libClause, libArgs := citationLibraryClause(opts)
	args := append([]any{doi}, libArgs...)
	row := s.queryRowContext(ctx, selectColumns+` WHERE doi=? AND hidden=0`+libClause+` ORDER BY id ASC LIMIT 1`, args...)
	b, err := scanBook(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Book{}, ErrNotFound
	}
	return b, err
}

// FindBookByArxivID returns the first book with the given arXiv id.
func (s *Store) FindBookByArxivID(ctx context.Context, arxivID string, opts CitationLookupOpts) (models.Book, error) {
	arxivID = strings.TrimSpace(arxivID)
	if arxivID == "" {
		return models.Book{}, ErrNotFound
	}
	libClause, libArgs := citationLibraryClause(opts)
	args := append([]any{arxivID}, libArgs...)
	row := s.queryRowContext(ctx, selectColumns+` WHERE arxiv_id=? AND hidden=0`+libClause+` ORDER BY id ASC LIMIT 1`, args...)
	b, err := scanBook(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Book{}, ErrNotFound
	}
	return b, err
}

// FindBookByPubmedID returns the first book with the given PMID.
func (s *Store) FindBookByPubmedID(ctx context.Context, pmid string, opts CitationLookupOpts) (models.Book, error) {
	pmid = strings.TrimSpace(pmid)
	if pmid == "" {
		return models.Book{}, ErrNotFound
	}
	libClause, libArgs := citationLibraryClause(opts)
	args := append([]any{pmid}, libArgs...)
	row := s.queryRowContext(ctx, selectColumns+` WHERE pubmed_id=? AND hidden=0`+libClause+` ORDER BY id ASC LIMIT 1`, args...)
	b, err := scanBook(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Book{}, ErrNotFound
	}
	return b, err
}

// FindBookByTitleAuthor finds a book by case-insensitive title and author within optional libraries.
// Author is required to avoid colliding titles across the catalog.
func (s *Store) FindBookByTitleAuthor(ctx context.Context, title, author string, opts CitationLookupOpts) (models.Book, error) {
	title = strings.TrimSpace(title)
	author = strings.TrimSpace(author)
	if title == "" || author == "" {
		return models.Book{}, ErrNotFound
	}
	libClause, libArgs := citationLibraryClause(opts)
	args := append([]any{title, author, "%" + author + "%"}, libArgs...)
	row := s.queryRowContext(ctx, selectColumns+`
 WHERE hidden=0 AND LOWER(title)=LOWER(?) AND (LOWER(author)=LOWER(?) OR author LIKE ?)`+libClause+`
 ORDER BY id ASC LIMIT 1`, args...)
	b, err := scanBook(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Book{}, ErrNotFound
	}
	return b, err
}
