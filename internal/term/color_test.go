package term_test

import (
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
	if got := term.Bold(nil, "x"); got != "x" {
		t.Fatalf("Bold without writer: %q", got)
	}
}
