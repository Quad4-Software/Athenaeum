// Package opds generates OPDS 1.2 Atom catalog feeds for e-reader clients.
package opds

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"athenaeum/internal/models"
)

const (
	nsAtom = "http://www.w3.org/2005/Atom"
	nsOPDS = "http://opds-spec.org/2010/catalog"
)

// FeedWriter renders OPDS Atom feeds.
type FeedWriter struct {
	BaseURL string
	Title   string
}

type feed struct {
	XMLName xml.Name `xml:"feed"`
	Xmlns   string   `xml:"xmlns,attr"`
	OPDS    string   `xml:"xmlns:opds,attr"`
	ID      string   `xml:"id"`
	Title   string   `xml:"title"`
	Updated string   `xml:"updated"`
	Links   []link   `xml:"link"`
	Entries []entry  `xml:"entry"`
}

type link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr,omitempty"`
}

type entry struct {
	ID      string  `xml:"id"`
	Title   string  `xml:"title"`
	Updated string  `xml:"updated"`
	Author  *author `xml:"author,omitempty"`
	Summary string  `xml:"summary,omitempty"`
	Links   []link  `xml:"link"`
}

type author struct {
	Name string `xml:"name"`
}

// WriteRootCatalog writes the top-level OPDS navigation feed.
func (w FeedWriter) WriteRootCatalog(out io.Writer) error {
	now := time.Now().UTC().Format(time.RFC3339)
	f := feed{
		Xmlns:   nsAtom,
		OPDS:    nsOPDS,
		ID:      w.BaseURL + "/opds/",
		Title:   w.Title,
		Updated: now,
		Links: []link{
			{Rel: "self", Href: w.BaseURL + "/opds/", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "start", Href: w.BaseURL + "/opds/", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
		Entries: []entry{
			{
				ID:      w.BaseURL + "/opds/recent",
				Title:   "Recent additions",
				Updated: now,
				Links: []link{
					{Rel: "subsection", Href: w.BaseURL + "/opds/recent", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
				},
			},
			{
				ID:      w.BaseURL + "/opds/series",
				Title:   "Series",
				Updated: now,
				Links: []link{
					{Rel: "subsection", Href: w.BaseURL + "/opds/series", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
				},
			},
			{
				ID:      w.BaseURL + "/opds/comics",
				Title:   "Comics",
				Updated: now,
				Links: []link{
					{Rel: "subsection", Href: w.BaseURL + "/opds/comics", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
				},
			},
			{
				ID:      w.BaseURL + "/opds/kindle",
				Title:   "Kindle books",
				Updated: now,
				Links: []link{
					{Rel: "subsection", Href: w.BaseURL + "/opds/kindle", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
				},
			},
			{
				Title:   "Search",
				Updated: now,
				Links: []link{
					{Rel: "search", Href: w.BaseURL + "/opds/search?q={searchTerms}", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
				},
			},
		},
	}
	enc := xml.NewEncoder(out)
	enc.Indent("", "  ")
	if _, err := out.Write([]byte(xml.Header)); err != nil {
		return err
	}
	return enc.Encode(f)
}

// WriteSeriesNavigation writes a feed listing all series.
func (w FeedWriter) WriteSeriesNavigation(out io.Writer, series []models.SeriesInfo) error {
	now := time.Now().UTC().Format(time.RFC3339)
	f := feed{
		Xmlns:   nsAtom,
		OPDS:    nsOPDS,
		ID:      w.BaseURL + "/opds/series",
		Title:   w.Title + " — Series",
		Updated: now,
		Links: []link{
			{Rel: "self", Href: w.BaseURL + "/opds/series", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "start", Href: w.BaseURL + "/opds/", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}
	for _, s := range series {
		href := w.BaseURL + "/opds/series/" + urlPathEscape(s.Name)
		f.Entries = append(f.Entries, entry{
			ID:      href,
			Title:   s.Name,
			Updated: now,
			Summary: fmt.Sprintf("%d books", s.Count),
			Links: []link{
				{Rel: "subsection", Href: href, Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
			},
		})
	}
	enc := xml.NewEncoder(out)
	enc.Indent("", "  ")
	if _, err := out.Write([]byte(xml.Header)); err != nil {
		return err
	}
	return enc.Encode(f)
}

func urlPathEscape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "/", "%2F"), " ", "%20")
}

// WriteAcquisitionFeed writes a feed of downloadable books.
func (w FeedWriter) WriteAcquisitionFeed(out io.Writer, books []models.Book, selfPath string, progress map[int64]models.Progress) error {
	now := time.Now().UTC().Format(time.RFC3339)
	f := feed{
		Xmlns:   nsAtom,
		OPDS:    nsOPDS,
		ID:      w.BaseURL + selfPath,
		Title:   w.Title,
		Updated: now,
		Links: []link{
			{Rel: "self", Href: w.BaseURL + selfPath, Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
			{Rel: "start", Href: w.BaseURL + "/opds/", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}
	for _, b := range books {
		p := progress[b.ID]
		f.Entries = append(f.Entries, bookEntry(w.BaseURL, b, p))
	}
	enc := xml.NewEncoder(out)
	enc.Indent("", "  ")
	if _, err := out.Write([]byte(xml.Header)); err != nil {
		return err
	}
	return enc.Encode(f)
}

func bookEntry(base string, b models.Book, p models.Progress) entry {
	updated := b.ModifiedAt.UTC().Format(time.RFC3339)
	if updated == "" {
		updated = time.Now().UTC().Format(time.RFC3339)
	}
	summary := strings.TrimSpace(b.Description)
	if p.Percent > 0 {
		pct := min(int(p.Percent*100), 100)
		if summary != "" {
			summary = fmt.Sprintf("%d%% read — %s", pct, summary)
		} else {
			summary = fmt.Sprintf("%d%% read", pct)
		}
	}
	e := entry{
		ID:      fmt.Sprintf("%s/opds/book/%d", base, b.ID),
		Title:   b.Title,
		Updated: updated,
		Summary: summary,
		Links: []link{
			{
				Rel:  "http://opds-spec.org/acquisition",
				Href: fmt.Sprintf("%s/api/books/%d/download", base, b.ID),
				Type: mimeForFormat(b.Format),
			},
			{
				Rel:  "http://opds-spec.org/acquisition/open-access",
				Href: fmt.Sprintf("%s/api/books/%d/file", base, b.ID),
				Type: mimeForFormat(b.Format),
			},
		},
	}
	if b.HasCover {
		coverURL := fmt.Sprintf("%s/api/books/%d/cover", base, b.ID)
		e.Links = append(e.Links, link{
			Rel:  "http://opds-spec.org/image/thumbnail",
			Href: coverURL,
			Type: "image/jpeg",
		})
		e.Links = append(e.Links, link{
			Rel:  "http://opds-spec.org/image",
			Href: coverURL,
			Type: "image/jpeg",
		})
	}
	if b.Author != "" {
		e.Author = &author{Name: b.Author}
	}
	return e
}

func mimeForFormat(format string) string {
	switch format {
	case models.FormatEPUB:
		return "application/epub+zip"
	case models.FormatPDF:
		return "application/pdf"
	case models.FormatMP3:
		return "audio/mpeg"
	case models.FormatM4B, models.FormatM4A:
		return "audio/mp4"
	case models.FormatOGG:
		return "audio/ogg"
	case models.FormatFLAC:
		return "audio/flac"
	case models.FormatMOBI, models.FormatAZW, models.FormatAZW3:
		return "application/x-mobipocket-ebook"
	case models.FormatCBZ:
		return "application/vnd.comicbook+zip"
	case models.FormatCBR:
		return "application/vnd.comicbook-rar"
	case models.FormatAudiobook:
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}

// SetXMLHeaders applies standard OPDS response headers.
func SetXMLHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
}
