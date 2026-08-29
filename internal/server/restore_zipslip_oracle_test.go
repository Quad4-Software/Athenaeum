package server

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/brand"
)

// PROVED_RESTORE_ZIP_SLIP
// Guarantee: restore skips zip entries that escape DataDir via .. path components.

func TestRestoreRejectsZipSlip(t *testing.T) {
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

	outside := filepath.Join(srv.cfg.DataDir, "..", "zip-slip-target")
	outside = filepath.Clean(outside)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	dbw, err := zw.Create(brand.DBFilename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dbw.Write([]byte("sqlite-placeholder")); err != nil {
		t.Fatal(err)
	}
	slip, err := zw.Create("../zip-slip-target")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := slip.Write([]byte("pwned")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	var body bytes.Buffer
	contentType := writeZipMultipart(&body, buf.Bytes())
	req := httptest.NewRequest(http.MethodPost, "/api/admin/restore", &body)
	req.Header.Set("Content-Type", contentType)
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("restore status=%d body=%s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(outside); err == nil {
		t.Fatalf("zip slip wrote outside data dir: %s", outside)
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	fmt.Println("PROVED_RESTORE_ZIP_SLIP: escape entry skipped")
}

func writeZipMultipart(buf *bytes.Buffer, zipBytes []byte) string {
	const boundary = "testdataboundary"
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString(`Content-Disposition: form-data; name="file"; filename="backup.zip"` + "\r\n")
	buf.WriteString("Content-Type: application/zip\r\n\r\n")
	buf.Write(zipBytes)
	buf.WriteString("\r\n--" + boundary + "--\r\n")
	return "multipart/form-data; boundary=" + boundary
}
