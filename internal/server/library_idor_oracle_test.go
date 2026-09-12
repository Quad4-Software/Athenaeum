package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

// PROVED_LIBRARY_RESTRICTED_CHAPTERS_IDOR
// Guarantee: users restricted to library A cannot read chapters for books in library B.

func TestRestrictedLibraryChaptersIDOROracle(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	adminHash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", adminHash, true); err != nil {
		t.Fatal(err)
	}
	userHash, _ := auth.HashPassword("longpassword")
	uid, err := store.CreateUser(ctx, "reader", userHash, false)
	if err != nil {
		t.Fatal(err)
	}

	libA := filepath.Join(srv.cfg.DataDir, "libA")
	libB := filepath.Join(srv.cfg.DataDir, "libB")
	if err := os.MkdirAll(libA, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(libB, 0o755); err != nil {
		t.Fatal(err)
	}
	a, err := store.CreateLibraryFull(ctx, models.LibraryCreate{Name: "A", MountPath: libA, Backend: models.LibraryBackendLocal})
	if err != nil {
		t.Fatal(err)
	}
	b, err := store.CreateLibraryFull(ctx, models.LibraryCreate{Name: "B", MountPath: libB, Backend: models.LibraryBackendLocal})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetUserLibraries(ctx, uid, []int64{a.ID}); err != nil {
		t.Fatal(err)
	}
	bookID, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: b.ID, Title: "Secret", Format: models.FormatEPUB, RelPath: "secret.epub",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceChapters(ctx, bookID, []models.Chapter{
		{Title: "Hidden Chapter", Index: 0},
	}); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, _, _ := loginUser(t, handler, "reader", "longpassword")
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/chapters", bookID), nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatalf("chapters IDOR allowed body=%s", rec.Body.String())
	}
	fmt.Println("PROVED_LIBRARY_RESTRICTED_CHAPTERS_IDOR: denied status=", rec.Code)
}

// PROVED_LIBRARY_RESTRICTED_WRITES_IDOR
// Guarantee: users restricted to library A cannot read or write progress,
// favorites, or collection membership for books in library B.

func TestRestrictedLibraryWritesIDOROracle(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	adminHash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", adminHash, true); err != nil {
		t.Fatal(err)
	}
	userHash, _ := auth.HashPassword("longpassword")
	uid, err := store.CreateUser(ctx, "reader", userHash, false)
	if err != nil {
		t.Fatal(err)
	}

	libA := filepath.Join(srv.cfg.DataDir, "libA")
	libB := filepath.Join(srv.cfg.DataDir, "libB")
	for _, dir := range []string{libA, libB} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	a, err := store.CreateLibraryFull(ctx, models.LibraryCreate{Name: "A", MountPath: libA, Backend: models.LibraryBackendLocal})
	if err != nil {
		t.Fatal(err)
	}
	b, err := store.CreateLibraryFull(ctx, models.LibraryCreate{Name: "B", MountPath: libB, Backend: models.LibraryBackendLocal})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetUserLibraries(ctx, uid, []int64{a.ID}); err != nil {
		t.Fatal(err)
	}
	bookID, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: b.ID, Title: "Secret", Format: models.FormatEPUB, RelPath: "secret.epub",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	col, err := store.CreateCollection(ctx, uid, "mine", "")
	if err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, _, _ := loginUser(t, handler, "reader", "longpassword")

	cases := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, fmt.Sprintf("/api/books/%d/progress", bookID), ""},
		{http.MethodPut, fmt.Sprintf("/api/books/%d/progress", bookID), `{"percent":0.5}`},
		{http.MethodGet, fmt.Sprintf("/api/books/%d/favorite", bookID), ""},
		{http.MethodPut, fmt.Sprintf("/api/books/%d/favorite", bookID), `{"favorite":true}`},
		{http.MethodPost, fmt.Sprintf("/api/collections/%d/books/%d", col.ID, bookID), ""},
		{http.MethodDelete, fmt.Sprintf("/api/collections/%d/books/%d", col.ID, bookID), ""},
	}
	for _, tc := range cases {
		var req *http.Request
		if tc.body == "" {
			req = httptest.NewRequest(tc.method, tc.path, nil)
		} else {
			req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		}
		req.AddCookie(session)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code == http.StatusOK || rec.Code == http.StatusNoContent {
			t.Fatalf("%s %s: IDOR allowed status=%d body=%s", tc.method, tc.path, rec.Code, rec.Body.String())
		}
	}
	fmt.Println("PROVED_LIBRARY_RESTRICTED_WRITES_IDOR: all denied")
}
