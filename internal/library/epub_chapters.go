package library

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"path"
	"strings"

	"golang.org/x/net/html"
)

// EPUBChapter is one spine document's narration-ready text.
type EPUBChapter struct {
	Title      string
	Paragraphs []string
}

// epubBlockTags mirrors the reader's narration selector: block-level
// elements whose text is spoken as separate utterances.
var epubBlockTags = map[string]bool{
	"p": true, "li": true, "h1": true, "h2": true, "h3": true, "h4": true,
	"h5": true, "h6": true, "blockquote": true, "dd": true, "dt": true,
	"figcaption": true, "td": true, "pre": true,
}

var epubSkipTags = map[string]bool{
	"script": true, "style": true, "nav": true, "svg": true,
	"head": true, "title": true, "audio": true, "video": true,
}

// ExtractEPUBChapters walks an EPUB's spine in reading order and returns
// chapter titles plus normalized paragraph text for each document that
// contains speakable content.
func ExtractEPUBChapters(filePath string) ([]EPUBChapter, error) {
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	files := make(map[string]*zip.File, len(zr.File))
	for _, f := range zr.File {
		files[f.Name] = f
	}

	opfPath, err := opfPathFrom(files)
	if err != nil {
		return nil, err
	}
	pkg, err := readOPF(files[opfPath])
	if err != nil {
		return nil, err
	}
	base := path.Dir(opfPath)

	byID := make(map[string]struct {
		href, mediaType, props string
	}, len(pkg.Manifest.Items))
	for _, it := range pkg.Manifest.Items {
		byID[it.ID] = struct {
			href, mediaType, props string
		}{it.Href, it.MediaType, it.Properties}
	}

	titles := epubTOCTitles(files, pkg, base)

	var chapters []EPUBChapter
	for _, ref := range pkg.Spine.ItemRefs {
		if ref.Linear == "no" {
			continue
		}
		item, ok := byID[ref.IDRef]
		if !ok || !isHTMLMedia(item.mediaType) {
			continue
		}
		if strings.Contains(item.props, "nav") {
			continue
		}
		full := path.Clean(path.Join(base, item.href))
		zf, ok := files[full]
		if !ok {
			continue
		}
		raw, err := readZipFile(zf)
		if err != nil {
			continue
		}
		paras, heading := htmlParagraphs(raw)
		if len(paras) == 0 {
			continue
		}
		title := titles[stripFragment(full)]
		if title == "" {
			title = heading
		}
		chapters = append(chapters, EPUBChapter{Title: title, Paragraphs: paras})
	}
	return chapters, nil
}

func isHTMLMedia(mt string) bool {
	return mt == "application/xhtml+xml" || mt == "text/html"
}

func stripFragment(p string) string {
	if i := strings.IndexByte(p, '#'); i >= 0 {
		return p[:i]
	}
	return p
}

// epubTOCTitles maps normalized content paths to their TOC label, from
// the EPUB2 NCX or the EPUB3 nav document.
func epubTOCTitles(files map[string]*zip.File, pkg opfPackage, base string) map[string]string {
	out := map[string]string{}
	for _, it := range pkg.Manifest.Items {
		full := path.Clean(path.Join(base, it.Href))
		switch {
		case it.MediaType == "application/x-dtbncx+xml":
			if zf, ok := files[full]; ok {
				if raw, err := readZipFile(zf); err == nil {
					mergeTitles(out, ncxTitles(raw, base))
				}
			}
		case strings.Contains(it.Properties, "nav"):
			if zf, ok := files[full]; ok {
				if raw, err := readZipFile(zf); err == nil {
					mergeTitles(out, navDocTitles(raw, base))
				}
			}
		}
	}
	return out
}

func mergeTitles(dst, src map[string]string) {
	for k, v := range src {
		if _, ok := dst[k]; !ok {
			dst[k] = v
		}
	}
}

// ncxTitles parses an NCX navMap into src->label pairs. Keys are joined
// against the NCX's own directory so they match normalized content paths.
func ncxTitles(raw []byte, base string) map[string]string {
	type navPoint struct {
		Label   string `xml:"navLabel>text"`
		Content struct {
			Src string `xml:"src,attr"`
		} `xml:"content"`
		Children []navPoint `xml:"navPoint"`
	}
	var navMap struct {
		Points []navPoint `xml:"navPoint"`
	}
	var doc struct {
		NavMap struct {
			Points []navPoint `xml:"navPoint"`
		} `xml:"navMap"`
	}
	if err := xml.Unmarshal(raw, &doc); err != nil {
		return nil
	}
	navMap.Points = doc.NavMap.Points

	out := map[string]string{}
	var walk func(pts []navPoint)
	walk = func(pts []navPoint) {
		for _, p := range pts {
			src := p.Content.Src
			if src != "" && p.Label != "" {
				full := path.Clean(path.Join(base, stripFragment(src)))
				if _, exists := out[full]; !exists {
					out[full] = strings.TrimSpace(p.Label)
				}
			}
			walk(p.Children)
		}
	}
	walk(navMap.Points)
	return out
}

// navDocTitles parses an EPUB3 nav document's toc links into
// href->label pairs.
func navDocTitles(raw []byte, base string) map[string]string {
	root, err := html.Parse(bytes.NewReader(raw))
	if err != nil {
		return nil
	}
	out := map[string]string{}
	var walk func(n *html.Node, inTOC bool)
	walk = func(n *html.Node, inTOC bool) {
		if n.Type == html.ElementNode {
			if n.Data == "nav" || n.Data == "section" || n.Data == "div" {
				if attr(n, "epub:type") == "toc" || attr(n, "type") == "toc" || attr(n, "role") == "doc-toc" {
					inTOC = true
				}
			}
			if inTOC && n.Data == "a" {
				href := attr(n, "href")
				label := strings.TrimSpace(textContent(n))
				if href != "" && label != "" {
					full := path.Clean(path.Join(base, stripFragment(href)))
					if _, exists := out[full]; !exists {
						out[full] = label
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, inTOC)
		}
	}
	walk(root, false)
	return out
}

// htmlParagraphs extracts normalized text from block-level elements.
// Returns the paragraphs and the first heading text, if any.
func htmlParagraphs(raw []byte) ([]string, string) {
	root, err := html.Parse(bytes.NewReader(raw))
	if err != nil {
		return nil, ""
	}
	var paras []string
	var heading string
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if epubSkipTags[n.Data] {
				return
			}
			if epubBlockTags[n.Data] {
				t := normalizeSpace(textContent(n))
				if t != "" {
					paras = append(paras, t)
					if heading == "" && strings.HasPrefix(n.Data, "h") && len(n.Data) == 2 {
						heading = t
					}
				}
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return paras, heading
}

func attr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name || strings.HasSuffix(a.Key, ":"+name) {
			return a.Val
		}
	}
	return ""
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(c *html.Node) {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		}
		for k := c.FirstChild; k != nil; k = k.NextSibling {
			walk(k)
		}
	}
	walk(n)
	return b.String()
}

func normalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
