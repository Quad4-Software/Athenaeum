package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func (s *Server) registerBookRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/books", s.handleListBooks)
	mux.HandleFunc("GET /api/books/{id}", s.handleGetBook)
	mux.HandleFunc("GET /api/books/{id}/cover", s.handleCover)
	mux.HandleFunc("GET /api/books/{id}/file", s.handleFile)
	mux.HandleFunc("GET /api/books/{id}/download", s.handleDownload)
	mux.HandleFunc("GET /api/books/{id}/progress", s.handleGetProgress)
	mux.HandleFunc("PUT /api/books/{id}/progress", s.handlePutProgress)
	mux.HandleFunc("GET /api/books/{id}/chapters", s.handleGetChapters)
}

func (s *Server) handleGetChapters(w http.ResponseWriter, r *http.Request) {
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	chapters, err := s.store.ListChapters(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if chapters == nil {
		chapters = []models.Chapter{}
	}
	writeJSON(w, http.StatusOK, chapters)
}

func (s *Server) handleListBooks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	query := models.BookQuery{
		Search:       q.Get("search"),
		Sort:         q.Get("sort"),
		Format:       q.Get("format"),
		Series:       q.Get("series"),
		Author:       q.Get("author"),
		LibraryID:    int64(atoiDefault(q.Get("library"), 0)),
		CollectionID: int64(atoiDefault(q.Get("collection"), 0)),
		UserID:       UserIDFromContext(r.Context()),
		Favorites:    q.Get("favorites") == "1" || q.Get("favorites") == "true",
		InProgress:   q.Get("inProgress") == "1" || q.Get("inProgress") == "true",
		Tag:          q.Get("tag"),
		Limit:        atoiDefault(q.Get("limit"), 60),
		Offset:       atoiDefault(q.Get("offset"), 0),
	}
	var err error
	query, err = s.applyBookAccess(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	page, err := s.store.ListBooks(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.attachTagsAndRatings(r.Context(), page.Items, query.UserID)
	writeJSON(w, http.StatusOK, page)
}

func (s *Server) attachTagsAndRatings(ctx context.Context, items []models.Book, userID int64) {
	if len(items) == 0 {
		return
	}
	ids := make([]int64, len(items))
	for i, b := range items {
		ids[i] = b.ID
	}
	tags, err := s.store.ListBookTagsBatch(ctx, ids)
	if err == nil {
		for i := range items {
			items[i].Tags = tags[items[i].ID]
		}
	}
	if userID <= 0 {
		return
	}
	ratings, err := s.store.RatingsBatch(ctx, userID, ids)
	if err == nil {
		for i := range items {
			items[i].UserRating = ratings[items[i].ID]
		}
	}
}

func (s *Server) handleGetBook(w http.ResponseWriter, r *http.Request) {
	book, err := s.bookByIDChecked(w, r)
	if err != nil {
		return
	}
	items := []models.Book{book}
	s.attachTagsAndRatings(r.Context(), items, UserIDFromContext(r.Context()))
	writeJSON(w, http.StatusOK, items[0])
}

func (s *Server) handleCover(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	book, err := s.store.GetBook(r.Context(), id)
	if errors.Is(err, storage.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !s.requireBookAccess(w, r, book) {
		return
	}
	f, info, err := openCoverFile(s.cfg.CoverDir(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, "cover", info.ModTime(), f)
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	s.serveBookFile(w, r, false)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	s.serveBookFile(w, r, true)
}

func (s *Server) serveBookFile(w http.ResponseWriter, r *http.Request, attachment bool) {
	book, err := s.bookByIDChecked(w, r)
	if err != nil {
		return
	}
	relPath := book.RelPath
	format := book.Format
	if book.Format == models.FormatAudiobook {
		tracks, err := s.store.ListAudiobookTracks(r.Context(), book.ID)
		if err != nil || len(tracks) == 0 {
			writeError(w, http.StatusNotFound, errors.New("audiobook tracks missing"))
			return
		}
		idx := atoiDefault(r.URL.Query().Get("track"), 0)
		if idx < 0 || idx >= len(tracks) {
			idx = 0
		}
		relPath = tracks[idx].RelPath
		format = tracks[idx].Format
	}
	s.serveLibraryFile(w, r, book.LibraryID, relPath, safeFilename(book), contentType(format), attachment)
}

func (s *Server) handleGetProgress(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	userID := UserIDFromContext(r.Context())
	p, err := s.store.GetProgress(r.Context(), userID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handlePutProgress(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var p models.Progress
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	p.BookID = id
	userID := UserIDFromContext(r.Context())
	if err := s.store.SaveProgress(r.Context(), userID, p); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	p.UserID = userID
	p.UpdatedAt = time.Now()
	writeJSON(w, http.StatusOK, p)
}
