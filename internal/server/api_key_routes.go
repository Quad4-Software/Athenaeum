package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

type createAPIKeyRequest struct {
	Name string `json:"name"`
}

func (s *Server) registerAPIKeyRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/api-keys", s.handleListAPIKeys)
	mux.HandleFunc("POST /api/auth/api-keys", s.handleCreateAPIKey)
	mux.HandleFunc("DELETE /api/auth/api-keys/{id}", s.handleDeleteAPIKey)
	mux.HandleFunc("GET /api/docs", s.handleAPIDocs)
	mux.HandleFunc("GET /api/openapi.json", s.handleOpenAPI)
	mux.HandleFunc("GET /docs", s.handleDocsUI)
	mux.HandleFunc("GET /docs/app.js", s.handleDocsAppJS)
}

func (s *Server) handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, errUnauthorized)
		return
	}
	targetID := u.ID
	if u.IsAdmin {
		if q := r.URL.Query().Get("userId"); q != "" {
			if id, err := strconv.ParseInt(q, 10, 64); err == nil && id > 0 {
				targetID = id
			}
		}
	}
	keys, err := s.store.ListAPIKeys(r.Context(), targetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if keys == nil {
		keys = []models.APIKey{}
	}
	writeJSON(w, http.StatusOK, keys)
}

func (s *Server) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, errUnauthorized)
		return
	}
	var req createAPIKeyRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	name := stringsTrim(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, errors.New("name is required"))
		return
	}
	if len(name) > 128 {
		writeError(w, http.StatusBadRequest, errors.New("name too long"))
		return
	}
	created, err := s.store.CreateAPIKey(r.Context(), u.ID, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, u.ID, u.Username, u.ID, u.Username, "apikey.create", created.Prefix)
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, errUnauthorized)
		return
	}
	keyID, ok := pathID(w, r)
	if !ok {
		return
	}
	var err error
	if u.IsAdmin && r.URL.Query().Get("userId") != "" {
		err = s.store.DeleteAPIKeyAdmin(r.Context(), keyID)
	} else {
		err = s.store.DeleteAPIKey(r.Context(), u.ID, keyID)
	}
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, errors.New("api key not found"))
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, u.ID, u.Username, u.ID, u.Username, "apikey.revoke", strconv.FormatInt(keyID, 10))
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, apiDocumentation())
}

func stringsTrim(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\n' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}
