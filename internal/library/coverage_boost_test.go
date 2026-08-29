package library

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func TestMergeAudiobookFoldersGroups(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.Open(filepath.Join(dir, "s.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	libDir := filepath.Join(dir, "lib")
	bookDir := filepath.Join(libDir, "Album")
	if err := os.MkdirAll(bookDir, 0o750); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	lib, err := store.CreateLibrary(ctx, "Main", libDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"01.mp3", "02.mp3"} {
		if _, err := store.UpsertBook(ctx, &models.Book{
			LibraryID: lib.ID,
			Title:     name,
			Author:    "Auth",
			Series:    "The Series",
			Format:    models.FormatMP3,
			RelPath:   filepath.ToSlash(filepath.Join("Album", name)),
			AbsPath:   filepath.Join(bookDir, name),
			FileSize:  100,
		}, 1); err != nil {
			t.Fatal(err)
		}
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	sc := New(store, filepath.Join(dir, "covers"), dir, log, 1)
	if err := sc.MergeAudiobookFolders(ctx, lib.ID); err != nil {
		t.Fatal(err)
	}
	page, err := store.ListBooks(ctx, models.BookQuery{Format: models.FormatAudiobook})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("audiobooks=%d", len(page.Items))
	}
	if page.Items[0].Title != "The Series" {
		t.Fatalf("title=%q", page.Items[0].Title)
	}
}

func TestWatcherScheduleFlush(t *testing.T) {
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
	ctx := context.Background()
	if err := store.EnsureDefaultLibrary(ctx, libDir); err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	sc := New(store, filepath.Join(dir, "covers"), dir, log, 1)
	completed := make(chan struct{}, 1)
	sc.SetOnComplete(func(ScanCompleteEvent) {
		select {
		case completed <- struct{}{}:
		default:
		}
	})
	w := NewWatcher(store, sc, log)
	w.debounce = 20 * time.Millisecond

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		w.Run(runCtx)
		close(done)
	}()
	time.Sleep(30 * time.Millisecond)
	book := filepath.Join(libDir, "watched.pdf")
	if err := os.WriteFile(book, []byte("%PDF-1.4 watched"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-completed:
	case <-time.After(3 * time.Second):
		t.Fatal("expected watcher-triggered scan completion")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("watcher hang")
	}
}

func TestIsImageBytes(t *testing.T) {
	jpeg := make([]byte, 16)
	jpeg[0], jpeg[1], jpeg[2] = 0xff, 0xd8, 0xff
	if !isImageBytes(jpeg) {
		t.Fatal("jpeg")
	}
	png := make([]byte, 16)
	copy(png, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	if !isImageBytes(png) {
		t.Fatal("png")
	}
	webp := make([]byte, 16)
	copy(webp[0:4], "RIFF")
	copy(webp[8:12], "WEBP")
	if !isImageBytes(webp) {
		t.Fatal("webp")
	}
	if isImageBytes([]byte("short")) || isImageBytes(make([]byte, 16)) {
		t.Fatal("non-image")
	}
}

func TestParseAudioChaptersFallback(t *testing.T) {
	path := writeTestID3WithChapters(t)
	ch := parseAudioChapters(path)
	if len(ch) == 0 {
		t.Fatal("expected chapters")
	}
}

func TestSearchMetadataWithMock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{
				"id": "g1",
				"volumeInfo": map[string]any{
					"title":   "Mock Book",
					"authors": []string{"Auth"},
				},
			}},
		})
	}))
	defer srv.Close()

	oldURL := googleBooksAPIURL
	googleBooksAPIURL = srv.URL
	t.Cleanup(func() { googleBooksAPIURL = oldURL })

	oldClient := sharedMetadataSearcher.client
	sharedMetadataSearcher.client = srv.Client()
	t.Cleanup(func() { sharedMetadataSearcher.client = oldClient })

	got := SearchMetadata(context.Background(), models.MetadataSearchQuery{
		Title:     "Mock Book",
		Providers: []string{"google"},
	})
	if len(got) == 0 || got[0].Title != "Mock Book" {
		t.Fatalf("got=%+v", got)
	}
	if SearchMetadata(context.Background(), models.MetadataSearchQuery{}) != nil {
		t.Fatal("empty query")
	}
}
