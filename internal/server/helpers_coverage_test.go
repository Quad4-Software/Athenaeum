package server

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"athenaeum/internal/models"
)

func TestContentType(t *testing.T) {
	cases := map[string]string{
		models.FormatEPUB: "application/epub+zip",
		models.FormatPDF:  "application/pdf",
		models.FormatMP3:  "audio/mpeg",
		models.FormatM4B:  "audio/mp4",
		models.FormatM4A:  "audio/mp4",
		models.FormatOGG:  "audio/ogg",
		models.FormatFLAC: "audio/flac",
		models.FormatMOBI: "application/x-mobipocket-ebook",
		models.FormatAZW:  "application/x-mobipocket-ebook",
		models.FormatAZW3: "application/x-mobipocket-ebook",
		models.FormatKFX:  "application/vnd.amazon.ebook",
		models.FormatCBZ:  "application/vnd.comicbook+zip",
		models.FormatCBR:  "application/vnd.comicbook-rar",
		"unknown":         "application/octet-stream",
	}
	for in, want := range cases {
		if got := contentType(in); got != want {
			t.Errorf("contentType(%q)=%q want %q", in, got, want)
		}
	}
}

func TestSafeFilenameAndAtoiDefault(t *testing.T) {
	if got := safeFilename(models.Book{Title: "Dune", Format: "epub"}); got != "Dune.epub" {
		t.Fatalf("got %q", got)
	}
	if got := safeFilename(models.Book{Format: "pdf"}); got != "book.pdf" {
		t.Fatalf("empty title got %q", got)
	}
	if atoiDefault("", 7) != 7 || atoiDefault("3", 7) != 3 || atoiDefault("x", 7) != 7 {
		t.Fatal("atoiDefault")
	}
}

func TestRequestBaseURL(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "http://example.test/x", nil)
	req.Host = "example.test"
	if got := s.requestBaseURL(req); got != "http://example.test" {
		t.Fatalf("got %q", got)
	}

	req.TLS = &tls.ConnectionState{}
	if got := s.requestBaseURL(req); got != "https://example.test" {
		t.Fatalf("tls got %q", got)
	}

	req = httptest.NewRequest(http.MethodGet, "http://internal/x", nil)
	req.Host = "internal"
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "pub.example")
	s.proxies.set("127.0.0.1")
	if got := s.requestBaseURL(req); got != "https://pub.example" {
		t.Fatalf("forwarded got %q", got)
	}
}
