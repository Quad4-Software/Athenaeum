package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

func TestGuestUserAndMetrics(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("adminpass123")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	csrf := fetchCSRF(t, handler)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username":"admin","password":"adminpass123"}`))
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", rec.Code, rec.Body.String())
	}
	sessionCookies := rec.Result().Cookies()

	req = httptest.NewRequest(http.MethodPut, "/api/admin/server", strings.NewReader(`{"metricsEnabled":true,"metricsAuth":true,"metricsUsername":"prom","metricsPassword":"metricspass123","trustedProxies":"127.0.0.1","corsEnabled":false,"cspEnabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	for _, c := range sessionCookies {
		req.AddCookie(c)
	}
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("save server config status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("metrics without auth status=%d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.SetBasicAuth("prom", "metricspass123")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics with auth status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "athenaeum_up 1") {
		t.Fatalf("metrics body missing athenaeum_up: %s", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/auth/users/guest", strings.NewReader(`{"expiresInHours":48}`))
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	for _, c := range sessionCookies {
		req.AddCookie(c)
	}
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create guest status=%d body=%s", rec.Code, rec.Body.String())
	}
	var creds models.GuestCredentials
	if err := json.NewDecoder(rec.Body).Decode(&creds); err != nil {
		t.Fatal(err)
	}
	if creds.Password == "" || creds.User.Username == "" || !creds.User.IsGuest {
		t.Fatalf("guest creds=%+v", creds)
	}

	csrf2 := fetchCSRF(t, handler)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username":"`+creds.User.Username+`","password":"`+creds.Password+`"}`))
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf2)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("guest login status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMetricsDisabled(t *testing.T) {
	srv, _ := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("disabled metrics status=%d", rec.Code)
	}
	_, _ = io.ReadAll(rec.Body)
}
