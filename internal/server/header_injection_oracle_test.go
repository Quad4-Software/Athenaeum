package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

// PROVED_SAFE_FILENAME_HEADER
// Guarantee: Content-Disposition / SMTP Subject from book titles
// cannot inject CR LF header fields or break out of quoted filenames.

func TestDownloadContentDispositionCRLFOracle(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}
	libDir := filepath.Join(srv.cfg.DataDir, "lib")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libDir, "a.epub"), []byte("epub"), 0o644); err != nil {
		t.Fatal(err)
	}
	lib, err := store.CreateLibrary(ctx, "Main", libDir)
	if err != nil {
		t.Fatal(err)
	}
	id, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     "Nice\r\nX-Injected: pwned",
		Format:    models.FormatEPUB,
		RelPath:   "a.epub",
	}, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, _, _ := loginUser(t, handler, "admin", "longpassword")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/download", id), nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("download status=%d body=%s", rec.Code, rec.Body.String())
	}

	raw := rec.Header().Get("Content-Disposition")
	injected := rec.Header().Get("X-Injected")
	if strings.Contains(raw, "\r") || strings.Contains(raw, "\n") || injected != "" {
		t.Fatalf("header injection disposition=%q X-Injected=%q", raw, injected)
	}
	fmt.Println("PROVED_SAFE_FILENAME_HEADER: CR/LF stripped from disposition=", raw)
}

func TestSafeFilenameQuoteBreakoutOracle(t *testing.T) {
	name := safeFilename(models.Book{Title: `evil"; filename="pwned`, Format: "epub"})
	if strings.Contains(name, `"`) {
		t.Fatalf("quotes remain in filename=%q", name)
	}
	fmt.Println("PROVED_SAFE_FILENAME_HEADER: quotes stripped filename=", name)
}

func TestSMTPSubjectHeaderInjectionOracle(t *testing.T) {
	msg := buildMIMEAttachment("from@x", "to@x", "Title\r\nBcc: evil@x", "a.epub", []byte("x"))
	raw := string(msg)
	if strings.Contains(raw, "\r\nBcc:") {
		t.Fatal("smtp subject header injection still present")
	}
	fmt.Println("PROVED_SAFE_FILENAME_HEADER: smtp subject sanitized")
}
