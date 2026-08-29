package library

import (
	"path/filepath"
	"strings"

	"athenaeum/internal/models"
)

// FormatFromExt maps a file extension to a stored format constant.
func FormatFromExt(p string) string {
	switch strings.ToLower(filepath.Ext(p)) {
	case ".epub":
		return models.FormatEPUB
	case ".pdf":
		return models.FormatPDF
	case ".mobi":
		return models.FormatMOBI
	case ".azw3":
		return models.FormatAZW3
	case ".azw":
		return models.FormatAZW
	case ".kfx":
		return models.FormatKFX
	case ".cbz":
		return models.FormatCBZ
	case ".cbr":
		return models.FormatCBR
	case ".mp3":
		return models.FormatMP3
	case ".m4b":
		return models.FormatM4B
	case ".m4a":
		return models.FormatM4A
	case ".ogg":
		return models.FormatOGG
	case ".flac":
		return models.FormatFLAC
	default:
		return ""
	}
}
