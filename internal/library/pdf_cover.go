package library

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func extractPDFCoverFromFile(filePath string) []byte {
	if img := extractPDFCoverPdfcpu(filePath); len(img) > 0 {
		return img
	}
	if img := extractPDFCoverPdftoppm(filePath); len(img) > 0 {
		return img
	}

	data, err := os.ReadFile(filePath) // #nosec G304 -- path is library-relative from scanner
	if err != nil {
		return nil
	}
	if len(data) > pdfScanLimit {
		data = data[:pdfScanLimit]
	}
	return extractPDFCoverHeuristic(data)
}

func extractPDFCoverPdfcpu(filePath string) []byte {
	f, err := os.Open(filePath) // #nosec G304 -- path is library-relative from scanner
	if err != nil {
		return nil
	}
	defer f.Close()

	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	pageMaps, err := api.ExtractImagesRaw(f, []string{"1", "2", "3"}, conf)
	if err != nil || len(pageMaps) == 0 {
		return nil
	}
	return pickPDFCoverImage(pageMaps)
}

type pdfCoverCandidate struct {
	data  []byte
	score int64
}

func pickPDFCoverImage(pageMaps []map[int]model.Image) []byte {
	var fallback *pdfCoverCandidate
	for _, pageMap := range pageMaps {
		pageBest := bestPDFCoverOnPage(pageMap)
		if pageBest == nil {
			continue
		}
		if len(pageBest.data) >= minCoverBytes {
			return pageBest.data
		}
		if fallback == nil || pageBest.score > fallback.score {
			fallback = pageBest
		}
	}
	if fallback != nil {
		return fallback.data
	}
	return nil
}

func bestPDFCoverOnPage(pageMap map[int]model.Image) *pdfCoverCandidate {
	var best *pdfCoverCandidate
	var fallback *pdfCoverCandidate
	for _, img := range pageMap {
		data, err := readPDFImageBytes(img)
		if err != nil || len(data) < 256 {
			continue
		}
		score := pdfCoverImageScore(img, len(data))
		cand := &pdfCoverCandidate{data: data, score: score}
		if len(data) >= minCoverBytes {
			if best == nil || score > best.score {
				best = cand
			}
		} else if fallback == nil || score > fallback.score {
			fallback = cand
		}
	}
	if best != nil {
		return best
	}
	return fallback
}

func pdfCoverImageScore(img model.Image, byteLen int) int64 {
	area := int64(img.Width) * int64(img.Height)
	if area <= 0 {
		area = int64(byteLen)
	}
	score := area
	if img.Thumb {
		score /= 2
	}
	return score
}

func readPDFImageBytes(img model.Image) ([]byte, error) {
	if img.Reader == nil {
		return nil, io.ErrUnexpectedEOF
	}
	if rs, ok := img.Reader.(io.ReadSeeker); ok {
		if _, err := rs.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	}
	return io.ReadAll(img.Reader)
}

func extractPDFCoverPdftoppm(filePath string) []byte {
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		return nil
	}
	prefix := filepath.Join(os.TempDir(), "athenaeum-cover-"+strconv.FormatInt(time.Now().UnixNano(), 10))
	cmd := exec.Command( // #nosec G204 -- fixed pdftoppm flags, user library path as input
		"pdftoppm",
		"-f", "1",
		"-l", "1",
		"-jpeg",
		"-singlefile",
		"-r", "150",
		filePath,
		prefix,
	)
	if err := cmd.Run(); err != nil {
		return nil
	}
	outPath := prefix + ".jpg"
	data, err := os.ReadFile(outPath) // #nosec G304 -- temp file we created
	_ = os.Remove(outPath)
	if err != nil || len(data) < 256 {
		return nil
	}
	return data
}
