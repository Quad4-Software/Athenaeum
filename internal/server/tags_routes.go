package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func (s *Server) registerTagRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/tags", s.handleListTags)
	mux.HandleFunc("POST /api/tags", s.handleCreateTag)
	mux.HandleFunc("GET /api/books/{id}/tags", s.handleListBookTags)
	mux.HandleFunc("PUT /api/books/{id}/tags", s.handleSetBookTags)
	mux.HandleFunc("POST /api/books/{id}/tags", s.handleAddBookTag)
	mux.HandleFunc("DELETE /api/books/{id}/tags/{tagId}", s.handleRemoveBookTag)
	mux.HandleFunc("GET /api/books/{id}/rating", s.handleGetRating)
	mux.HandleFunc("PUT /api/books/{id}/rating", s.handleSetRating)
}

func (s *Server) handleListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.store.ListTags(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if tags == nil {
		tags = []models.Tag{}
	}
	writeJSON(w, http.StatusOK, tags)
}

func (s *Server) handleCreateTag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, errors.New("name required"))
		return
	}
	tag, err := s.store.CreateTag(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, tag)
}

func (s *Server) handleListBookTags(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	names, err := s.store.ListBookTags(r.Context(), bookID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if names == nil {
		names = []string{}
	}
	writeJSON(w, http.StatusOK, names)
}

func (s *Server) handleSetBookTags(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	var req struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	names, err := s.store.SetBookTags(r.Context(), bookID, req.Tags)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if names == nil {
		names = []string{}
	}
	writeJSON(w, http.StatusOK, names)
}

func (s *Server) handleAddBookTag(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, errors.New("name required"))
		return
	}
	names, err := s.store.AddBookTag(r.Context(), bookID, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, names)
}

func (s *Server) handleRemoveBookTag(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	tagID, err := strconv.ParseInt(r.PathValue("tagId"), 10, 64)
	if err != nil || tagID <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid tag id"))
		return
	}
	if err := s.store.RemoveBookTag(r.Context(), bookID, tagID); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleGetRating(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	userID := UserIDFromContext(r.Context())
	rating, err := s.store.GetRating(r.Context(), userID, bookID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, rating)
}

func (s *Server) handleSetRating(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.bookByIDChecked(w, r); err != nil {
		return
	}
	var req struct {
		Rating int `json:"rating"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	userID := UserIDFromContext(r.Context())
	rating, err := s.store.SetRating(r.Context(), userID, bookID, req.Rating)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, rating)
}
