package server

import (
	"net/http"

	"athenaeum/internal/system"
	"athenaeum/internal/version"
)

func (s *Server) registerSystemRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/system/stats", s.handleSystemStats)
}

func (s *Server) handleSystemStats(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	paths := []string{s.cfg.DataDir, s.cfg.LibraryDir}
	libs, err := s.store.ListLibraries(r.Context())
	if err == nil {
		seen := map[string]struct{}{s.cfg.DataDir: {}, s.cfg.LibraryDir: {}}
		for _, lib := range libs {
			if lib.MountPath == "" {
				continue
			}
			if _, ok := seen[lib.MountPath]; ok {
				continue
			}
			seen[lib.MountPath] = struct{}{}
			paths = append(paths, lib.MountPath)
		}
	}
	stats := system.ReadStats(paths)
	stats.Version = version.Version
	stats.WebVersion = version.WebVersion
	writeJSON(w, http.StatusOK, stats)
}
