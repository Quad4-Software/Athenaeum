package library

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"athenaeum/internal/libfs"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

// MissingBook is a catalog entry whose file is absent on disk.
type MissingBook struct {
	ID        int64  `json:"id"`
	LibraryID int64  `json:"libraryId"`
	Title     string `json:"title"`
	RelPath   string `json:"relPath"`
}

// IntegrityReport summarizes catalog vs filesystem consistency.
type IntegrityReport struct {
	TotalBooks   int           `json:"totalBooks"`
	MissingCount int           `json:"missingCount"`
	MissingFiles []MissingBook `json:"missingFiles"`
	OrphanCovers int           `json:"orphanCovers"`
}

// MaintenanceStatus reports progress for long-running admin tasks.
type MaintenanceStatus struct {
	Running      bool       `json:"running"`
	Task         string     `json:"task,omitempty"`
	Total        int64      `json:"total"`
	Done         int64      `json:"done"`
	Updated      int64      `json:"updated"`
	Skipped      int64      `json:"skipped"`
	Failed       int64      `json:"failed"`
	CurrentTitle string     `json:"currentTitle,omitempty"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	FinishedAt   *time.Time `json:"finishedAt,omitempty"`
}

// Maintenance runs admin catalog maintenance jobs.
type Maintenance struct {
	store    *storage.Store
	coverDir string
	log      *slog.Logger

	running  atomic.Bool
	progress maintenanceProgress
	wg       sync.WaitGroup
}

type maintenanceProgress struct {
	task         atomic.Value
	total        atomic.Int64
	done         atomic.Int64
	updated      atomic.Int64
	skipped      atomic.Int64
	failed       atomic.Int64
	currentTitle atomic.Value
	startedAt    atomic.Value
	finishedAt   atomic.Value
}

// NewMaintenance creates a maintenance job runner.
func NewMaintenance(store *storage.Store, coverDir string, log *slog.Logger) *Maintenance {
	return &Maintenance{store: store, coverDir: coverDir, log: log}
}

// Running reports whether a background maintenance job is active.
func (m *Maintenance) Running() bool { return m.running.Load() }

// Status returns the current or most recent job snapshot.
func (m *Maintenance) Status() MaintenanceStatus {
	st := MaintenanceStatus{Running: m.running.Load()}
	if v := m.progress.task.Load(); v != nil {
		st.Task = v.(string)
	}
	st.Total = m.progress.total.Load()
	st.Done = m.progress.done.Load()
	st.Updated = m.progress.updated.Load()
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

// VerifyIntegrity checks that indexed books and cover cache entries are consistent.
func VerifyIntegrity(ctx context.Context, store *storage.Store, coverDir string) (IntegrityReport, error) {
	books, err := store.ListBooksForMetadata(ctx, 0, nil)
	if err != nil {
		return IntegrityReport{}, err
	}

	fsCache := make(map[int64]libfs.LibraryFS)
	report := IntegrityReport{TotalBooks: len(books)}

	for _, book := range books {
		fs, ok := fsCache[book.LibraryID]
		if !ok {
			fs, err = store.OpenLibraryFS(ctx, book.LibraryID)
			if err != nil {
				return IntegrityReport{}, err
			}
			fsCache[book.LibraryID] = fs
		}
		missing, err := bookFileMissing(ctx, store, fs, book)
		if err != nil {
			return IntegrityReport{}, err
		}
		if missing {
			report.MissingCount++
			if len(report.MissingFiles) < 100 {
				report.MissingFiles = append(report.MissingFiles, MissingBook{
					ID:        book.ID,
					LibraryID: book.LibraryID,
					Title:     book.Title,
					RelPath:   book.RelPath,
				})
			}
		}
	}

	orphan, err := countOrphanCovers(ctx, store, coverDir)
	if err != nil {
		return IntegrityReport{}, err
	}
	report.OrphanCovers = orphan
	if report.MissingFiles == nil {
		report.MissingFiles = []MissingBook{}
	}
	return report, nil
}

func bookFileMissing(ctx context.Context, store *storage.Store, fs libfs.LibraryFS, book models.Book) (bool, error) {
	if book.Format == models.FormatAudiobook {
		tracks, err := store.ListAudiobookTracks(ctx, book.ID)
		if err != nil {
			return false, err
		}
		if len(tracks) == 0 {
			return true, nil
		}
		for _, tr := range tracks {
			if _, err := fs.Stat(ctx, tr.RelPath); err != nil {
				if errors.Is(err, libfs.ErrNotExist) {
					return true, nil
				}
				return false, err
			}
		}
		return false, nil
	}
	_, err := fs.Stat(ctx, book.RelPath)
	if err == nil {
		return false, nil
	}
	if errors.Is(err, libfs.ErrNotExist) {
		return true, nil
	}
	return false, err
}

func countOrphanCovers(ctx context.Context, store *storage.Store, coverDir string) (int, error) {
	entries, err := os.ReadDir(coverDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	var orphan int
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".img") {
			continue
		}
		id, err := strconv.ParseInt(strings.TrimSuffix(ent.Name(), ".img"), 10, 64)
		if err != nil || id <= 0 {
			orphan++
			continue
		}
		if _, err := store.GetBook(ctx, id); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				orphan++
				continue
			}
			return 0, err
		}
	}
	return orphan, nil
}

// PruneMissingBooks removes catalog rows whose files are missing on disk.
func PruneMissingBooks(ctx context.Context, store *storage.Store) (int, error) {
	books, err := store.ListBooksForMetadata(ctx, 0, nil)
	if err != nil {
		return 0, err
	}
	fsCache := make(map[int64]libfs.LibraryFS)
	var removed int
	for _, book := range books {
		fs, ok := fsCache[book.LibraryID]
		if !ok {
			fs, err = store.OpenLibraryFS(ctx, book.LibraryID)
			if err != nil {
				return removed, err
			}
			fsCache[book.LibraryID] = fs
		}
		missing, err := bookFileMissing(ctx, store, fs, book)
		if err != nil {
			return removed, err
		}
		if !missing {
			continue
		}
		if err := store.DeleteBook(ctx, book.ID); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

// CleanupOrphanCovers deletes cached cover images with no matching book row.
func CleanupOrphanCovers(ctx context.Context, store *storage.Store, coverDir string) (int, error) {
	entries, err := os.ReadDir(coverDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	var removed int
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".img") {
			continue
		}
		id, err := strconv.ParseInt(strings.TrimSuffix(ent.Name(), ".img"), 10, 64)
		if err != nil || id <= 0 {
			if rmErr := os.Remove(filepath.Join(coverDir, ent.Name())); rmErr == nil {
				removed++
			}
			continue
		}
		if _, err := store.GetBook(ctx, id); errors.Is(err, storage.ErrNotFound) {
			if rmErr := os.Remove(filepath.Join(coverDir, ent.Name())); rmErr == nil {
				removed++
			}
		} else if err != nil {
			return removed, err
		}
	}
	return removed, nil
}

// Wait blocks until in-flight maintenance jobs finish or ctx is done.
func (m *Maintenance) Wait(ctx context.Context) error {
	return waitGroup(ctx, &m.wg)
}

// StartRegenerateCovers launches background cover re-extraction for non-edited covers.
// ctx should outlive the HTTP request (process/job lifetime).
func (m *Maintenance) StartRegenerateCovers(ctx context.Context, libraryID int64) bool {
	if !m.running.CompareAndSwap(false, true) {
		return false
	}
	m.wg.Go(func() {
		m.runRegenerateCovers(ctx, libraryID)
	})
	return true
}

func (m *Maintenance) runRegenerateCovers(ctx context.Context, libraryID int64) {
	defer m.running.Store(false)

	now := time.Now()
	m.progress.task.Store("regenerate_covers")
	m.progress.total.Store(0)
	m.progress.done.Store(0)
	m.progress.updated.Store(0)
	m.progress.skipped.Store(0)
	m.progress.failed.Store(0)
	m.progress.currentTitle.Store("")
	m.progress.startedAt.Store(now)

	books, err := m.store.ListBooksForMetadata(ctx, libraryID, nil)
	if err != nil {
		m.log.Error("regenerate covers list books", "err", err)
		finished := time.Now()
		m.progress.finishedAt.Store(finished)
		return
	}

	m.progress.total.Store(int64(len(books)))
	fsCache := &libraryFSCache{m: make(map[int64]libfs.LibraryFS)}

	const workers = 3
	jobs := make(chan models.Book, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for book := range jobs {
				m.regenerateOneCover(ctx, book, fsCache)
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

type libraryFSCache struct {
	mu sync.Mutex
	m  map[int64]libfs.LibraryFS
}

func (c *libraryFSCache) get(ctx context.Context, store *storage.Store, libraryID int64) (libfs.LibraryFS, error) {
	c.mu.Lock()
	fs, ok := c.m[libraryID]
	c.mu.Unlock()
	if ok {
		return fs, nil
	}
	fs, err := store.OpenLibraryFS(ctx, libraryID)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.m[libraryID] = fs
	c.mu.Unlock()
	return fs, nil
}

func (m *Maintenance) regenerateOneCover(ctx context.Context, book models.Book, cache *libraryFSCache) {
	defer m.progress.done.Add(1)
	m.progress.currentTitle.Store(book.Title)

	if book.CoverEdited {
		m.progress.skipped.Add(1)
		return
	}
	if book.Format == models.FormatAudiobook || book.Format == models.FormatKFX {
		m.progress.skipped.Add(1)
		return
	}

	fs, err := cache.get(ctx, m.store, book.LibraryID)
	if err != nil {
		m.log.Warn("regenerate covers mount", "libraryId", book.LibraryID, "err", err)
		m.progress.failed.Add(1)
		return
	}

	var path string
	var cleanup func()
	if fs.Backend() == libfs.BackendLocal {
		full, err := fs.LocalAbsPath(book.RelPath)
		if err != nil {
			m.progress.skipped.Add(1)
			return
		}
		if _, err := os.Stat(full); err != nil {
			m.progress.skipped.Add(1)
			return
		}
		path = full
		cleanup = func() {}
	} else {
		tmpDir := filepath.Join(m.coverDir, "..", "tmp")
		path, err = libfs.Materialize(ctx, fs, book.RelPath, tmpDir)
		if err != nil {
			m.progress.skipped.Add(1)
			return
		}
		cleanup = func() { _ = os.Remove(path) }
	}
	defer cleanup()

	data := ExtractCoverFromFile(path, book.Format)
	if len(data) == 0 {
		if err := m.store.SetBookHasCover(ctx, book.ID, false); err != nil {
			m.log.Warn("regenerate covers clear flag", "id", book.ID, "err", err)
		}
		m.progress.skipped.Add(1)
		return
	}

	if err := writeCoverBytes(m.coverDir, book.ID, data); err != nil {
		m.log.Warn("regenerate covers write", "id", book.ID, "err", err)
		m.progress.failed.Add(1)
		return
	}
	if err := m.store.SetBookHasCover(ctx, book.ID, true); err != nil {
		m.log.Warn("regenerate covers update", "id", book.ID, "err", err)
		m.progress.failed.Add(1)
		return
	}
	m.progress.updated.Add(1)
}
