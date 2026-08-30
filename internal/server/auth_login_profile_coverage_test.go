package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

func TestAuthLoginProfileCoverage(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)

	do := func(method, path string, body any, sess, cs *http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		var rdr *bytes.Reader
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			rdr = bytes.NewReader(b)
		} else {
			rdr = bytes.NewReader(nil)
		}
		req := httptest.NewRequest(method, path, rdr)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if sess != nil {
			req.AddCookie(sess)
		}
		if method != http.MethodGet && method != http.MethodHead && cs != nil {
			withCSRF(req, cs)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	regCSRF := fetchCSRF(t, handler)
	rec := do(http.MethodPost, "/api/auth/register-public", map[string]any{
		"username": "pubuser", "password": "longpassword",
	}, nil, regCSRF)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("register disabled status=%d", rec.Code)
	}

	rec = do(http.MethodPut, "/api/auth/settings", models.AuthSettings{AllowRegistration: true}, session, csrf)
	if rec.Code != http.StatusOK {
		t.Fatalf("enable registration status=%d body=%s", rec.Code, rec.Body.String())
	}

	regCSRF = fetchCSRF(t, handler)
	rec = do(http.MethodPost, "/api/auth/register-public", map[string]any{
		"username": "x", "password": "longpassword",
	}, nil, regCSRF)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("register short name status=%d", rec.Code)
	}
	regCSRF = fetchCSRF(t, handler)
	rec = do(http.MethodPost, "/api/auth/register-public", map[string]any{
		"username": "pubuser", "password": "short",
	}, nil, regCSRF)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("register weak password status=%d", rec.Code)
	}

	regCSRF = fetchCSRF(t, handler)
	rec = do(http.MethodPost, "/api/auth/register-public", map[string]any{
		"username": "pubuser", "password": "longpassword",
	}, nil, regCSRF)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register public status=%d body=%s", rec.Code, rec.Body.String())
	}
	var pubUser models.User
	if err := json.NewDecoder(rec.Body).Decode(&pubUser); err != nil {
		t.Fatal(err)
	}
	var pubSession, pubCSRF *http.Cookie
	for _, c := range rec.Result().Cookies() {
		switch c.Name {
		case auth.SessionCookie:
			pubSession = c
		case auth.CSRFCookie:
			pubCSRF = c
		}
	}
	if pubSession == nil || pubCSRF == nil {
		t.Fatal("missing cookies after public register")
	}

	regCSRF = fetchCSRF(t, handler)
	rec = do(http.MethodPost, "/api/auth/register-public", map[string]any{
		"username": "pubuser", "password": "longpassword",
	}, nil, regCSRF)
	if rec.Code != http.StatusConflict {
		t.Fatalf("register taken status=%d", rec.Code)
	}

	rec = do(http.MethodPut, "/api/auth/profile", map[string]any{"username": "x"}, pubSession, pubCSRF)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("profile short status=%d", rec.Code)
	}
	rec = do(http.MethodPut, "/api/auth/profile", map[string]any{"username": "pubuser"}, pubSession, pubCSRF)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile same name status=%d", rec.Code)
	}
	rec = do(http.MethodPut, "/api/auth/profile", map[string]any{"username": "admin"}, pubSession, pubCSRF)
	if rec.Code != http.StatusConflict {
		t.Fatalf("profile taken status=%d", rec.Code)
	}
	rec = do(http.MethodPut, "/api/auth/profile", map[string]any{"username": "pubrenamed"}, pubSession, pubCSRF)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile rename status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = do(http.MethodPut, "/api/auth/password", map[string]any{
		"currentPassword": "wrong", "newPassword": "longerpassword",
	}, pubSession, pubCSRF)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("change password wrong current status=%d", rec.Code)
	}
	rec = do(http.MethodPut, "/api/auth/password", map[string]any{
		"currentPassword": "longpassword", "newPassword": "short",
	}, pubSession, pubCSRF)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("change password weak status=%d", rec.Code)
	}
	rec = do(http.MethodPut, "/api/auth/password", map[string]any{
		"currentPassword": "longpassword", "newPassword": "longerpassword",
	}, pubSession, pubCSRF)
	if rec.Code != http.StatusOK {
		t.Fatalf("change password status=%d body=%s", rec.Code, rec.Body.String())
	}

	loginCSRF := fetchCSRF(t, handler)
	rec = do(http.MethodPost, "/api/auth/login", map[string]any{
		"username": "pubrenamed", "password": "longerpassword",
	}, nil, loginCSRF)
	if rec.Code != http.StatusOK {
		t.Fatalf("login after password change status=%d body=%s", rec.Code, rec.Body.String())
	}

	loginCSRF = fetchCSRF(t, handler)
	rec = do(http.MethodPost, "/api/auth/login", map[string]any{
		"username": "", "password": "",
	}, nil, loginCSRF)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("login empty status=%d", rec.Code)
	}
	loginCSRF = fetchCSRF(t, handler)
	rec = do(http.MethodPost, "/api/auth/login", map[string]any{
		"username": "pubrenamed", "password": "nope-wrong-password",
	}, nil, loginCSRF)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("login bad password status=%d", rec.Code)
	}

	_ = store
	_ = pubUser
}
