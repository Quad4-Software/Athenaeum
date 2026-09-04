package library

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"athenaeum/internal/models"
)

// BibEntry is one parsed BibTeX record.
type BibEntry struct {
	Type   string
	Key    string
	Fields map[string]string
}

var bibEntryStartRe = regexp.MustCompile(`(?i)@([a-z]+)\s*\{\s*([^,\s]*)\s*,`)

// ParseBibTeX parses a .bib document into entries. Nested braces are handled.
func ParseBibTeX(data string) []BibEntry {
	var out []BibEntry
	i := 0
	for i < len(data) {
		loc := bibEntryStartRe.FindStringSubmatchIndex(data[i:])
		if loc == nil {
			break
		}
		abs := i + loc[0]
		typ := data[i+loc[2] : i+loc[3]]
		key := data[i+loc[4] : i+loc[5]]
		braceStart := strings.IndexByte(data[abs:], '{')
		if braceStart < 0 {
			i = abs + 1
			continue
		}
		start := abs + braceStart
		end := findMatchingBrace(data, start)
		if end < 0 {
			break
		}
		body := data[start+1 : end]
		// Drop cite key prefix before first comma.
		if idx := strings.IndexByte(body, ','); idx >= 0 {
			body = body[idx+1:]
		}
		fields := parseBibFields(body)
		out = append(out, BibEntry{
			Type:   strings.ToLower(strings.TrimSpace(typ)),
			Key:    strings.TrimSpace(key),
			Fields: fields,
		})
		if len(out) >= maxBibImportEntries {
			break
		}
		i = end + 1
	}
	return out
}

func findMatchingBrace(s string, openIdx int) int {
	depth := 0
	for i := openIdx; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		case '"':
			// Skip quoted strings at top of field values when scanning braces.
			i++
			for i < len(s) && s[i] != '"' {
				if s[i] == '\\' && i+1 < len(s) {
					i += 2
					continue
				}
				i++
			}
		}
	}
	return -1
}

func parseBibFields(body string) map[string]string {
	fields := make(map[string]string)
	i := 0
	for i < len(body) {
		for i < len(body) && (body[i] == ',' || unicode.IsSpace(rune(body[i]))) {
			i++
		}
		if i >= len(body) {
			break
		}
		keyStart := i
		for i < len(body) && body[i] != '=' {
			i++
		}
		if i >= len(body) {
			break
		}
		key := strings.ToLower(strings.TrimSpace(body[keyStart:i]))
		i++ // skip =
		for i < len(body) && unicode.IsSpace(rune(body[i])) {
			i++
		}
		if i >= len(body) {
			break
		}
		var val string
		switch body[i] {
		case '{':
			end := findMatchingBrace(body, i)
			if end < 0 {
				return fields
			}
			val = body[i+1 : end]
			i = end + 1
		case '"':
			i++
			start := i
			for i < len(body) && body[i] != '"' {
				if body[i] == '\\' && i+1 < len(body) {
					i += 2
					continue
				}
				i++
			}
			val = body[start:i]
			if i < len(body) {
				i++
			}
		default:
			start := i
			for i < len(body) && body[i] != ',' {
				i++
			}
			val = strings.TrimSpace(body[start:i])
		}
		if key != "" {
			fields[key] = bibUnescape(strings.TrimSpace(val))
		}
	}
	return fields
}

func bibUnescape(s string) string {
	s = strings.ReplaceAll(s, `\&`, "&")
	s = strings.ReplaceAll(s, `\%`, "%")
	s = strings.ReplaceAll(s, `\_`, "_")
	s = strings.ReplaceAll(s, `\{`, "{")
	s = strings.ReplaceAll(s, `\}`, "}")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func bibField(e BibEntry, keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(e.Fields[k]); v != "" {
			return v
		}
	}
	return ""
}

