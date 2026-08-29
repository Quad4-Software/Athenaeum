package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"athenaeum/internal/storage"
)

func (s *Server) registerFavoriteRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/favorites", s.handleListFavorites)
	mux.HandleFunc("GET /api/books/{id}/favorite", s.handleGetFavorite)
	mux.HandleFunc("PUT /api/books/{id}/favorite", s.handleSetFavorite)
}

func (s *Server) handleListFavorites(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	ids, err := s.store.ListFavoriteIDs(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if ids == nil {
		ids = []int64{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ids": ids})
}

func (s *Server) handleGetFavorite(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	userID := UserIDFromContext(r.Context())
	fav, err := s.store.IsFavorite(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("book not found"))
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"favorite": fav})
}

func (s *Server) handleSetFavorite(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body struct {
		Favorite bool `json:"favorite"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	userID := UserIDFromContext(r.Context())
	if err := s.store.SetFavorite(r.Context(), userID, id, body.Favorite); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("book not found"))
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"favorite": body.Favorite})
}
