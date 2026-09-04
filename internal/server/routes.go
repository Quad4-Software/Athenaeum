package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
	"athenaeum/internal/system"
	"athenaeum/internal/telemetry"
	"athenaeum/internal/version"
)

var jsonEncodeBufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// registerAPI registers all JSON and file endpoints under /api.
func (s *Server) registerAPI(mux *http.ServeMux) {
	s.registerAuthRoutes(mux)
	s.registerUserLibraryRoutes(mux)
	s.registerCollectionRoutes(mux)
	s.registerLibraryRoutes(mux)
	s.registerUploadRoutes(mux)
	s.registerOPDSRoutes(mux)
	s.registerFavoriteRoutes(mux)
	s.registerTagRoutes(mux)
	s.registerShareRoutes(mux)
	s.registerKosyncRoutes(mux)
	s.registerSMTPRoutes(mux)
	s.registerTTSRoutes(mux)
	s.registerOfflineRoutes(mux)
	s.registerContentIndexRoutes(mux)
	s.registerFSRoutes(mux)
	s.registerInviteRoutes(mux)
	s.registerWebhookRoutes(mux)
	s.registerPocketIDRoutes(mux)

	mux.HandleFunc("GET /api/health", s.handleHealth)
	s.registerLibraryScanRoutes(mux)
	s.registerBookRoutes(mux)

	s.registerBookEditRoutes(mux)
	s.registerBibTeXRoutes(mux)
	s.registerMetadataJobRoutes(mux)
	s.registerSystemRoutes(mux)
	s.registerSettingsRoutes(mux)
	s.registerGuestRoutes(mux)
	s.registerBookmarkRoutes(mux)
	s.registerMediaRoutes(mux)
	s.registerAdminBackupRoutes(mux)
	s.registerAdminTaskRoutes(mux)
	s.registerI18nRoutes(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{"status": "ok"}
	if err := s.store.PingDB(r.Context()); err != nil {
		resp["status"] = "degraded"
		resp["database"] = "error"
	} else {
		resp["database"] = "ok"
	}
	scan := s.scanner.Status()
	resp["scanning"] = scan.Scanning
	if scan.FinishedAt != nil {
		resp["lastScan"] = scan.FinishedAt.Format(time.RFC3339)
	}
	if st := system.DiskFree(s.cfg.DataDir); st >= 0 {
		resp["diskFreeBytes"] = st
	}
	resp["version"] = version.Version
	resp["webVersion"] = version.WebVersion
	if tel := telemetry.PublicFromConfig(s.cfg); tel.SentryDSN != "" {
		resp["telemetry"] = tel
	}
	writeJSON(w, http.StatusOK, resp)
}

// attachTagsAndRatings fills in Tags and UserRating for a page of books using
// batched lookups, ignoring lookup errors since these fields are supplemental.
func (s *Server) bookByID(w http.ResponseWriter, r *http.Request) (models.Book, error) {
	id, ok := pathID(w, r)
	if !ok {
		return models.Book{}, errors.New("bad id")
	}
	book, err := s.store.GetBook(r.Context(), id)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, errors.New("book not found"))
		return models.Book{}, err
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return models.Book{}, err
	}
	return book, nil
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid id"))
		return 0, false
	}
	return id, true
}

func contentType(format string) string {
	switch format {
	case models.FormatEPUB:
		return "application/epub+zip"
	case models.FormatPDF:
		return "application/pdf"
	case models.FormatMP3:
		return "audio/mpeg"
	case models.FormatM4B, models.FormatM4A:
		return "audio/mp4"
	case models.FormatOGG:
		return "audio/ogg"
	case models.FormatFLAC:
		return "audio/flac"
	case models.FormatMOBI, models.FormatAZW, models.FormatAZW3:
		return "application/x-mobipocket-ebook"
	case models.FormatKFX:
		return "application/vnd.amazon.ebook"
	case models.FormatCBZ:
		return "application/vnd.comicbook+zip"
	case models.FormatCBR:
		return "application/vnd.comicbook-rar"
	default:
		return "application/octet-stream"
	}
}

func safeFilename(b models.Book) string {
	name := sanitizeFilenameToken(b.Title)
	format := sanitizeFilenameToken(b.Format)
	if format == "" {
		format = "bin"
	}
	return name + "." + format
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	buf := jsonEncodeBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	encErr := json.NewEncoder(buf).Encode(v)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if encErr != nil {
		jsonEncodeBufPool.Put(buf)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"encode failed"}`))
		return
	}
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
	jsonEncodeBufPool.Put(buf)
}

func writeError(w http.ResponseWriter, status int, err error) {
	if status >= http.StatusInternalServerError && err != nil {
		telemetry.CaptureException(err)
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
