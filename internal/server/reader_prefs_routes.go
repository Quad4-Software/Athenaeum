package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleGetReaderPrefs(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	prefs, err := s.store.GetReaderPrefs(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, prefs)
}

func (s *Server) handlePutReaderPrefs(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prefs map[string]any `json:"prefs"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	userID := UserIDFromContext(r.Context())
	prefs, err := s.store.SaveReaderPrefs(r.Context(), userID, req.Prefs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, prefs)
}
