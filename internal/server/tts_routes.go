package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"athenaeum/internal/models"
)

const (
	maxTTSTextBytes  = 8000
	maxTTSAudioBytes = 8 << 20
)

func (s *Server) registerTTSRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admin/tts", s.handleGetTTS)
	mux.HandleFunc("PUT /api/admin/tts", s.handlePutTTS)
	mux.HandleFunc("POST /api/admin/tts/test", s.handleTestTTS)
	mux.HandleFunc("GET /api/tts/status", s.handleTTSStatus)
	mux.HandleFunc("GET /api/tts/voices", s.handleTTSVoices)
	mux.HandleFunc("POST /api/tts/synthesize", s.handleTTSSynthesize)
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
	if err := pingTTS(ctx, cfg); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "Kokoro sidecar reachable"})
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
	voices, err := listTTSVoices(ctx, cfg)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
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
	audio, contentType, err := synthesizeTTS(ctx, cfg, text, voice, speed)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if contentType == "" {
		contentType = "audio/wav"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(audio) // #nosec G705 -- TTS audio bytes Content-Type is audio/*
}

func (s *Server) requireEnabledTTS(w http.ResponseWriter, r *http.Request) (models.TTSSettings, error) {
	cfg, err := s.store.GetTTSSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return cfg, err
	}
	if !cfg.Enabled || cfg.BaseURL == "" {
		err := errors.New("kokoro TTS is not configured")
		writeError(w, http.StatusBadRequest, err)
		return cfg, err
	}
	return cfg, nil
}

type ttsVoice struct {
	ID    string `json:"id"`
	Label string `json:"label,omitempty"`
	Lang  string `json:"lang,omitempty"`
}

func pingTTS(ctx context.Context, cfg models.TTSSettings) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.BaseURL+"/health", nil)
	if err != nil {
		return err
	}
	applyTTSAuth(req, cfg)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("sidecar health returned %d", res.StatusCode)
	}
	return nil
}

func listTTSVoices(ctx context.Context, cfg models.TTSSettings) ([]ttsVoice, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.BaseURL+"/voices", nil)
	if err != nil {
		return nil, err
	}
	applyTTSAuth(req, cfg)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("sidecar voices returned %d: %s", res.StatusCode, truncateErr(body))
	}
	var parsed struct {
		Voices []ttsVoice `json:"voices"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Voices) == 0 {
		return []ttsVoice{{ID: cfg.DefaultVoice, Label: cfg.DefaultVoice}}, nil
	}
	return parsed.Voices, nil
}

func synthesizeTTS(ctx context.Context, cfg models.TTSSettings, text, voice string, speed float64) ([]byte, string, error) {
	payload, err := json.Marshal(map[string]any{
		"text":  text,
		"voice": voice,
		"speed": speed,
	})
	if err != nil {
		return nil, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.BaseURL+"/v1/audio/speech", bytes.NewReader(payload))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/wav, audio/*, application/octet-stream")
	applyTTSAuth(req, cfg)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, maxTTSAudioBytes+1))
	if err != nil {
		return nil, "", err
	}
	if len(body) > maxTTSAudioBytes {
		return nil, "", errors.New("audio response too large")
	}
	if res.StatusCode >= 300 {
		return nil, "", fmt.Errorf("sidecar synthesize returned %d: %s", res.StatusCode, truncateErr(body))
	}
	if len(body) == 0 {
		return nil, "", errors.New("empty audio response")
	}
	ct := res.Header.Get("Content-Type")
	return body, ct, nil
}

func applyTTSAuth(req *http.Request, cfg models.TTSSettings) {
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}
}

func truncateErr(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 200 {
		return s[:200] + "…"
	}
	return s
}
