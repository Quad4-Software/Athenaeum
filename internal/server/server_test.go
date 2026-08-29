package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/config"
	"athenaeum/internal/library"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func testServer(t *testing.T) (*Server, *storage.Store) {
	t.Helper()
	dir := t.TempDir()
	store, err := storage.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	cfg := config.Config{DataDir: dir, LibraryDir: filepath.Join(dir, "lib")}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	scanner := library.New(store, cfg.CoverDir(), cfg.TempDir(), log, 2)
	srv, err := New(context.Background(), cfg, store, scanner, log)
	if err != nil {
		t.Fatal(err)
	}
	return srv, store
}

func TestHealthAndBooks(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	if _, err := store.UpsertBook(ctx, &models.Book{Title: "Test", Format: models.FormatPDF, RelPath: "t.pdf"}, 1); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("health status=%d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/books", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	var page models.BookPage
	if err := json.NewDecoder(rec.Body).Decode(&page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 {
		t.Errorf("total=%d", page.Total)
	}
}

func TestHealthTelemetry(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	cfg := config.Config{
		DataDir:           dir,
		LibraryDir:        filepath.Join(dir, "lib"),
		SentryDSN:         "https://secret@glitchtip.example.com/1",
		SentryDSNPublic:   "https://public@glitchtip.example.com/2",
		SentryEnvironment: "test",
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	scanner := library.New(store, cfg.CoverDir(), cfg.TempDir(), log, 2)
	srv, err := New(context.Background(), cfg, store, scanner, log)
	if err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	var body struct {
		Telemetry struct {
			SentryDSN   string `json:"sentryDsn"`
			Environment string `json:"environment"`
		} `json:"telemetry"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Telemetry.SentryDSN != cfg.SentryDSNPublic {
		t.Fatalf("sentryDsn=%q", body.Telemetry.SentryDSN)
	}
	if body.Telemetry.Environment != "test" {
		t.Fatalf("environment=%q", body.Telemetry.Environment)
	}
}

func TestAuthFlow(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "bob", hash, true); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	csrf := fetchCSRF(t, handler)

	// Protected route without auth
	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	// Login
	body := bytes.NewBufferString(`{"username":"bob","password":"longpassword"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", rec.Code, rec.Body.String())
	}
	var sessionCookie, refreshCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		switch c.Name {
		case auth.SessionCookie:
			sessionCookie = c
		case auth.RefreshCookie:
			refreshCookie = c
		}
	}
	if sessionCookie == nil || refreshCookie == nil {
		t.Fatalf("expected session and refresh cookies, got %v", rec.Result().Cookies())
	}
	var loginCSRF *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == auth.CSRFCookie {
			loginCSRF = c
		}
	}
	if loginCSRF == nil {
		t.Fatal("expected csrf cookie after login")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(sessionCookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("me status=%d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(refreshCookie)
	withCSRF(req, loginCSRF)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestOPDSRoot(t *testing.T) {
	srv, _ := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/opds/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Recent additions") {
		t.Fatalf("body=%s", rec.Body.String())
	}
}

func TestCollectionsAPI(t *testing.T) {
	srv, _ := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	csrf := fetchCSRF(t, handler)

	req := httptest.NewRequest(http.MethodPost, "/api/collections", bytes.NewBufferString(`{"name":"Sci-Fi"}`))
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/collections", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	var cols []models.Collection
	if err := json.NewDecoder(rec.Body).Decode(&cols); err != nil {
		t.Fatal(err)
	}
	if len(cols) != 1 || cols[0].Name != "Sci-Fi" {
		t.Fatalf("cols=%+v", cols)
	}
}

func TestAPIKeyAuth(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	userID, err := store.CreateUser(ctx, "keyuser", hash, false)
	if err != nil {
		t.Fatal(err)
	}
	created, err := store.CreateAPIKey(ctx, userID, "test")
	if err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+created.Key)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("docs status=%d", rec.Code)
	}
}
