package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func (s *Server) registerKosyncRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /kosync/users/auth", s.handleKosyncAuth)
	mux.HandleFunc("PUT /kosync/syncs/progress", s.handleKosyncPutProgress)
	mux.HandleFunc("GET /kosync/syncs/progress/{document}", s.handleKosyncGetProgress)
}

func (s *Server) kosyncUser(w http.ResponseWriter, r *http.Request) (models.User, bool) {
	user, pass, ok := r.BasicAuth()
	if !ok || user == "" {
		w.Header().Set("WWW-Authenticate", `Basic realm="kosync"`)
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return models.User{}, false
	}
	u, hash, err := s.store.GetUserByUsername(r.Context(), user)
	if err != nil || !auth.CheckPassword(hash, pass) {
		writeError(w, http.StatusUnauthorized, errors.New("invalid credentials"))
		return models.User{}, false
	}
	return u, true
}

func (s *Server) handleKosyncAuth(w http.ResponseWriter, r *http.Request) {
	u, ok := s.kosyncUser(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"username": u.Username,
	})
}

func (s *Server) handleKosyncPutProgress(w http.ResponseWriter, r *http.Request) {
	u, ok := s.kosyncUser(w, r)
	if !ok {
		return
	}
	var req struct {
		Document   string  `json:"document"`
		Progress   string  `json:"progress"`
		Percentage float64 `json:"percentage"`
		Device     string  `json:"device"`
		DeviceID   string  `json:"device_id"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Document == "" {
		writeError(w, http.StatusBadRequest, errors.New("document required"))
		return
	}
	ts := time.Now().Unix()
	if err := s.store.SaveKosyncProgress(r.Context(), u.ID, req.Document, req.Progress, req.Percentage, req.Device, req.DeviceID, ts); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"document": req.Document, "timestamp": ts})
}

func (s *Server) handleKosyncGetProgress(w http.ResponseWriter, r *http.Request) {
	u, ok := s.kosyncUser(w, r)
	if !ok {
		return
	}
	doc := r.PathValue("document")
	d, err := s.store.GetKosyncProgress(r.Context(), u.ID, doc)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}
