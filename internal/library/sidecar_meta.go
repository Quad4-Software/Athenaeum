package library

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"athenaeum/internal/models"
)

type sidecarFields struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Series      string  `json:"series"`
	SeriesIndex float64 `json:"seriesIndex"`
	Description string  `json:"description"`
	Language    string  `json:"language"`
	ISBN        string  `json:"isbn"`
	ASIN        string  `json:"asin"`
}

// sidecarMetadata reads metadata.json or a companion .opf next to a book file.
func sidecarMetadata(filePath string) sidecarFields {
	dir := filepath.Dir(filePath)
	base := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	candidates := []string{
		filepath.Join(dir, base+".json"),
		filepath.Join(dir, base+".metadata.json"),
		filepath.Join(dir, "metadata.json"),
		filepath.Join(dir, base+".opf"),
		filepath.Join(dir, "metadata.opf"),
	}
	for _, p := range candidates {
		if meta, ok := readSidecarFile(p); ok {
			return meta
		}
	}
	return sidecarFields{}
}

func readSidecarFile(path string) (sidecarFields, bool) {
	data, err := os.ReadFile(path) // #nosec G304
	if err != nil || len(data) == 0 {
		return sidecarFields{}, false
	}
	if strings.HasSuffix(strings.ToLower(path), ".opf") {
		meta, err := parseOPFBytes(data)
		if err != nil {
			return sidecarFields{}, false
		}
		return sidecarFields{
			Title:       meta.Title,
			Author:      meta.Author,
			Series:      meta.Series,
			SeriesIndex: meta.SeriesIndex,
			Description: meta.Description,
			Language:    meta.Language,
		}, true
	}
	var raw sidecarFields
	if err := json.Unmarshal(data, &raw); err != nil {
		return sidecarFields{}, false
	}
	return raw, true
}

func parseOPFBytes(data []byte) (epubMeta, error) {
	var pkg opfPackage
	if err := xml.Unmarshal(normalizeXMLDecl(data), &pkg); err != nil {
		return epubMeta{}, err
	}
	meta := epubMeta{Language: pkg.Metadata.Language, Description: pkg.Metadata.Description}
	if len(pkg.Metadata.Titles) > 0 {
		meta.Title = strings.TrimSpace(pkg.Metadata.Titles[0])
	}
	if len(pkg.Metadata.Creators) > 0 {
		meta.Author = strings.TrimSpace(pkg.Metadata.Creators[0])
	}
	for _, m := range pkg.Metadata.Metas {
		switch {
		case m.Property == "belongs-to-collection":
			meta.Series = strings.TrimSpace(m.Value)
		case m.Property == "group-position" && meta.SeriesIndex == 0:
			meta.SeriesIndex, _ = strconv.ParseFloat(strings.TrimSpace(m.Value), 64)
		case m.Name == "calibre:series":
			meta.Series = strings.TrimSpace(m.Content)
		case m.Name == "calibre:series_index":
			meta.SeriesIndex, _ = strconv.ParseFloat(strings.TrimSpace(m.Content), 64)
		}
	}
	return meta, nil
}

func applySidecarMeta(book *models.Book, side sidecarFields) (isbn, asin string) {
	if side.Title != "" {
		book.Title = side.Title
	}
	if side.Author != "" {
		book.Author = side.Author
	}
	if side.Series != "" {
		book.Series = side.Series
	}
	if side.SeriesIndex > 0 {
		book.SeriesIndex = side.SeriesIndex
	}
	if side.Description != "" {
		book.Description = side.Description
	}
	if side.Language != "" {
		book.Language = side.Language
	}
	return strings.TrimSpace(side.ISBN), strings.TrimSpace(side.ASIN)
}

func applyLookupMeta(book *models.Book, fields sidecarFields) {
	if fields.Title != "" && book.Title == "" {
		book.Title = fields.Title
	}
	if fields.Author != "" && book.Author == "" {
		book.Author = fields.Author
	}
	if fields.Description != "" && book.Description == "" {
		book.Description = fields.Description
	}
	if fields.Language != "" && book.Language == "" {
		book.Language = fields.Language
	}
	if fields.Series != "" && book.Series == "" {
		book.Series = fields.Series
	}
	if fields.SeriesIndex > 0 && book.SeriesIndex == 0 {
		book.SeriesIndex = fields.SeriesIndex
	}
}
