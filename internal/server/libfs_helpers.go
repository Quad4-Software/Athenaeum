package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"athenaeum/internal/libfs"
)

func (s *Server) openLibraryFS(ctx context.Context, libraryID int64) (libfs.LibraryFS, error) {
	s.libFSMu.Lock()
	if fs, ok := s.libFS[libraryID]; ok {
		s.libFSMu.Unlock()
		return fs, nil
	}
	s.libFSMu.Unlock()

	fs, err := s.store.OpenLibraryFS(ctx, libraryID)
	if err != nil {
		return nil, err
	}

	s.libFSMu.Lock()
	if existing, ok := s.libFS[libraryID]; ok {
		s.libFSMu.Unlock()
		return existing, nil
	}
	s.libFS[libraryID] = fs
	s.libFSMu.Unlock()
	return fs, nil
}

func (s *Server) invalidateLibraryFS(libraryID int64) {
	s.libFSMu.Lock()
	delete(s.libFS, libraryID)
	s.libFSMu.Unlock()
}

func (s *Server) serveLibraryFile(w http.ResponseWriter, r *http.Request, libraryID int64, relPath, filename, ctype string, attachment bool) {
	fs, err := s.openLibraryFS(r.Context(), libraryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	info, err := fs.Stat(r.Context(), relPath)
	if err != nil {
		if errors.Is(err, libfs.ErrNotExist) {
			writeError(w, http.StatusNotFound, errors.New("file missing on disk"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	f, err := fs.Open(r.Context(), relPath)
	if err != nil {
		if errors.Is(err, libfs.ErrNotExist) {
			writeError(w, http.StatusNotFound, errors.New("file missing on disk"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Accept-Ranges", "bytes")
	if attachment {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	}
	mod := info.ModTime
	if mod.IsZero() {
		mod = time.Now()
	}
	http.ServeContent(w, r, filename, mod, f)
}

// materializeBookFile returns a local filesystem path for path-based tools.
// cleanup must be called when finished (no-op for local mounts).
func (s *Server) materializeBookFile(ctx context.Context, libraryID int64, relPath string) (path string, cleanup func(), err error) {
	fs, err := s.openLibraryFS(ctx, libraryID)
	if err != nil {
		return "", nil, err
	}
	noop := func() {}
	if fs.Backend() == libfs.BackendLocal {
		full, err := fs.LocalAbsPath(relPath)
		if err != nil {
			return "", nil, err
		}
		if _, err := os.Stat(full); err != nil {
			return "", nil, err
		}
		return full, noop, nil
	}
	tmp, err := libfs.Materialize(ctx, fs, relPath, s.cfg.TempDir())
	if err != nil {
		return "", nil, err
	}
	return tmp, func() { _ = os.Remove(tmp) }, nil
}

func copyReaderToLibrary(ctx context.Context, fs libfs.LibraryFS, relPath string, r io.Reader, size int64) error {
	if parent := parentRel(relPath); parent != "" {
		if err := fs.MkdirAll(ctx, parent); err != nil {
			return err
		}
	}
	return fs.Write(ctx, relPath, r, size)
}

func parentRel(relPath string) string {
	for i := len(relPath) - 1; i >= 0; i-- {
		if relPath[i] == '/' {
			return relPath[:i]
		}
	}
	return ""
}
