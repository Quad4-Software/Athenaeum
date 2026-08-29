package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"athenaeum/internal/models"
)

func TestTagsRoutes(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	id, err := store.UpsertBook(ctx, &models.Book{Title: "Dune", Author: "Herbert", Format: models.FormatEPUB, RelPath: "dune.epub"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	csrf := fetchCSRF(t, handler)

	putBody, _ := json.Marshal(map[string][]string{"tags": {"sci-fi", "favorite"}})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d/tags", id), bytes.NewReader(putBody))
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("put tags status=%d body=%s", rec.Code, rec.Body.String())
	}
	var tags []string
	if err := json.NewDecoder(rec.Body).Decode(&tags); err != nil {
		t.Fatal(err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %+v", tags)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list tags status=%d body=%s", rec.Code, rec.Body.String())
	}
	var allTags []models.Tag
	if err := json.NewDecoder(rec.Body).Decode(&allTags); err != nil {
		t.Fatal(err)
	}
	if len(allTags) != 2 {
		t.Fatalf("expected 2 global tags, got %+v", allTags)
	}

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d", id), nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get book status=%d body=%s", rec.Code, rec.Body.String())
	}
	var book models.Book
	if err := json.NewDecoder(rec.Body).Decode(&book); err != nil {
		t.Fatal(err)
	}
	if len(book.Tags) != 2 {
		t.Fatalf("expected book detail to include 2 tags, got %+v", book.Tags)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/books?tag=favorite", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("filter by tag status=%d body=%s", rec.Code, rec.Body.String())
	}
	var page models.BookPage
	if err := json.NewDecoder(rec.Body).Decode(&page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 {
		t.Fatalf("expected 1 book for tag filter, got %+v", page)
	}

	var tagID int64
	for _, tag := range allTags {
		if tag.Name == "favorite" {
			tagID = tag.ID
		}
	}
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/books/%d/tags/%d", id, tagID), nil)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete tag status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRatingRoutes(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	id, err := store.UpsertBook(ctx, &models.Book{Title: "Dune", Format: models.FormatEPUB, RelPath: "dune.epub"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	csrf := fetchCSRF(t, handler)

	body, _ := json.Marshal(map[string]int{"rating": 4})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d/rating", id), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("put rating status=%d body=%s", rec.Code, rec.Body.String())
	}
	var rating models.BookRating
	if err := json.NewDecoder(rec.Body).Decode(&rating); err != nil {
		t.Fatal(err)
	}
	if rating.Rating != 4 {
		t.Fatalf("expected rating 4, got %+v", rating)
	}

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/rating", id), nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get rating status=%d body=%s", rec.Code, rec.Body.String())
	}

	invalid, _ := json.Marshal(map[string]int{"rating": 9})
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d/rating", id), bytes.NewReader(invalid))
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for invalid rating, got %d", rec.Code)
	}
}

func TestReaderPrefsRoutes(t *testing.T) {
	srv, _ := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	csrf := fetchCSRF(t, handler)

	body, _ := json.Marshal(map[string]any{"prefs": map[string]any{"theme": "night", "fontPct": 110}})
	req := httptest.NewRequest(http.MethodPut, "/api/auth/reader-prefs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("put reader prefs status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/reader-prefs", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get reader prefs status=%d body=%s", rec.Code, rec.Body.String())
	}
	var prefs models.ReaderPrefs
	if err := json.NewDecoder(rec.Body).Decode(&prefs); err != nil {
		t.Fatal(err)
	}
	if prefs.Prefs["theme"] != "night" {
		t.Fatalf("unexpected reader prefs: %+v", prefs)
	}
}
