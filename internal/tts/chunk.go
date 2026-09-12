package tts

import (
	"strings"
	"unicode/utf8"
)

// ChunkParagraphs joins paragraphs into synthesis-sized chunks that end on
// sentence boundaries where possible. No chunk exceeds max runes.
func ChunkParagraphs(paras []string, max int) []string {
	if max <= 0 {
		max = JobChunkChars
	}
	var out []string
	var buf strings.Builder

	flush := func() {
		if s := strings.TrimSpace(buf.String()); s != "" {
			out = append(out, s)
		}
		buf.Reset()
	}

	for _, p := range paras {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		for utf8.RuneCountInString(p) > max {
			if buf.Len() > 0 {
				flush()
			}
			head, tail := splitSentence(p, max)
			if head == "" {
				head = string([]rune(p)[:max])
				tail = string([]rune(p)[max:])
			}
			out = append(out, head)
			p = strings.TrimSpace(tail)
		}
		if p == "" {
			continue
		}
		if buf.Len() > 0 && utf8.RuneCountInString(buf.String())+1+utf8.RuneCountInString(p) > max {
			flush()
		}
		if buf.Len() > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(p)
	}
	flush()
	return out
}

// splitSentence finds a cut near max runes at a sentence or clause
// boundary, falling back to a word boundary.
func splitSentence(s string, max int) (head, tail string) {
	runes := []rune(s)
	if len(runes) <= max {
		return s, ""
	}
	region := string(runes[:max])
	// Prefer the last sentence terminator inside the window.
	cut := -1
	for i := len(region) - 1; i >= 0; i-- {
		c := region[i]
		if c == '.' || c == '!' || c == '?' || c == ';' || c == ':' {
			cut = i + 1
			break
		}
	}
	// Otherwise split on the last space.
	if cut < 0 {
		cut = strings.LastIndexByte(region, ' ')
	}
	if cut <= 0 {
		return "", s
	}
	runeCut := utf8.RuneCountInString(region[:cut])
	head = strings.TrimSpace(region[:cut])
	tail = strings.TrimSpace(string(runes[runeCut:]))
	return head, tail
}
