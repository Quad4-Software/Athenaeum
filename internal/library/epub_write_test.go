package library

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestXMLEsc(t *testing.T) {
	got := xmlEsc(`a&b<c>d"e`)
	want := `a&amp;b&lt;c&gt;d&quot;e`
	if got != want {
		t.Fatalf("xmlEsc = %q, want %q", got, want)
	}
}

func TestWriteSimpleEPUB(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out.epub")
	if err := writeSimpleEPUB(dest, `T & A`, "Author", "<p>hi</p>"); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(dest)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Size() < 100 {
		t.Fatalf("epub too small: %d", fi.Size())
	}
	zr, err := zip.OpenReader(dest)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	for _, need := range []string{"mimetype", "META-INF/container.xml", "OEBPS/content.opf", "OEBPS/chapter.xhtml"} {
		if !names[need] {
			t.Errorf("missing %s", need)
		}
	}
}

func TestMinimalEPUBFromHTML(t *testing.T) {
	data, err := MinimalEPUBFromHTML("Title", "Author", "<p>body</p>")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 100 {
		t.Fatalf("len=%d", len(data))
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range zr.File {
		if f.Name == "mimetype" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("missing mimetype")
	}
}

func TestIsCalibreAvailable(t *testing.T) {
	_ = IsCalibreAvailable()
}
