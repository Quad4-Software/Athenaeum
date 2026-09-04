package server

import (
	"net/http"
	"time"

	"athenaeum/internal/models"
)

func (s *Server) registerLibraryScanRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/library/stats", s.handleStats)
	mux.HandleFunc("GET /api/library/scan/status", s.handleScanStatus)
	mux.HandleFunc("POST /api/library/scan", s.handleScan)
	mux.HandleFunc("GET /api/series", s.handleListSeries)
	mux.HandleFunc("GET /api/authors", s.handleListAuthors)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	libraryID := int64(atoiDefault(r.URL.Query().Get("library"), 0))
	stats, err := s.store.Stats(r.Context(), libraryID, UserIDFromContext(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if u, ok := UserFromContext(r.Context()); !ok || !u.IsAdmin {
		stats.UserCount = 0
	}
	scan := s.scanner.Status()
	stats.Scanning = scan.Scanning
	if scan.FinishedAt != nil {
		stats.LastScanAt = scan.FinishedAt.Format(time.RFC3339)
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handleScanStatus(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermManageLibrary); !ok {
		return
	}
	writeJSON(w, http.StatusOK, s.scanner.Status())
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermManageLibrary); !ok {
		return
	}
	go func() {
		if err := s.scanner.Scan(s.jobsCtx); err != nil {
			s.log.Error("background scan failed", "err", err)
		}
	}()
	writeJSON(w, http.StatusAccepted, map[string]bool{"started": true})
}

func (s *Server) handleListSeries(w http.ResponseWriter, r *http.Request) {
	libraryID := int64(atoiDefault(r.URL.Query().Get("library"), 0))
	libID, libIDs, err := s.libraryFilterIDs(r.Context(), libraryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	series, err := s.store.ListSeries(r.Context(), libID, libIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if series == nil {
		series = []models.SeriesInfo{}
	}
	writeJSON(w, http.StatusOK, series)
}

func (s *Server) handleListAuthors(w http.ResponseWriter, r *http.Request) {
	libraryID := int64(atoiDefault(r.URL.Query().Get("library"), 0))
	libID, libIDs, err := s.libraryFilterIDs(r.Context(), libraryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	authors, err := s.store.ListAuthors(r.Context(), libID, libIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if authors == nil {
		authors = []models.AuthorInfo{}
	}
	writeJSON(w, http.StatusOK, authors)
}
