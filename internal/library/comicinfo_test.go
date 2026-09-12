package library

import (
	"archive/zip"
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func writeCBZ(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

const fullComicInfoXML = `<?xml version="1.0"?>
<ComicInfo>
	<Title>The Big Showdown</Title>
	<Series>Hero Tales</Series>
	<Number>7</Number>
	<Writer>Jane Creator</Writer>
	<Summary>A hero does things.</Summary>
	<Publisher>Indie Press</Publisher>
	<Genre>Action, Fantasy</Genre>
	<Tags>favorites, classic</Tags>
	<Year>2011</Year>
	<LanguageISO>en</LanguageISO>
	<Manga>YesAndRightToLeft</Manga>
	<AgeRating>Mature 17+</AgeRating>
</ComicInfo>`

func TestParseCBZComicInfo(t *testing.T) {
	page := map[string]string{"page1.jpg": "fakejpeg"}

	cases := []struct {
		name  string
		files map[string]string
		want  comicInfo
	}{
		{
			name:  "all fields",
			files: mergeMaps(page, map[string]string{"ComicInfo.xml": fullComicInfoXML}),
			want: comicInfo{
				Title:            "The Big Showdown",
				Series:           "Hero Tales",
				SeriesIndex:      7,
				Author:           "Jane Creator",
				Description:      "A hero does things.",
				Publisher:        "Indie Press",
				Tags:             []string{"Action", "Fantasy", "favorites", "classic"},
				PublishedYear:    2011,
				Language:         "en",
				ReadingDirection: "rtl",
				AgeRating:        "Mature 17+",
			},
		},
		{
			name:  "fractional number",
			files: mergeMaps(page, map[string]string{"ComicInfo.xml": `<ComicInfo><Number>5.5</Number></ComicInfo>`}),
			want:  comicInfo{SeriesIndex: 5.5, ReadingDirection: "ltr"},
		},
		{
			name:  "suffixed number",
			files: mergeMaps(page, map[string]string{"ComicInfo.xml": `<ComicInfo><Number>5a</Number></ComicInfo>`}),
			want:  comicInfo{SeriesIndex: 5, ReadingDirection: "ltr"},
		},
		{
			name:  "non-numeric number",
			files: mergeMaps(page, map[string]string{"ComicInfo.xml": `<ComicInfo><Number>abc</Number></ComicInfo>`}),
			want:  comicInfo{ReadingDirection: "ltr"},
		},
		{
			name:  "manga yes",
			files: mergeMaps(page, map[string]string{"ComicInfo.xml": `<ComicInfo><Manga>Yes</Manga></ComicInfo>`}),
			want:  comicInfo{ReadingDirection: "rtl"},
		},
		{
			name:  "manga no",
			files: mergeMaps(page, map[string]string{"ComicInfo.xml": `<ComicInfo><Manga>No</Manga></ComicInfo>`}),
			want:  comicInfo{ReadingDirection: "ltr"},
		},
		{
			name:  "manga unknown",
			files: mergeMaps(page, map[string]string{"ComicInfo.xml": `<ComicInfo><Manga>Unknown</Manga></ComicInfo>`}),
			want:  comicInfo{ReadingDirection: "ltr"},
		},
		{
			name:  "malformed xml yields zero info",
			files: mergeMaps(page, map[string]string{"ComicInfo.xml": `<ComicInfo><Title>oops`}),
			want:  comicInfo{},
		},
		{
			name:  "missing file yields zero info",
			files: page,
			want:  comicInfo{},
		},
		{
			name:  "lowercase filename",
			files: mergeMaps(page, map[string]string{"comicinfo.xml": `<ComicInfo><Title>lower</Title></ComicInfo>`}),
			want:  comicInfo{Title: "lower", ReadingDirection: "ltr"},
		},
		{
			name:  "nested member is ignored",
			files: mergeMaps(page, map[string]string{"meta/ComicInfo.xml": `<ComicInfo><Title>nested</Title></ComicInfo>`}),
			want:  comicInfo{},
		},
		{
			name:  "empty elements stay zero",
			files: mergeMaps(page, map[string]string{"ComicInfo.xml": `<ComicInfo><Title></Title><Year>junk</Year></ComicInfo>`}),
			want:  comicInfo{ReadingDirection: "ltr"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "test.cbz")
			writeCBZ(t, path, tc.files)
			meta := parseCBZ(path)
			if !reflect.DeepEqual(meta.Info, tc.want) {
				t.Errorf("info = %+v, want %+v", meta.Info, tc.want)
			}
			if meta.PageCount != 1 {
				t.Errorf("page count = %d, want 1", meta.PageCount)
			}
		})
	}
}

func mergeMaps(a, b map[string]string) map[string]string {
	out := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func TestScannerAppliesComicInfoOnce(t *testing.T) {
	ctx := context.Background()
	libDir := t.TempDir()
	coverDir := filepath.Join(t.TempDir(), "covers")

	cbzPath := filepath.Join(libDir, "hero.cbz")
	writeCBZ(t, cbzPath, map[string]string{
		"page1.jpg":     "fakejpeg",
		"ComicInfo.xml": fullComicInfoXML,
	})

	store, err := storage.Open(filepath.Join(t.TempDir(), "s.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.EnsureDefaultLibrary(ctx, libDir); err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	sc := New(store, coverDir, t.TempDir(), log, 2)

	if err := sc.Scan(ctx); err != nil {
		t.Fatalf("scan: %v", err)
	}

	book, err := store.GetBookByPath(ctx, 1, "hero.cbz")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if book.Title != "The Big Showdown" {
		t.Errorf("title = %q, want %q", book.Title, "The Big Showdown")
	}
	if book.Author != "Jane Creator" {
		t.Errorf("author = %q, want %q", book.Author, "Jane Creator")
	}
	if book.Series != "Hero Tales" || book.SeriesIndex != 7 {
		t.Errorf("series = %q #%v, want Hero Tales #7", book.Series, book.SeriesIndex)
	}
	if book.Publisher != "Indie Press" {
		t.Errorf("publisher = %q, want %q", book.Publisher, "Indie Press")
	}
	if book.ReadingDirection != models.ReadingDirectionRTL {
		t.Errorf("reading direction = %q, want rtl", book.ReadingDirection)
	}
	if book.AgeRating != "Mature 17+" {
		t.Errorf("age rating = %q, want %q", book.AgeRating, "Mature 17+")
	}
	if book.PublishedYear != 2011 || book.Language != "en" || book.Description != "A hero does things." {
		t.Errorf("year/lang/desc = %d/%q/%q", book.PublishedYear, book.Language, book.Description)
	}

	wantTags := []string{"Action", "classic", "Fantasy", "favorites"}
	gotTags, err := store.ListBookTags(ctx, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotTags, wantTags) {
		t.Fatalf("tags = %v, want %v", gotTags, wantTags)
	}

	// User replaces tags, then the archive changes and is rescanned.
	if _, err := store.SetBookTags(ctx, book.ID, []string{"mine"}); err != nil {
		t.Fatal(err)
	}
	writeCBZ(t, cbzPath, map[string]string{
		"page1.jpg": "fakejpeg2",
		"page2.jpg": "fakejpeg3",
		"ComicInfo.xml": `<?xml version="1.0"?>
<ComicInfo>
	<Title>Volume Two</Title>
	<Genre>Action, Fantasy</Genre>
</ComicInfo>`,
	})
	if err := sc.Scan(ctx); err != nil {
		t.Fatalf("rescan: %v", err)
	}
	book, err = store.GetBookByPath(ctx, 1, "hero.cbz")
	if err != nil {
		t.Fatal(err)
	}
	if book.Title != "Volume Two" {
		t.Fatalf("title after rescan = %q, want %q (file was not reprocessed)", book.Title, "Volume Two")
	}
	gotTags, err = store.ListBookTags(ctx, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotTags, []string{"mine"}) {
		t.Errorf("tags after rescan = %v, want [mine]", gotTags)
	}
}
