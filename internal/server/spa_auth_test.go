package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"athenaeum/internal/auth"
)

func TestSPAAccessibleWithoutAuth(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("secretpass")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"/", "/login", "/assets/", "/favicon.ico", "/manifest.webmanifest"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Accept", "text/html")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code == http.StatusUnauthorized {
			t.Fatalf("%s status=401 body=%s", path, rec.Body.String())
		}
	}
}

func TestSPAIndexIsHTML(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("secretpass")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "text/html")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<html") && !strings.Contains(body, "<!DOCTYPE") {
		t.Fatalf("expected HTML, got %q", body[:min(80, len(body))])
	}
}
