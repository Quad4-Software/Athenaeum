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
	"strings"
	"testing"

	"athenaeum/internal/models"
)

func TestSendBookSMTPDisabled(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)
	ctx := context.Background()

	libDir := filepath.Join(srv.cfg.DataDir, "smtp-lib")
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}
	pdfPath := filepath.Join(libDir, "send.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4 sendme"), 0o640); err != nil {
		t.Fatal(err)
	}
	lib, err := store.CreateLibrary(ctx, "SMTP Lib", libDir)
	if err != nil {
		t.Fatal(err)
	}
	bookID, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Send Me",
		Format:    models.FormatPDF,
		RelPath:   "send.pdf",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string]any{"to": "reader@example.com"})
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/books/%d/send", bookID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("smtp disabled status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "smtp") {
		t.Fatalf("body=%s", rec.Body.String())
	}

	body, _ = json.Marshal(map[string]any{"kindle": true})
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/books/%d/send", bookID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("kindle missing status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSendSMTPTextAndMIME(t *testing.T) {
	msg := buildMIMEAttachment("from@x", "to@x", "Subject", "file.pdf", []byte("payload-data-here"))
	if !strings.Contains(string(msg), "Content-Transfer-Encoding: base64") {
		t.Fatalf("mime=%s", msg)
	}
	if !strings.Contains(string(msg), "filename=\"file.pdf\"") {
		t.Fatal("missing filename")
	}

	err := sendSMTPText(models.SMTPSettings{}, "to@x", "subj", "body")
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("err=%v", err)
	}
	err = sendSMTPText(models.SMTPSettings{Enabled: true, Host: "", FromAddr: "a@b.c"}, "to@x", "subj", "body")
	if err == nil {
		t.Fatal("expected host error")
	}
	err = sendSMTPText(models.SMTPSettings{
		Enabled: true, Host: "127.0.0.1", Port: 1, FromAddr: "a@b.c",
	}, "to@x", "subj", "body")
	if err == nil {
		t.Fatal("expected dial error")
	}
}
