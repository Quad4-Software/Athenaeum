package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"athenaeum/internal/auth"
)

func TestAPIUnauthorizedHTMLRedirect(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d want redirect, body=%s", rec.Code, rec.Body.String())
	}
	loc := rec.Header().Get("Location")
	if loc == "" || loc[:6] != "/login" {
		t.Fatalf("location=%q", loc)
	}
}

func TestAPIUnauthorizedJSON(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAPIForbiddenAndNotFoundHTMLRedirect(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	adminHash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", adminHash, true); err != nil {
		t.Fatal(err)
	}
	userHash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "reader", userHash, false); err != nil {
		t.Fatal(err)
	}
	session, _, _ := loginUser(t, handler, "reader", "longpassword")

	req := httptest.NewRequest(http.MethodGet, "/api/admin/server", nil)
	req.AddCookie(session)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("forbidden status=%d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/error/forbidden" {
		t.Fatalf("forbidden location=%q", loc)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/missing-resource", nil)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	rec = httptest.NewRecorder()
	if !redirectAPIError(rec, req, http.StatusNotFound) {
		t.Fatal("expected not-found html redirect")
	}
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/error/not-found" {
		t.Fatalf("not found redirect code=%d loc=%q", rec.Code, rec.Header().Get("Location"))
	}
	if redirectAPIError(httptest.NewRecorder(), req, http.StatusBadRequest) {
		t.Fatal("non-authz statuses should not redirect")
	}
}

func TestPrefersHTMLMatrix(t *testing.T) {
	cases := []struct {
		name string
		hdr  map[string]string
		want bool
	}{
		{"navigate", map[string]string{"Sec-Fetch-Mode": "navigate"}, true},
		{"xhr", map[string]string{"X-Requested-With": "XMLHttpRequest", "Accept": "text/html"}, false},
		{"json only", map[string]string{"Accept": "application/json"}, false},
		{"html", map[string]string{"Accept": "text/html"}, true},
		{"empty", map[string]string{}, false},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
		for k, v := range tc.hdr {
			req.Header.Set(k, v)
		}
		if got := prefersHTML(req); got != tc.want {
			t.Fatalf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}
