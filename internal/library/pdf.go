package library

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type pdfMeta struct {
	Title     string
	Author    string
	CoverData []byte
}

var (
	pdfLiteralRe = regexp.MustCompile(`/([A-Za-z]+)\s*\(([^)]*)\)`)
	pdfHexRe     = regexp.MustCompile(`/([A-Za-z]+)\s*<([0-9A-Fa-f\s]+)>`)
)

const pdfScanLimit = 8 << 20

// parsePDF extracts document info and the best available cover thumbnail.
func parsePDF(filePath string) pdfMeta {
	name := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	fallback := pdfMeta{Title: titleFromFilename(name)}

	if side := sidecarCover(filePath); len(side) > 0 {
		meta := pdfInfoFromFile(filePath, fallback)
		meta.CoverData = side
		return meta
	}

	data, err := os.ReadFile(filePath) // #nosec G304 -- path is library-relative from scanner
	if err != nil {
		return fallback
	}
	if len(data) > pdfScanLimit {
		data = data[:pdfScanLimit]
	}

	meta := pdfInfoFromPDF(data)
	if meta.Title == "" {
		meta.Title = fallback.Title
	}
	meta.CoverData = extractPDFCoverFromFile(filePath)
	return meta
}

func pdfInfoFromFile(filePath string, fallback pdfMeta) pdfMeta {
	data, err := os.ReadFile(filePath) // #nosec G304 -- path is library-relative from scanner
	if err != nil {
		return fallback
	}
	if len(data) > pdfScanLimit {
		data = data[:pdfScanLimit]
	}
	meta := pdfInfoFromPDF(data)
	if meta.Title == "" {
		meta.Title = fallback.Title
	}
	return meta
}

const minCoverBytes = 2048
const firstPageScanBytes = 150 << 10
const metadataScanBytes = 512 << 10

var coverKeywords = []string{"/Artwork", "/Cover", "/Thumb", "/Poster"}
var xmpGImgRe = regexp.MustCompile(`(?i)<xmpGImg:image>([A-Za-z0-9+/=\s]+)</xmpGImg:image>`)

var (
	coverKeywordsLower [][]byte
	pdfLowerPool       sync.Pool
)

func init() {
	coverKeywordsLower = make([][]byte, len(coverKeywords))
	for i, kw := range coverKeywords {
		coverKeywordsLower[i] = bytes.ToLower([]byte(kw))
	}
}

func borrowPDFLower(n int) []byte {
	if v := pdfLowerPool.Get(); v != nil {
		if b, ok := v.([]byte); ok && cap(b) >= n {
			return b[:n]
		}
	}
	return make([]byte, n)
}

func releasePDFLower(b []byte) {
	pdfLowerPool.Put(b[:0])
}

func asciiLowerPDF(dst, src []byte) {
	for i, b := range src {
		if b >= 'A' && b <= 'Z' {
			dst[i] = b + ('a' - 'A')
		} else {
			dst[i] = b
		}
	}
}

// extractPDFCover picks cover art from PDF metadata first, then the first page.
func extractPDFCoverHeuristic(data []byte) []byte {
	if img := extractPDFMetadataArtwork(data); len(img) >= minCoverBytes {
		return img
	}
	if img := extractPDFFirstPageImage(data); len(img) >= minCoverBytes {
		return img
	}
	if img := extractPDFMetadataArtwork(data); len(img) > 0 {
		return img
	}
	return extractPDFFirstPageImage(data)
}

func extractPDFMetadataArtwork(data []byte) []byte {
	if img := xmpGImgCover(data); len(img) > 0 {
		return img
	}
	for _, region := range metadataScanRegions(data) {
		if img := imageNearKeywordsIn(region); len(img) > 0 {
			return img
		}
	}
	return nil
}

func metadataScanRegions(data []byte) [][]byte {
	const headBytes = 128 << 10
	n := len(data)
	if n <= headBytes+metadataScanBytes {
		return [][]byte{data}
	}
	return [][]byte{data[:headBytes], data[n-metadataScanBytes:]}
}

func xmpGImgCover(data []byte) []byte {
	m := xmpGImgRe.FindSubmatch(data)
	if len(m) < 2 {
		return nil
	}
	raw := bytes.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, m[1])
	decoded, err := base64.StdEncoding.DecodeString(string(raw))
	if err != nil || len(decoded) < 256 {
		return nil
	}
	return decoded
}

func extractPDFFirstPageImage(data []byte) []byte {
	chunk := firstPageChunk(data)
	if img := firstJPEGIn(chunk); len(img) > 0 {
		return img
	}
	return firstPNGIn(chunk, len(chunk))
}

func firstPageChunk(data []byte) []byte {
	start := indexFirstPageMarker(data)
	if start < 0 {
		end := min(firstPageScanBytes, len(data))
		return data[:end]
	}
	end := len(data)
	rest := data[start+1:]
	for _, pat := range [][]byte{[]byte("/Type /Page"), []byte("/Type/Page")} {
		if next := bytes.Index(rest, pat); next >= 0 {
			if cand := start + 1 + next; cand < end {
				end = cand
			}
		}
	}
	if maxEnd := start + firstPageScanBytes; end > maxEnd {
		end = maxEnd
	}
	if end > len(data) {
		end = len(data)
	}
	return data[start:end]
}

func indexFirstPageMarker(data []byte) int {
	if i := bytes.Index(data, []byte("/Type /Page")); i >= 0 {
		return i
	}
	return bytes.Index(data, []byte("/Type/Page"))
}

