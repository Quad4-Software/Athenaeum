package server

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/models"
)

func writeMinimalEPUBBytes(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	files := map[string]string{
		"mimetype": "application/epub+zip",
		"META-INF/container.xml": `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`,
		"OEBPS/content.opf": `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Upload Tales</dc:title>
    <dc:creator>Uploader</dc:creator>
    <dc:language>en</dc:language>
  </metadata>
  <manifest>
    <item id="c1" href="chap.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="c1"/>
  </spine>
</package>`,
		"OEBPS/chap.xhtml": `<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml"><body><p>Hello upload chapter text.</p></body></html>`,
	}
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(w, content); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestUploadFinalizeFullFlow(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)

	libDir := filepath.Join(srv.cfg.DataDir, "upload-lib")
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}

	doJSON := func(method, path string, body any) *httptest.ResponseRecorder {
		t.Helper()
		var rdr io.Reader
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			rdr = bytes.NewReader(b)
		}
		req := httptest.NewRequest(method, path, rdr)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.AddCookie(session)
		if method != http.MethodGet && method != http.MethodHead {
			withCSRF(req, csrf)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	rec := doJSON(http.MethodPost, "/api/libraries", map[string]any{
		"name": "Upload Lib", "mountPath": libDir, "backend": "local",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create lib status=%d body=%s", rec.Code, rec.Body.String())
	}
	var lib models.Library
	if err := json.NewDecoder(rec.Body).Decode(&lib); err != nil {
		t.Fatal(err)
	}

	payload := writeMinimalEPUBBytes(t)
	rec = doJSON(http.MethodPost, fmt.Sprintf("/api/libraries/%d/uploads", lib.ID), map[string]any{
		"relPath": "incoming/book.epub", "totalSize": len(payload),
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create upload status=%d body=%s", rec.Code, rec.Body.String())
	}
	var sess models.UploadSession
	if err := json.NewDecoder(rec.Body).Decode(&sess); err != nil {
		t.Fatal(err)
	}

	rec = doJSON(http.MethodGet, fmt.Sprintf("/api/libraries/%d/uploads/%s", lib.ID, sess.ID), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get upload status=%d", rec.Code)
	}

	mid := len(payload) / 2
	patch := func(start, end int, chunk []byte) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodPatch,
			fmt.Sprintf("/api/libraries/%d/uploads/%s", lib.ID, sess.ID),
			bytes.NewReader(chunk))
		req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(payload)))
		req.AddCookie(session)
		withCSRF(req, csrf)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	rec = patch(0, mid-1, payload[:mid])
	if rec.Code != http.StatusOK {
		t.Fatalf("patch mid status=%d body=%s", rec.Code, rec.Body.String())
	}
	var midSess models.UploadSession
	if err := json.NewDecoder(rec.Body).Decode(&midSess); err != nil {
		t.Fatal(err)
	}
	if midSess.Done || midSess.Offset != int64(mid) {
		t.Fatalf("mid sess done=%v offset=%d", midSess.Done, midSess.Offset)
	}

	rec = patch(mid, len(payload)-1, payload[mid:])
	if rec.Code != http.StatusOK {
		t.Fatalf("patch final status=%d body=%s", rec.Code, rec.Body.String())
	}
	var done models.UploadSession
	if err := json.NewDecoder(rec.Body).Decode(&done); err != nil {
		t.Fatal(err)
	}
	if !done.Done || done.BookID <= 0 {
		t.Fatalf("finalize done=%v bookID=%d body offset=%d", done.Done, done.BookID, done.Offset)
	}

	dest := filepath.Join(libDir, "incoming", "book.epub")
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("library file missing: %v", err)
	}

	rec = doJSON(http.MethodPost, fmt.Sprintf("/api/libraries/%d/uploads", lib.ID), map[string]any{
		"relPath": "../escape.epub", "totalSize": 10,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("escape path status=%d", rec.Code)
	}
	rec = doJSON(http.MethodPost, fmt.Sprintf("/api/libraries/%d/uploads", lib.ID), map[string]any{
		"relPath": "nope.txt", "totalSize": 10,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad ext status=%d", rec.Code)
	}
	rec = doJSON(http.MethodPost, fmt.Sprintf("/api/libraries/%d/uploads", lib.ID), map[string]any{
		"relPath": "ok.pdf", "totalSize": 0,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("zero size status=%d", rec.Code)
	}

	delPayload := []byte("%PDF-1.4 tiny")
	rec = doJSON(http.MethodPost, fmt.Sprintf("/api/libraries/%d/uploads", lib.ID), map[string]any{
		"relPath": "to-delete.pdf", "totalSize": len(delPayload),
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create delete upload status=%d", rec.Code)
	}
	var delSess models.UploadSession
	if err := json.NewDecoder(rec.Body).Decode(&delSess); err != nil {
		t.Fatal(err)
	}
	delReq := httptest.NewRequest(http.MethodDelete,
		fmt.Sprintf("/api/libraries/%d/uploads/%s", lib.ID, delSess.ID), nil)
	delReq.AddCookie(session)
	withCSRF(delReq, csrf)
	delRec := httptest.NewRecorder()
	handler.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("delete upload status=%d body=%s", delRec.Code, delRec.Body.String())
	}
}

func TestParseContentRangeErrors(t *testing.T) {
	cases := []string{
		"",
		"bytes ",
		"bytes 0-1",
		"bytes 0-1/x",
		"bytes 0-1/0",
		"bytes abc-1/10",
		"bytes 5-1/10",
		"bytes 0-x/10",
	}
	for _, h := range cases {
		if _, _, _, err := parseContentRange(h); err == nil {
			t.Fatalf("expected error for %q", h)
		}
	}
	start, end, total, err := parseContentRange("bytes 2-9/20")
	if err != nil || start != 2 || end != 9 || total != 20 {
		t.Fatalf("got %d-%d/%d err=%v", start, end, total, err)
	}
}

func TestSanitizeUploadRelPathEdges(t *testing.T) {
	if _, err := sanitizeUploadRelPath(""); err == nil {
		t.Fatal("empty")
	}
	if _, err := sanitizeUploadRelPath(".."); err == nil {
		t.Fatal("dotdot")
	}
	if _, err := sanitizeUploadRelPath("a/../../b.pdf"); err == nil {
		t.Fatal("nested traversal")
	}
	got, err := sanitizeUploadRelPath("/Books/Sub/file.EPUB")
	if err != nil || got != "Books/Sub/file.EPUB" {
		t.Fatalf("got %q err=%v", got, err)
	}
	if uploadFormatFromExt("x.PDF") == "" {
		t.Fatal("pdf ext")
	}
}
