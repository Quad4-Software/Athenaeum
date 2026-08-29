package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/auth"
)

func TestI18nLocales(t *testing.T) {
	srv, _ := testServer(t)
	dir := filepath.Join(srv.cfg.DataDir, "i18n")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "de.json"), []byte(`{"$name":"Deutsch","nav":{"allBooks":"Alle"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte(`{"n":1}`), 0o644); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/i18n/locales", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Locales []struct {
			Code   string `json:"code"`
			Source string `json:"source"`
		} `json:"locales"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Locales) < 2 {
		t.Fatalf("locales=%d", len(body.Locales))
	}
}

func TestI18nLocaleAndTemplate(t *testing.T) {
	srv, _ := testServer(t)
	dir := filepath.Join(srv.cfg.DataDir, "i18n")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "de.json"), []byte(`{"nav":{"allBooks":"Alle"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/i18n/en", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("en status=%d", rec.Code)
	}
	var en map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&en); err != nil {
		t.Fatal(err)
	}
	if en["nav.allBooks"] == "" {
		t.Fatal("missing en key")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/i18n/de", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("de status=%d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/i18n/template", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("template status=%d", rec.Code)
	}
	var tmpl map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&tmpl); err != nil {
		t.Fatal(err)
	}
	if tmpl["nav.allBooks"] != "" {
		t.Fatalf("template value should be empty, got %q", tmpl["nav.allBooks"])
	}

	req = httptest.NewRequest(http.MethodGet, "/api/i18n/missing", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing status=%d", rec.Code)
	}
}

func TestI18nPublicWhenAuthRequired(t *testing.T) {
	srv, store := testServer(t)
	ctx := t.Context()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "bob", hash, true); err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/i18n/locales", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
