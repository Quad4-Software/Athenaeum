package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

// PROVED_LIBRARY_MUTATE_AUTHZ
// Guarantee: non-admin users without manage_library cannot create libraries
// or browse the host filesystem via /api/fs/browse.

func TestNonAdminCannotCreateLibrary(t *testing.T) {
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
	session, _, csrf := loginUser(t, handler, "reader", "longpassword")

	body, _ := json.Marshal(map[string]string{
		"name":      "Evil",
		"mountPath": t.TempDir(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/libraries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("create library status=%d body=%s want 403", rec.Code, rec.Body.String())
	}
	fmt.Println("PROVED_LIBRARY_MUTATE_AUTHZ: create denied for default user")
}

func TestNonAdminCannotBrowseFS(t *testing.T) {
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
	libDir := filepath.Join(srv.cfg.DataDir, "lib")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, _, _ := loginUser(t, handler, "reader", "longpassword")

	req := httptest.NewRequest(http.MethodGet, "/api/fs/browse", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("fs browse status=%d body=%s want 403", rec.Code, rec.Body.String())
	}
	fmt.Println("PROVED_LIBRARY_MUTATE_AUTHZ: fs browse denied for default user")
}

func TestAdminCanCreateLibrary(t *testing.T) {
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
	mount := t.TempDir()
	body, _ := json.Marshal(map[string]string{"name": "Main", "mountPath": mount})
	req := httptest.NewRequest(http.MethodPost, "/api/libraries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("admin create status=%d body=%s", rec.Code, rec.Body.String())
	}
	var lib models.Library
	if err := json.NewDecoder(rec.Body).Decode(&lib); err != nil {
		t.Fatal(err)
	}
	if lib.Name != "Main" {
		t.Fatalf("name=%q", lib.Name)
	}
}
