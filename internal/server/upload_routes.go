package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"athenaeum/internal/library"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func (s *Server) registerUploadRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/libraries/{id}/uploads", s.handleCreateUpload)
	mux.HandleFunc("GET /api/libraries/{id}/uploads/{uploadId}", s.handleGetUpload)
	mux.HandleFunc("PATCH /api/libraries/{id}/uploads/{uploadId}", s.handlePatchUpload)
	mux.HandleFunc("DELETE /api/libraries/{id}/uploads/{uploadId}", s.handleDeleteUpload)
}

type createUploadBody struct {
	RelPath   string `json:"relPath"`
	TotalSize int64  `json:"totalSize"`
}

func (s *Server) handleCreateUpload(w http.ResponseWriter, r *http.Request) {
	libID, ok := libraryPathID(w, r)
	if !ok {
		return
	}
	if !s.requireLibraryAccess(w, r, libID) {
		return
	}
	required, err := s.store.AuthRequired(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	userID := UserIDFromContext(r.Context())
	if required && userID == 0 {
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var body createUploadBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	rel, err := sanitizeUploadRelPath(body.RelPath)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if body.TotalSize <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("totalSize must be positive"))
		return
	}
	if s.cfg.UploadMaxBytes > 0 && body.TotalSize > s.cfg.UploadMaxBytes {
		writeError(w, http.StatusRequestEntityTooLarge, fmt.Errorf("upload exceeds limit of %d bytes", s.cfg.UploadMaxBytes))
		return
	}
	format := uploadFormatFromExt(rel)
	if format == "" {
		writeError(w, http.StatusBadRequest, errors.New("unsupported file type"))
		return
	}
	id, err := newUploadID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := os.MkdirAll(s.cfg.UploadDir(), 0o750); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	sess := models.UploadSession{
		ID:        id,
		LibraryID: libID,
		UserID:    userID,
		RelPath:   rel,
		TotalSize: body.TotalSize,
	}
	if err := s.store.CreateUploadSession(r.Context(), sess); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := writeUploadPart(s.cfg.UploadDir(), id, nil); err != nil {
		_ = s.store.DeleteUploadSession(r.Context(), id)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	sess, _ = s.store.GetUploadSession(r.Context(), id)
	writeJSON(w, http.StatusCreated, sess)
}

func (s *Server) handleGetUpload(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.loadUploadSession(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

func (s *Server) handlePatchUpload(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.loadUploadSession(w, r)
	if !ok {
		return
	}
	if sess.Done {
		writeError(w, http.StatusConflict, errors.New("upload already completed"))
		return
	}
	start, end, total, err := parseContentRange(r.Header.Get("Content-Range"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if total != sess.TotalSize {
		writeError(w, http.StatusBadRequest, errors.New("total size mismatch"))
		return
	}
	if start != sess.Offset {
		writeError(w, http.StatusConflict, fmt.Errorf("expected offset %d, got %d", sess.Offset, start))
		return
	}
	chunkLen := end - start + 1
	if chunkLen <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid content range"))
		return
	}
	if end >= sess.TotalSize {
		writeError(w, http.StatusBadRequest, errors.New("range exceeds total size"))
		return
	}
	f, err := openUploadPart(s.cfg.UploadDir(), sess.ID, os.O_WRONLY)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		_ = f.Close()
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	limited := io.LimitReader(r.Body, chunkLen)
	written, err := io.Copy(f, limited)
	_ = f.Close()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if written != chunkLen {
		writeError(w, http.StatusBadRequest, errors.New("short write"))
		return
	}
	newOffset := start + written
	if err := s.store.UpdateUploadOffset(r.Context(), sess.ID, newOffset); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if newOffset < sess.TotalSize {
		sess.Offset = newOffset
		writeJSON(w, http.StatusOK, sess)
		return
	}
	bookID, err := s.finalizeUpload(r, sess)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	sess.Offset = newOffset
	sess.Done = true
	sess.BookID = bookID
	writeJSON(w, http.StatusOK, sess)
}

func (s *Server) handleDeleteUpload(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.loadUploadSession(w, r)
	if !ok {
		return
	}
	_ = removeUploadPart(s.cfg.UploadDir(), sess.ID)
	if err := s.store.DeleteUploadSession(r.Context(), sess.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) loadUploadSession(w http.ResponseWriter, r *http.Request) (models.UploadSession, bool) {
	libID, ok := libraryPathID(w, r)
	if !ok {
		return models.UploadSession{}, false
	}
	if !s.requireLibraryAccess(w, r, libID) {
		return models.UploadSession{}, false
	}
	uploadID := r.PathValue("uploadId")
	if uploadID == "" {
		writeError(w, http.StatusBadRequest, errors.New("missing upload id"))
		return models.UploadSession{}, false
	}
	sess, err := s.store.GetUploadSession(r.Context(), uploadID)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return models.UploadSession{}, false
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return models.UploadSession{}, false
	}
	if sess.LibraryID != libID {
		writeError(w, http.StatusNotFound, storage.ErrNotFound)
		return models.UploadSession{}, false
	}
	userID := UserIDFromContext(r.Context())
	if user, ok := UserFromContext(r.Context()); ok {
		if sess.UserID != user.ID && !user.IsAdmin {
			writeError(w, http.StatusForbidden, errors.New("upload access denied"))
			return models.UploadSession{}, false
		}
	} else if sess.UserID != userID {
		writeError(w, http.StatusForbidden, errors.New("upload access denied"))
		return models.UploadSession{}, false
	}
	return sess, true
}

func (s *Server) finalizeUpload(r *http.Request, sess models.UploadSession) (int64, error) {
	fs, err := s.openLibraryFS(r.Context(), sess.LibraryID)
	if err != nil {
		return 0, err
	}
	part, err := openUploadPart(s.cfg.UploadDir(), sess.ID, os.O_RDONLY)
	if err != nil {
		return 0, err
	}
	defer part.Close()
	info, err := part.Stat()
	if err != nil {
		return 0, err
	}
	if err := copyReaderToLibrary(r.Context(), fs, sess.RelPath, part, info.Size()); err != nil {
		return 0, err
	}
	_ = removeUploadPart(s.cfg.UploadDir(), sess.ID)
	bookID, err := s.scanner.IndexFile(r.Context(), sess.LibraryID, sess.RelPath)
	if err != nil {
		return 0, err
	}
	if err := s.store.CompleteUploadSession(r.Context(), sess.ID, bookID); err != nil {
		return bookID, err
	}
	if user, ok := UserFromContext(r.Context()); ok {
		s.logAudit(r, user.ID, user.Username, 0, "", "book.upload",
			fmt.Sprintf("library=%d path=%s book=%d", sess.LibraryID, sess.RelPath, bookID))
		s.emitWebhook(models.WebhookEventBookUpload, map[string]any{
			"bookId":    bookID,
			"libraryId": sess.LibraryID,
			"path":      sess.RelPath,
		})
	}
	return bookID, nil
}

func newUploadID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func sanitizeUploadRelPath(rel string) (string, error) {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" || strings.Contains(rel, "..") {
		return "", errors.New("invalid relPath")
	}
	clean := filepath.Clean(rel)
	if clean == "." || strings.HasPrefix(clean, "..") {
		return "", errors.New("invalid relPath")
	}
	return filepath.ToSlash(clean), nil
}

func parseContentRange(h string) (start, end, total int64, err error) {
	if !strings.HasPrefix(h, "bytes ") {
		return 0, 0, 0, errors.New("missing Content-Range header")
	}
	parts := strings.Split(strings.TrimPrefix(h, "bytes "), "/")
	if len(parts) != 2 {
		return 0, 0, 0, errors.New("invalid Content-Range")
	}
	total, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil || total <= 0 {
		return 0, 0, 0, errors.New("invalid total in Content-Range")
	}
	rangeParts := strings.Split(parts[0], "-")
	if len(rangeParts) != 2 {
		return 0, 0, 0, errors.New("invalid range in Content-Range")
	}
	start, err = strconv.ParseInt(rangeParts[0], 10, 64)
	if err != nil || start < 0 {
		return 0, 0, 0, errors.New("invalid range start")
	}
	end, err = strconv.ParseInt(rangeParts[1], 10, 64)
	if err != nil || end < start {
		return 0, 0, 0, errors.New("invalid range end")
	}
	return start, end, total, nil
}

func uploadFormatFromExt(rel string) string {
	return library.FormatFromExt(rel)
}
