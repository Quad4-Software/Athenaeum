package server

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"athenaeum/internal/models"
)

const (
	webhookMaxAttempts = 3
	webhookHTTPTimeout = httpClientTimeout
)

func (s *Server) emitWebhook(event string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	go s.dispatchWebhooks(s.jobsCtx, event, data)
}

func (s *Server) dispatchWebhooks(ctx context.Context, event string, data map[string]any) {
	hooks, err := s.store.ListEnabledWebhooksForEvent(ctx, event)
	if err != nil {
		s.log.Warn("list webhooks failed", "event", event, "err", err)
		return
	}
	if len(hooks) == 0 {
		return
	}
	id, err := newWebhookDeliveryID()
	if err != nil {
		s.log.Warn("webhook id failed", "err", err)
		return
	}
	envelope := map[string]any{
		"id":        id,
		"event":     event,
		"createdAt": time.Now().UTC().Format(time.RFC3339),
		"data":      data,
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		s.log.Warn("webhook marshal failed", "err", err)
		return
	}
	for _, wh := range hooks {
		s.deliverWebhook(ctx, wh, event, payload)
	}
}

func (s *Server) deliverWebhook(ctx context.Context, wh models.Webhook, event string, payload []byte) {
	sig := signWebhookPayload(wh.Secret, payload)
	var lastErr string
	var statusCode int
	success := false
	attempts := 0
	if err := validateWebhookURL(wh.URL); err != nil {
		now := time.Now()
		_, _ = s.store.InsertWebhookDelivery(ctx, models.WebhookDelivery{
			WebhookID: wh.ID, Event: event, Payload: string(payload),
			Success: false, Attempts: 0, LastError: err.Error(), CreatedAt: now,
		})
		return
	}
	client := webhookHTTPClient()
	for attempt := 1; attempt <= webhookMaxAttempts; attempt++ {
		attempts = attempt
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader(payload))
		if err != nil {
			lastErr = err.Error()
			break
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Athenaeum-Webhook/1")
		if sig != "" {
			req.Header.Set("X-Athenaeum-Signature", "sha256="+sig)
		}
		res, err := client.Do(req)
		if err != nil {
			lastErr = err.Error()
			time.Sleep(time.Duration(attempt*attempt) * 200 * time.Millisecond)
			continue
		}
		_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
		_ = res.Body.Close()
		statusCode = res.StatusCode
		if res.StatusCode >= 200 && res.StatusCode < 300 {
			success = true
			lastErr = ""
			break
		}
		lastErr = fmt.Sprintf("status %d", res.StatusCode)
		time.Sleep(time.Duration(attempt*attempt) * 200 * time.Millisecond)
	}
	now := time.Now()
	d := models.WebhookDelivery{
		WebhookID:  wh.ID,
		Event:      event,
		Payload:    string(payload),
		StatusCode: statusCode,
		Success:    success,
		Attempts:   attempts,
		LastError:  lastErr,
		CreatedAt:  now,
	}
	if success {
		d.DeliveredAt = &now
	}
	if _, err := s.store.InsertWebhookDelivery(ctx, d); err != nil {
		s.log.Warn("webhook delivery log failed", "err", err)
	}
}

func signWebhookPayload(secret string, payload []byte) string {
	if secret == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func newWebhookDeliveryID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// VerifyWebhookSignature checks an X-Athenaeum-Signature header value.
func VerifyWebhookSignature(secret, header string, payload []byte) bool {
	want := "sha256=" + signWebhookPayload(secret, payload)
	return hmac.Equal([]byte(want), []byte(header))
}
