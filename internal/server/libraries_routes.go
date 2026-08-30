package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"athenaeum/internal/libfs"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func (s *Server) registerLibraryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/libraries", s.handleListLibraries)
	mux.HandleFunc("POST /api/libraries", s.handleCreateLibrary)
	mux.HandleFunc("POST /api/libraries/test-s3", s.handleTestS3)
	mux.HandleFunc("PUT /api/libraries/reorder", s.handleReorderLibraries)
	mux.HandleFunc("GET /api/libraries/{id}", s.handleGetLibrary)
	mux.HandleFunc("PUT /api/libraries/{id}", s.handleUpdateLibrary)
	mux.HandleFunc("DELETE /api/libraries/{id}", s.handleDeleteLibrary)
	mux.HandleFunc("POST /api/libraries/{id}/scan", s.handleScanLibrary)
}

func (s *Server) handleListLibraries(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	var libs []models.Library
	var err error
	if ok {
		libs, err = s.store.ListLibrariesForUser(r.Context(), user)
	} else {
		libs, err = s.store.ListLibraries(r.Context())
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if libs == nil {
		libs = []models.Library{}
	}
	writeJSON(w, http.StatusOK, libs)
}

func (s *Server) handleGetLibrary(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryPathID(w, r)
	if !ok {
		return
	}
	if !s.requireLibraryAccess(w, r, id) {
		return
	}
	lib, err := s.store.GetLibrary(r.Context(), id)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, lib)
}

type libraryBody struct {
	Name      string                 `json:"name"`
	MountPath string                 `json:"mountPath"`
	Backend   string                 `json:"backend"`
	S3        *models.LibraryS3Input `json:"s3,omitempty"`
}

func (s *Server) handleCreateLibrary(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermManageLibrary); !ok {
		return
	}
	var body libraryBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	in := models.LibraryCreate{
		Name:      body.Name,
		MountPath: body.MountPath,
		Backend:   body.Backend,
		S3:        body.S3,
	}
	if strings.TrimSpace(in.Backend) == "" {
		in.Backend = models.LibraryBackendLocal
	}
	if in.Backend == models.LibraryBackendLocal && (body.Name == "" || body.MountPath == "") {
		writeError(w, http.StatusBadRequest, errors.New("name and mountPath are required"))
		return
	}
	if in.Backend == models.LibraryBackendS3 && body.Name == "" {
		writeError(w, http.StatusBadRequest, errors.New("name is required"))
		return
	}
	lib, err := s.store.CreateLibraryFull(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, lib)
}

func (s *Server) handleUpdateLibrary(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermManageLibrary); !ok {
		return
	}
	id, ok := libraryPathID(w, r)
	if !ok {
		return
	}
	var body libraryBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	in := models.LibraryCreate{
		Name:      body.Name,
		MountPath: body.MountPath,
		Backend:   body.Backend,
		S3:        body.S3,
	}
	if strings.TrimSpace(in.Backend) == "" {
		in.Backend = models.LibraryBackendLocal
	}
	if in.Backend == models.LibraryBackendLocal && (body.Name == "" || body.MountPath == "") {
		writeError(w, http.StatusBadRequest, errors.New("name and mountPath are required"))
		return
	}
	if in.Backend == models.LibraryBackendS3 && body.Name == "" {
		writeError(w, http.StatusBadRequest, errors.New("name is required"))
		return
	}
	lib, err := s.store.UpdateLibraryFull(r.Context(), id, in)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	s.invalidateLibraryFS(id)
	writeJSON(w, http.StatusOK, lib)
}

func (s *Server) handleTestS3(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermManageLibrary); !ok {
		return
	}
	var body models.LibraryS3Input
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	cfg := libfs.S3Config{
		Endpoint:     body.Endpoint,
		Region:       body.Region,
		Bucket:       body.Bucket,
		Prefix:       body.Prefix,
		AccessKey:    body.AccessKey,
		SecretKey:    body.SecretKey,
		UsePathStyle: body.UsePathStyle,
		TLS:          body.TLS,
	}
	if err := libfs.TestS3(r.Context(), cfg); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDeleteLibrary(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermManageLibrary); !ok {
		return
	}
	id, ok := libraryPathID(w, r)
	if !ok {
		return
	}
	if err := s.store.DeleteLibrary(r.Context(), id); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	s.invalidateLibraryFS(id)
	w.WriteHeader(http.StatusNoContent)
}

type reorderBody struct {
	IDs []int64 `json:"ids"`
}

func (s *Server) handleReorderLibraries(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermManageLibrary); !ok {
		return
	}
	var body reorderBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if len(body.IDs) == 0 {
		writeError(w, http.StatusBadRequest, errors.New("ids are required"))
		return
	}
	if err := s.store.ReorderLibraries(r.Context(), body.IDs); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleScanLibrary(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermManageLibrary); !ok {
		return
	}
	id, ok := libraryPathID(w, r)
	if !ok {
		return
	}
	if !s.requireLibraryAccess(w, r, id) {
		return
	}
	if _, err := s.store.GetLibrary(r.Context(), id); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	go func() {
		if err := s.scanner.ScanLibrary(s.jobsCtx, id); err != nil {
			s.log.Error("library scan failed", "libraryId", id, "err", err)
		}
	}()
	writeJSON(w, http.StatusAccepted, map[string]bool{"started": true})
}

func libraryPathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid library id"))
		return 0, false
	}
	return id, true
}
