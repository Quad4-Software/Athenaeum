package library

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func writeSimpleEPUB(destPath, title, author, bodyHTML string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o750); err != nil {
		return err
	}
	f, err := os.Create(destPath) // #nosec G304
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	now := time.Now()
	files := map[string]string{
		"mimetype":               "application/epub+zip",
		"META-INF/container.xml": containerXML,
		"OEBPS/content.opf":      fmt.Sprintf(opfTemplate, xmlEsc(title), xmlEsc(author), now.Format(time.RFC3339)),
		"OEBPS/nav.xhtml":        fmt.Sprintf(navTemplate, xmlEsc(title)),
		"OEBPS/chapter.xhtml":    fmt.Sprintf(xhtmlTemplate, xmlEsc(title), bodyHTML),
		"OEBPS/toc.ncx":          fmt.Sprintf(ncxTemplate, xmlEsc(title)),
	}
	for name, content := range files {
		method := zip.Deflate
		if name == "mimetype" {
			method = zip.Store
		}
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:   name,
			Method: method,
		})
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(content)); err != nil {
			return err
		}
	}
	return zw.Close()
}

func xmlEsc(s string) string {
	repl := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return repl.Replace(s)
}

const containerXML = `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`

const opfTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>%s</dc:title>
    <dc:creator>%s</dc:creator>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">%s</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="chapter" href="chapter.xhtml" media-type="application/xhtml+xml"/>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
  </manifest>
  <spine>
    <itemref idref="chapter"/>
  </spine>
</package>`

const navTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head><title>%s</title></head>
<body><nav epub:type="toc"><ol><li><a href="chapter.xhtml">Start</a></li></ol></nav></body>
</html>`

const xhtmlTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>%s</title></head>
<body>%s</body>
</html>`

const ncxTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1">
<head><meta name="dtb:uid" content="athenaeum-convert"/></head>
<docTitle><text>%s</text></docTitle>
<navMap><navPoint id="n1"><navLabel><text>Start</text></navLabel><content src="chapter.xhtml"/></navPoint></navMap>
</ncx>`

// ConvertWithCalibre runs ebook-convert when calibre is installed.
func convertWithCalibre(src, dest, target string) error {
	return runCalibreConvert(src, dest, target)
}

func runCalibreConvert(src, dest, target string) error {
	// implemented in convert.go
	return calibreConvert(src, dest, target)
}

// IsCalibreAvailable reports whether ebook-convert is on PATH.
func IsCalibreAvailable() bool {
	return calibreAvailable()
}

// MinimalEPUBFromHTML builds EPUB bytes in memory.
func MinimalEPUBFromHTML(title, author, body string) ([]byte, error) {
	tmp := filepath.Join(os.TempDir(), "athenaeum-epub-build")
	dest := filepath.Join(tmp, "out.epub")
	_ = os.RemoveAll(tmp)
	if err := os.MkdirAll(tmp, 0o700); err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	if err := writeSimpleEPUB(dest, title, author, body); err != nil {
		return nil, err
	}
	return os.ReadFile(dest) // #nosec G304
}
