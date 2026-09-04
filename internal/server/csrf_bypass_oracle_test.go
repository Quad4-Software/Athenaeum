package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"athenaeum/internal/auth"
)

// PROVED_CSRF_BASIC_HEADER_BYPASS
// Guarantee: a session-authenticated browser request must still require CSRF
// even if the attacker adds a fake Authorization: Basic header.
// Before the fix, usesExternalAuth returned true on header shape alone and
// skipped CSRF while withAuth still trusted the session cookie.

func TestCSRFBypassViaFakeBasicHeaderOracle(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, err := auth.HashPassword("longpassword")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateUser(ctx, "victim", hash, true); err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, _, _ := loginUser(t, handler, "victim", "longpassword")

	body := bytes.NewBufferString(`{"allowRegistration":true}`)
	req := httptest.NewRequest(http.MethodPut, "/api/auth/settings", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic YTpi") // fake Basic, not a real user
	req.AddCookie(session)
	// Deliberately omit X-CSRF-Token.
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatalf("CSRF bypassed via fake Basic header status=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 csrf failure got %d body=%s", rec.Code, rec.Body.String())
	}
	fmt.Println("PROVED_CSRF_BASIC_HEADER_BYPASS: fake Basic no longer skips CSRF with session cookie")
}

func TestCSRFBypassViaFakeAPIKeyHeaderOracle(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, err := auth.HashPassword("longpassword")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateUser(ctx, "victim", hash, true); err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, _, _ := loginUser(t, handler, "victim", "longpassword")

	body := bytes.NewBufferString(`{"allowRegistration":true}`)
	req := httptest.NewRequest(http.MethodPut, "/api/auth/settings", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", auth.APIKeyPrefix+"not-a-real-key")
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatalf("CSRF bypassed via fake API key header status=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 csrf failure got %d body=%s", rec.Code, rec.Body.String())
	}
	fmt.Println("PROVED_CSRF_API_KEY_HEADER_BYPASS: fake API key no longer skips CSRF with session cookie")
}

func TestAPIKeyWithoutSessionSkipsCSRF(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, err := auth.HashPassword("longpassword")
	if err != nil {
		t.Fatal(err)
	}
	uid, err := store.CreateUser(ctx, "apier", hash, true)
	if err != nil {
		t.Fatal(err)
	}
	created, err := store.CreateAPIKey(ctx, uid, "cli")
	if err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	body := bytes.NewBufferString(`{"allowRegistration":true}`)
	req := httptest.NewRequest(http.MethodPut, "/api/auth/settings", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", created.Key)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("API key without session should skip CSRF status=%d body=%s", rec.Code, rec.Body.String())
	}
}
