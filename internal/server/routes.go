package server

import (
	"bytes"
	"context"
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
	mux.HandleFunc("GET /api/library/stats", s.handleStats)
	mux.HandleFunc("GET /api/library/scan/status", s.handleScanStatus)
	mux.HandleFunc("POST /api/library/scan", s.handleScan)
	mux.HandleFunc("GET /api/series", s.handleListSeries)
	mux.HandleFunc("GET /api/authors", s.handleListAuthors)

	mux.HandleFunc("GET /api/books", s.handleListBooks)
	mux.HandleFunc("GET /api/books/{id}", s.handleGetBook)
	mux.HandleFunc("GET /api/books/{id}/cover", s.handleCover)
	mux.HandleFunc("GET /api/books/{id}/file", s.handleFile)
	mux.HandleFunc("GET /api/books/{id}/download", s.handleDownload)

	mux.HandleFunc("GET /api/books/{id}/progress", s.handleGetProgress)
	mux.HandleFunc("PUT /api/books/{id}/progress", s.handlePutProgress)

	mux.HandleFunc("GET /api/books/{id}/chapters", s.handleGetChapters)

	s.registerBookEditRoutes(mux)
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

func (s *Server) handleGetChapters(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.store.GetBook(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("book not found"))
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	chapters, err := s.store.ListChapters(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if chapters == nil {
		chapters = []models.Chapter{}
	}
	writeJSON(w, http.StatusOK, chapters)
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
	writeJSON(w, http.StatusOK, s.scanner.Status())
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) handleListBooks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	query := models.BookQuery{
		Search:       q.Get("search"),
		Sort:         q.Get("sort"),
		Format:       q.Get("format"),
		Series:       q.Get("series"),
		Author:       q.Get("author"),
		LibraryID:    int64(atoiDefault(q.Get("library"), 0)),
		CollectionID: int64(atoiDefault(q.Get("collection"), 0)),
		UserID:       UserIDFromContext(r.Context()),
		Favorites:    q.Get("favorites") == "1" || q.Get("favorites") == "true",
		InProgress:   q.Get("inProgress") == "1" || q.Get("inProgress") == "true",
		Tag:          q.Get("tag"),
		Limit:        atoiDefault(q.Get("limit"), 60),
		Offset:       atoiDefault(q.Get("offset"), 0),
	}
	var err error
	query, err = s.applyBookAccess(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	page, err := s.store.ListBooks(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.attachTagsAndRatings(r.Context(), page.Items, query.UserID)
	writeJSON(w, http.StatusOK, page)
}

// attachTagsAndRatings fills in Tags and UserRating for a page of books using
// batched lookups, ignoring lookup errors since these fields are supplemental.
func (s *Server) attachTagsAndRatings(ctx context.Context, items []models.Book, userID int64) {
	if len(items) == 0 {
		return
	}
	ids := make([]int64, len(items))
	for i, b := range items {
		ids[i] = b.ID
	}
	tags, err := s.store.ListBookTagsBatch(ctx, ids)
	if err == nil {
		for i := range items {
			items[i].Tags = tags[items[i].ID]
		}
	}
	if userID <= 0 {
		return
	}
	ratings, err := s.store.RatingsBatch(ctx, userID, ids)
	if err == nil {
		for i := range items {
			items[i].UserRating = ratings[items[i].ID]
		}
	}
}

func (s *Server) handleGetBook(w http.ResponseWriter, r *http.Request) {
	book, err := s.bookByIDChecked(w, r)
	if err != nil {
		return
	}
	items := []models.Book{book}
	s.attachTagsAndRatings(r.Context(), items, UserIDFromContext(r.Context()))
	writeJSON(w, http.StatusOK, items[0])
}

func (s *Server) handleCover(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	book, err := s.store.GetBook(r.Context(), id)
	if errors.Is(err, storage.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !s.requireBookAccess(w, r, book) {
		return
	}
	f, info, err := openCoverFile(s.cfg.CoverDir(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, "cover", info.ModTime(), f)
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	s.serveBookFile(w, r, false)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	s.serveBookFile(w, r, true)
}

func (s *Server) serveBookFile(w http.ResponseWriter, r *http.Request, attachment bool) {
	book, err := s.bookByIDChecked(w, r)
	if err != nil {
		return
	}
	relPath := book.RelPath
	format := book.Format
	if book.Format == models.FormatAudiobook {
		tracks, err := s.store.ListAudiobookTracks(r.Context(), book.ID)
		if err != nil || len(tracks) == 0 {
			writeError(w, http.StatusNotFound, errors.New("audiobook tracks missing"))
			return
		}
		idx := atoiDefault(r.URL.Query().Get("track"), 0)
		if idx < 0 || idx >= len(tracks) {
			idx = 0
		}
		relPath = tracks[idx].RelPath
		format = tracks[idx].Format
	}
	s.serveLibraryFile(w, r, book.LibraryID, relPath, safeFilename(book), contentType(format), attachment)
}

func (s *Server) handleGetProgress(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	userID := UserIDFromContext(r.Context())
	p, err := s.store.GetProgress(r.Context(), userID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handlePutProgress(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var p models.Progress
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	p.BookID = id
	userID := UserIDFromContext(r.Context())
	if err := s.store.SaveProgress(r.Context(), userID, p); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	p.UserID = userID
	p.UpdatedAt = time.Now()
	writeJSON(w, http.StatusOK, p)
}

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
	name := b.Title
	if name == "" {
		name = "book"
	}
	return name + "." + b.Format
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
