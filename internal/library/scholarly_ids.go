package library

import (
	"regexp"
	"strings"
)

var (
	doiRe = regexp.MustCompile(`(?i)\b(?:doi[:\s]*)?(?:https?://(?:dx\.)?doi\.org/)?(10\.\d{4,9}/[-._;()/:A-Z0-9]+)\b`)
	// Bare DOI pasted into a form field (no surrounding text).
	doiBareRe = regexp.MustCompile(`(?i)^10\.\d{4,9}/[-._;()/:A-Z0-9]+$`)
	// New-style arXiv IDs (YYMM.NNNNN) and old-style (archive/YYMMNNN).
	arxivRe = regexp.MustCompile(`(?i)\b(?:arXiv[:\s]*)?(?:https?://arxiv\.org/(?:abs|pdf)/)?((?:\d{4}\.\d{4,5}(?:v\d+)?)|(?:[a-z\-]+(?:\.[A-Z]{2})?/\d{7})(?:v\d+)?)\b`)
	pmidRe  = regexp.MustCompile(`(?i)\b(?:pmid[:\s]+)(\d{1,8})\b`)
)

// NormalizeDOI strips URL prefixes and returns a bare DOI, or empty if invalid.
func NormalizeDOI(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	m := doiRe.FindStringSubmatch(raw)
	if len(m) >= 2 {
		return strings.TrimRight(m[1], ".,;)")
	}
	raw = strings.TrimRight(raw, ".,;)")
	if doiBareRe.MatchString(raw) {
		return raw
	}
	return ""
}

// NormalizeArxivID returns a canonical arXiv id without version suffix when possible.
func NormalizeArxivID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	m := arxivRe.FindStringSubmatch(raw)
	if len(m) < 2 {
		return ""
	}
	id := m[1]
	if i := strings.LastIndex(strings.ToLower(id), "v"); i > 0 {
		rest := id[i+1:]
		if rest != "" && isAllDigits(rest) {
			id = id[:i]
		}
	}
	return id
}

// NormalizePubmedID returns digits-only PMID or empty.
func NormalizePubmedID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if isAllDigits(raw) && len(raw) <= 8 {
		return raw
	}
	m := pmidRe.FindStringSubmatch(raw)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func findDOIInText(s string) string {
	m := doiRe.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return NormalizeDOI(m[1])
}

func findArxivInText(s string) string {
	m := arxivRe.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return NormalizeArxivID(m[1])
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
