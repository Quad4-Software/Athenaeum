package library

import (
	"strings"
	"testing"

	"athenaeum/internal/models"
)

func TestNormalizeDOI(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"10.1000/xyz123", "10.1000/xyz123"},
		{"doi:10.1000/xyz123", "10.1000/xyz123"},
		{"https://doi.org/10.1000/xyz123", "10.1000/xyz123"},
		{"https://dx.doi.org/10.1000/xyz123.", "10.1000/xyz123"},
		{"10.", ""},
		{"10.1234", ""},
		{"not-a-doi", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := NormalizeDOI(c.in); got != c.want {
			t.Errorf("NormalizeDOI(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeArxivID(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"2301.12345", "2301.12345"},
		{"arXiv:2301.12345v2", "2301.12345"},
		{"https://arxiv.org/abs/2301.12345", "2301.12345"},
		{"hep-th/9901001", "hep-th/9901001"},
		{"nonsense", ""},
	}
	for _, c := range cases {
		if got := NormalizeArxivID(c.in); got != c.want {
			t.Errorf("NormalizeArxivID(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizePubmedID(t *testing.T) {
	if got := NormalizePubmedID("PMID: 12345678"); got != "12345678" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizePubmedID("42"); got != "42" {
		t.Fatalf("got %q", got)
	}
}

func TestParseBibTeXAndRoundTrip(t *testing.T) {
	raw := `
@article{smith2020,
  title = {A Sample Paper},
  author = {Smith, Jane and Doe, John},
  journal = {Nature},
  year = {2020},
  volume = {12},
  number = {3},
  pages = {10--20},
  doi = {10.1000/sample},
}
`
	entries := ParseBibTeX(raw)
	if len(entries) != 1 {
		t.Fatalf("entries=%d", len(entries))
	}
	side := BibEntryToSidecar(entries[0])
	if side.Title != "A Sample Paper" {
		t.Fatalf("title=%q", side.Title)
	}
	if side.DOI != "10.1000/sample" {
		t.Fatalf("doi=%q", side.DOI)
	}
	if side.Journal != "Nature" || side.PublishedYear != 2020 {
		t.Fatalf("journal/year=%q/%d", side.Journal, side.PublishedYear)
	}
	if !strings.Contains(side.Author, "Smith") {
		t.Fatalf("author=%q", side.Author)
	}

	book := models.Book{
		Title:         side.Title,
		Author:        side.Author,
		DOI:           side.DOI,
		Journal:       side.Journal,
		Volume:        side.Volume,
		Issue:         side.Issue,
		Pages:         side.Pages,
		PublishedYear: side.PublishedYear,
		Description:   "abstract text",
	}
	out := FormatBibTeX(book)
	if !strings.Contains(out, "@article{") {
		t.Fatalf("missing @article: %s", out)
	}
	if !strings.Contains(out, "10.1000/sample") {
		t.Fatalf("missing doi: %s", out)
	}
	round := ParseBibTeX(out)
	if len(round) != 1 || round[0].Fields["title"] == "" {
		t.Fatalf("round-trip failed: %+v", round)
	}
}

func TestFindDOIInText(t *testing.T) {
	text := "See doi:10.1038/nature12373 for details."
	if got := findDOIInText(text); got != "10.1038/nature12373" {
		t.Fatalf("got %q", got)
	}
}

func TestMatchToBookUpdateCitation(t *testing.T) {
	u := MatchToBookUpdate(models.MetadataMatch{
		Title:         "T",
		Author:        "A",
		DOI:           "10.1000/x",
		ArxivID:       "2301.12345",
		PubmedID:      "99",
		Journal:       "J",
		Volume:        "1",
		Issue:         "2",
		Pages:         "3-4",
		PublishedYear: 2021,
	})
	if u.DOI != "10.1000/x" || u.Journal != "J" || u.PublishedYear != 2021 {
		t.Fatalf("%+v", u)
	}
}

func TestIsPaper(t *testing.T) {
	if models.IsPaper(models.Book{Title: "x"}) {
		t.Fatal("expected not paper")
	}
	if !models.IsPaper(models.Book{DOI: "10.1000/x"}) {
		t.Fatal("expected paper")
	}
}
