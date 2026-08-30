package library

import (
	"os"
	"path/filepath"
	"strings"

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

// ResolveBookAbsPath returns a path under mount for relPath, or empty if
// the relative path would escape the mount.
func ResolveBookAbsPath(mount, relPath string) string {
	mount = filepath.Clean(mount)
	relPath = filepath.ToSlash(strings.TrimPrefix(filepath.Clean(filepath.FromSlash(relPath)), string(filepath.Separator)))
	if relPath == "." || relPath == "" {
		return mount
	}
	if relPath == ".." || strings.HasPrefix(relPath, "../") {
		return ""
	}
	full := filepath.Join(mount, filepath.FromSlash(relPath))
	rel, err := filepath.Rel(mount, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return ""
	}
	root := mount
	if resolved, err := filepath.EvalSymlinks(mount); err == nil {
		root = filepath.Clean(resolved)
	}
	check := full
	if _, err := os.Lstat(full); err != nil {
		check = filepath.Dir(full)
	}
	for {
		resolved, err := filepath.EvalSymlinks(check)
		if err == nil {
			resolved = filepath.Clean(resolved)
			if resolved != root && !strings.HasPrefix(resolved, root+string(filepath.Separator)) {
				return ""
			}
			return full
		}
		parent := filepath.Dir(check)
		if parent == check {
			return ""
		}
		check = parent
	}
}