// BibEntryToSidecar maps a BibTeX entry onto sidecar/citation fields.
func BibEntryToSidecar(e BibEntry) sidecarFields {
	authors := bibField(e, "author")
	authors = strings.ReplaceAll(authors, " and ", ", ")
	yearStr := bibField(e, "year")
	year, _ := strconv.Atoi(yearStr)
	doi := NormalizeDOI(bibField(e, "doi"))
	arxiv := NormalizeArxivID(bibField(e, "arxiv", "eprint"))
	if arxiv == "" && strings.EqualFold(bibField(e, "archiveprefix", "archive_prefix"), "arXiv") {
		arxiv = NormalizeArxivID(bibField(e, "eprint"))
	}
	pmid := NormalizePubmedID(bibField(e, "pmid", "pubmed"))
	journal := bibField(e, "journal", "booktitle", "venue")
	return sidecarFields{
		Title:         bibField(e, "title"),
		Author:        authors,
		Description:   bibField(e, "abstract", "annote"),
		DOI:           doi,
		ArxivID:       arxiv,
		PubmedID:      pmid,
		Journal:       journal,
		Volume:        bibField(e, "volume"),
		Issue:         bibField(e, "number", "issue"),
		Pages:         bibField(e, "pages"),
		PublishedYear: year,
		Language:      bibField(e, "language"),
	}
}

// FormatBibTeX renders a book as a single BibTeX @article (or @misc) entry.
func FormatBibTeX(b models.Book) string {
	entryType := "article"
	if b.Journal == "" && b.DOI == "" && b.ArxivID != "" {
		entryType = "misc"
	}
	key := bibCiteKey(b)
	var sb strings.Builder
	fmt.Fprintf(&sb, "@%s{%s,\n", entryType, key)
	writeBibField(&sb, "title", b.Title)
	writeBibField(&sb, "author", bibAuthorsFromString(b.Author))
	writeBibField(&sb, "journal", b.Journal)
	if b.PublishedYear > 0 {
		writeBibField(&sb, "year", strconv.Itoa(b.PublishedYear))
	}
	writeBibField(&sb, "volume", b.Volume)
	writeBibField(&sb, "number", b.Issue)
	writeBibField(&sb, "pages", b.Pages)
	writeBibField(&sb, "doi", b.DOI)
	if b.ArxivID != "" {
		writeBibField(&sb, "eprint", b.ArxivID)
		writeBibField(&sb, "archivePrefix", "arXiv")
	}
	writeBibField(&sb, "pmid", b.PubmedID)
	if b.Description != "" {
		writeBibField(&sb, "abstract", b.Description)
	}
	sb.WriteString("}\n")
	return sb.String()
}

// FormatBibTeXMulti joins several books into one .bib document.
func FormatBibTeXMulti(books []models.Book) string {
	var sb strings.Builder
	for i, b := range books {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(FormatBibTeX(b))
	}
	return sb.String()
}

func writeBibField(sb *strings.Builder, key, val string) {
	val = strings.TrimSpace(val)
	if val == "" {
		return
	}
	val = strings.ReplaceAll(val, `{`, "")
	val = strings.ReplaceAll(val, `}`, "")
	val = strings.ReplaceAll(val, "%", `\%`)
	val = strings.ReplaceAll(val, "\r\n", " ")
	val = strings.ReplaceAll(val, "\n", " ")
	val = strings.ReplaceAll(val, "\r", " ")
	val = strings.Join(strings.Fields(val), " ")
	fmt.Fprintf(sb, "  %s = {%s},\n", key, val)
}

func bibCiteKey(b models.Book) string {
	if b.DOI != "" {
		parts := strings.Split(b.DOI, "/")
		last := parts[len(parts)-1]
		last = regexp.MustCompile(`[^A-Za-z0-9]+`).ReplaceAllString(last, "")
		if last != "" {
			return last
		}
	}
	if b.ArxivID != "" {
		return "arxiv" + strings.ReplaceAll(b.ArxivID, ".", "")
	}
	if b.PubmedID != "" {
		return "pmid" + b.PubmedID
	}
	author := "unknown"
	if b.Author != "" {
		author = strings.Fields(b.Author)[0]
		author = regexp.MustCompile(`[^A-Za-z0-9]+`).ReplaceAllString(author, "")
	}
	year := "nd"
	if b.PublishedYear > 0 {
		year = strconv.Itoa(b.PublishedYear)
	}
	return strings.ToLower(author) + year
}

func bibAuthorsFromString(author string) string {
	author = strings.TrimSpace(author)
	if author == "" {
		return ""
	}
	parts := strings.Split(author, ",")
	if len(parts) == 1 {
		return author
	}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return strings.Join(parts, " and ")
}
