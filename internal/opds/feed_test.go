package opds

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"athenaeum/internal/models"
)

func TestWriteRootCatalog(t *testing.T) {
	var buf bytes.Buffer
	w := FeedWriter{BaseURL: "http://localhost:8080", Title: "Athenaeum"}
	if err := w.WriteRootCatalog(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"<feed", "Recent additions", "opds-spec.org", "/opds/recent"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestWriteAcquisitionFeed(t *testing.T) {
	var buf bytes.Buffer
	w := FeedWriter{BaseURL: "http://localhost:8080", Title: "Recent"}
	books := []models.Book{{
		ID: 3, Title: "Dune", Author: "Herbert", Format: models.FormatEPUB,
		ModifiedAt: time.Now(), HasCover: true, Description: "Sand",
	}}
	progress := map[int64]models.Progress{3: {Percent: 0.42}}
	if err := w.WriteAcquisitionFeed(&buf, books, "/opds/recent", progress); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Dune") || !strings.Contains(out, "/api/books/3/download") {
		t.Fatalf("unexpected feed:\n%s", out)
	}
	if !strings.Contains(out, "42% read") {
		t.Fatalf("missing progress summary:\n%s", out)
	}
}

func TestWriteSeriesNavigation(t *testing.T) {
	var buf bytes.Buffer
	w := FeedWriter{BaseURL: "http://localhost:8080", Title: "Athenaeum"}
	series := []models.SeriesInfo{
		{Name: "Dune Saga", Count: 3},
		{Name: "Space / Opera", Count: 1},
	}
	if err := w.WriteSeriesNavigation(&buf, series); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Dune Saga") || !strings.Contains(out, "3 books") {
		t.Fatalf("series feed:\n%s", out)
	}
	if !strings.Contains(out, urlPathEscape("Space / Opera")) {
		t.Fatalf("escaped series path missing:\n%s", out)
	}
}

func TestURLPathEscape(t *testing.T) {
	if got := urlPathEscape("a/b c"); got != "a%2Fb%20c" {
		t.Fatalf("got %q", got)
	}
}

func TestSetXMLHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	SetXMLHeaders(rec)
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "atom+xml") {
		t.Fatalf("content-type %q", ct)
	}
	if rec.Header().Get("Cache-Control") != "no-cache" {
		t.Fatal("cache-control")
	}
}

func TestMimeForFormat(t *testing.T) {
	cases := map[string]string{
		models.FormatEPUB:      "application/epub+zip",
		models.FormatPDF:       "application/pdf",
		models.FormatMP3:       "audio/mpeg",
		models.FormatM4B:       "audio/mp4",
		models.FormatM4A:       "audio/mp4",
		models.FormatOGG:       "audio/ogg",
		models.FormatFLAC:      "audio/flac",
		models.FormatMOBI:      "application/x-mobipocket-ebook",
		models.FormatAZW:       "application/x-mobipocket-ebook",
		models.FormatAZW3:      "application/x-mobipocket-ebook",
		models.FormatCBZ:       "application/vnd.comicbook+zip",
		models.FormatCBR:       "application/vnd.comicbook-rar",
		models.FormatAudiobook: "audio/mpeg",
		"unknown":              "application/octet-stream",
	}
	for format, want := range cases {
		if got := mimeForFormat(format); got != want {
			t.Fatalf("mimeForFormat(%q)=%q want %q", format, got, want)
		}
	}
}
