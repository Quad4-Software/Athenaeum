package backup

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func buildZip(t *testing.T, entries []struct {
	name    string
	data    string
	mode    os.FileMode
	setMode bool
}) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		h := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		if e.setMode {
			h.SetMode(e.mode)
		} else {
			h.SetMode(0o600)
		}
		w, err := zw.CreateHeader(h)
		if err != nil {
			t.Fatalf("create header %s: %v", e.name, err)
		}
		if _, err := w.Write([]byte(e.data)); err != nil {
			t.Fatalf("write %s: %v", e.name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func TestRestoreExtractsIntoDataDir(t *testing.T) {
	dir := t.TempDir()
	z := buildZip(t, []struct {
		name    string
		data    string
		mode    os.FileMode
		setMode bool
	}{
		{name: "athenaeum.db", data: "db-bytes"},
		{name: "covers/a.png", data: "png"},
	})
	if err := Restore(bytes.NewReader(z), dir, "athenaeum.db"); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "covers", "a.png"))
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if string(b) != "png" {
		t.Fatalf("unexpected content %q", b)
	}
}

func TestRestoreRejectsTraversalAndLinks(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "evil.txt")
	cases := []string{
		"../evil.txt",
		"../../evil.txt",
		"/abs.txt",
		"..\\evil.txt",
	}
	var entries []struct {
		name    string
		data    string
		mode    os.FileMode
		setMode bool
	}
	entries = append(entries, struct {
		name    string
		data    string
		mode    os.FileMode
		setMode bool
	}{name: "athenaeum.db", data: "db"})
	for _, n := range cases {
		entries = append(entries, struct {
			name    string
			data    string
			mode    os.FileMode
			setMode bool
		}{name: n, data: "evil"})
	}
	// Symlink entry must be skipped, not written as a link.
	entries = append(entries, struct {
		name    string
		data    string
		mode    os.FileMode
		setMode bool
	}{name: "link.txt", data: "/etc/passwd", mode: os.ModeSymlink | 0o777, setMode: true})
	z := buildZip(t, entries)
	if err := Restore(bytes.NewReader(z), dir, "athenaeum.db"); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatalf("traversal entry escaped to %s", outside)
	}
	entriesOut, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entriesOut {
		names = append(names, e.Name())
	}
	for _, n := range names {
		if strings.Contains(n, "evil") || n == "link.txt" {
			t.Fatalf("unexpected extracted entry %s", n)
		}
	}
}

func TestRestoreRequiresDBEntry(t *testing.T) {
	dir := t.TempDir()
	z := buildZip(t, []struct {
		name    string
		data    string
		mode    os.FileMode
		setMode bool
	}{
		{name: "other.bin", data: "x"},
	})
	if err := Restore(bytes.NewReader(z), dir, "athenaeum.db"); err == nil {
		t.Fatal("expected error for archive without db entry")
	}
}
