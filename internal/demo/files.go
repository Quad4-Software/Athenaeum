package demo

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func writeMinimalEPUB(path, title, author string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	f, err := os.Create(path) // #nosec G304 -- path under demo library root
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	mw, err := zw.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	})
	if err != nil {
		return err
	}
	if _, err := mw.Write([]byte("application/epub+zip")); err != nil {
		return err
	}

	cw, err := zw.Create("META-INF/container.xml")
	if err != nil {
		return err
	}
	_, err = cw.Write([]byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>
`))
	if err != nil {
		return err
	}

	ow, err := zw.Create("OEBPS/content.opf")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(ow, `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">demo-%s</dc:identifier>
    <dc:title>%s</dc:title>
    <dc:creator>%s</dc:creator>
    <dc:language>en</dc:language>
  </metadata>
  <manifest>
    <item id="chapter" href="chapter.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="chapter"/>
  </spine>
</package>
`, filepath.Base(path), xmlEscape(title), xmlEscape(author))
	if err != nil {
		return err
	}

	xw, err := zw.Create("OEBPS/chapter.xhtml")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(xw, `<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml">
  <head><title>%s</title></head>
  <body>
    <h1>%s</h1>
    <p>Generated demo chapter for %s by %s.</p>
    <p>This file exists so the library scanner and reader have something to open.</p>
  </body>
</html>
`, xmlEscape(title), xmlEscape(title), xmlEscape(title), xmlEscape(author))
	return err
}

func writeMinimalPDF(path, title, author string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	body := fmt.Sprintf(`%%PDF-1.4
1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj
2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj
3 0 obj<</Type/Page/MediaBox[0 0 612 792]/Parent 2 0 R>>endobj
trailer<</Root 1 0 R/Info<</Title (%s)/Author (%s)>>>>
%%%%EOF
`, pdfEscape(title), pdfEscape(author))
	return os.WriteFile(path, []byte(body), 0o640) // #nosec G306 -- demo PDF under library tree
}

func writeMinimalCBZ(path, title string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	f, err := os.Create(path) // #nosec G304 -- path under demo library root
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()
	png, err := EncodeCoverBuffer(title, "demo")
	if err != nil {
		return err
	}
	w, err := zw.Create("001.png")
	if err != nil {
		return err
	}
	_, err = w.Write(png)
	return err
}

// Tiny silent-ish MPEG frame so audio formats have a non-empty file.
var tinyMP3 = []byte{
	0xFF, 0xFB, 0x90, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

func writeMinimalAudio(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	payload := bytesRepeat(tinyMP3, 64)
	return os.WriteFile(path, payload, 0o640) // #nosec G306 -- demo audio under library tree
}

func writeAudiobookSet(dir, title string) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	for i, name := range []string{"01 - Opening.mp3", "02 - Kiln.mp3", "03 - Closing.mp3"} {
		_ = i
		if err := writeMinimalAudio(filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	meta := fmt.Sprintf("title=%s\n", title)
	return os.WriteFile(filepath.Join(dir, "metadata.txt"), []byte(meta), 0o640) // #nosec G306 -- demo metadata under library tree
}

func writeMinimalMOBI(path, title, author string) error {
	// MOBI is complex. Write a stub the indexer can still register by extension.
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	body := fmt.Sprintf("DEMO-MOBI\nTitle: %s\nAuthor: %s\n", title, author)
	return os.WriteFile(path, []byte(body), 0o640) // #nosec G306 -- demo MOBI stub under library tree
}

func writeMediaFile(libraryDir string, e Entry) (relPath, absPath string, err error) {
	switch e.Format {
	case "epub":
		relPath = filepath.Join("demo", e.Slug+".epub")
		absPath = filepath.Join(libraryDir, relPath)
		err = writeMinimalEPUB(absPath, e.Title, e.Author)
	case "pdf":
		relPath = filepath.Join("demo", e.Slug+".pdf")
		absPath = filepath.Join(libraryDir, relPath)
		err = writeMinimalPDF(absPath, e.Title, e.Author)
	case "cbz":
		relPath = filepath.Join("demo", e.Slug+".cbz")
		absPath = filepath.Join(libraryDir, relPath)
		err = writeMinimalCBZ(absPath, e.Title)
	case "mobi":
		relPath = filepath.Join("demo", e.Slug+".mobi")
		absPath = filepath.Join(libraryDir, relPath)
		err = writeMinimalMOBI(absPath, e.Title, e.Author)
	case "mp3", "m4b":
		relPath = filepath.Join("demo", e.Slug+"."+e.Format)
		absPath = filepath.Join(libraryDir, relPath)
		err = writeMinimalAudio(absPath)
	case "audiobook":
		relPath = filepath.Join("demo", e.Slug)
		absPath = filepath.Join(libraryDir, relPath)
		err = writeAudiobookSet(absPath, e.Title)
	default:
		relPath = filepath.Join("demo", e.Slug+".epub")
		absPath = filepath.Join(libraryDir, relPath)
		err = writeMinimalEPUB(absPath, e.Title, e.Author)
	}
	return relPath, absPath, err
}

func xmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return r.Replace(s)
}

func pdfEscape(s string) string {
	r := strings.NewReplacer("(", "\\(", ")", "\\)", "\\", "\\\\")
	return r.Replace(s)
}

func bytesRepeat(b []byte, n int) []byte {
	out := make([]byte, 0, len(b)*n)
	for range n {
		out = append(out, b...)
	}
	return out
}
