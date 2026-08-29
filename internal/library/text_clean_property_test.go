package library

import (
	"strings"
	"testing"
	"testing/quick"
	"unicode/utf8"
)

func TestPropertyCleanDisplayTextIdempotent(t *testing.T) {
	fn := func(in string) bool {
		once := CleanDisplayText(in)
		return CleanDisplayText(once) == once && utf8.ValidString(once) && !strings.ContainsRune(once, 0)
	}
	if err := quick.Check(fn, &quick.Config{MaxCount: 500}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyCleanSeriesNameIdempotent(t *testing.T) {
	fn := func(in string) bool {
		once := CleanSeriesName(in)
		return CleanSeriesName(once) == once && utf8.ValidString(once)
	}
	if err := quick.Check(fn, &quick.Config{MaxCount: 400}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyCleanBookTitleNeverEmptyWhenPathHasName(t *testing.T) {
	fn := func(title string) bool {
		path := "/library/Readable Title.pdf"
		got := CleanBookTitle(title, path)
		return got != "" && utf8.ValidString(got)
	}
	if err := quick.Check(fn, &quick.Config{MaxCount: 200}); err != nil {
		t.Fatal(err)
	}
}
