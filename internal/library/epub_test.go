package library

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

// writeTestEPUB builds a minimal but valid EPUB on disk and returns its path.
func writeTestEPUB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "book.epub")

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)

	files := map[string]string{
		"META-INF/container.xml": `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`,
		"OEBPS/content.opf": `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Driven Tales</dc:title>
    <dc:creator>Ada Lovelace</dc:creator>
    <dc:language>en</dc:language>
    <dc:description>A short test book.</dc:description>
    <meta name="calibre:series" content="The Series"/>
    <meta name="calibre:series_index" content="3"/>
  </metadata>
  <manifest>
    <item id="cover" href="cover.png" media-type="image/png" properties="cover-image"/>
  </manifest>
</package>`,
		"OEBPS/cover.png": "PNGDATA",
	}

	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create %s: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("zip write %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return path
}

func TestParseEPUB(t *testing.T) {
	path := writeTestEPUB(t)

	meta, err := parseEPUB(path)
	if err != nil {
		t.Fatalf("parseEPUB: %v", err)
	}

	if meta.Title != "Test Driven Tales" {
		t.Errorf("title = %q, want %q", meta.Title, "Test Driven Tales")
	}
	if meta.Author != "Ada Lovelace" {
		t.Errorf("author = %q, want %q", meta.Author, "Ada Lovelace")
	}
	if meta.Language != "en" {
		t.Errorf("language = %q, want %q", meta.Language, "en")
	}
	if meta.Series != "The Series" {
		t.Errorf("series = %q, want %q", meta.Series, "The Series")
	}
	if meta.SeriesIndex != 3 {
		t.Errorf("seriesIndex = %v, want 3", meta.SeriesIndex)
	}
	if string(meta.CoverData) != "PNGDATA" {
		t.Errorf("cover data = %q, want %q", string(meta.CoverData), "PNGDATA")
	}
}

func TestParseEPUBInvalid(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.epub")
	if err := os.WriteFile(bad, []byte("not a zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := parseEPUB(bad); err == nil {
		t.Fatal("expected error for invalid epub, got nil")
	}
}

func writeTestEPUBXML11(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "book-xml11.epub")

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)

	files := map[string]string{
		"META-INF/container.xml": `<?xml version="1.1" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`,
		"OEBPS/content.opf": `<?xml version="1.1" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>XML 1.1 Book</dc:title>
    <dc:creator>Test Author</dc:creator>
    <dc:language>en</dc:language>
  </metadata>
  <manifest>
    <item id="cover" href="cover.png" media-type="image/png" properties="cover-image"/>
  </manifest>
</package>`,
		"OEBPS/cover.png": "PNGDATA",
	}

	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create %s: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("zip write %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return path
}

func TestParseEPUBXML11(t *testing.T) {
	path := writeTestEPUBXML11(t)

	meta, err := parseEPUB(path)
	if err != nil {
		t.Fatalf("parseEPUB: %v", err)
	}
	if meta.Title != "XML 1.1 Book" {
		t.Errorf("title = %q, want %q", meta.Title, "XML 1.1 Book")
	}
	if meta.Author != "Test Author" {
		t.Errorf("author = %q, want %q", meta.Author, "Test Author")
	}
}

func TestTitleFromFilename(t *testing.T) {
	cases := map[string]string{
		"the_great_gatsby": "the great gatsby",
		"clean.code.draft": "clean code draft",
		"Already Readable": "Already Readable",
	}
	for in, want := range cases {
		if got := titleFromFilename(in); got != want {
			t.Errorf("titleFromFilename(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFormatFromExt(t *testing.T) {
	cases := map[string]string{
		"a.epub":      "epub",
		"b.EPUB":      "epub",
		"c.pdf":       "pdf",
		"d.txt":       "",
		"noext":       "",
		"path/to.pdf": "pdf",
	}
	for in, want := range cases {
		if got := formatFromExt(in); got != want {
			t.Errorf("formatFromExt(%q) = %q, want %q", in, got, want)
		}
	}
}
