package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLooksLikeStaticAsset(t *testing.T) {
	cases := map[string]bool{
		"assets/index-abc.js": true,
		"favicon.svg":         true,
		"sw.js":               true,
		"book/12":             false,
		"error/offline":       false,
		"settings/library":    false,
		"":                    false,
	}
	for name, want := range cases {
		if got := looksLikeStaticAsset(name); got != want {
			t.Fatalf("%q: got %v want %v", name, got, want)
		}
	}
}

func TestMissingAssetReturnsHTML404(t *testing.T) {
	handler, err := spaHandler("")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/assets/missing-chunk-xyz.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("Content-Type=%q want text/html", ct)
	}
	if !strings.Contains(rec.Body.String(), "Not found") {
		t.Fatalf("body=%q", rec.Body.String())
	}
}

func TestSPAUnknownRouteStillServesIndex(t *testing.T) {
	handler, err := spaHandler("")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/book/99", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d want 200", rec.Code)
	}
}

func TestRecoverMiddlewareHTML(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := recoverMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	req := httptest.NewRequest(http.MethodGet, "/library", nil)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("Content-Type=%q", rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(rec.Body.String(), "Something went wrong") {
		t.Fatalf("body=%q", rec.Body.String())
	}
}

func TestRecoverMiddlewareJSON(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := recoverMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("Content-Type=%q", rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(rec.Body.String(), "internal server error") {
		t.Fatalf("body=%q", rec.Body.String())
	}
}
