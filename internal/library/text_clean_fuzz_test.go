package library

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func FuzzCleanDisplayText(f *testing.F) {
	seeds := []string{
		"",
		"Hello",
		"  Hello   World  ",
		"Coffee\x00migraine",
		"Line\u200bone",
		"Bad\uFFFDtext",
		"a\tb",
		strings.Repeat("x", 4096),
		"\xff\xfe",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		got := CleanDisplayText(in)
		if !utf8.ValidString(got) {
			t.Fatalf("invalid utf8: %q", got)
		}
		if strings.ContainsRune(got, 0) {
			t.Fatalf("nul in output: %q", got)
		}
		if again := CleanDisplayText(got); again != got {
			t.Fatalf("not idempotent: %q -> %q -> %q", in, got, again)
		}
	})
}

func FuzzIsGarbledText(f *testing.F) {
	for _, s := range []string{"", "Masters of Doom", "1\uFFFD\uFFFDbS", "a b c d"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		_ = IsGarbledText(in)
		cleaned := CleanDisplayText(in)
		_ = IsGarbledText(cleaned)
	})
}
