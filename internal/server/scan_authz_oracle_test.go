package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

// PROVED_SCAN_REQUIRES_MANAGE_LIBRARY
// Guarantee: default readers without manage_library cannot start a library scan.

func TestScanRequiresManageLibraryOracle(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodPost, "/api/library/scan", nil)
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("global scan status=%d want 403 body=%s", rec.Code, rec.Body.String())
	}

	libDir := filepath.Join(srv.cfg.DataDir, "scanlib")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatal(err)
	}
	lib, err := store.CreateLibraryFull(ctx, models.LibraryCreate{
		Name: "scanlib", MountPath: libDir, Backend: models.LibraryBackendLocal,
	})
	if err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/libraries/%d/scan", lib.ID), nil)
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("library scan status=%d want 403 body=%s", rec.Code, rec.Body.String())
	}
	fmt.Println("PROVED_SCAN_REQUIRES_MANAGE_LIBRARY: reader denied scan")
}
