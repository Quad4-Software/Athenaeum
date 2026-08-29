package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"athenaeum/internal/library"
	"athenaeum/internal/models"
)

func TestBookEditMetadataCoversAndDelete(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)
	ctx := context.Background()

	libDir := filepath.Join(srv.cfg.DataDir, "edit-lib")
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(srv.cfg.CoverDir(), 0o750); err != nil {
		t.Fatal(err)
	}

	epubBytes := writeMinimalEPUBBytes(t)
	epubRel := "edit-book.epub"
	if err := os.WriteFile(filepath.Join(libDir, epubRel), epubBytes, 0o640); err != nil {
		t.Fatal(err)
	}

	createLib := map[string]any{"name": "Edit Lib", "mountPath": libDir, "backend": "local"}
	body, _ := json.Marshal(createLib)
	req := httptest.NewRequest(http.MethodPost, "/api/libraries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create lib status=%d body=%s", rec.Code, rec.Body.String())
	}
	var lib models.Library
	if err := json.NewDecoder(rec.Body).Decode(&lib); err != nil {
		t.Fatal(err)
	}

	bookID, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Edit Me",
		Author:    "Author",
		Format:    models.FormatEPUB,
		RelPath:   epubRel,
		FileSize:  int64(len(epubBytes)),
	}, 1)
	if err != nil {
		t.Fatal(err)
	}

	doJSON := func(method, path string, payload any) *httptest.ResponseRecorder {
		t.Helper()
		var rdr io.Reader
		if payload != nil {
			b, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			rdr = bytes.NewReader(b)
		}
		req := httptest.NewRequest(method, path, rdr)
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.AddCookie(session)
		if method != http.MethodGet && method != http.MethodHead {
			withCSRF(req, csrf)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	rec = doJSON(http.MethodGet, "/api/metadata/providers", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("providers status=%d", rec.Code)
	}

	rec = doJSON(http.MethodPost, fmt.Sprintf("/api/books/%d/metadata/search", bookID), models.MetadataSearchQuery{
		Title: "Edit Me", Author: "Author",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("metadata search status=%d body=%s", rec.Code, rec.Body.String())
	}
	var search models.MetadataSearchResult
	if err := json.NewDecoder(rec.Body).Decode(&search); err != nil {
		t.Fatal(err)
	}

	tinyPNG := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x00, 0x03, 0x00, 0x01, 0x00, 0x05, 0xfe,
		0xd4, 0xef, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45,
		0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
	restoreCover := library.SwapCoverHTTPClient(&http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"image/png"}},
				Body:       io.NopCloser(bytes.NewReader(tinyPNG)),
				Request:    r,
			}, nil
		}),
	})
	t.Cleanup(restoreCover)

	rec = doJSON(http.MethodPost, fmt.Sprintf("/api/books/%d/metadata/apply", bookID), models.MetadataApplyRequest{
		Match: models.MetadataMatch{
			Title:    "Applied Title",
			Author:   "Applied Author",
			Series:   "Applied Series",
			Language: "en",
			CoverURL: "https://example.com/cover.png",
		},
		ApplyCover: true,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("metadata apply status=%d body=%s", rec.Code, rec.Body.String())
	}
	var applied models.Book
	if err := json.NewDecoder(rec.Body).Decode(&applied); err != nil {
		t.Fatal(err)
	}
	if applied.Title != "Applied Title" {
		t.Fatalf("title=%q", applied.Title)
	}

	rec = doJSON(http.MethodPost, fmt.Sprintf("/api/books/%d/metadata/apply", bookID), models.MetadataApplyRequest{
		Match: models.MetadataMatch{Title: ""},
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty title apply status=%d", rec.Code)
	}

	rec = doJSON(http.MethodPut, fmt.Sprintf("/api/books/%d/cover-from-url", bookID), map[string]string{
		"url": "https://example.com/other.png",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("cover-from-url status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doJSON(http.MethodPut, fmt.Sprintf("/api/books/%d/cover-from-url", bookID), map[string]string{
		"url": "http://127.0.0.1/blocked.png",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("blocked cover url status=%d", rec.Code)
	}

	var coverBuf bytes.Buffer
	mw := multipart.NewWriter(&coverBuf)
	part, err := mw.CreateFormFile("cover", "fresh.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("multipart-cover-bytes")); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	coverReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d/cover", bookID), &coverBuf)
	coverReq.Header.Set("Content-Type", mw.FormDataContentType())
	coverReq.AddCookie(session)
	withCSRF(coverReq, csrf)
	coverRec := httptest.NewRecorder()
	handler.ServeHTTP(coverRec, coverReq)
	if coverRec.Code != http.StatusOK {
		t.Fatalf("put cover status=%d body=%s", coverRec.Code, coverRec.Body.String())
	}

	rawCover := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d/cover", bookID), bytes.NewReader([]byte("raw-cover")))
	rawCover.Header.Set("Content-Type", "application/octet-stream")
	rawCover.AddCookie(session)
	withCSRF(rawCover, csrf)
	rawRec := httptest.NewRecorder()
	handler.ServeHTTP(rawRec, rawCover)
	if rawRec.Code != http.StatusOK {
		t.Fatalf("raw put cover status=%d", rawRec.Code)
	}

	rec = doJSON(http.MethodDelete, fmt.Sprintf("/api/books/%d/cover", bookID), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete cover status=%d body=%s", rec.Code, rec.Body.String())
	}

	path, cleanup, err := srv.materializeBookFile(ctx, lib.ID, epubRel)
	if err != nil {
		t.Fatal(err)
	}
	cleanup()
	if path == "" {
		t.Fatal("empty materialize path")
	}
	chunks, err := extractEPUBText(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected epub text chunks")
	}

	rec = doJSON(http.MethodPost, "/api/admin/content-index", nil)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("content-index status=%d", rec.Code)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		ids, err := store.SearchBookContentIDs(ctx, "upload")
		if err == nil {
			for _, id := range ids {
				if id == bookID {
					goto indexed
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
indexed:

	rec = doJSON(http.MethodDelete, fmt.Sprintf("/api/books/%d", bookID), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete book status=%d body=%s", rec.Code, rec.Body.String())
	}
	if _, err := store.GetBook(ctx, bookID); err == nil {
		t.Fatal("book still present after delete")
	}
	if _, err := os.Stat(filepath.Join(libDir, epubRel)); !os.IsNotExist(err) {
		t.Fatalf("epub file should be removed err=%v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
