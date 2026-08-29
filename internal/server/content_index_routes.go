package server

import (
	"archive/zip"
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"athenaeum/internal/models"
)

var htmlTagRe = regexp.MustCompile(`(?is)<[^>]*>`)
var spaceRe = regexp.MustCompile(`\s+`)

func (s *Server) registerContentIndexRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/admin/content-index", s.handleContentIndex)
}

func (s *Server) handleContentIndex(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	go s.runContentIndex()
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "started"})
}

func (s *Server) runContentIndex() {
	ctx := context.Background()
	page, err := s.store.ListBooks(ctx, models.BookQuery{Format: models.FormatEPUB, Limit: 5000})
	if err != nil {
		return
	}
	for _, book := range page.Items {
		if book.Format != models.FormatEPUB {
			continue
		}
		path, cleanup, err := s.materializeBookFile(ctx, book.LibraryID, book.RelPath)
		if err != nil {
			continue
		}
		chunks, err := extractEPUBText(path)
		cleanup()
		if err != nil || len(chunks) == 0 {
			continue
		}
		_ = s.store.ReplaceBookContent(ctx, book.ID, chunks)
	}
}

func extractEPUBText(path string) ([]string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	var chunks []string
	var buf strings.Builder
	const maxChunks = 200
	const chunkSize = 2000

	flush := func() {
		text := strings.TrimSpace(buf.String())
		buf.Reset()
		if text == "" {
			return
		}
		for len(text) > 0 && len(chunks) < maxChunks {
			if utf8.RuneCountInString(text) <= chunkSize {
				chunks = append(chunks, text)
				return
			}
			runes := []rune(text)
			chunks = append(chunks, string(runes[:chunkSize]))
			text = string(runes[chunkSize:])
		}
	}

	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		if !strings.HasSuffix(name, ".xhtml") && !strings.HasSuffix(name, ".html") && !strings.HasSuffix(name, ".htm") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		raw, err := io.ReadAll(io.LimitReader(rc, 2<<20))
		_ = rc.Close()
		if err != nil {
			continue
		}
		text := htmlTagRe.ReplaceAllString(string(raw), " ")
		text = spaceRe.ReplaceAllString(text, " ")
		buf.WriteString(text)
		buf.WriteByte(' ')
		if buf.Len() > chunkSize*4 {
			flush()
		}
		if len(chunks) >= maxChunks {
			break
		}
	}
	flush()
	return chunks, nil
}
