package logging_test

import (
	"log/slog"
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
