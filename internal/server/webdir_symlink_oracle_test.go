package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// PROVED_WEBDIR_SYMLINK_JAIL
// Guarantee claimed: spaHandler(webDir) must not serve symlink targets
// outside the configured web directory.

func TestWebDirSymlinkEscapeOracle(t *testing.T) {
	base := t.TempDir()
	web := filepath.Join(base, "web")
	secretDir := filepath.Join(base, "secret")
	if err := os.MkdirAll(web, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(secretDir, 0o755); err != nil {
		t.Fatal(err)
	}
	secretPath := filepath.Join(secretDir, "secret.txt")
	if err := os.WriteFile(secretPath, []byte("TOPSECRET"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(web, "index.html"), []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secretPath, filepath.Join(web, "leak.txt")); err != nil {
		t.Fatal(err)
	}

	h, err := spaHandler(web)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/leak.txt", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	body, _ := io.ReadAll(rec.Body)
	if rec.Code == http.StatusOK && strings.Contains(string(body), "TOPSECRET") {
		fmt.Println("PROVED_WEBDIR_SYMLINK_JAIL: served outside webDir via symlink")
		return
	}
	t.Fatalf("not vulnerable: status=%d body=%q", rec.Code, body)
}
