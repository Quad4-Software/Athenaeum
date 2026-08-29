package library

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	reSeriesBracketTag   = regexp.MustCompile(`(?i)\[[^\]]*\]\s*`)
	reSeriesIndexSuffix  = regexp.MustCompile(`(?i)(?:,\s*)?(?:#\s*|no\.?\s*|book\s*|vol\.?\s*)\d+(?:\.\d+)?\s*$`)
	reSeriesYearSuffix   = regexp.MustCompile(`-\d{4}$`)
	reSeriesARNCode      = regexp.MustCompile(`(?i)^ARN[A-Z0-9]+$`)
	reSeriesNumericJunk  = regexp.MustCompile(`(?i)^\d+(?:[-_][a-z]?\d+)?$`)
	reSeriesCatalogToken = regexp.MustCompile(`(?i)^[A-Z]{2,}\d+[_-][A-Z0-9]+$`)
)

// CleanSeriesName strips site tags, normalizes spacing, and rejects junk catalog tokens.
func CleanSeriesName(s string) string {
	s = CleanDisplayText(s)
	if s == "" {
		return ""
	}
	if !seriesNameNeedsClean(s) {
		if isInvalidSeriesName(s) {
			return ""
		}
		return s
	}

	for {
		prev := s
		s = reSeriesBracketTag.ReplaceAllString(s, "")
		s = trimSpace(s)
		if s == prev {
			break
		}
	}

	s = strings.ReplaceAll(s, "_", " ")
	s = reSeriesIndexSuffix.ReplaceAllString(s, "")
	s = reSeriesYearSuffix.ReplaceAllString(s, "")
	if isInvalidSeriesName(s) {
		return ""
	}
	s = strings.ReplaceAll(s, "-", " ")
	s = trimSpace(s)
	s = strings.Trim(s, ",;")
	s = collapseFields(s)

	if isInvalidSeriesName(s) {
		return ""
	}
	return s
}

func seriesNameNeedsClean(s string) bool {
	if strings.ContainsAny(s, "[]_") {
		return true
	}
	if strings.Contains(s, "-") && reSeriesYearSuffix.MatchString(s) {
		return true
	}
	if reSeriesBracketTag.MatchString(s) {
		return true
	}
	if reSeriesIndexSuffix.MatchString(s) {
		return true
	}
	return fieldsNeedCollapse(s)
}

func fieldsNeedCollapse(s string) bool {
	prevSpace := false
	for _, r := range s {
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

func collapseFields(s string) string {
	if !fieldsNeedCollapse(s) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	needSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
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

func isInvalidSeriesName(s string) bool {
	if s == "" {
		return true
	}
	if reSeriesNumericJunk.MatchString(s) {
		return true
	}
	if IsGarbledText(s) {
		return true
	}
	if !seriesNameNeedsCompactCheck(s) {
		return false
	}
	compact := seriesCompactToken(s)
	if reSeriesARNCode.MatchString(compact) {
		return true
	}
	if reSeriesCatalogToken.MatchString(compact) {
		return true
	}
	return false
}

func seriesNameNeedsCompactCheck(s string) bool {
	if len(s) >= 3 {
		switch {
		case s[0] == 'A' || s[0] == 'a':
			if s[1] == 'R' || s[1] == 'r' {
				if s[2] == 'N' || s[2] == 'n' {
					return true
				}
			}
		}
	}
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			return true
		}
	}
	return false
}

func seriesCompactToken(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case ' ', '-', '_', '.':
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
