package library

import (
	"strings"
	"testing"

	"athenaeum/internal/models"
)

func TestFormatFromExtAll(t *testing.T) {
	cases := map[string]string{
		"book.EPUB":  models.FormatEPUB,
		"/a/b.pdf":   models.FormatPDF,
		"x.mobi":     models.FormatMOBI,
		"x.azw3":     models.FormatAZW3,
		"x.azw":      models.FormatAZW,
		"x.kfx":      models.FormatKFX,
		"x.cbz":      models.FormatCBZ,
		"x.cbr":      models.FormatCBR,
		"x.mp3":      models.FormatMP3,
		"x.m4b":      models.FormatM4B,
		"x.m4a":      models.FormatM4A,
		"x.ogg":      models.FormatOGG,
		"x.flac":     models.FormatFLAC,
		"readme.txt": "",
		"noext":      "",
	}
	for in, want := range cases {
		if got := FormatFromExt(in); got != want {
			t.Errorf("FormatFromExt(%q)=%q want %q", in, got, want)
		}
	}
}

func TestMetadataProvidersAndMatchUpdate(t *testing.T) {
	providers := MetadataProviders()
	if len(providers) == 0 {
		t.Fatal("expected providers")
	}
	upd := MatchToBookUpdate(models.MetadataMatch{
		Title:       "  Title  ",
		Author:      " Author ",
		Series:      "  spaced   series  ",
		SeriesIndex: 2,
		Language:    " en ",
		Description: " desc ",
	})
	if upd.Title != "Title" || upd.Author != "Author" || upd.Language != "en" {
		t.Fatalf("update=%+v", upd)
	}
	if upd.Series == "" || upd.SeriesIndex != 2 {
		t.Fatalf("series fields=%+v", upd)
	}
}

func TestGoogleBooksQueryHelpers(t *testing.T) {
	if got := googleBooksQuery("", "", "978123"); got != "isbn:978123" {
		t.Fatalf("isbn query=%q", got)
	}
	got := googleBooksQuery("The Martian", "Andy Weir", "")
	if !strings.Contains(got, "intitle:") || !strings.Contains(got, "inauthor:") {
		t.Fatalf("query=%q", got)
	}
	if quoteQueryTerm("one") != "one" {
		t.Fatal("single term")
	}
	if quoteQueryTerm("two words") != `"two words"` {
		t.Fatal("multi word term")
	}
	if bestGoogleCover("", " http://img ", "https://other") != "https://img" {
		t.Fatalf("cover=%q", bestGoogleCover("", " http://img ", "https://other"))
	}
	if bestGoogleCover("https://x&edge=curl") != "https://x" {
		t.Fatal("edge=curl strip")
	}
	if bestGoogleCover() != "" {
		t.Fatal("empty covers")
	}
	if parsePublishedYear("2020-01-01") != 2020 {
		t.Fatal("year parse")
	}
	if parsePublishedYear("abc") != 0 || parsePublishedYear("0999") != 0 {
		t.Fatal("invalid year")
	}
}
