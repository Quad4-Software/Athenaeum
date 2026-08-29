package server

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"athenaeum/internal/assets"
)

func TestServiceWorkerNoCache(t *testing.T) {
	handler, err := spaHandler("")
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"/sw.js", "/manifest.webmanifest"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status=%d", path, rec.Code)
		}
		cc := rec.Header().Get("Cache-Control")
		if cc == "" || cc == "public, max-age=31536000, immutable" {
			t.Fatalf("%s Cache-Control=%q want no-store policy", path, cc)
		}
	}
}

func TestHashedAssetsImmutable(t *testing.T) {
	handler, err := spaHandler("")
	if err != nil {
		t.Fatal(err)
	}

	sub, err := fs.Sub(assets.DistFS, "dist")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := fs.ReadDir(sub, "assets")
	if err != nil {
		t.Fatal(err)
	}
	var sample string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "index-") && strings.HasSuffix(name, ".css") {
			sample = "assets/" + name
			break
		}
	}
	if sample == "" {
		t.Fatal("no hashed index-*.css asset in embedded dist")
	}

	req := httptest.NewRequest(http.MethodGet, "/"+sample, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d path=%s", rec.Code, sample)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("Cache-Control=%q", got)
	}
}

func TestSPAHandlerWebDir(t *testing.T) {
	dir := t.TempDir()
	index := "<!doctype html><title>external</title>"
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(index), 0o600); err != nil {
		t.Fatal(err)
	}
	assetsDir := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assetsDir, 0o750); err != nil {
		t.Fatal(err)
	}
	cssName := "index-test.css"
	if err := os.WriteFile(filepath.Join(assetsDir, cssName), []byte("body{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	handler, err := spaHandler(dir)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/book/1", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("spa fallback status=%d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "external") {
		t.Fatalf("body=%q", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/assets/"+cssName, nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("asset status=%d", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("Cache-Control=%q", got)
	}
}

func TestSPAHandlerWebDirMissingIndex(t *testing.T) {
	dir := t.TempDir()
	if _, err := spaHandler(dir); err == nil {
		t.Fatal("expected error for missing index.html")
	}
}

func TestSPAHandlerWebDirMissing(t *testing.T) {
	if _, err := spaHandler(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected error for missing web-dir")
	}
}
