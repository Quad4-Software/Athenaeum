package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
	"athenaeum/internal/tts"
)

const (
	maxTTSTextBytes = tts.MaxTextBytes
	// maxPendingJobsPerUser caps queued/running jobs per account.
	maxPendingJobsPerUser = 20
)

// ttsResponseFormats are the values accepted for the admin response_format
// setting. The sidecar must be able to encode them.
var ttsResponseFormats = map[string]bool{
	"mp3": true, "wav": true, "opus": true, "flac": true,
	"m4a": true, "aac": true, "pcm": true,
}

func (s *Server) registerTTSRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admin/tts", s.handleGetTTS)
	mux.HandleFunc("PUT /api/admin/tts", s.handlePutTTS)
	mux.HandleFunc("POST /api/admin/tts/test", s.handleTestTTS)
	mux.HandleFunc("GET /api/tts/status", s.handleTTSStatus)
	mux.HandleFunc("GET /api/tts/voices", s.handleTTSVoices)
	mux.HandleFunc("POST /api/tts/synthesize", s.handleTTSSynthesize)
	mux.HandleFunc("GET /api/tts/prefs", s.handleGetTTSPrefs)
	mux.HandleFunc("PUT /api/tts/prefs", s.handlePutTTSPrefs)
	mux.HandleFunc("GET /api/tts/jobs", s.handleListTTSJobs)
	mux.HandleFunc("POST /api/tts/jobs", s.handleCreateTTSJob)
	mux.HandleFunc("POST /api/tts/jobs/{id}/cancel", s.handleCancelTTSJob)
	mux.HandleFunc("POST /api/tts/jobs/{id}/retry", s.handleRetryTTSJob)
	mux.HandleFunc("DELETE /api/tts/jobs/{id}", s.handleDeleteTTSJob)
}

func (s *Server) handleGetTTS(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	cfg, err := s.store.GetTTSSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, cfg.Public())
}

func (s *Server) handlePutTTS(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	var cfg models.TTSSettings
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	cfg.DefaultVoice = strings.TrimSpace(cfg.DefaultVoice)
	cfg.Model = strings.TrimSpace(cfg.Model)
	cfg.ResponseFormat = strings.ToLower(strings.TrimSpace(cfg.ResponseFormat))
	if cfg.ResponseFormat != "" && !ttsResponseFormats[cfg.ResponseFormat] {
		writeError(w, http.StatusBadRequest, fmt.Errorf("responseFormat must be one of mp3, wav, opus, flac, m4a, aac, pcm"))
		return
	}
	if cfg.APIKey == "" {
		existing, err := s.store.GetTTSSettings(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		cfg.APIKey = existing.APIKey
	}
	if cfg.Enabled {
		if cfg.BaseURL == "" {
			writeError(w, http.StatusBadRequest, errors.New("baseUrl is required when TTS is enabled"))
			return
		}
		if _, err := url.ParseRequestURI(cfg.BaseURL); err != nil {
			writeError(w, http.StatusBadRequest, errors.New("baseUrl must be a valid URL"))
			return
		}
	}
	if cfg.TimeoutSec < 5 {
		cfg.TimeoutSec = 5
	}
	if cfg.TimeoutSec > 300 {
		cfg.TimeoutSec = 300
	}
	if err := s.store.SaveTTSSettings(r.Context(), cfg); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.kickTTS()
	writeJSON(w, http.StatusOK, cfg.Public())
}

func (s *Server) handleTestTTS(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	cfg, err := s.store.GetTTSSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !cfg.Enabled || cfg.BaseURL == "" {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "message": "TTS is not enabled"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(cfg.TimeoutSec)*time.Second)
	defer cancel()
	if err := tts.NewClient(cfg).Ping(ctx); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "TTS sidecar reachable"})
}

func (s *Server) handleTTSStatus(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.store.GetTTSSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, models.TTSStatus{
		Enabled:      cfg.Enabled && cfg.BaseURL != "",
		DefaultVoice: cfg.DefaultVoice,
	})
}

func (s *Server) handleTTSVoices(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.requireEnabledTTS(w, r)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(cfg.TimeoutSec)*time.Second)
	defer cancel()
	voices, err := tts.NewClient(cfg).Voices(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if len(voices) == 0 {
		voices = []tts.Voice{{ID: cfg.DefaultVoice, Label: cfg.DefaultVoice}}
	}
	writeJSON(w, http.StatusOK, map[string]any{"voices": voices})
}

