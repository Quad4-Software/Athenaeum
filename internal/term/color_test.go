package term_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"athenaeum/internal/term"
)

func TestParseMode(t *testing.T) {
	cases := map[string]term.Mode{
		"":       term.ModeAuto,
		"auto":   term.ModeAuto,
		"always": term.ModeAlways,
		"never":  term.ModeNever,
		"on":     term.ModeAlways,
		"off":    term.ModeNever,
	}
	for in, want := range cases {
		got, err := term.ParseMode(in)
		if err != nil {
			t.Fatalf("ParseMode(%q): %v", in, err)
		}
		if got != want {
			t.Fatalf("ParseMode(%q)=%v want %v", in, got, want)
		}
	}
	if _, err := term.ParseMode("rainbow"); err == nil {
		t.Fatal("expected error for invalid mode")
	}
}

func TestWrapNoColor(t *testing.T) {
	term.Apply(term.ModeNever)
	t.Cleanup(func() { term.Apply(term.ModeAuto) })
	if got := term.Bold(nil, "x"); got != "x" {
		t.Fatalf("Bold without writer: %q", got)
	}
}

func TestColorHelpersForced(t *testing.T) {
	term.Apply(term.ModeAlways)
	t.Cleanup(func() { term.Apply(term.ModeAuto) })

	w := os.Stdout
	if !term.Enabled(w) {
		t.Fatal("expected color enabled for stdout")
	}
	checks := []struct {
		name string
		got  string
	}{
		{"Red", term.Red(w, "r")},
		{"Cyan", term.Cyan(w, "c")},
		{"Magenta", term.Magenta(w, "m")},
		{"Header", term.Header(w, "h")},
		{"Flag", term.Flag(w, "f")},
		{"Error", term.Error(w, "e")},
		{"Warn", term.Warn(w, "w")},
		{"Dim", term.Dim(w, "d")},
		{"Yellow", term.Yellow(w, "y")},
	}
	for _, tc := range checks {
		if !strings.Contains(tc.got, "\x1b[") || !strings.HasSuffix(tc.got, "\x1b[0m") {
			t.Fatalf("%s = %q missing ansi", tc.name, tc.got)
		}
	}
	if term.Red(w, "") != "" {
		t.Fatal("empty text should stay empty")
	}

	var buf bytes.Buffer
	term.Fprint(&buf, "a")
	term.Fprintln(&buf, "b")
	term.Fprintf(&buf, "%s", "c")
	if buf.String() != "ab\nc" {
		t.Fatalf("print helpers=%q", buf.String())
	}
}

func TestApplyForceAndNoColor(t *testing.T) {
	t.Setenv("FORCE_COLOR", "1")
	t.Setenv("NO_COLOR", "")
	t.Setenv("ATHENAEUM_NO_COLOR", "")
	term.Apply(term.ModeAuto)
	t.Cleanup(func() { term.Apply(term.ModeAuto) })
	if !term.Enabled(os.Stdout) {
		t.Fatal("FORCE_COLOR should enable")
	}

	t.Setenv("FORCE_COLOR", "")
	t.Setenv("NO_COLOR", "1")
	term.Apply(term.ModeAuto)
	if term.Enabled(os.Stdout) {
		t.Fatal("NO_COLOR should disable")
	}
}
