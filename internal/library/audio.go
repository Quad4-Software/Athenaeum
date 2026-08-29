package library

import (
	"os"
	"path/filepath"
	"strings"

	"athenaeum/internal/models"
	"github.com/dhowden/tag"
)

// audioMeta holds metadata extracted from audio file tags (ID3, Vorbis, MP4).
type audioMeta struct {
	Title     string
	Author    string
	Album     string
	ISBN      string
	ASIN      string
	CoverData []byte
	Chapters  []models.Chapter
}

// parseAudio reads embedded tags from an audiobook file, falling back to the
// filename when tags are missing or unreadable.
func parseAudio(filePath string) audioMeta {
	fallback := audioMeta{
		Title: titleFromFilename(strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))),
	}

	f, err := os.Open(filePath) // #nosec G304 -- path is library-relative from scanner
	if err != nil {
		return fallback
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return fallback
	}

	meta := audioMeta{
		Title:  firstNonEmpty(m.Title(), m.Album()),
		Author: firstNonEmpty(m.Artist(), m.AlbumArtist(), m.Composer()),
		Album:  m.Album(),
	}
	if meta.Title == "" {
		meta.Title = fallback.Title
	}

	if pic := m.Picture(); pic != nil && len(pic.Data) > 0 {
		meta.CoverData = pic.Data
	}
	meta.ISBN, meta.ASIN = parseTagIdentifiers(m)
	meta.Chapters = parseAudioChapters(filePath)
	return meta
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}
