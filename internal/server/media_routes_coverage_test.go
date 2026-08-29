package server

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/library"
	"athenaeum/internal/models"
)

func writeTestCBZ(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("001.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte{0xff, 0xd8, 0xff, 0xd9}); err != nil {
		t.Fatal(err)
	}
	w2, err := zw.Create("002.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w2.Write([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMediaComicPages(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	libDir := filepath.Join(srv.cfg.DataDir, "comics")
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}
	cbzPath := filepath.Join(libDir, "issue.cbz")
	writeTestCBZ(t, cbzPath)

	lib, err := store.CreateLibrary(ctx, "Comics", libDir)
	if err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(cbzPath)
	comicID, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Issue One",
		Author:    "Artist",
		Format:    models.FormatCBZ,
		RelPath:   "issue.cbz",
		FileSize:  info.Size(),
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	pdfID, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Not Comic",
		Format:    models.FormatPDF,
		RelPath:   "missing.pdf",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/pages", comicID), nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("manifest status=%d body=%s", rec.Code, rec.Body.String())
	}
	var manifest models.ComicManifest
	if err := json.NewDecoder(rec.Body).Decode(&manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Total < 1 || len(manifest.Pages) < 1 {
		t.Fatalf("manifest=%+v", manifest)
	}

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/pages/0", comicID), nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("page status=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") == "" || rec.Body.Len() == 0 {
		t.Fatalf("content-type=%q len=%d", rec.Header().Get("Content-Type"), rec.Body.Len())
	}

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/pages/bad", comicID), nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad page status=%d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/pages", pdfID), nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("non-comic status=%d", rec.Code)
	}
}

func TestMediaAudiobookTracks(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	libDir := filepath.Join(srv.cfg.DataDir, "audio")
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}
	lib, err := store.CreateLibrary(ctx, "Audio", libDir)
	if err != nil {
		t.Fatal(err)
	}
	setID, err := store.UpsertAudiobookSet(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Set",
		Author:    "Narrator",
		Format:    models.FormatAudiobook,
		RelPath:   "set/.athenaeum-set",
		AbsPath:   filepath.Join(libDir, "set"),
	}, []models.AudiobookTrack{
		{Index: 0, RelPath: "set/01.mp3", Title: "One", Format: models.FormatMP3, FileSize: 10},
		{Index: 1, RelPath: "set/02.mp3", Title: "Two", Format: models.FormatMP3, FileSize: 20},
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/tracks", setID), nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("tracks status=%d body=%s", rec.Code, rec.Body.String())
	}
	var tracks []models.AudiobookTrack
	if err := json.NewDecoder(rec.Body).Decode(&tracks); err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 2 || tracks[0].Title != "One" {
		t.Fatalf("tracks=%+v", tracks)
	}

	pdfID, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID, Title: "PDF", Format: models.FormatPDF, RelPath: "x.pdf",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/tracks", pdfID), nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("non-audiobook status=%d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/books/999999/tracks", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing book status=%d", rec.Code)
	}
}

func TestMediaMobiSections(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	libDir := filepath.Join(srv.cfg.DataDir, "mobi")
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libDir, "book.mobi"), []byte("not-a-real-mobi"), 0o640); err != nil {
		t.Fatal(err)
	}
	lib, err := store.CreateLibrary(ctx, "Mobi", libDir)
	if err != nil {
		t.Fatal(err)
	}
	mobiID, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Mobi Book",
		Format:    models.FormatMOBI,
		RelPath:   "book.mobi",
		FileSize:  16,
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	pdfID, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "PDF",
		Format:    models.FormatPDF,
		RelPath:   "x.pdf",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/mobi-sections", mobiID), nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("mobi sections status=%d body=%s", rec.Code, rec.Body.String())
	}
	var secs []models.MobiSection
	if err := json.NewDecoder(rec.Body).Decode(&secs); err != nil {
		t.Fatal(err)
	}

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/mobi-sections", pdfID), nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("non-mobi status=%d", rec.Code)
	}
}

func TestMediaConvertBookWithoutCalibre(t *testing.T) {
	if library.IsCalibreAvailable() {
		t.Skip("calibre installed")
	}
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)
	ctx := context.Background()

	libDir := filepath.Join(srv.cfg.DataDir, "convert")
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libDir, "book.pdf"), []byte("%PDF-1.4 convert"), 0o640); err != nil {
		t.Fatal(err)
	}
	lib, err := store.CreateLibrary(ctx, "Convert", libDir)
	if err != nil {
		t.Fatal(err)
	}
	bookID, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Convert Me",
		Format:    models.FormatPDF,
		RelPath:   "book.pdf",
		FileSize:  16,
	}, 1)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/books/%d/convert?target=pdf", bookID), nil)
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("convert without calibre status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/books/%d/convert?target=docx", bookID), nil)
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad target status=%d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/books/%d/convert", bookID), nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden {
		t.Fatalf("unauth convert status=%d", rec.Code)
	}
}
