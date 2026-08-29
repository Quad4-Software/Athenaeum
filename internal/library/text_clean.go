package library

import (
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"athenaeum/internal/models"
)

func CleanDisplayText(s string) string {
	s = trimSpace(s)
	if s == "" {
		return ""
	}
	if !displayTextNeedsClean(s) {
		return s
	}
	return cleanDisplayTextSlow(s)
}

func displayTextNeedsClean(s string) bool {
	if !utf8.ValidString(s) {
		return true
	}
	prevSpace := false
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		if displayRuneNeedsClean(r) {
			return true
		}
		if unicode.IsSpace(r) {
			if prevSpace {
				return true
			}
			prevSpace = true
			continue
		}
		prevSpace = false
	}
	return false
}

func displayRuneNeedsClean(r rune) bool {
	switch {
	case r == 0 || (r < 0x20 && r != '\t'):
		return true
	case r == '\uFEFF', r == '\u200B', r == '\u200C', r == '\u200D', r == '\u2060':
		return true
	case r == '\uFFFD':
		return true
	case unicode.IsControl(r):
		return true
	default:
		return false
	}
}

func normalizeDisplayRune(r rune) (rune, bool) {
	if displayRuneNeedsClean(r) {
		return ' ', true
	}
	if unicode.IsSpace(r) {
		return ' ', true
	}
	return r, false
}

func cleanDisplayTextSlow(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	needSpace := false
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			if b.Len() > 0 {
				needSpace = true
			}
			continue
		}
		i += size
		r, spaceLike := normalizeDisplayRune(r)
		if spaceLike {
			if b.Len() > 0 {
				needSpace = true
			}
			continue
		}
		if needSpace {
			b.WriteByte(' ')
			needSpace = false
		}
		b.WriteRune(r)
	}
	return b.String()
}

func IsGarbledText(s string) bool {
	s = trimSpace(s)
	if s == "" {
		return true
	}
	if !utf8.ValidString(s) {
		return true
	}

	letters := 0
	weird := 0
	for _, r := range s {
		switch {
		case r == '\uFFFD':
			weird++
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			letters++
		case unicode.IsSpace(r):
			letters++
		case isTitlePunct(r):
			letters++
		case r > 127 && unicode.IsPrint(r):
			letters++
		default:
			weird++
		}
	}
	total := letters + weird
	if total == 0 || letters == 0 {
		return true
	}
	if float64(weird)/float64(total) > 0.15 {
		return true
	}
	return hasBrokenWords(s)
}

func hasBrokenWords(s string) bool {
	wordCount := 0
	wordRunes := 0
	inWord := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if inWord {
				wordCount++
				if wordRunes > 3 {
					return false
				}
				inWord = false
				wordRunes = 0
			}
			continue
		}
		inWord = true
		wordRunes++
	}
	if inWord {
		wordCount++
		if wordRunes > 3 {
			return false
		}
	}
	return wordCount >= 3
}

func isTitlePunct(r rune) bool {
	switch r {
	case '-', '–', '—', ':', '.', ',', '\'', '"', '&', '(', ')', '!', '?', '/', ';', '+', '#', '@':
		return true
	default:
		return false
	}
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end {
		r, size := utf8.DecodeRuneInString(s[start:])
		if !unicode.IsSpace(r) {
			break
		}
		start += size
	}
	for end > start {
		r, sz := utf8.DecodeLastRuneInString(s[:end])
		if !unicode.IsSpace(r) {
			break
		}
		end -= sz
	}
	if start == 0 && end == len(s) {
		return s
	}
	return s[start:end]
}

func titleFromRelPath(relPath string) string {
	base := strings.TrimSuffix(filepath.Base(relPath), filepath.Ext(relPath))
	return titleFromFilename(base)
}

func CleanBookTitle(title, filePath string) string {
	cleaned := CleanDisplayText(title)
	if cleaned != "" && !IsGarbledText(cleaned) {
		return cleaned
	}

	meta := parseFilenameMeta(filePath)
	if meta.Title != "" {
		if fb := CleanDisplayText(meta.Title); fb != "" && !IsGarbledText(fb) {
			return fb
		}
	}
	if fb := CleanDisplayText(titleFromRelPath(filePath)); fb != "" && !IsGarbledText(fb) {
		return fb
	}
	if cleaned != "" {
		return cleaned
	}
	return CleanDisplayText(filepath.Base(filePath))
}

func CleanAuthorName(author, filePath string) string {
	cleaned := CleanDisplayText(author)
	if cleaned != "" && !IsGarbledText(cleaned) {
		return cleaned
	}
	meta := parseFilenameMeta(filePath)
	if meta.Author != "" {
		if fb := CleanDisplayText(meta.Author); fb != "" && !IsGarbledText(fb) {
			return fb
		}
	}
	if IsGarbledText(cleaned) {
		return ""
	}
	return cleaned
}

func normalizeBookText(book *models.Book, filePath string) {
	if book == nil {
		return
	}
	path := filePath
	if path == "" {
		path = book.RelPath
	}
	book.Title = CleanBookTitle(book.Title, path)
	book.Author = CleanAuthorName(book.Author, path)
	book.Series = CleanSeriesName(book.Series)
}
