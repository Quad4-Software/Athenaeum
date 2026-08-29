package library

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

// MetadataMatcher runs background metadata identification jobs.
type MetadataMatcher struct {
	store    *storage.Store
	coverDir string
	log      *slog.Logger

	running  atomic.Bool
	progress metadataMatchProgress
	wg       sync.WaitGroup
}

// NewMetadataMatcher creates a matcher for bulk metadata jobs.
func NewMetadataMatcher(store *storage.Store, coverDir string, log *slog.Logger) *MetadataMatcher {
	return &MetadataMatcher{store: store, coverDir: coverDir, log: log}
}

type metadataMatchProgress struct {
	total        atomic.Int64
	done         atomic.Int64
	matched      atomic.Int64
	skipped      atomic.Int64
	failed       atomic.Int64
	currentTitle atomic.Value
	startedAt    atomic.Value
	finishedAt   atomic.Value
}

// MetadataMatchStatus reports live metadata match progress.
type MetadataMatchStatus struct {
	Running      bool       `json:"running"`
	Total        int64      `json:"total"`
	Done         int64      `json:"done"`
	Matched      int64      `json:"matched"`
	Skipped      int64      `json:"skipped"`
	Failed       int64      `json:"failed"`
	CurrentTitle string     `json:"currentTitle,omitempty"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	FinishedAt   *time.Time `json:"finishedAt,omitempty"`
}

// MetadataAutoMatchRequest selects books for automatic metadata matching.
type MetadataAutoMatchRequest struct {
	BookIDs    []int64 `json:"bookIds,omitempty"`
	LibraryID  int64   `json:"libraryId,omitempty"`
	ApplyCover bool    `json:"applyCover"`
}

// Running reports whether a metadata job is active.
func (m *MetadataMatcher) Running() bool { return m.running.Load() }

// Status returns the current or most recent job snapshot.
func (m *MetadataMatcher) Status() MetadataMatchStatus {
	st := MetadataMatchStatus{Running: m.running.Load()}
	st.Total = m.progress.total.Load()
	st.Done = m.progress.done.Load()
	st.Matched = m.progress.matched.Load()
	st.Skipped = m.progress.skipped.Load()
	st.Failed = m.progress.failed.Load()
	if v := m.progress.currentTitle.Load(); v != nil {
		if title := v.(string); title != "" {
			st.CurrentTitle = title
		}
	}
	if v := m.progress.startedAt.Load(); v != nil {
		t := v.(time.Time)
		st.StartedAt = &t
	}
	if !st.Running {
		if v := m.progress.finishedAt.Load(); v != nil {
			t := v.(time.Time)
			st.FinishedAt = &t
		}
	}
	return st
}

// Wait blocks until in-flight metadata jobs finish or ctx is done.
func (m *MetadataMatcher) Wait(ctx context.Context) error {
	return waitGroup(ctx, &m.wg)
}

// Start launches a metadata matching job in the background.
// ctx should outlive the HTTP request (process/job lifetime).
func (m *MetadataMatcher) Start(ctx context.Context, req MetadataAutoMatchRequest) bool {
	if !m.running.CompareAndSwap(false, true) {
		return false
	}
	m.wg.Go(func() {
		m.run(ctx, req)
	})
	return true
}

func (m *MetadataMatcher) run(ctx context.Context, req MetadataAutoMatchRequest) {
	defer m.running.Store(false)

	now := time.Now()
	m.progress.total.Store(0)
	m.progress.done.Store(0)
	m.progress.matched.Store(0)
	m.progress.skipped.Store(0)
	m.progress.failed.Store(0)
	m.progress.currentTitle.Store("")
	m.progress.startedAt.Store(now)

	books, err := m.store.ListBooksForMetadata(ctx, req.LibraryID, req.BookIDs)
	if err != nil {
		m.log.Error("metadata match list books", "err", err)
		finished := time.Now()
		m.progress.finishedAt.Store(finished)
		return
	}

	m.progress.total.Store(int64(len(books)))

	const workers = 3
	jobs := make(chan models.Book, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for book := range jobs {
				m.matchOne(ctx, book, req.ApplyCover)
			}
		})
	}

	for _, book := range books {
		if ctx.Err() != nil {
			break
		}
		jobs <- book
	}
	close(jobs)
	wg.Wait()

	finished := time.Now()
	m.progress.finishedAt.Store(finished)
	m.progress.currentTitle.Store("")
}

func (m *MetadataMatcher) matchOne(ctx context.Context, book models.Book, applyCover bool) {
	defer m.progress.done.Add(1)
	m.progress.currentTitle.Store(book.Title)

	matches := SearchMetadata(ctx, models.MetadataSearchQuery{
		Title:  book.Title,
		Author: book.Author,
	})
	match, ok := BestMetadataMatch(book, matches)
	if !ok {
		m.progress.skipped.Add(1)
		return
	}

	update := MatchToBookUpdate(match)
	if update.Title == "" {
		m.progress.skipped.Add(1)
		return
	}

	if _, err := m.store.UpdateBookMetadata(ctx, book.ID, update); err != nil {
		m.log.Warn("metadata match update failed", "id", book.ID, "err", err)
		m.progress.failed.Add(1)
		return
	}

	if applyCover && match.CoverURL != "" {
		if data, err := FetchCoverImage(ctx, match.CoverURL); err == nil {
			_ = writeCoverBytes(m.coverDir, book.ID, data)
			_ = m.store.SetBookCover(ctx, book.ID, true)
		}
	}

	m.progress.matched.Add(1)
}

func writeCoverBytes(coverDir string, id int64, data []byte) error {
	return os.WriteFile(CoverPath(coverDir, id), data, 0o600)
}

// CleanStoredSeriesNames normalizes series names for all books in the database.
func CleanStoredSeriesNames(ctx context.Context, store *storage.Store) (int64, error) {
	books, err := store.ListBooksForMetadata(ctx, 0, nil)
	if err != nil {
		return 0, err
	}
	var updated int64
	for _, b := range books {
		if b.Series == "" || b.MetaEdited {
			continue
		}
		cleaned := CleanSeriesName(b.Series)
		if cleaned == b.Series {
			continue
		}
		if _, err := store.UpdateBookMetadata(ctx, b.ID, models.BookUpdate{
			Title:       b.Title,
			Author:      b.Author,
			Series:      cleaned,
			SeriesIndex: b.SeriesIndex,
			Language:    b.Language,
			Description: b.Description,
		}); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}

// CleanStoredBookText normalizes titles and authors, falling back to filename parsing when metadata is garbled.
func CleanStoredBookText(ctx context.Context, store *storage.Store) (int64, error) {
	books, err := store.ListBooksForMetadata(ctx, 0, nil)
	if err != nil {
		return 0, err
	}
	var updated int64
	for _, b := range books {
		if b.MetaEdited {
			continue
		}
		path := b.RelPath
		if path == "" {
			path = b.AbsPath
		}
		newTitle := CleanBookTitle(b.Title, path)
		newAuthor := CleanAuthorName(b.Author, path)
		newSeries := CleanSeriesName(b.Series)
		if newTitle == b.Title && newAuthor == b.Author && newSeries == b.Series {
			continue
		}
		if _, err := store.UpdateBookMetadata(ctx, b.ID, models.BookUpdate{
			Title:       newTitle,
			Author:      newAuthor,
			Series:      newSeries,
			SeriesIndex: b.SeriesIndex,
			Language:    b.Language,
			Description: b.Description,
		}); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}
