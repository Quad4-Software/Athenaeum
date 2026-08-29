package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func (s *Server) registerBookmarkRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/books/{id}/bookmarks", s.handleListBookmarks)
	mux.HandleFunc("POST /api/books/{id}/bookmarks", s.handleCreateBookmark)
	mux.HandleFunc("DELETE /api/books/{id}/bookmarks/{bookmarkId}", s.handleDeleteBookmark)
	mux.HandleFunc("GET /api/books/{id}/highlights", s.handleListHighlights)
	mux.HandleFunc("POST /api/books/{id}/highlights", s.handleCreateHighlight)
	mux.HandleFunc("DELETE /api/books/{id}/highlights/{highlightId}", s.handleDeleteHighlight)
	mux.HandleFunc("POST /api/books/{id}/reading-time", s.handleAddReadingTime)
	mux.HandleFunc("GET /api/stats/reading", s.handleReadingStats)
}

func (s *Server) handleListBookmarks(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	userID := UserIDFromContext(r.Context())
	items, err := s.store.ListBookmarks(r.Context(), userID, bookID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateBookmark(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	var b models.Bookmark
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&b); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if b.Location == "" {
		writeError(w, http.StatusBadRequest, errors.New("location required"))
		return
	}
	b.BookID = bookID
	userID := UserIDFromContext(r.Context())
	id, err := s.store.CreateBookmark(r.Context(), userID, b)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	b.ID = id
	b.UserID = userID
	writeJSON(w, http.StatusCreated, b)
}

func (s *Server) handleDeleteBookmark(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	bmID, err := strconv.ParseInt(r.PathValue("bookmarkId"), 10, 64)
	if err != nil || bmID <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid bookmark id"))
		return
	}
	userID := UserIDFromContext(r.Context())
	if err := s.store.DeleteBookmark(r.Context(), userID, bmID); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = bookID
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleListHighlights(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	userID := UserIDFromContext(r.Context())
	items, err := s.store.ListHighlights(r.Context(), userID, bookID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateHighlight(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	var h models.Highlight
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&h); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if h.Location == "" {
		writeError(w, http.StatusBadRequest, errors.New("location required"))
		return
	}
	h.BookID = bookID
	userID := UserIDFromContext(r.Context())
	id, err := s.store.CreateHighlight(r.Context(), userID, h)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	h.ID = id
	h.UserID = userID
	writeJSON(w, http.StatusCreated, h)
}

func (s *Server) handleDeleteHighlight(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	hlID, err := strconv.ParseInt(r.PathValue("highlightId"), 10, 64)
	if err != nil || hlID <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid highlight id"))
		return
	}
	userID := UserIDFromContext(r.Context())
	if err := s.store.DeleteHighlight(r.Context(), userID, hlID); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = bookID
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleAddReadingTime(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	var req struct {
		Seconds int64 `json:"seconds"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	userID := UserIDFromContext(r.Context())
	if err := s.store.AddReadSeconds(r.Context(), userID, bookID, req.Seconds); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleReadingStats(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	st, err := s.store.ReadingStats(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}
