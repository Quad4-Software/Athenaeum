package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

func TestListAndRevokeSessions(t *testing.T) {
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
	session, _, csrf := loginUser(t, handler, "bob", "longpassword")

	req := httptest.NewRequest(http.MethodGet, "/api/auth/sessions", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list sessions status=%d body=%s", rec.Code, rec.Body.String())
	}
	var sessions []models.UserSession
	if err := json.NewDecoder(rec.Body).Decode(&sessions); err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || !sessions[0].Current {
		t.Fatalf("sessions=%+v", sessions)
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/auth/sessions/"+sessions[0].ID, nil)
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("revoke session status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(session)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked session should be invalid, status=%d", rec.Code)
	}
}

func TestRevokeOtherAndAllSessions(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	uid, err := store.CreateUser(ctx, "bob", hash, true)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, _, csrf := loginUser(t, handler, "bob", "longpassword")

	access2, err := auth.NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	refresh2, err := auth.NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	sessID2, err := auth.NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if err := store.CreateUserSession(ctx, models.SessionCreate{
		SessionID:      sessID2,
		UserID:         uid,
		AccessToken:    access2,
		RefreshToken:   refresh2,
		AccessExpires:  now.Add(15 * time.Minute),
		RefreshExpires: now.Add(30 * 24 * time.Hour),
		IP:             "127.0.0.1",
		UserAgent:      "other-device",
		Device:         "Other",
		AuthMethod:     "password",
	}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/auth/sessions", nil)
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("revoke other status=%d body=%s", rec.Code, rec.Body.String())
	}
	var otherRes map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&otherRes); err != nil {
		t.Fatal(err)
	}
	if otherRes["revoked"] != float64(1) {
		t.Fatalf("revoked others=%v want 1", otherRes["revoked"])
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(session)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("current session should remain, status=%d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/auth/sessions?all=true", nil)
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("revoke all status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(session)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("all sessions revoked should invalidate current, status=%d", rec.Code)
	}
}

func TestAuthMethodsWithoutOIDC(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/auth/methods", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("methods status=%d", rec.Code)
	}
	var methods models.AuthMethods
	if err := json.NewDecoder(rec.Body).Decode(&methods); err != nil {
		t.Fatal(err)
	}
	if !methods.AuthEnabled || !methods.LoginLocal || methods.LoginOIDC {
		t.Fatalf("methods=%+v", methods)
	}
}

func TestOIDCConfigAdminOnly(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	adminHash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", adminHash, true); err != nil {
		t.Fatal(err)
	}
	userHash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "reader", userHash, false); err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, _, _ := loginUser(t, handler, "reader", "longpassword")

	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/config", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-admin oidc config status=%d", rec.Code)
	}
}

func TestSaveOIDCConfig(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, _, csrf := loginUser(t, handler, "admin", "longpassword")

	body := bytes.NewBufferString(`{
		"enabled": true,
		"loginLocal": true,
		"issuerUrl": "https://auth.example.com",
		"authorizeUrl": "https://auth.example.com/authorize",
		"tokenUrl": "https://auth.example.com/token",
		"clientId": "reader",
		"clientSecret": "secret-value",
		"buttonText": "Sign in with Example",
		"matchBy": "email",
		"autoRegister": true
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/auth/oidc/config", body)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("save oidc config status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/methods", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	var methods models.AuthMethods
	if err := json.NewDecoder(rec.Body).Decode(&methods); err != nil {
		t.Fatal(err)
	}
	if !methods.LoginOIDC || methods.OIDCButtonText != "Sign in with Example" {
		t.Fatalf("methods=%+v", methods)
	}
}
