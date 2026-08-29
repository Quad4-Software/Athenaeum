package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func (s *Server) registerShareRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/books/{id}/share", s.handleCreateShare)
	mux.HandleFunc("GET /api/books/{id}/share", s.handleListShares)
	mux.HandleFunc("DELETE /api/books/{id}/share/{shareId}", s.handleDeleteShare)
	mux.HandleFunc("GET /api/share/{token}", s.handleShareMeta)
	mux.HandleFunc("GET /share/{token}/download", s.handleShareDownload)
}

func (s *Server) handleCreateShare(w http.ResponseWriter, r *http.Request) {
	book, err := s.bookByIDChecked(w, r)
	if err != nil {
		return
	}
	var req struct {
		ExpiresInHours int   `json:"expiresInHours"`
		MaxDownloads   int64 `json:"maxDownloads"`
	}
	_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<12)).Decode(&req)
	var expires *time.Time
	if req.ExpiresInHours > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresInHours) * time.Hour)
		expires = &t
	}
	userID := UserIDFromContext(r.Context())
	sl, err := s.store.CreateShareLink(r.Context(), book.ID, userID, expires, req.MaxDownloads)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	base := s.requestBaseURL(r)
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":        sl.ID,
		"token":     sl.Token,
		"url":       base + "/share/" + sl.Token + "/download",
		"expiresAt": sl.ExpiresAt,
	})
}

func (s *Server) handleListShares(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	items, err := s.store.ListSharesForBook(r.Context(), bookID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if items == nil {
		items = []models.ShareLink{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleDeleteShare(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	shareID, err := strconv.ParseInt(r.PathValue("shareId"), 10, 64)
	if err != nil || shareID <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid share id"))
		return
	}
	if err := s.store.DeleteShareLink(r.Context(), bookID, shareID); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleShareMeta(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	sl, err := s.store.GetShareLinkByToken(r.Context(), token)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if sl.ExpiresAt != nil && time.Now().After(*sl.ExpiresAt) {
		writeError(w, http.StatusGone, errors.New("share link expired"))
		return
	}
	book, err := s.store.GetBook(r.Context(), sl.BookID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, models.ShareLinkMeta{
		Token:     sl.Token,
		BookTitle: book.Title,
		Author:    book.Author,
		Format:    book.Format,
		FileSize:  book.FileSize,
		ExpiresAt: sl.ExpiresAt,
	})
}

func (s *Server) handleShareDownload(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	sl, err := s.store.GetShareLinkByToken(r.Context(), token)
	if errors.Is(err, storage.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if sl.ExpiresAt != nil && time.Now().After(*sl.ExpiresAt) {
		writeError(w, http.StatusGone, errors.New("share link expired"))
		return
	}
	ok, err := s.store.TryIncrementShareDownload(r.Context(), sl.ID, sl.MaxDownloads)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !ok {
		writeError(w, http.StatusGone, errors.New("download limit reached"))
		return
	}
	book, err := s.store.GetBook(r.Context(), sl.BookID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	s.serveLibraryFile(w, r, book.LibraryID, book.RelPath, safeFilename(book), contentType(book.Format), true)
}
