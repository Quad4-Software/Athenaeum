package library

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"athenaeum/internal/storage"
)

func TestWatcherRunCancel(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.Open(filepath.Join(dir, "s.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	libDir := filepath.Join(dir, "lib")
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := store.EnsureDefaultLibrary(context.Background(), libDir); err != nil {
		t.Fatal(err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	sc := New(store, filepath.Join(dir, "covers"), dir, log, 1)
	w := NewWatcher(store, sc, log)
	if w.debounce != 2*time.Second {
		t.Fatalf("debounce=%v", w.debounce)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not stop")
	}
}
