package opds

import (
	"bytes"
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
		ModifiedAt: time.Now(), HasCover: true,
	}}
	if err := w.WriteAcquisitionFeed(&buf, books, "/opds/recent", nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Dune") || !strings.Contains(out, "/api/books/3/download") {
		t.Fatalf("unexpected feed:\n%s", out)
	}
}

func TestMimeForFormat(t *testing.T) {
	if mimeForFormat(models.FormatMP3) != "audio/mpeg" {
		t.Fatal("mp3 mime")
	}
}
