package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadEnvFileSetsUnsetVariables(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("ATHENAEUM_ADDR=:9090\nATHENAEUM_LIBRARY=/books\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ATHENAEUM_ADDR", "")

	if err := loadEnvFile(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("ATHENAEUM_ADDR"); got != ":9090" {
		t.Fatalf("ATHENAEUM_ADDR=%q", got)
	}
	if got := os.Getenv("ATHENAEUM_LIBRARY"); got != "/books" {
		t.Fatalf("ATHENAEUM_LIBRARY=%q", got)
	}
}

func TestLoadEnvFileDoesNotOverrideExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("ATHENAEUM_ADDR=:9090\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ATHENAEUM_ADDR", ":7070")

	if err := loadEnvFile(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("ATHENAEUM_ADDR"); got != ":7070" {
		t.Fatalf("ATHENAEUM_ADDR=%q want :7070", got)
	}
}

func TestParseUsesEnvAndFlags(t *testing.T) {
	t.Setenv("ATHENAEUM_ADDR", ":9001")
	t.Setenv("ATHENAEUM_DATA", "")
	t.Setenv("ATHENAEUM_LIBRARY", "")

	cfg, err := Parse([]string{"--addr", ":9002", "--data", "./testdata-data", "--library", "./testdata-lib"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":9002" {
		t.Fatalf("addr=%q", cfg.Addr)
	}
	if !filepath.IsAbs(cfg.DataDir) || !filepath.IsAbs(cfg.LibraryDir) {
		t.Fatalf("paths should be absolute: data=%s library=%s", cfg.DataDir, cfg.LibraryDir)
	}
}

func TestParseLoadsDotenvBeforeDefaults(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("ATHENAEUM_ADDR=:8088\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ATHENAEUM_ADDR", "")
	t.Setenv("ATHENAEUM_DATA", filepath.Join(dir, "data"))
	t.Setenv("ATHENAEUM_LIBRARY", filepath.Join(dir, "library"))

	cfg, err := Parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":8088" {
		t.Fatalf("addr=%q want :8088", cfg.Addr)
	}
}

func TestParseScanWorkersDefault(t *testing.T) {
	t.Setenv("ATHENAEUM_SCAN_WORKERS", "")
	t.Setenv("ATHENAEUM_DATA", t.TempDir())
	t.Setenv("ATHENAEUM_LIBRARY", t.TempDir())

	cfg, err := Parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ScanWorkers != 2 {
		t.Fatalf("ScanWorkers=%d want 2", cfg.ScanWorkers)
	}
}

func TestParseScanWorkersEnv(t *testing.T) {
	t.Setenv("ATHENAEUM_SCAN_WORKERS", "4")
	t.Setenv("ATHENAEUM_DATA", t.TempDir())
	t.Setenv("ATHENAEUM_LIBRARY", t.TempDir())

	cfg, err := Parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ScanWorkers != 4 {
		t.Fatalf("ScanWorkers=%d want 4", cfg.ScanWorkers)
	}
}

func TestParseWebDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ATHENAEUM_DATA", t.TempDir())
	t.Setenv("ATHENAEUM_LIBRARY", t.TempDir())
	t.Setenv("ATHENAEUM_WEB_DIR", "")

	cfg, err := Parse([]string{"--web-dir", dir})
	if err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WebDir != abs {
		t.Fatalf("WebDir=%q want %q", cfg.WebDir, abs)
	}
}

func TestParseWebDirEmptyUsesEmbedded(t *testing.T) {
	t.Setenv("ATHENAEUM_DATA", t.TempDir())
	t.Setenv("ATHENAEUM_LIBRARY", t.TempDir())
	t.Setenv("ATHENAEUM_WEB_DIR", "")

	cfg, err := Parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WebDir != "" {
		t.Fatalf("WebDir=%q want empty", cfg.WebDir)
	}
}

func TestUsesPostgresAndLogFileDir(t *testing.T) {
	for _, driver := range []string{"postgres", "PostgreSQL", "pg", " PG "} {
		if !(Config{DatabaseDriver: driver}).UsesPostgres() {
			t.Fatalf("UsesPostgres(%q) = false", driver)
		}
	}
	if (Config{DatabaseDriver: "sqlite"}).UsesPostgres() {
		t.Fatal("sqlite should not be postgres")
	}
	if (Config{}).LogFileDir() != "" {
		t.Fatal("empty log file dir")
	}
	got := (Config{LogFile: "/var/log/athenaeum/app.log"}).LogFileDir()
	if got != "/var/log/athenaeum" {
		t.Fatalf("LogFileDir=%q", got)
	}
}

func TestPrintHelp(t *testing.T) {
	var buf bytes.Buffer
	PrintHelp(&buf)
	out := buf.String()
	if !strings.Contains(out, "Usage:") || !strings.Contains(out, "--addr") {
		t.Fatalf("help missing expected sections: %s", out[:min(200, len(out))])
	}
}
