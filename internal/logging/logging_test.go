package logging_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/logging"
)

func TestParseLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"":      slog.LevelInfo,
		"info":  slog.LevelInfo,
		"debug": slog.LevelDebug,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
	for in, want := range cases {
		got, err := logging.ParseLevel(in)
		if err != nil {
			t.Fatalf("ParseLevel(%q): %v", in, err)
		}
		if got != want {
			t.Fatalf("ParseLevel(%q)=%v want %v", in, got, want)
		}
	}
	if _, err := logging.ParseLevel("trace"); err == nil {
		t.Fatal("expected error")
	}
}

func TestSetupStderrOnly(t *testing.T) {
	closer, err := logging.Setup(logging.Options{Level: "info"})
	if err != nil {
		t.Fatal(err)
	}
	if err := closer.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSetupWithFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	closer, err := logging.Setup(logging.Options{Level: "debug", File: path, AddSource: true})
	if err != nil {
		t.Fatal(err)
	}
	slog.Info("coverage")
	if err := closer.Close(); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(path)
	if err != nil || st.Size() == 0 {
		t.Fatalf("log file missing or empty: %v", err)
	}
}

func TestSetupBadLevel(t *testing.T) {
	if _, err := logging.Setup(logging.Options{Level: "nope"}); err == nil {
		t.Fatal("expected error")
	}
}
