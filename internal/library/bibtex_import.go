package library

import (
	"context"
	"errors"
	"strings"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

const maxBibImportEntries = 5000

// BibImportResult summarises a BibTeX import against existing library files.
type BibImportResult struct {
	Matched         int      `json:"matched"`
	Updated         int      `json:"updated"`
	Unmatched       int      `json:"unmatched"`
	Skipped         int      `json:"skipped"`
	UnmatchedTitles []string `json:"unmatchedTitles,omitempty"`
}

// BibImportOptions scopes which libraries import may update.
type BibImportOptions struct {
	LibraryIDs []int64 // empty = unrestricted (admins / auth disabled)
}

type bibBookStore interface {
	FindBookByDOI(ctx context.Context, doi string, opts storage.CitationLookupOpts) (models.Book, error)
	FindBookByArxivID(ctx context.Context, arxivID string, opts storage.CitationLookupOpts) (models.Book, error)
	FindBookByPubmedID(ctx context.Context, pmid string, opts storage.CitationLookupOpts) (models.Book, error)
	FindBookByTitleAuthor(ctx context.Context, title, author string, opts storage.CitationLookupOpts) (models.Book, error)
	UpdateBookMetadata(ctx context.Context, id int64, u models.BookUpdate) (models.Book, error)
}

// ImportBibTeX matches BibTeX entries to existing books and applies citation metadata.
func ImportBibTeX(ctx context.Context, store bibBookStore, data string, opts BibImportOptions) (BibImportResult, error) {
	entries := ParseBibTeX(data)
	if len(entries) > maxBibImportEntries {
		entries = entries[:maxBibImportEntries]
	}
	lookup := storage.CitationLookupOpts{LibraryIDs: opts.LibraryIDs}
	var result BibImportResult
	for _, e := range entries {
		side := BibEntryToSidecar(e)
		if side.Title == "" && side.DOI == "" && side.ArxivID == "" && side.PubmedID == "" {
			result.Skipped++
			continue
		}
		book, err := matchBibEntry(ctx, store, side, lookup)
		if errors.Is(err, storage.ErrNotFound) {
			result.Unmatched++
			if side.Title != "" && len(result.UnmatchedTitles) < 20 {
				result.UnmatchedTitles = append(result.UnmatchedTitles, side.Title)
			}
			continue
		}
		if err != nil {
			return result, err
		}
		result.Matched++
		if book.MetaEdited {
			result.Skipped++
			continue
		}
		update := bookToUpdate(book)
		mergeCitationUpdate(&update, side)
		if _, err := store.UpdateBookMetadata(ctx, book.ID, update); err != nil {
			return result, err
		}
		result.Updated++
	}
	return result, nil
}

func matchBibEntry(ctx context.Context, store bibBookStore, side sidecarFields, opts storage.CitationLookupOpts) (models.Book, error) {
	if side.DOI != "" {
		if b, err := store.FindBookByDOI(ctx, side.DOI, opts); err == nil {
			return b, nil
		} else if !errors.Is(err, storage.ErrNotFound) {
			return models.Book{}, err
		}
	}
	if side.ArxivID != "" {
		if b, err := store.FindBookByArxivID(ctx, side.ArxivID, opts); err == nil {
			return b, nil
		} else if !errors.Is(err, storage.ErrNotFound) {
			return models.Book{}, err
		}
	}
	if side.PubmedID != "" {
		if b, err := store.FindBookByPubmedID(ctx, side.PubmedID, opts); err == nil {
			return b, nil
		} else if !errors.Is(err, storage.ErrNotFound) {
			return models.Book{}, err
		}
	}
	if side.Title != "" && side.Author != "" {
		return store.FindBookByTitleAuthor(ctx, side.Title, side.Author, opts)
	}
	return models.Book{}, storage.ErrNotFound
}

func bookToUpdate(b models.Book) models.BookUpdate {
	return models.BookUpdate{
		Title:         b.Title,
		Author:        b.Author,
		Series:        b.Series,
		SeriesIndex:   b.SeriesIndex,
		Language:      b.Language,
		Description:   b.Description,
		DOI:           b.DOI,
		ArxivID:       b.ArxivID,
		PubmedID:      b.PubmedID,
		Journal:       b.Journal,
		Volume:        b.Volume,
		Issue:         b.Issue,
		Pages:         b.Pages,
		PublishedYear: b.PublishedYear,
	}
}

func mergeCitationUpdate(u *models.BookUpdate, side sidecarFields) {
	// Fill title/author only when empty so BibTeX does not clobber good PDF metadata.
	if side.Title != "" && strings.TrimSpace(u.Title) == "" {
		u.Title = side.Title
	}
	if side.Author != "" && strings.TrimSpace(u.Author) == "" {
		u.Author = side.Author
	}
	if side.Description != "" && u.Description == "" {
		u.Description = side.Description
	}
	if side.Language != "" && u.Language == "" {
		u.Language = side.Language
	}
	if d := NormalizeDOI(side.DOI); d != "" {
		u.DOI = d
	}
	if a := NormalizeArxivID(side.ArxivID); a != "" {
		u.ArxivID = a
	}
	if p := NormalizePubmedID(side.PubmedID); p != "" {
		u.PubmedID = p
	}
	if side.Journal != "" {
		u.Journal = side.Journal
	}
	if side.Volume != "" {
		u.Volume = side.Volume
	}
	if side.Issue != "" {
		u.Issue = side.Issue
	}
	if side.Pages != "" {
		u.Pages = side.Pages
	}
	if side.PublishedYear > 0 {
		u.PublishedYear = side.PublishedYear
	}
}

// BookUpdateFromSidecar is used by tests and import helpers.
func BookUpdateFromSidecar(side sidecarFields) models.BookUpdate {
	return models.BookUpdate{
		Title:         strings.TrimSpace(side.Title),
		Author:        strings.TrimSpace(side.Author),
		Description:   strings.TrimSpace(side.Description),
		Language:      strings.TrimSpace(side.Language),
		DOI:           NormalizeDOI(side.DOI),
		ArxivID:       NormalizeArxivID(side.ArxivID),
		PubmedID:      NormalizePubmedID(side.PubmedID),
		Journal:       strings.TrimSpace(side.Journal),
		Volume:        strings.TrimSpace(side.Volume),
		Issue:         strings.TrimSpace(side.Issue),
		Pages:         strings.TrimSpace(side.Pages),
		PublishedYear: side.PublishedYear,
	}
}
