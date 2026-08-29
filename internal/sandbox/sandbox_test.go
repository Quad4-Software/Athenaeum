package sandbox_test

import (
	"testing"

	"athenaeum/internal/sandbox"
)

func TestParseMode(t *testing.T) {
	cases := map[string]sandbox.Mode{
		"":       sandbox.ModeOff,
		"off":    sandbox.ModeOff,
		"try":    sandbox.ModeTry,
		"strict": sandbox.ModeStrict,
		"on":     sandbox.ModeStrict,
	}
	for in, want := range cases {
		got, err := sandbox.ParseMode(in)
		if err != nil {
			t.Fatalf("ParseMode(%q): %v", in, err)
		}
		if got != want {
			t.Fatalf("ParseMode(%q)=%v want %v", in, got, want)
		}
	}
	if _, err := sandbox.ParseMode("maybe"); err == nil {
		t.Fatal("expected error")
	}
}
