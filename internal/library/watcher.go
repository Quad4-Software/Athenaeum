package library

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
	"github.com/fsnotify/fsnotify"
)

// Watcher reacts to filesystem changes under library mounts.
type Watcher struct {
	store    *storage.Store
	scanner  *Scanner
	log      *slog.Logger
	debounce time.Duration

	mu      sync.Mutex
	pending map[int64]struct{}
	timer   *time.Timer
}

// NewWatcher creates a filesystem watcher with debounced rescans.
func NewWatcher(store *storage.Store, scanner *Scanner, log *slog.Logger) *Watcher {
	return &Watcher{
		store:    store,
		scanner:  scanner,
		log:      log,
		debounce: 2 * time.Second,
		pending:  make(map[int64]struct{}),
	}
}

// Run watches library mount paths until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		w.log.Warn("fsnotify unavailable", "err", err)
		return
	}
	defer fw.Close()

	if err := w.watchAll(ctx, fw); err != nil {
		w.log.Warn("initial watch setup failed", "err", err)
	}

	go w.refreshMounts(ctx, fw)

	for {
		select {
		case <-ctx.Done():
			w.mu.Lock()
			if w.timer != nil {
				w.timer.Stop()
				w.timer = nil
			}
			w.pending = make(map[int64]struct{})
			w.mu.Unlock()
			return
		case ev, ok := <-fw.Events:
			if !ok {
				return
			}
			if ev.Has(fsnotify.Create) || ev.Has(fsnotify.Write) || ev.Has(fsnotify.Remove) || ev.Has(fsnotify.Rename) {
				w.schedule(ctx, ev.Name)
			}
		case err, ok := <-fw.Errors:
			if !ok {
				return
			}
			w.log.Warn("fsnotify error", "err", err)
		}
	}
}

func (w *Watcher) watchAll(ctx context.Context, fw *fsnotify.Watcher) error {
	libs, err := w.store.ListLibraries(ctx)
	if err != nil {
		return err
	}
	for _, lib := range libs {
		if lib.Backend != "" && lib.Backend != models.LibraryBackendLocal {
			continue
		}
		if lib.MountPath != "" && !strings.HasPrefix(lib.MountPath, "s3://") {
			_ = w.addRecursive(fw, lib.MountPath)
		}
	}
	return nil
}

func (w *Watcher) refreshMounts(ctx context.Context, fw *fsnotify.Watcher) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = w.watchAll(ctx, fw)
		}
	}
}

func (w *Watcher) addRecursive(fw *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") && path != root {
			return filepath.SkipDir
		}
		_ = fw.Add(path)
		return nil
	})
}

func (w *Watcher) schedule(ctx context.Context, absPath string) {
	libs, err := w.store.ListLibraries(ctx)
	if err != nil {
		return
	}
	for _, lib := range libs {
		if lib.Backend != "" && lib.Backend != models.LibraryBackendLocal {
			continue
		}
		if lib.MountPath == "" || strings.HasPrefix(lib.MountPath, "s3://") {
			continue
		}
		rel, err := filepath.Rel(lib.MountPath, absPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}
		w.mu.Lock()
		w.pending[lib.ID] = struct{}{}
		if w.timer != nil {
			w.timer.Stop()
		}
		w.timer = time.AfterFunc(w.debounce, func() {
			w.flush(ctx)
		})
		w.mu.Unlock()
		return
	}
}

func (w *Watcher) flush(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	w.mu.Lock()
	pending := w.pending
	w.pending = make(map[int64]struct{})
	w.mu.Unlock()
	if len(pending) == 0 || w.scanner.Scanning() {
		return
	}
	for libID := range pending {
		if ctx.Err() != nil {
			return
		}
		if err := w.scanner.ScanLibrary(ctx, libID); err != nil {
			w.log.Warn("watch scan failed", "library", libID, "err", err)
		}
	}
}
