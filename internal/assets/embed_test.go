package assets_test

import (
	"io"
	"io/fs"
	"strings"
	"testing"

	"athenaeum/internal/assets"
)

func TestDistFSFallsBackToStubIndex(t *testing.T) {
	// Prefer built index when present; otherwise fallback shell.
	f, err := assets.DistFS.Open("dist/index.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	raw, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, "Athenaeum") {
		t.Fatalf("unexpected index shell: %q", body[:min(80, len(body))])
	}
}

func TestDistFSKeepsTrackedIcons(t *testing.T) {
	if _, err := fs.Stat(assets.DistFS, "dist/favicon.svg"); err != nil {
		t.Fatal(err)
	}
}
