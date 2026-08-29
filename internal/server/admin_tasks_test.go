package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

func TestAdminVerifyIntegrity(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}

	libDir := filepath.Join(srv.cfg.DataDir, "lib")
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}
	lib, err := store.CreateLibrary(ctx, "Main", libDir)
	if err != nil {
		t.Fatal(err)
	}
	libID := lib.ID
	if _, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: libID,
		Title:     "Missing",
		Format:    models.FormatPDF,
		RelPath:   "missing.pdf",
	}, 1); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	csrf := fetchCSRF(t, handler)

	login := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"username":"admin","password":"longpassword"}`))
	withCSRF(login, csrf)
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, login)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status=%d", loginRec.Code)
	}
	var sessionCookie *http.Cookie
	var loginCSRF *http.Cookie
	for _, c := range loginRec.Result().Cookies() {
		switch c.Name {
		case auth.SessionCookie:
			sessionCookie = c
		case auth.CSRFCookie:
			loginCSRF = c
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/admin/tasks/verify", nil)
	req.AddCookie(sessionCookie)
	withCSRF(req, loginCSRF)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("verify status=%d body=%s", rec.Code, rec.Body.String())
	}
	var report struct {
		MissingCount int `json:"missingCount"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&report); err != nil {
		t.Fatal(err)
	}
	if report.MissingCount != 1 {
		t.Fatalf("missing=%d", report.MissingCount)
	}
}

func TestAdminTasksRequireAdmin(t *testing.T) {
	srv, _ := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/admin/tasks/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
}
