package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

// PROVED_SHARE_MAX_DOWNLOAD_RACE
// Guarantee claimed: MaxDownloads is honored under concurrent downloaders.

func TestShareMaxDownloadsRaceOracle(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}
	libDir := filepath.Join(srv.cfg.DataDir, "lib")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libDir, "a.epub"), []byte("epub-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	lib, err := store.CreateLibrary(ctx, "Main", libDir)
	if err != nil {
		t.Fatal(err)
	}
	bookID, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Race",
		Format:    models.FormatEPUB,
		RelPath:   "a.epub",
	}, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	max := int64(1)
	sl, err := store.CreateShareLink(ctx, bookID, 1, nil, max)
	if err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	var okCount atomic.Int64
	var wg sync.WaitGroup
	for range 32 {
		wg.Go(func() {
			req := httptest.NewRequest(http.MethodGet, "/share/"+sl.Token+"/download", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code == http.StatusOK {
				okCount.Add(1)
			}
		})
	}
	wg.Wait()

	got := okCount.Load()
	if got > max {
		t.Fatalf("MaxDownloads not honored under concurrency: got %d > max %d", got, max)
	}
	fmt.Println("PROVED_SHARE_MAX_DOWNLOAD_RACE: concurrent successes=", got, "max=", max)
}
