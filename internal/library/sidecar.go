package library

import (
	"os"
	"path/filepath"
	"strings"
)

var sidecarNames = []string{
	"cover.jpg", "cover.jpeg", "cover.png", "cover.webp",
	"folder.jpg", "folder.jpeg", "folder.png",
	"artwork.jpg", "artwork.jpeg", "artwork.png",
	"Front.jpg", "Front.jpeg", "Front.png",
}

// sidecarCover returns image bytes from a companion file next to the book.
func sidecarCover(filePath string) []byte {
	dir := filepath.Dir(filePath)
	base := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	candidates := append([]string{
		base + ".jpg", base + ".jpeg", base + ".png", base + ".webp",
	}, sidecarNames...)

	for _, name := range candidates {
		p := filepath.Join(dir, name)
		data, err := os.ReadFile(p) // #nosec G304 -- library-relative path
		if err != nil || len(data) < 256 {
			continue
		}
		if isImageBytes(data) {
			return data
		}
	}
	return nil
}

func isImageBytes(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	if data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff {
		return true
	}
	pngSig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if len(data) >= len(pngSig) && string(data[:len(pngSig)]) == string(pngSig) {
		return true
	}
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return true
	}
	return false
}
