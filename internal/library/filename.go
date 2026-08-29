package library

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"athenaeum/internal/models"
)

var (
	reYearTag    = regexp.MustCompile(`^\[?\(?(\d{4})\)?\]?[\s._-]*`)
	reSeriesNum  = regexp.MustCompile(`(?i)^(.+?)[\s._-]+(?:#|no\.?|vol\.?|book)?[\s._-]*(\d{1,3})(?:[\s._-]|$)`)
	reAuthorDash = regexp.MustCompile(`^(.+?)\s*[-–—]\s*(.+)$`)
)

type filenameMeta struct {
	Title       string
	Author      string
	Series      string
	SeriesIndex float64
}

func parseFilenameMeta(filePath string) filenameMeta {
	base := filepath.Base(filePath)
	if dot := strings.LastIndexByte(base, '.'); dot > 0 {
		base = base[:dot]
	}
	base = trimSpace(base)
	if base == "" {
		return filenameMeta{}
	}

	name := base
	if reYearTag.MatchString(name) {
		name = reYearTag.ReplaceAllString(name, "")
		name = trimSpace(name)
	}

	if m := reSeriesNum.FindStringSubmatch(name); len(m) == 3 {
		idx, _ := strconv.ParseFloat(m[2], 64)
		rest := CleanSeriesName(trimSpace(m[1]))
		title, author := splitAuthorTitle(rest)
		series := rest
		if title != "" && title != rest {
			series = CleanSeriesName(strings.TrimSuffix(rest, title))
			series = strings.Trim(series, " -–—")
		}
		return filenameMeta{
			Title:       title,
			Author:      author,
			Series:      series,
			SeriesIndex: idx,
		}
	}

	title, author := splitAuthorTitle(name)
	return filenameMeta{Title: title, Author: author}
}

func splitAuthorTitle(name string) (title, author string) {
	name = trimSpace(name)
	if name == "" {
		return "", ""
	}

	if m := reAuthorDash.FindStringSubmatch(name); len(m) == 3 {
		left := trimSpace(m[1])
		right := trimSpace(m[2])
		if looksLikeAuthor(left) && !looksLikeAuthor(right) {
			return right, left
		}
		if looksLikeAuthor(right) && !looksLikeAuthor(left) {
			return left, right
		}
		if len(left) <= len(right) {
			return right, left
		}
		return left, right
	}

	return titleFromFilename(name), ""
}

func looksLikeAuthor(s string) bool {
	s = trimSpace(s)
	if s == "" {
		return false
	}
	if strings.Contains(s, ",") {
		return true
	}
	wordCount := 0
	wordStart := -1
	for i := 0; i <= len(s); i++ {
		if i < len(s) && s[i] != ' ' && s[i] != '\t' {
			if wordStart < 0 {
				wordStart = i
			}
			continue
		}
		if wordStart < 0 {
			continue
		}
		wordCount++
		if wordCount > 4 {
			return false
		}
		if s[wordStart] < 'A' || s[wordStart] > 'Z' {
			return false
		}
		wordStart = -1
	}
	return wordCount >= 2 && wordCount <= 4
}

func titleFromFilename(name string) string {
	if strings.ContainsAny(name, "_.") {
		name = strings.ReplaceAll(name, "_", " ")
		name = strings.ReplaceAll(name, ".", " ")
	}
	return CleanDisplayText(name)
}

func applyFilenameMeta(book *models.Book, meta filenameMeta) {
	if meta.Title != "" && book.Title == "" {
		book.Title = meta.Title
	}
	if meta.Author != "" && book.Author == "" {
		book.Author = meta.Author
	}
	if meta.Series != "" && book.Series == "" {
		book.Series = CleanSeriesName(meta.Series)
	}
	if meta.SeriesIndex > 0 && book.SeriesIndex == 0 {
		book.SeriesIndex = meta.SeriesIndex
	}
}
