package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"athenaeum/internal/library"
)

func (s *Server) registerAdminTaskRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admin/tasks/status", s.handleMaintenanceStatus)
	mux.HandleFunc("POST /api/admin/tasks/verify", s.handleVerifyIntegrity)
	mux.HandleFunc("POST /api/admin/tasks/prune-missing", s.handlePruneMissing)
	mux.HandleFunc("POST /api/admin/tasks/cleanup-covers", s.handleCleanupOrphanCovers)
	mux.HandleFunc("POST /api/admin/tasks/regenerate-covers", s.handleRegenerateCovers)
	mux.HandleFunc("POST /api/admin/tasks/cleanup-series", s.handleAdminCleanupSeries)
	mux.HandleFunc("POST /api/admin/tasks/cleanup-text", s.handleAdminCleanupText)
}

func (s *Server) handleMaintenanceStatus(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	writeJSON(w, http.StatusOK, s.maintenance.Status())
}

func (s *Server) handleVerifyIntegrity(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	report, err := library.VerifyIntegrity(r.Context(), s.store, s.cfg.CoverDir())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handlePruneMissing(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	removed, err := library.PruneMissingBooks(r.Context(), s.store)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"removed": removed})
}

func (s *Server) handleCleanupOrphanCovers(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	removed, err := library.CleanupOrphanCovers(r.Context(), s.store, s.cfg.CoverDir())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"removed": removed})
}

func (s *Server) handleRegenerateCovers(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	var req struct {
		LibraryID int64 `json:"libraryId"`
	}
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
	}
	if !s.maintenance.StartRegenerateCovers(s.jobsCtx, req.LibraryID) {
		writeError(w, http.StatusConflict, errors.New("maintenance task already running"))
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]bool{"ok": true})
}

func (s *Server) handleAdminCleanupSeries(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	updated, err := library.CleanStoredSeriesNames(r.Context(), s.store)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"updated": updated})
}

func (s *Server) handleAdminCleanupText(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	updated, err := library.CleanStoredBookText(r.Context(), s.store)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"updated": updated})
}
