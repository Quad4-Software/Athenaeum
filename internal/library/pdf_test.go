package library

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestParsePDFInfo(t *testing.T) {
	pdf := []byte(`%PDF-1.4
1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj
trailer<</Info 3 0 R/Root 1 0 R>>
3 0 obj<</Title (The Martian)/Author (Andy Weir)>>endobj
%%EOF`)
	path := filepath.Join(t.TempDir(), "martian.pdf")
	if err := os.WriteFile(path, pdf, 0o644); err != nil {
		t.Fatal(err)
	}
	meta := parsePDF(path)
	if meta.Title != "The Martian" {
		t.Errorf("title = %q", meta.Title)
	}
	if meta.Author != "Andy Weir" {
		t.Errorf("author = %q", meta.Author)
	}
}

func TestExtractPDFCoverXMP(t *testing.T) {
	jpeg := makeTestJPEG(3000)
	b64 := base64.StdEncoding.EncodeToString(jpeg)
	xmp := []byte(`<x:xmpmeta><rdf:RDF><rdf:Description><xmpGImg:image>` + b64 + `</xmpGImg:image></rdf:Description></rdf:RDF></x:xmpmeta>`)
	data := append([]byte("%PDF-1.4\n"), xmp...)
	got := extractPDFCoverHeuristic(data)
	if len(got) != len(jpeg) {
		t.Fatalf("xmp cover len=%d want %d", len(got), len(jpeg))
	}
}

func TestExtractPDFCoverFirstPageBeforeBody(t *testing.T) {
	jpeg := makeTestJPEG(minCoverBytes)

	data := append([]byte("%PDF-1.4 stream\n"), jpeg...)
	data = append(data, []byte("\nendstream")...)
	got := extractPDFCoverHeuristic(data)
	if len(got) < minCoverBytes {
		t.Fatalf("cover len=%d", len(got))
	}
}

func makeTestJPEG(size int) []byte {
	jpeg := make([]byte, size)
	jpeg[0], jpeg[1], jpeg[2] = 0xff, 0xd8, 0xff
	jpeg[len(jpeg)-2], jpeg[len(jpeg)-1] = 0xff, 0xd9
	return jpeg
}

func TestImageNearKeywordsIn(t *testing.T) {
	const total = 712804
	data := make([]byte, total)
	meta := data[total-metadataScanBytes:]
	copy(meta[len(meta)-7:], []byte("/Cover"))

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("imageNearKeywordsIn panicked: %v", r)
		}
	}()
	if img := imageNearKeywordsIn(data); img != nil {
		_ = len(img)
	}
	_ = imageNearKeywordsIn(meta)
}

func TestExtractPDFCoverPrefersMetadataOverBody(t *testing.T) {
	cover := makeTestJPEG(3000)
	body := makeTestJPEG(minCoverBytes * 3)

	data := make([]byte, 900<<10)
	copy(data[len(data)-len(cover)-7:], []byte("/Cover"))
	copy(data[len(data)-len(cover):], cover)
	copy(data[200<<10:], body)

	got := extractPDFCoverHeuristic(data)
	if len(got) != len(cover) {
		t.Fatalf("cover len=%d want metadata cover %d", len(got), len(cover))
	}
}

func TestExtractPDFCoverFirstPageNotLargest(t *testing.T) {
	small := makeTestJPEG(minCoverBytes)
	large := makeTestJPEG(minCoverBytes * 4)

	var data []byte
	data = append(data, []byte("%PDF-1.4 /Type /Page stream\n")...)
	data = append(data, small...)
	data = append(data, []byte("\nstream\n")...)
	data = append(data, large...)
	data = append(data, []byte("\n/Type /Page\n")...)

	got := extractPDFCoverHeuristic(data)
	if len(got) != len(small) {
		t.Fatalf("cover len=%d want first-page image %d", len(got), len(small))
	}
}

func TestExtractPDFCoverIgnoresTailJPEGWithoutMetadata(t *testing.T) {
	page := makeTestJPEG(minCoverBytes)
	tail := makeTestJPEG(minCoverBytes * 5)

	data := make([]byte, 700<<10)
	copy(data[8<<10:], page)
	copy(data[len(data)-len(tail):], tail)

	got := extractPDFCoverHeuristic(data)
	if len(got) != len(page) {
		t.Fatalf("cover len=%d want first-page image %d, not tail %d", len(got), len(page), len(tail))
	}
}

func TestFirstPNGInCorruptChunkLen(t *testing.T) {
	data := make([]byte, 524288)
	sig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	copy(data[500000:], sig)
	off := 500008
	copy(data[off+4:off+8], []byte("IHDR"))
	off += 12
	data[off] = 0xff
	data[off+1] = 0xff
	data[off+2] = 0xff
	data[off+3] = 0xff
	copy(data[off+4:off+8], []byte("IEND"))

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("firstPNGIn panicked: %v", r)
		}
	}()
	_ = firstPNGIn(data, len(data))
}

func TestPdfUnescapeLiteral(t *testing.T) {
	got := pdfUnescapeLiteral(`Hello \(world\)`)
	if got != "Hello (world)" {
		t.Errorf("got %q", got)
	}
}

func TestExtractPDFCoverPrefersFirstPageOverLargerLaterPage(t *testing.T) {
	path := "/home/user1/Documents/book/007.pdf"
	if _, err := os.Stat(path); err != nil {
		t.Skip("reference PDF not available")
	}
	got := extractPDFCoverFromFile(path)
	if len(got) < minCoverBytes {
		t.Fatalf("cover len=%d", len(got))
	}
	// Page 1 embedded image is ~204KB; page 2 is ~850KB. Cover must come from page 1.
	if len(got) > 300<<10 {
		t.Fatalf("cover len=%d looks like page 2 image, want first-page cover", len(got))
	}
}
