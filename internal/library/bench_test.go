package library

import (
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/models"
)

func TestClampScanWorkers(t *testing.T) {
	if got := ClampScanWorkers(0); got != defaultScanWorkers {
		t.Fatalf("ClampScanWorkers(0)=%d want %d", got, defaultScanWorkers)
	}
	if got := ClampScanWorkers(4); got != 4 {
		t.Fatalf("ClampScanWorkers(4)=%d", got)
	}
	if got := ClampScanWorkers(99); got != maxScanWorkers {
		t.Fatalf("ClampScanWorkers(99)=%d want %d", got, maxScanWorkers)
	}
}

func BenchmarkParsePDFInfo(b *testing.B) {
	pdf := []byte(`%PDF-1.4
1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj
trailer<</Info 3 0 R/Root 1 0 R>>
3 0 obj<</Title (Benchmark Book)/Author (Bench Author)>>endobj
%%EOF`)
	path := filepath.Join(b.TempDir(), "bench.pdf")
	if err := os.WriteFile(path, pdf, 0o644); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parsePDF(path)
	}
}

func BenchmarkParseEPUB(b *testing.B) {
	path := writeTestEPUB(&testing.T{})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := parseEPUB(path); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkImageNearKeywords(b *testing.B) {
	data := make([]byte, 712804)
	copy(data[len(data)-7:], []byte("/Cover"))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = imageNearKeywordsIn(data)
	}
}

func BenchmarkCleanDisplayTextClean(b *testing.B) {
	s := "Masters of Doom"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CleanDisplayText(s)
	}
}

func BenchmarkCleanDisplayTextDirty(b *testing.B) {
	s := "  Coffee consumption and\x00migraine  "
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CleanDisplayText(s)
	}
}

func BenchmarkIsGarbledTextClean(b *testing.B) {
	s := "Masters of Doom"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsGarbledText(s)
	}
}

func BenchmarkIsGarbledTextGarbled(b *testing.B) {
	s := "1\uFFFD\uFFFD\uFFFD\uFFFDbS\\b~"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsGarbledText(s)
	}
}

func BenchmarkCleanBookTitleClean(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CleanBookTitle("Masters of Doom", "/books/Masters of Doom.pdf")
	}
}

func BenchmarkCleanBookTitleGarbled(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CleanBookTitle("1\uFFFD\uFFFD\uFFFD\uFFFDbS\\b~", "/books/Masters of Doom.pdf")
	}
}

func BenchmarkCleanSeriesName(b *testing.B) {
	s := "The Expanse"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CleanSeriesName(s)
	}
}

func BenchmarkParseFilenameMeta(b *testing.B) {
	path := "/books/Asimov, Isaac - Foundation.epub"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseFilenameMeta(path)
	}
}

func BenchmarkNormalizeBookText(b *testing.B) {
	book := &models.Book{
		Title:  "Masters of Doom",
		Author: "David Kushner",
		Series: "Gaming History",
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizeBookText(book, "/books/Masters of Doom.pdf")
	}
}

func BenchmarkExtractPDFMetadataArtwork(b *testing.B) {
	jpeg := make([]byte, 3000)
	jpeg[0], jpeg[1], jpeg[2] = 0xff, 0xd8, 0xff
	jpeg[len(jpeg)-2], jpeg[len(jpeg)-1] = 0xff, 0xd9
	data := append([]byte("%PDF-1.4 /Cover\n"), jpeg...)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractPDFMetadataArtwork(data)
	}
}
