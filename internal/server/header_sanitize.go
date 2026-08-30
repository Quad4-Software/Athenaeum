package server

import (
	"strings"
	"unicode"
)

// sanitizeHeaderValue strips CR/LF and other controls so attacker-controlled
// titles cannot inject MIME or HTTP header fields.
func sanitizeHeaderValue(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\r' || r == '\n' || r == '\x00' {
			continue
		}
		if unicode.IsControl(r) {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

func sanitizeFilenameToken(s string) string {
	s = sanitizeHeaderValue(s)
	s = strings.ReplaceAll(s, `"`, "")
	s = strings.ReplaceAll(s, `/`, "_")
	s = strings.ReplaceAll(s, `\`, "_")
	if s == "" {
		return "book"
	}
	return s
}
