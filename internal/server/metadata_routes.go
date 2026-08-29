package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"athenaeum/internal/library"
	"athenaeum/internal/models"
)

func (s *Server) registerMetadataJobRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/library/metadata/match", s.handleMetadataAutoMatch)
	mux.HandleFunc("GET /api/library/metadata/match/status", s.handleMetadataMatchStatus)
	mux.HandleFunc("POST /api/library/series/cleanup", s.handleSeriesCleanup)
}

func (s *Server) handleMetadataAutoMatch(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermEditMetadata); !ok {
		return
	}
	var req library.MetadataAutoMatchRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<18)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if !s.metadataMatcher.Start(s.jobsCtx, req) {
		writeError(w, http.StatusConflict, errors.New("metadata match already running"))
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]bool{"ok": true})
}

func (s *Server) handleMetadataMatchStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.metadataMatcher.Status())
}

func (s *Server) handleSeriesCleanup(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermManageLibrary); !ok {
		return
	}
	updated, err := library.CleanStoredSeriesNames(r.Context(), s.store)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"updated": updated})
}
