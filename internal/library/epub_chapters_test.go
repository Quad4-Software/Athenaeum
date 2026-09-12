package library

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func writeNarrationEPUB(t *testing.T, files map[string]string) string {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "book.epub")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

const testContainerXML = `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles><rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles>
</container>`

const testOPF = `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="c1" href="ch1.xhtml" media-type="application/xhtml+xml"/>
    <item id="c2" href="ch2.xhtml" media-type="application/xhtml+xml"/>
    <item id="img" href="cover.jpg" media-type="image/jpeg"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
    <itemref idref="c1"/>
    <itemref idref="c2"/>
  </spine>
</package>`

const testNav = `<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops"><body>
<nav epub:type="toc"><ol>
<li><a href="ch1.xhtml">Chapter One</a></li>
<li><a href="ch2.xhtml">Chapter Two</a></li>
</ol></nav></body></html>`

const testCh1 = `<html><body>
<h1>One</h1>
<p>  It was a   bright day. </p>
<p>Second paragraph.</p>
<script>var x = 1;</script>
</body></html>`

const testCh2 = `<html><body>
<p>More text here.</p>
</body></html>`

func TestExtractEPUBChapters(t *testing.T) {
	path := writeNarrationEPUB(t, map[string]string{
		"META-INF/container.xml": testContainerXML,
		"OEBPS/content.opf":      testOPF,
		"OEBPS/nav.xhtml":        testNav,
		"OEBPS/ch1.xhtml":        testCh1,
		"OEBPS/ch2.xhtml":        testCh2,
		"OEBPS/cover.jpg":        "fake",
	})

	chapters, err := ExtractEPUBChapters(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d: %#v", len(chapters), chapters)
	}
	if chapters[0].Title != "Chapter One" {
		t.Fatalf("expected nav title, got %q", chapters[0].Title)
	}
	if chapters[1].Title != "Chapter Two" {
		t.Fatalf("expected nav title, got %q", chapters[1].Title)
	}
	want := []string{"One", "It was a bright day.", "Second paragraph."}
	if len(chapters[0].Paragraphs) != len(want) {
		t.Fatalf("paragraphs: %#v", chapters[0].Paragraphs)
	}
	for i, p := range want {
		if chapters[0].Paragraphs[i] != p {
			t.Fatalf("paragraph %d = %q, want %q", i, chapters[0].Paragraphs[i], p)
		}
	}
	if chapters[1].Paragraphs[0] != "More text here." {
		t.Fatalf("unexpected ch2: %#v", chapters[1].Paragraphs)
	}
}

func TestExtractEPUBChaptersFallbackTitle(t *testing.T) {
	// Without a nav doc, chapter titles fall back to the first heading.
	opf := `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0">
  <manifest>
    <item id="c1" href="text/c1.html" media-type="application/xhtml+xml"/>
  </manifest>
  <spine><itemref idref="c1"/></spine>
</package>`
	container := `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles><rootfile full-path="content.opf" media-type="application/oebps-package+xml"/></rootfiles>
</container>`
	path := writeNarrationEPUB(t, map[string]string{
		"META-INF/container.xml": container,
		"content.opf":            opf,
		"text/c1.html":           `<html><body><h2>Prologue</h2><p>Text.</p></body></html>`,
	})
	chapters, err := ExtractEPUBChapters(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(chapters) != 1 || chapters[0].Title != "Prologue" {
		t.Fatalf("expected heading fallback title, got %#v", chapters)
	}
}
