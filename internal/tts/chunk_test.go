package tts

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestChunkParagraphsJoinsShortParas(t *testing.T) {
	got := ChunkParagraphs([]string{"One.", "Two.", "Three."}, 100)
	if len(got) != 1 || got[0] != "One. Two. Three." {
		t.Fatalf("got %#v", got)
	}
}

func TestChunkParagraphsRespectsMax(t *testing.T) {
	paras := []string{"aaaa", "bbbb", "cccc"}
	got := ChunkParagraphs(paras, 10)
	if len(got) != 2 {
		t.Fatalf("expected 2 chunks, got %#v", got)
	}
	for _, c := range got {
		if utf8.RuneCountInString(c) > 10 {
			t.Fatalf("chunk over max: %q", c)
		}
	}
}

func TestChunkParagraphsSplitsLongParagraphOnSentence(t *testing.T) {
	long := "First sentence. Second sentence. Third sentence. Fourth sentence."
	got := ChunkParagraphs([]string{long}, 40)
	if len(got) < 2 {
		t.Fatalf("expected split, got %#v", got)
	}
	if got[0] != "First sentence. Second sentence." {
		t.Fatalf("unexpected first chunk %q", got[0])
	}
	var joined []string
	joined = append(joined, got...)
	if strings.Join(joined, " ") != long {
		t.Fatalf("content lost: %#v", joined)
	}
}

func TestChunkParagraphsHardSplitsWithoutBoundaries(t *testing.T) {
	long := strings.Repeat("x", 100)
	got := ChunkParagraphs([]string{long}, 30)
	if len(got) < 4 {
		t.Fatalf("expected hard splits, got %#v", got)
	}
	for _, c := range got {
		if utf8.RuneCountInString(c) > 30 {
			t.Fatalf("chunk over max: %q", c)
		}
	}
	if n := utf8.RuneCountInString(strings.Join(got, "")); n != 100 {
		t.Fatalf("content lost: %d runes", n)
	}
}
