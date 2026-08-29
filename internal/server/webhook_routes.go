package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func (s *Server) registerWebhookRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admin/webhooks", s.handleListWebhooks)
	mux.HandleFunc("POST /api/admin/webhooks", s.handleCreateWebhook)
	mux.HandleFunc("GET /api/admin/webhooks/{id}", s.handleGetWebhook)
	mux.HandleFunc("PUT /api/admin/webhooks/{id}", s.handleUpdateWebhook)
	mux.HandleFunc("DELETE /api/admin/webhooks/{id}", s.handleDeleteWebhook)
	mux.HandleFunc("GET /api/admin/webhooks/{id}/deliveries", s.handleListWebhookDeliveries)
	mux.HandleFunc("POST /api/admin/webhooks/{id}/test", s.handleTestWebhook)
}

func (s *Server) handleListWebhooks(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	list, err := s.store.ListWebhooks(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]models.WebhookPublic, 0, len(list))
	for _, wh := range list {
		out = append(out, wh.Public())
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleCreateWebhook(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	var req struct {
		URL     string   `json:"url"`
		Secret  string   `json:"secret"`
		Events  []string `json:"events"`
		Enabled *bool    `json:"enabled"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	url := strings.TrimSpace(req.URL)
	if url == "" || !strings.HasPrefix(url, "http") {
		writeError(w, http.StatusBadRequest, errors.New("url must be an http(s) endpoint"))
		return
	}
	events := sanitizeWebhookEvents(req.Events)
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	wh, err := s.store.CreateWebhook(r.Context(), models.Webhook{
		URL:     url,
		Secret:  req.Secret,
		Events:  events,
		Enabled: enabled,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, wh.Public())
}

func (s *Server) handleGetWebhook(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	wh, err := s.store.GetWebhook(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, wh.Public())
}

func (s *Server) handleUpdateWebhook(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	existing, err := s.store.GetWebhook(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	var req struct {
		URL     string   `json:"url"`
		Secret  string   `json:"secret"`
		Events  []string `json:"events"`
		Enabled *bool    `json:"enabled"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if url := strings.TrimSpace(req.URL); url != "" {
		existing.URL = url
	}
	if req.Secret != "" {
		existing.Secret = req.Secret
	}
	if req.Events != nil {
		existing.Events = sanitizeWebhookEvents(req.Events)
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}
	if err := s.store.UpdateWebhook(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, existing.Public())
}

func (s *Server) handleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := s.store.DeleteWebhook(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "deleted"})
}

func (s *Server) handleListWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	list, err := s.store.ListWebhookDeliveries(r.Context(), id, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if list == nil {
		list = []models.WebhookDelivery{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleTestWebhook(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	wh, err := s.store.GetWebhook(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"id":        "test",
		"event":     models.WebhookEventPing,
		"createdAt": time.Now().UTC().Format(time.RFC3339),
		"data":      map[string]any{"webhookId": wh.ID, "message": "test"},
	})
	s.deliverWebhook(r.Context(), wh, models.WebhookEventPing, payload)
	writeJSON(w, http.StatusOK, map[string]string{"ok": "sent"})
}

func sanitizeWebhookEvents(events []string) []string {
	allowed := make(map[string]struct{}, len(models.WebhookEventsV1))
	for _, e := range models.WebhookEventsV1 {
		allowed[e] = struct{}{}
	}
	var out []string
	seen := map[string]struct{}{}
	for _, e := range events {
		e = strings.TrimSpace(e)
		if _, ok := allowed[e]; !ok {
			continue
		}
		if _, ok := seen[e]; ok {
			continue
		}
		seen[e] = struct{}{}
		out = append(out, e)
	}
	if out == nil {
		out = []string{}
	}
	return out
}
