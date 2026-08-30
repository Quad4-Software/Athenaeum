package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"athenaeum/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

func loginUser(t *testing.T, handler http.Handler, username, password string) (*http.Cookie, *http.Cookie, *http.Cookie) {
	t.Helper()
	csrf := fetchCSRF(t, handler)
	body := bytes.NewBufferString(`{"username":"` + username + `","password":"` + password + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", rec.Code, rec.Body.String())
	}
	var session, refresh, csrfCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		switch c.Name {
		case auth.SessionCookie:
			session = c
		case auth.RefreshCookie:
			refresh = c
		case auth.CSRFCookie:
			csrfCookie = c
		}
	}
	if session == nil || refresh == nil || csrfCookie == nil {
		t.Fatalf("missing cookies after login: %v", rec.Result().Cookies())
	}
	return session, refresh, csrfCookie
}

func TestRegisterBlockedBeforeSetup(t *testing.T) {
	srv, _ := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	csrf := fetchCSRF(t, handler)
	body := bytes.NewBufferString(`{"username":"intruder","password":"longpassword"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("register before setup status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSetupBlockedAfterBootstrap(t *testing.T) {
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
	csrf := fetchCSRF(t, handler)
	body := bytes.NewBufferString(`{"username":"other","password":"longpassword1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/setup", body)
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("setup after bootstrap status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCSRFRequiredOnMutatingAPI(t *testing.T) {
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
	session, _, _ := loginUser(t, handler, "bob", "longpassword")

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("logout without csrf status=%d", rec.Code)
	}
}

func TestNonAdminCannotListUsers(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/auth/users", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-admin list users status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPasswordChangeRevokesOldSession(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("oldpassword1")
	if _, err := store.CreateUser(ctx, "bob", hash, true); err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	oldSession, _, csrf := loginUser(t, handler, "bob", "oldpassword1")

	body := bytes.NewBufferString(`{"currentPassword":"oldpassword1","newPassword":"newpassword1"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/auth/password", body)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(oldSession)
	withCSRF(req, csrf)
	changeRec := httptest.NewRecorder()
	handler.ServeHTTP(changeRec, req)
	if changeRec.Code != http.StatusOK {
		t.Fatalf("change password status=%d body=%s", changeRec.Code, changeRec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(oldSession)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("old session should be invalid, status=%d", rec.Code)
	}

	var newSession *http.Cookie
	for _, c := range changeRec.Result().Cookies() {
		if c.Name == auth.SessionCookie && c.Value != "" {
			newSession = c
		}
	}
	if newSession == nil {
		t.Fatal("expected new session cookie after password change")
	}
	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(newSession)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("new session after password change status=%d", rec.Code)
	}
}

func TestRefreshFailureKeepsValidSession(t *testing.T) {
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
	_, refresh, csrf := loginUser(t, handler, "bob", "longpassword")

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(refresh)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh status=%d body=%s", rec.Code, rec.Body.String())
	}

	var newSession *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == auth.SessionCookie && c.Value != "" {
			newSession = c
		}
	}
	if newSession == nil {
		t.Fatal("expected new session cookie after refresh")
	}

	req = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(refresh)
	req.AddCookie(newSession)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("reused refresh token status=%d, want 401", rec.Code)
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == auth.SessionCookie && c.MaxAge < 0 {
			t.Fatal("valid session cookie should not be cleared when refresh token reuse fails")
		}
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(newSession)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("session should remain valid after refresh reuse failure, status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRefreshTokenRotation(t *testing.T) {
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
	_, refresh, csrf := loginUser(t, handler, "bob", "longpassword")

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(refresh)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(refresh)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("reused refresh token status=%d, want 401", rec.Code)
	}
}

func TestFailedRefreshPreservesCSRFCookie(t *testing.T) {
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
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("refresh without token status=%d, want 401", rec.Code)
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == auth.CSRFCookie && c.MaxAge < 0 {
			t.Fatal("failed refresh must not clear CSRF cookie needed for login")
		}
	}

	body := bytes.NewBufferString(`{"username":"bob","password":"wrong-password"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	loginReq.Header.Set("Content-Type", "application/json")
	withCSRF(loginReq, csrf)
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code == http.StatusForbidden {
		t.Fatalf("login CSRF rejected after failed refresh: %s", loginRec.Body.String())
	}
	if loginRec.Code != http.StatusUnauthorized {
		t.Fatalf("login status=%d body=%s, want 401 bad credentials", loginRec.Code, loginRec.Body.String())
	}
}

func TestLogoutWithExpiredSessionCookie(t *testing.T) {
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
	_, refresh, csrf := loginUser(t, handler, "bob", "longpassword")

	expiredSession := &http.Cookie{Name: auth.SessionCookie, Value: "expired-token"}
	_ = store.CreateSession(ctx, "expired-token", uid, time.Now().Add(-time.Hour))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(expiredSession)
	req.AddCookie(refresh)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout with expired session status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(refresh)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("refresh after logout status=%d, want 401", rec.Code)
	}
}

func TestOPDSRequiresAuthWhenEnabled(t *testing.T) {
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
	req := httptest.NewRequest(http.MethodGet, "/opds/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("opds without auth status=%d, want 401", rec.Code)
	}
}

func TestInvalidLoginDoesNotLeakUserExistence(t *testing.T) {
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

	for _, body := range []string{
		`{"username":"nobody","password":"longpassword"}`,
		`{"username":"bob","password":"wrongpassword"}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		withCSRF(req, csrf)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("login body=%s status=%d", body, rec.Code)
		}
		var errBody map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&errBody); err != nil {
			t.Fatal(err)
		}
		if errBody["error"] != "invalid username or password" {
			t.Fatalf("unexpected error message: %q", errBody["error"])
		}
	}
}

func TestLoginUpgradesLegacyBcryptHash(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	legacy, err := bcrypt.GenerateFromPassword([]byte("longpassword"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	u, err := store.CreateUser(ctx, "legacy", string(legacy), true)
	if err != nil {
		t.Fatal(err)
	}
	if u == 0 {
		t.Fatal("expected created user id")
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	loginUser(t, handler, "legacy", "longpassword")

	_, hash, err := store.GetUserByUsername(ctx, "legacy")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("expected argon2id upgrade after login, got %q", hash)
	}
	if !auth.CheckPassword(hash, "longpassword") {
		t.Fatal("upgraded hash should still verify")
	}
}
