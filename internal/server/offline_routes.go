package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) registerOfflineRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/offline", s.handleListOffline)
	mux.HandleFunc("POST /api/offline", s.handleAddOffline)
	mux.HandleFunc("DELETE /api/offline", s.handleRemoveOffline)
}

func (s *Server) handleListOffline(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	ids, err := s.store.ListOfflineGrants(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if ids == nil {
		ids = []int64{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"bookIds": ids})
}

func (s *Server) handleAddOffline(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BookIDs []int64 `json:"bookIds"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	userID := UserIDFromContext(r.Context())
	if err := s.store.AddOfflineGrants(r.Context(), userID, req.BookIDs); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	ids, _ := s.store.ListOfflineGrants(r.Context(), userID)
	writeJSON(w, http.StatusOK, map[string]any{"bookIds": ids})
}

func (s *Server) handleRemoveOffline(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BookIDs []int64 `json:"bookIds"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	userID := UserIDFromContext(r.Context())
	if err := s.store.RemoveOfflineGrants(r.Context(), userID, req.BookIDs); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	ids, _ := s.store.ListOfflineGrants(r.Context(), userID)
	writeJSON(w, http.StatusOK, map[string]any{"bookIds": ids})
}