func imageNearKeywordsIn(data []byte) []byte {
	n := len(data)
	if n == 0 {
		return nil
	}
	lower := borrowPDFLower(n)
	asciiLowerPDF(lower, data)
	defer releasePDFLower(lower)

	for _, kwLower := range coverKeywordsLower {
		pos := bytes.Index(lower, kwLower)
		if pos < 0 {
			continue
		}
		window := pos
		if window > 4096 {
			window -= 4096
		}
		end := min(pos+8192, n)
		if window >= end {
			continue
		}
		chunk := data[window:end]
		if jpeg := firstJPEGIn(chunk); len(jpeg) > 0 {
			return jpeg
		}
		if png := firstPNGIn(chunk, len(chunk)); len(png) > 0 {
			return png
		}
	}
	return nil
}

func firstJPEGIn(data []byte) []byte {
	i := bytes.Index(data, []byte{0xff, 0xd8, 0xff})
	if i < 0 {
		return nil
	}
	end := bytes.Index(data[i+2:], []byte{0xff, 0xd9})
	if end < 0 {
		return nil
	}
	end += i + 2 + 2
	if end > len(data) {
		return nil
	}
	return data[i:end]
}

func firstPNGIn(data []byte, limit int) []byte {
	if limit > len(data) {
		limit = len(data)
	}
	sig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	start := bytes.Index(data[:limit], sig)
	if start < 0 {
		return nil
	}
	i := start + 8
	for i+8 <= limit {
		if i+4 > limit {
			break
		}
		chunkLen := int(data[i])<<24 | int(data[i+1])<<16 | int(data[i+2])<<8 | int(data[i+3])
		if chunkLen < 0 {
			break
		}
		i += 4
		if i+4 > limit {
			break
		}
		chunkType := string(data[i : i+4])
		i += 4
		next := i + chunkLen + 4
		if next < i || next > limit {
			break
		}
		i = next
		if chunkType == "IEND" {
			if i < start || i > limit {
				break
			}
			return data[start:i]
		}
	}
	return nil
}

func pdfInfoFromPDF(data []byte) pdfMeta {
	var meta pdfMeta
	tail := data
	if len(tail) > 256<<10 {
		tail = tail[len(tail)-256<<10:]
	}
	applyPDFInfoMatches(&meta, pdfLiteralRe.FindAllSubmatch(tail, -1))
	if meta.Title == "" && meta.Author == "" {
		applyPDFInfoMatches(&meta, pdfLiteralRe.FindAllSubmatch(data, -1))
	}
	if meta.Title == "" && meta.Author == "" {
		applyPDFInfoHex(&meta, pdfHexRe.FindAllSubmatch(tail, -1))
	}
	return meta
}

func applyPDFInfoMatches(meta *pdfMeta, matches [][][]byte) {
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		key := string(m[1])
		val := pdfUnescapeLiteral(string(m[2]))
		switch key {
		case "Title":
			if meta.Title == "" {
				meta.Title = val
			}
		case "Author":
			if meta.Author == "" {
				meta.Author = val
			}
		}
	}
}

func applyPDFInfoHex(meta *pdfMeta, matches [][][]byte) {
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		key := string(m[1])
		val := decodePDFHexString(string(m[2]))
		switch key {
		case "Title":
			if meta.Title == "" {
				meta.Title = val
			}
		case "Author":
			if meta.Author == "" {
				meta.Author = val
			}
		}
	}
}

func pdfUnescapeLiteral(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\r`, "\r")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = strings.ReplaceAll(s, `\(`, "(")
	s = strings.ReplaceAll(s, `\)`, ")")
	s = strings.ReplaceAll(s, `\\`, `\`)
	return CleanDisplayText(s)
}

func decodePDFHexString(hex string) string {
	hex = strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			return r
		}
		return -1
	}, hex)
	if len(hex) < 4 {
		return ""
	}
	if len(hex)%2 != 0 {
		hex = hex[:len(hex)-1]
	}
	b := make([]byte, len(hex)/2)
	for i := range b {
		n, err := strconv.ParseUint(hex[i*2:i*2+2], 16, 8)
		if err != nil {
			return ""
		}
		b[i] = byte(n)
	}
	if len(b) >= 2 && b[0] == 0xFE && b[1] == 0xFF {
		return CleanDisplayText(decodeUTF16BE(b[2:]))
	}
	latin := CleanDisplayText(strings.TrimSpace(string(b)))
	if len(b) >= 4 && len(b)%2 == 0 && pdfHexLooksUTF16(b) {
		utf16 := CleanDisplayText(decodeUTF16BE(b))
		if !IsGarbledText(utf16) && (IsGarbledText(latin) || len(utf16) > len(latin)/2) {
			return utf16
		}
	}
	return latin
}

func pdfHexLooksUTF16(b []byte) bool {
	pairs := len(b) / 2
	if pairs < 2 {
		return false
	}
	asciiUTF16 := 0
	for i := 0; i+1 < len(b); i += 2 {
		if b[i] == 0 && b[i+1] != 0 && b[i+1] < 0x80 {
			asciiUTF16++
		}
	}
	return asciiUTF16*2 >= pairs
}

func decodeUTF16BE(b []byte) string {
	if len(b) < 2 {
		return ""
	}
	var runes []rune
	for i := 0; i+1 < len(b); i += 2 {
		runes = append(runes, rune(b[i])<<8|rune(b[i+1]))
	}
	return strings.TrimSpace(string(runes))
}