func (s *Server) handleTTSSynthesize(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.requireEnabledTTS(w, r)
	if err != nil {
		return
	}
	var req struct {
		Text  string  `json:"text"`
		Voice string  `json:"voice"`
		Speed float64 `json:"speed"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxTTSTextBytes+1024)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	text := strings.TrimSpace(req.Text)
	if text == "" {
		writeError(w, http.StatusBadRequest, errors.New("text is required"))
		return
	}
	if len(text) > maxTTSTextBytes {
		writeError(w, http.StatusBadRequest, fmt.Errorf("text exceeds %d bytes", maxTTSTextBytes))
		return
	}
	voice := strings.TrimSpace(req.Voice)
	if voice == "" {
		voice = cfg.DefaultVoice
	}
	speed := req.Speed
	if speed <= 0 {
		speed = 1
	}
	if speed < 0.5 {
		speed = 0.5
	}
	if speed > 2 {
		speed = 2
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(cfg.TimeoutSec)*time.Second)
	defer cancel()
	audio, contentType, err := tts.NewClient(cfg).Synthesize(ctx, text, voice, speed)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if contentType == "" {
		contentType = mimeOctetStream
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(audio) // #nosec G705 -- TTS audio bytes Content-Type is audio/*
}

// handleGetTTSPrefs returns the caller's narration preferences.
func (s *Server) handleGetTTSPrefs(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermRead); !ok {
		return
	}
	prefs, err := s.store.GetTTSUserPrefs(r.Context(), UserIDFromContext(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, prefs)
}

// handlePutTTSPrefs saves narration preferences including the daily window.
func (s *Server) handlePutTTSPrefs(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermRead); !ok {
		return
	}
	var p models.TTSUserPrefs
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<15)).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	p.UserID = UserIDFromContext(r.Context())
	p.Voice = strings.TrimSpace(p.Voice)
	p.SchedStart = strings.TrimSpace(p.SchedStart)
	p.SchedEnd = strings.TrimSpace(p.SchedEnd)
	if p.Speed <= 0 {
		p.Speed = 1
	}
	if p.Speed < 0.5 || p.Speed > 2 {
		writeError(w, http.StatusBadRequest, errors.New("speed must be between 0.5 and 2"))
		return
	}
	if !tts.ValidSchedule(p) {
		writeError(w, http.StatusBadRequest, errors.New("schedStart and schedEnd must be HH:MM"))
		return
	}
	if err := s.store.SaveTTSUserPrefs(r.Context(), p); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.kickTTS()
	writeJSON(w, http.StatusOK, p)
}

// handleListTTSJobs returns the caller's jobs; admins may pass ?all=1.
func (s *Server) handleListTTSJobs(w http.ResponseWriter, r *http.Request) {
	u, ok := s.requirePermission(w, r, models.PermRead)
	if !ok {
		return
	}
	all := u.IsAdmin && r.URL.Query().Get("all") == "1"
	jobs, err := s.store.ListTTSJobs(r.Context(), UserIDFromContext(r.Context()), all)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"jobs": jobs})
}

// handleCreateTTSJob queues a whole-book narration job.
func (s *Server) handleCreateTTSJob(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermRead); !ok {
		return
	}
	var req struct {
		BookID int64   `json:"bookId"`
		Voice  string  `json:"voice"`
		Speed  float64 `json:"speed"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<15)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	ctx := r.Context()
	cfg, err := s.store.GetTTSSettings(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !cfg.Enabled || cfg.BaseURL == "" {
		writeError(w, http.StatusBadRequest, errors.New("TTS is not configured"))
		return
	}
	book, err := s.store.GetBook(ctx, req.BookID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("book not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !s.requireBookAccess(w, r, book) {
		return
	}
	if book.Format != models.FormatEPUB {
		writeError(w, http.StatusBadRequest, errors.New("only EPUB books can be narrated"))
		return
	}
	userID := UserIDFromContext(ctx)
	if n, err := s.store.CountActiveTTSJobsForBook(ctx, book.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	} else if n > 0 {
		writeError(w, http.StatusConflict, errors.New("an audiobook job for this book is already queued or running"))
		return
	}
	if n, err := s.store.CountPendingTTSJobs(ctx, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	} else if n >= maxPendingJobsPerUser {
		writeError(w, http.StatusConflict, fmt.Errorf("job limit reached (%d pending)", maxPendingJobsPerUser))
		return
	}
	prefs, err := s.store.GetTTSUserPrefs(ctx, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	voice := strings.TrimSpace(req.Voice)
	if voice == "" {
		voice = prefs.Voice
	}
	if voice == "" {
		voice = cfg.DefaultVoice
	}
	speed := req.Speed
	if speed <= 0 {
		speed = prefs.Speed
	}
	if speed <= 0 {
		speed = 1
	}
	if speed < 0.5 {
		speed = 0.5
	}
	if speed > 2 {
		speed = 2
	}
	runAt := int64(0)
	if !tts.InWindow(time.Now(), prefs) {
		runAt = tts.NextWindowStart(time.Now(), prefs).Unix()
	}
	id, err := s.store.CreateTTSJob(ctx, models.TTSJob{
		UserID: userID, BookID: book.ID, Voice: voice, Speed: speed, RunAt: runAt,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.kickTTS()
	job, err := s.store.GetTTSJob(ctx, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, job)
}

// ttsJobForUser loads a job the caller may act on (owner or admin).
func (s *Server) ttsJobForUser(w http.ResponseWriter, r *http.Request) (models.TTSJob, bool) {
	u, ok := s.requirePermission(w, r, models.PermRead)
	if !ok {
		return models.TTSJob{}, false
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid job id"))
		return models.TTSJob{}, false
	}
	job, err := s.store.GetTTSJob(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("job not found"))
			return models.TTSJob{}, false
		}
		writeError(w, http.StatusInternalServerError, err)
		return models.TTSJob{}, false
	}
	if !u.IsAdmin && job.UserID != u.ID && u.ID != 0 {
		writeError(w, http.StatusForbidden, errForbidden)
		return models.TTSJob{}, false
	}
	return job, true
}

func (s *Server) handleCancelTTSJob(w http.ResponseWriter, r *http.Request) {
	job, ok := s.ttsJobForUser(w, r)
	if !ok {
		return
	}
	if job.Status != models.TTSJobQueued && job.Status != models.TTSJobRunning {
		writeError(w, http.StatusConflict, errors.New("job is not active"))
		return
	}
	if err := s.store.FinishTTSJob(r.Context(), job.ID, models.TTSJobCancelled, ""); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleRetryTTSJob(w http.ResponseWriter, r *http.Request) {
	job, ok := s.ttsJobForUser(w, r)
	if !ok {
		return
	}
	if job.Status == models.TTSJobQueued || job.Status == models.TTSJobRunning {
		writeError(w, http.StatusConflict, errors.New("job is already active"))
		return
	}
	prefs, err := s.store.GetTTSUserPrefs(r.Context(), job.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	runAt := int64(0)
	if !tts.InWindow(time.Now(), prefs) {
		runAt = tts.NextWindowStart(time.Now(), prefs).Unix()
	}
	if err := s.store.RetryTTSJob(r.Context(), job.ID, runAt); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.kickTTS()
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDeleteTTSJob(w http.ResponseWriter, r *http.Request) {
	job, ok := s.ttsJobForUser(w, r)
	if !ok {
		return
	}
	if err := s.store.DeleteTTSJob(r.Context(), job.ID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusConflict, errors.New("job is still active or missing"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if s.ttsWorker != nil {
		s.ttsWorker.DropStaging(job.ID)
	}
	writeJSON(w, http.StatusNoContent, nil)
}

// kickTTS nudges the job worker when it is running.
func (s *Server) kickTTS() {
	if s.ttsWorker != nil {
		s.ttsWorker.Kick()
	}
}

func (s *Server) requireEnabledTTS(w http.ResponseWriter, r *http.Request) (models.TTSSettings, error) {
	cfg, err := s.store.GetTTSSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return cfg, err
	}
	if !cfg.Enabled || cfg.BaseURL == "" {
		err := errors.New("TTS sidecar is not configured")
		writeError(w, http.StatusBadRequest, err)
		return cfg, err
	}
	return cfg, nil
}
