// Package library scans the filesystem for books, extracts metadata and
// covers, and persists them through the storage layer. It is built to
// handle very large libraries by walking lazily and parsing files
// concurrently while skipping unchanged files.
package library

import (
	"context"
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

const (
	defaultScanWorkers = 2
	maxScanWorkers     = 32
)

// Scanner indexes mounted library directories into the store.
type Scanner struct {
	store          *storage.Store
	coverDir       string
	tempDir        string
	log            *slog.Logger
	defaultWorkers int
	onComplete     func(ScanCompleteEvent)

	scanning atomic.Bool
	progress scanProgress
	wg       sync.WaitGroup
}

// ScanCompleteEvent summarizes a finished library scan.
type ScanCompleteEvent struct {
	LibraryID int64
	Indexed   int64
	Skipped   int64
	Pruned    int64
}

// SetOnComplete registers a callback invoked after a successful scan finishes.
func (s *Scanner) SetOnComplete(fn func(ScanCompleteEvent)) {
	s.onComplete = fn
}

// New creates a Scanner that reads library roots from the database.
// tempDir holds short-lived downloads when indexing remote (S3) objects.
func New(store *storage.Store, coverDir, tempDir string, log *slog.Logger, workers int) *Scanner {
	if tempDir == "" {
		tempDir = filepath.Join(coverDir, "..", "tmp")
	}
	return &Scanner{
		store:          store,
		coverDir:       coverDir,
		tempDir:        tempDir,
		log:            log,
		defaultWorkers: ClampScanWorkers(workers),
	}
}

// ClampScanWorkers bounds the configured worker count.
func ClampScanWorkers(n int) int {
	if n < 1 {
		return defaultScanWorkers
	}
	if n > maxScanWorkers {
		return maxScanWorkers
	}
	return n
}

// Scanning reports whether a scan is currently in progress.
func (s *Scanner) Scanning() bool { return s.scanning.Load() }

// Wait blocks until in-flight scans finish or ctx is done.
func (s *Scanner) Wait(ctx context.Context) error {
	return waitGroup(ctx, &s.wg)
}

// Status returns the current or most recent scan progress snapshot.
func (s *Scanner) Status() models.ScanStatus {
	st := models.ScanStatus{Scanning: s.scanning.Load()}
	st.Indexed = s.progress.indexed.Load()
	st.Skipped = s.progress.skipped.Load()
	if v := s.progress.currentPath.Load(); v != nil {
		st.CurrentPath = v.(string)
	}
	if v := s.progress.libraryName.Load(); v != nil {
		st.LibraryName = v.(string)
	}
	if v := s.progress.startedAt.Load(); v != nil {
		t := v.(time.Time)
		st.StartedAt = &t
	}
	if v := s.progress.finishedAt.Load(); v != nil {
		t := v.(time.Time)
		st.FinishedAt = &t
	}
	return st
}

type scanProgress struct {
	indexed     atomic.Int64
	skipped     atomic.Int64
	currentPath atomic.Value
	libraryName atomic.Value
	startedAt   atomic.Value
	finishedAt  atomic.Value
}

type job struct {
	libraryID int64
	root      string
	absPath   string
	relPath   string
	size      int64
	mtime     time.Time
	backend   string
}

// Scan walks every configured library mount and indexes new or changed files.
func (s *Scanner) Scan(ctx context.Context) error {
	return s.ScanLibrary(ctx, 0)
}

// ScanLibrary scans one library by id, or all libraries when libraryID is 0.
func (s *Scanner) ScanLibrary(ctx context.Context, libraryID int64) error {
	if !s.scanning.CompareAndSwap(false, true) {
		s.log.Info("scan already in progress; skipping")
		return nil
	}
	s.wg.Add(1)
	defer func() {
		s.scanning.Store(false)
		s.wg.Done()
	}()

	now := time.Now()
	s.progress.indexed.Store(0)
	s.progress.skipped.Store(0)
	s.progress.currentPath.Store("")
	s.progress.libraryName.Store("")
	s.progress.startedAt.Store(now)

	if err := os.MkdirAll(s.coverDir, 0o750); err != nil {
		return err
	}

	libs, err := s.store.ListLibraries(ctx)
	if err != nil {
		return err
	}
	if libraryID > 0 {
		filtered := libs[:0]
		for _, lib := range libs {
			if lib.ID == libraryID {
				filtered = append(filtered, lib)
				break
			}
		}
		if len(filtered) == 0 {
			return storage.ErrNotFound
		}
		libs = filtered
	}

	var totalIndexed, totalSkipped, totalPruned int64
	for _, lib := range libs {
		if lib.MountPath == "" {
			s.log.Warn("library mount path empty, skipping", "library", lib.Name, "id", lib.ID)
			continue
		}
		fs, err := s.store.OpenLibraryFS(ctx, lib.ID)
		if err != nil {
			s.log.Warn("library mount unavailable, skipping", "library", lib.Name, "path", lib.MountPath, "err", err)
			continue
		}
		indexed, skipped, pruned, err := s.scanRoot(ctx, lib, fs)
		if err != nil {
			return err
		}
		totalIndexed += indexed
		totalSkipped += skipped
		totalPruned += pruned
	}

	s.log.Info("library scan finished",
		"indexed", totalIndexed, "skipped", totalSkipped, "pruned", totalPruned, "libraries", len(libs))
	s.progress.finishedAt.Store(time.Now())
	s.progress.currentPath.Store("")
	if err := s.store.EnsureAutoCollections(ctx); err != nil {
		s.log.Warn("auto collections sync failed", "err", err)
	}
	if cb := s.onComplete; cb != nil {
		cb(ScanCompleteEvent{
			LibraryID: libraryID,
			Indexed:   totalIndexed,
			Skipped:   totalSkipped,
			Pruned:    totalPruned,
		})
	}
	return nil
}

func (s *Scanner) scanWorkers(ctx context.Context) int {
	cfg, err := s.store.GetServerConfig(ctx, false)
	if err == nil && cfg.ScanWorkers > 0 {
		return ClampScanWorkers(cfg.ScanWorkers)
	}
	return s.defaultWorkers
}

func (s *Scanner) scanRoot(ctx context.Context, lib models.Library, fs libfs.LibraryFS) (indexed, skipped, pruned int64, err error) {
	workers := s.scanWorkers(ctx)
	s.log.Info("library scan started", "library", lib.Name, "root", fs.RootLabel(), "backend", fs.Backend(), "workers", workers)
	s.progress.libraryName.Store(lib.Name)

	jobs := make(chan job, 256)
	seen := &sync.Map{}
	var idx, skip atomic.Int64

	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	walkDone := make(chan error, 1)
	go func() {
		walkDone <- fs.Walk(ctx, func(info libfs.FileInfo) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			format := FormatFromExt(info.RelPath)
			if format == "" {
				return nil
			}
			abs := info.RelPath
			if fs.Backend() == libfs.BackendLocal {
				abs = filepath.Join(fs.RootLabel(), filepath.FromSlash(info.RelPath))
			}
			jobs <- job{
				libraryID: lib.ID,
				root:      fs.RootLabel(),
				absPath:   abs,
				relPath:   info.RelPath,
				size:      info.Size,
				mtime:     info.ModTime,
				backend:   fs.Backend(),
			}
			return nil
		})
		close(jobs)
	}()

	for j := range jobs {
		wg.Add(1)
		go func(j job) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			s.indexFile(ctx, fs, j, &idx, &skip, seen)
		}(j)
	}
	wg.Wait()

	walkErr := <-walkDone
	s.progress.indexed.Add(idx.Load())
	s.progress.skipped.Add(skip.Load())
	if walkErr != nil {
		return 0, 0, 0, walkErr
	}

	if err := s.MergeAudiobookFolders(ctx, lib.ID); err != nil {
		s.log.Warn("audiobook merge failed", "library", lib.ID, "err", err)
	}

	keep := make(map[string]struct{})
	seen.Range(func(k, _ any) bool {
		keep[k.(string)] = struct{}{}
		return true
	})
	setPaths, err := s.store.ListAudiobookSetRelPaths(ctx, lib.ID)
	if err != nil {
		return idx.Load(), skip.Load(), 0, err
	}
	for _, p := range setPaths {
		keep[p] = struct{}{}
	}
	removed, err := s.store.PrunePaths(ctx, lib.ID, keep)
	if err != nil {
		return idx.Load(), skip.Load(), 0, err
	}
	return idx.Load(), skip.Load(), int64(removed), nil
}

