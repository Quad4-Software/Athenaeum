package storage

import (
	"testing"
)

func TestBuildFTSQuery(t *testing.T) {
	cases := map[string]string{
		"":            "",
		"  ":          "",
		"dune":        `"dune"*`,
		"left hand":   `"left"* AND "hand"*`,
		`say "hello"`: `"say"* AND "hello"*`,
	}
	for in, want := range cases {
		if got := buildFTSQuery(in, " AND "); got != want {
			t.Errorf("buildFTSQuery(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildPostgresTSQuery(t *testing.T) {
	cases := map[string]string{
		"":          "",
		"dune":      "dune:*",
		"left hand": "left:* & hand:*",
	}
	for in, want := range cases {
		if got := buildPostgresTSQuery(in, " & "); got != want {
			t.Errorf("buildPostgresTSQuery(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFTSResultCap(t *testing.T) {
	if maxFTSResults < 1 || maxFTSResults > 5000 {
		t.Fatalf("maxFTSResults out of expected range: %d", maxFTSResults)
	}
}
