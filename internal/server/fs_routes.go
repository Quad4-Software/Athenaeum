package server

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"athenaeum/internal/library"
	"athenaeum/internal/models"
)

func (s *Server) registerFSRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/fs/browse", s.handleFSBrowse)
}

func (s *Server) handleFSBrowse(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermManageLibrary); !ok {
		return
	}
	path := r.URL.Query().Get("path")
	result, err := library.BrowseDirs(s.fsRoots(r.Context()), path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) fsRoots(ctx context.Context) []string {
	seen := map[string]struct{}{}
	var roots []string
	add := func(p string) {
		p = filepath.Clean(p)
		if p == "" || p == "." {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		st, err := os.Stat(p)
		if err != nil || !st.IsDir() {
			return
		}
		seen[p] = struct{}{}
		roots = append(roots, p)
	}

	add(s.cfg.LibraryDir)
	if home, err := os.UserHomeDir(); err == nil {
		add(home)
	}
	if libs, err := s.store.ListLibraries(ctx); err == nil {
		for _, lib := range libs {
			add(lib.MountPath)
		}
	}
	return roots
}