func (s *Scanner) indexFile(ctx context.Context, fs libfs.LibraryFS, j job, indexed, skipped *atomic.Int64, seen *sync.Map) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Error("index file panic", "path", j.relPath, "err", r)
			seen.Store(j.relPath, struct{}{})
		}
	}()
	s.progress.currentPath.Store(j.relPath)
	if s.process(ctx, fs, j, indexed, skipped) {
		seen.Store(j.relPath, struct{}{})
	}
}

func (s *Scanner) process(ctx context.Context, fs libfs.LibraryFS, j job, indexed, skipped *atomic.Int64) bool {
	mtime := j.mtime.Unix()
	size := j.size

	if oldM, oldS, ok, err := s.store.FileState(ctx, j.libraryID, j.relPath); err == nil && ok {
		if oldM == mtime && oldS == size {
			skipped.Add(1)
			return true
		}
	}

	parsePath := j.absPath
	if j.backend == libfs.BackendS3 || (fs != nil && fs.Backend() == libfs.BackendS3) {
		tmp, err := libfs.Materialize(ctx, fs, j.relPath, s.tempDir)
		if err != nil {
			s.log.Warn("materialize failed", "path", j.relPath, "err", err)
			return false
		}
		parsePath = tmp
		defer os.Remove(tmp)
	}

	format := FormatFromExt(j.relPath)
	absPath := j.absPath
	if j.backend == libfs.BackendS3 {
		absPath = strings.TrimSuffix(j.root, "/") + "/" + j.relPath
	}
	book := &models.Book{
		LibraryID: j.libraryID,
		Format:    format,
		RelPath:   j.relPath,
		AbsPath:   absPath,
		FileSize:  size,
	}

	var coverData []byte
	var chapters []models.Chapter
	var isbn, asin string

	switch format {
	case models.FormatEPUB:
		if meta, err := parseEPUB(parsePath); err == nil {
			applyEPUB(book, meta)
			coverData = meta.CoverData
		} else {
			s.log.Warn("epub parse failed", "path", j.relPath, "err", err)
		}
	case models.FormatPDF:
		meta := parsePDF(parsePath)
		book.Title = meta.Title
		book.Author = meta.Author
		coverData = meta.CoverData
		book.HasCover = len(coverData) > 0
	case models.FormatMOBI, models.FormatAZW3, models.FormatAZW:
		meta := parseMobiFamily(parsePath)
		book.Title = meta.Title
		book.Author = meta.Author
		book.Language = meta.Language
		book.Description = meta.Description
		coverData = meta.CoverData
		book.HasCover = len(coverData) > 0
	case models.FormatKFX:
		book.Title = filepath.Base(j.relPath)
	case models.FormatCBZ, models.FormatCBR:
		meta := parseComic(parsePath)
		coverData = meta.CoverData
		book.HasCover = len(coverData) > 0
	default:
		if models.IsAudio(format) {
			meta := parseAudio(parsePath)
			book.Title = meta.Title
			book.Author = meta.Author
			if meta.Album != "" && book.Series == "" {
				book.Series = CleanSeriesName(meta.Album)
			}
			coverData = meta.CoverData
			if len(coverData) == 0 && j.backend == libfs.BackendLocal {
				coverData = sidecarCover(parsePath)
			}
			book.HasCover = len(coverData) > 0
			chapters = meta.Chapters
			isbn, asin = meta.ISBN, meta.ASIN
		}
	}

	if j.backend == libfs.BackendLocal {
		side := sidecarMetadata(parsePath)
		sbISBN, sbASIN := applySidecarMeta(book, side)
		isbn = mergeIdentifiers(isbn, sbISBN)
		asin = mergeIdentifiers(asin, sbASIN)
	}
	applyFilenameMeta(book, parseFilenameMeta(j.relPath))
	enrichBookMetadata(ctx, book, isbn, asin)
	normalizeBookText(book, j.relPath)

	if book.Title == "" {
		book.Title = filepath.Base(j.relPath)
	}

	skipCover := false
	if existing, err := s.store.GetBookByPath(ctx, j.libraryID, j.relPath); err == nil {
		book.MetaEdited = existing.MetaEdited
		book.CoverEdited = existing.CoverEdited
		if existing.CoverEdited {
			skipCover = true
			book.HasCover = existing.HasCover
			coverData = nil
		}
	}

	id, err := s.store.UpsertBook(ctx, book, mtime)
	if err != nil {
		s.log.Error("upsert failed", "path", j.relPath, "err", err)
		return false
	}

	if len(coverData) > 0 && !skipCover {
		s.writeCover(id, coverData)
	}

	if err := s.store.ReplaceChapters(ctx, id, chapters); err != nil {
		s.log.Warn("chapter save failed", "id", id, "err", err)
	}

	if hash, err := FileSHA256(parsePath); err == nil && hash != "" {
		if err := s.store.ApplyContentHash(ctx, id, hash); err != nil {
			s.log.Warn("content hash failed", "id", id, "err", err)
		}
	}

	indexed.Add(1)
	return true
}

func applyEPUB(b *models.Book, m epubMeta) {
	b.Title = m.Title
	b.Author = m.Author
	b.Language = m.Language
	b.Description = m.Description
	b.Series = CleanSeriesName(m.Series)
	b.SeriesIndex = m.SeriesIndex
	b.HasCover = len(m.CoverData) > 0
}

func (s *Scanner) writeCover(id int64, data []byte) {
	path := filepath.Join(s.coverDir, coverFilename(id))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		s.log.Warn("cover write failed", "id", id, "err", err)
	}
}

func coverFilename(id int64) string {
	return strconv.FormatInt(id, 10) + ".img"
}

// CoverPath returns the absolute path to a book's cached cover image.
func CoverPath(coverDir string, id int64) string {
	return filepath.Join(coverDir, coverFilename(id))
}

// formatFromExt is deprecated; use FormatFromExt.
func formatFromExt(p string) string { return FormatFromExt(p) }
