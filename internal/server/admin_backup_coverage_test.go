package server

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/brand"
	"athenaeum/internal/models"
)

func TestAdminBackupExportImportRestore(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)

	coverDir := srv.cfg.CoverDir()
	if err := os.MkdirAll(coverDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(coverDir, "1.img"), []byte("cover-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	i18nDir := srv.cfg.I18nDir()
	if err := os.MkdirAll(i18nDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(i18nDir, "custom.json"), []byte(`{"ok":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(srv.cfg.DBPath(), []byte("sqlite-placeholder-for-backup"), 0o600); err != nil {
		t.Fatal(err)
	}

	do := func(method, path string, body io.Reader, contentType string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, body)
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		req.AddCookie(session)
		if method != http.MethodGet && method != http.MethodHead {
			withCSRF(req, csrf)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	rec := do(http.MethodGet, "/api/admin/backup", nil, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("backup status=%d body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/zip" {
		t.Fatalf("content-type=%q", ct)
	}
	backupZip := append([]byte(nil), rec.Body.Bytes()...)
	zr, err := zip.NewReader(bytes.NewReader(backupZip), int64(len(backupZip)))
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	if !names[brand.DBFilename] {
		t.Fatalf("backup missing %s entries=%v", brand.DBFilename, names)
	}
	if !names["covers/1.img"] {
		t.Fatalf("backup missing cover entries=%v", names)
	}
	if !names["config.json"] {
		t.Fatal("backup missing config.json")
	}

	rec = do(http.MethodGet, "/api/admin/config/export", nil, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("export status=%d", rec.Code)
	}
	var exported configExport
	if err := json.NewDecoder(rec.Body).Decode(&exported); err != nil {
		t.Fatal(err)
	}

	importBody, err := json.Marshal(map[string]any{
		"server":    exported.Server,
		"oidc":      exported.OIDC,
		"libraries": exported.Libraries,
	})
	if err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodPost, "/api/admin/config/import", bytes.NewReader(importBody), "application/json")
	if rec.Code != http.StatusOK {
		t.Fatalf("import status=%d body=%s", rec.Code, rec.Body.String())
	}
	var okResp map[string]bool
	if err := json.NewDecoder(rec.Body).Decode(&okResp); err != nil {
		t.Fatal(err)
	}
	if !okResp["ok"] {
		t.Fatalf("import resp=%v", okResp)
	}

	var restoreBuf bytes.Buffer
	mw := multipart.NewWriter(&restoreBuf)
	part, err := mw.CreateFormFile("file", "backup.zip")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(backupZip); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodPost, "/api/admin/restore", &restoreBuf, mw.FormDataContentType())
	if rec.Code != http.StatusOK {
		t.Fatalf("restore status=%d body=%s", rec.Code, rec.Body.String())
	}
	var restoreResp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&restoreResp); err != nil {
		t.Fatal(err)
	}
	if restoreResp["status"] != "restored" {
		t.Fatalf("restore resp=%v", restoreResp)
	}

	rec = do(http.MethodPost, "/api/admin/restore", bytes.NewReader(nil), "multipart/form-data; boundary=x")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("restore missing file status=%d", rec.Code)
	}
}

func TestAddZipFileDirect(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if err := addZipFile(zw, "nested/f.txt", path); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if len(zr.File) != 1 || zr.File[0].Name != "nested/f.txt" {
		t.Fatalf("files=%v", zr.File)
	}
}

func TestConfigImportUpdatesLibraryName(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)

	libDir := filepath.Join(srv.cfg.DataDir, "cfg-lib")
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"name": "Original", "mountPath": libDir, "backend": "local",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/libraries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create lib status=%d", rec.Code)
	}
	var lib models.Library
	if err := json.NewDecoder(rec.Body).Decode(&lib); err != nil {
		t.Fatal(err)
	}

	importBody, _ := json.Marshal(map[string]any{
		"libraries": []map[string]any{{"id": lib.ID, "name": "Renamed Via Import"}},
		"server": map[string]any{
			"metricsEnabled": false, "cspEnabled": false, "autoScanEnabled": false,
		},
	})
	req = httptest.NewRequest(http.MethodPost, "/api/admin/config/import", bytes.NewReader(importBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("import status=%d body=%s", rec.Code, rec.Body.String())
	}
	updated, err := store.GetLibrary(t.Context(), lib.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Renamed Via Import" {
		t.Fatalf("name=%q", updated.Name)
	}
}
