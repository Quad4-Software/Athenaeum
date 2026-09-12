package library

import (
	"encoding/xml"
	"io"
	"regexp"
	"strconv"
	"strings"

	"athenaeum/internal/models"
)

// comicInfoName is the archive-root member ComicRack uses for metadata.
const comicInfoName = "comicinfo.xml"

// maxComicInfoBytes bounds ComicInfo.xml reads; the member is a small XML
// sidecar and anything larger is treated as junk.
const maxComicInfoBytes = 1 << 20

// comicInfo is the subset of the ComicRack ComicInfo schema indexed on books.
type comicInfo struct {
	Title            string
	Series           string
	SeriesIndex      float64
	Author           string
	Description      string
	Publisher        string
	Tags             []string
	PublishedYear    int
	Language         string
	ReadingDirection string
	AgeRating        string
}

type comicInfoXML struct {
	Title       string `xml:"Title"`
	Series      string `xml:"Series"`
	Number      string `xml:"Number"`
	Writer      string `xml:"Writer"`
	Summary     string `xml:"Summary"`
	Publisher   string `xml:"Publisher"`
	Genre       string `xml:"Genre"`
	Tags        string `xml:"Tags"`
	Year        string `xml:"Year"`
	LanguageISO string `xml:"LanguageISO"`
	Manga       string `xml:"Manga"`
	AgeRating   string `xml:"AgeRating"`
}

// parseComicInfo decodes a ComicInfo.xml stream. Missing or malformed input
// returns the zero value so a bad sidecar never fails a scan.
func parseComicInfo(r io.Reader) comicInfo {
	data, err := io.ReadAll(io.LimitReader(r, maxComicInfoBytes))
	if err != nil || len(data) == 0 {
		return comicInfo{}
	}
	var raw comicInfoXML
	if err := xml.Unmarshal(normalizeXMLDecl(data), &raw); err != nil {
		return comicInfo{}
	}
	return comicInfo{
		Title:            strings.TrimSpace(raw.Title),
		Series:           strings.TrimSpace(raw.Series),
		SeriesIndex:      parseComicNumber(raw.Number),
		Author:           strings.TrimSpace(raw.Writer),
		Description:      strings.TrimSpace(raw.Summary),
		Publisher:        strings.TrimSpace(raw.Publisher),
		Tags:             comicInfoTagList(raw.Genre, raw.Tags),
		PublishedYear:    comicInfoYear(raw.Year),
		Language:         strings.TrimSpace(raw.LanguageISO),
		ReadingDirection: comicReadingDirection(raw.Manga),
		AgeRating:        strings.TrimSpace(raw.AgeRating),
	}
}

var comicNumberPrefix = regexp.MustCompile(`^-?\d+(?:\.\d+)?`)

// parseComicNumber reads the leading numeric portion of a ComicInfo Number,
// tolerating values like 5, 5.5 or 5a.
func parseComicNumber(s string) float64 {
	m := comicNumberPrefix.FindString(strings.TrimSpace(s))
	if m == "" {
		return 0
	}
	n, _ := strconv.ParseFloat(m, 64)
	return n
}

func comicInfoYear(s string) int {
	y, _ := strconv.Atoi(strings.TrimSpace(s))
	return y
}

// comicInfoTagList merges the comma-separated Genre and Tags fields.
func comicInfoTagList(genre, tags string) []string {
	var out []string
	for _, field := range []string{genre, tags} {
		for _, part := range strings.Split(field, ",") {
			if name := strings.TrimSpace(part); name != "" {
				out = append(out, name)
			}
		}
	}
	return out
}

func comicReadingDirection(manga string) string {
	switch strings.TrimSpace(manga) {
	case "Yes", "YesAndRightToLeft":
		return models.ReadingDirectionRTL
	default:
		return models.ReadingDirectionLTR
	}
}

// applyComicInfo copies parsed ComicInfo fields onto a book. Fields absent
// from the XML are left untouched so other metadata sources can fill them.
func applyComicInfo(b *models.Book, info comicInfo) {
	if info.Title != "" {
		b.Title = info.Title
	}
	if info.Series != "" {
		b.Series = CleanSeriesName(info.Series)
	}
	if info.SeriesIndex != 0 {
		b.SeriesIndex = info.SeriesIndex
	}
	if info.Author != "" {
		b.Author = info.Author
	}
	if info.Description != "" {
		b.Description = info.Description
	}
	if info.Publisher != "" {
		b.Publisher = info.Publisher
	}
	if info.PublishedYear > 0 {
		b.PublishedYear = info.PublishedYear
	}
	if info.Language != "" {
		b.Language = info.Language
	}
	if info.ReadingDirection != "" {
		b.ReadingDirection = info.ReadingDirection
	}
	if info.AgeRating != "" {
		b.AgeRating = info.AgeRating
	}
}
