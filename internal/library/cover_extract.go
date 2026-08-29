package library

import (
	"path/filepath"

	"athenaeum/internal/models"
)

// ExtractCoverFromFile reads embedded or sidecar cover art from a book file.
func ExtractCoverFromFile(absPath, format string) []byte {
	switch format {
	case models.FormatEPUB:
		if meta, err := parseEPUB(absPath); err == nil {
			return meta.CoverData
		}
	case models.FormatPDF:
		return parsePDF(absPath).CoverData
	case models.FormatMOBI, models.FormatAZW3, models.FormatAZW:
		return parseMobiFamily(absPath).CoverData
	case models.FormatCBZ, models.FormatCBR:
		return parseComic(absPath).CoverData
	default:
		if models.IsAudio(format) {
			meta := parseAudio(absPath)
			if len(meta.CoverData) > 0 {
				return meta.CoverData
			}
			return sidecarCover(absPath)
		}
	}
	if side := sidecarCover(absPath); len(side) > 0 {
		return side
	}
	return nil
}

// ResolveBookAbsPath returns the on-disk path for a catalog book.
func ResolveBookAbsPath(mount, relPath string) string {
	return filepath.Join(mount, relPath)
}
