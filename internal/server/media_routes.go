package server

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"athenaeum/internal/library"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func (s *Server) registerMediaRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/books/{id}/pages", s.handleComicManifest)
	mux.HandleFunc("GET /api/books/{id}/pages/{page}", s.handleComicPage)
	mux.HandleFunc("GET /api/books/{id}/mobi-sections", s.handleMobiSections)
	mux.HandleFunc("GET /api/books/{id}/tracks", s.handleAudiobookTracks)
	mux.HandleFunc("POST /api/books/{id}/convert", s.handleConvertBook)
}

func (s *Server) handleComicManifest(w http.ResponseWriter, r *http.Request) {
	book, err := s.bookByID(w, r)
	if err != nil {
		return
	}
	if !models.IsComic(book.Format) {
		writeError(w, http.StatusBadRequest, errors.New("not a comic"))
		return
	}
	path, cleanup, err := s.materializeBookFile(r.Context(), book.LibraryID, book.RelPath)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	defer cleanup()
	total, pages := library.ComicManifest(path)
	out := models.ComicManifest{Total: total, Pages: make([]models.ComicPage, len(pages))}
	for i, p := range pages {
		out.Pages[i] = models.ComicPage{Index: i, Name: p.Name, MimeType: p.Mime}
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleComicPage(w http.ResponseWriter, r *http.Request) {
	book, err := s.bookByID(w, r)
	if err != nil {
		return
	}
	if !models.IsComic(book.Format) {
		writeError(w, http.StatusBadRequest, errors.New("not a comic"))
		return
	}
	idx, err := strconv.Atoi(r.PathValue("page"))
	if err != nil || idx < 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid page"))
		return
	}
	path, cleanup, err := s.materializeBookFile(r.Context(), book.LibraryID, book.RelPath)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	defer cleanup()
	data, mime, err := library.OpenComicPage(path, idx)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	_, _ = w.Write(data)
}

func (s *Server) handleMobiSections(w http.ResponseWriter, r *http.Request) {
	book, err := s.bookByID(w, r)
	if err != nil {
		return
	}
	if !models.IsMobiFamily(book.Format) {
		writeError(w, http.StatusBadRequest, errors.New("not a mobi family book"))
		return
	}
	path, cleanup, err := s.materializeBookFile(r.Context(), book.LibraryID, book.RelPath)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	defer cleanup()
	secs := library.MobiSections(path)
	out := make([]models.MobiSection, len(secs))
	for i, sec := range secs {
		out[i] = models.MobiSection{Index: i, Title: sec.Title, HTML: sec.HTML}
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleAudiobookTracks(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	book, err := s.store.GetBook(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("book not found"))
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	if book.Format != models.FormatAudiobook {
		writeError(w, http.StatusBadRequest, errors.New("not a multi-file audiobook"))
		return
	}
	tracks, err := s.store.ListAudiobookTracks(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if tracks == nil {
		tracks = []models.AudiobookTrack{}
	}
	writeJSON(w, http.StatusOK, tracks)
}

func (s *Server) handleConvertBook(w http.ResponseWriter, r *http.Request) {
	book, err := s.bookByID(w, r)
	if err != nil {
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	target := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("target")))
	if target == "" {
		target = "epub"
	}
	if target != "epub" && target != "pdf" {
		writeError(w, http.StatusBadRequest, errors.New("target must be epub or pdf"))
		return
	}
	srcPath, cleanup, err := s.materializeBookFile(r.Context(), book.LibraryID, book.RelPath)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	defer cleanup()
	outDir := s.cfg.TempDir()
	dest, err := library.ConvertBook(srcPath, outDir, target)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	relOut := strings.TrimSuffix(book.RelPath, filepath.Ext(book.RelPath)) + "." + target
	fs, err := s.openLibraryFS(r.Context(), book.LibraryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	f, err := os.Open(dest) // #nosec G304 G703 -- dest from ConvertBook under TempDir
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := copyReaderToLibrary(r.Context(), fs, relOut, f, info.Size()); err != nil {
		_ = f.Close()
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = f.Close()
	_ = os.Remove(dest) // #nosec G703 -- dest from ConvertBook under TempDir

	bookID, idxErr := s.scanner.IndexFile(r.Context(), book.LibraryID, relOut)
	msg := "converted and indexed"
	if library.IsCalibreAvailable() {
		msg = "converted with calibre and indexed"
	}
	if idxErr != nil {
		s.log.Warn("convert index", "path", relOut, "err", idxErr)
		msg = "converted (index failed: " + idxErr.Error() + ")"
	}
	writeJSON(w, http.StatusOK, models.ConvertResult{
		TargetFormat: target,
		OutputPath:   relOut,
		BookID:       bookID,
		Message:      msg,
	})
}
