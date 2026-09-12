// Package tts implements the OpenAI-compatible speech client used for
// server-side narration, plus the queued whole-book audiobook worker.
package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"athenaeum/internal/models"
)

const (
	// MaxTextBytes caps a single synthesis request body.
	MaxTextBytes = 8000
	// MaxAudioBytes caps one synthesized audio response.
	MaxAudioBytes = 8 << 20
	// JobChunkChars splits chapter text into per-request pieces.
	JobChunkChars = 1800
	// JobAudioFormat is the output format requested for generated books.
	JobAudioFormat = "mp3"
)

// Voice is one narrator voice reported by the sidecar.
type Voice struct {
	ID    string `json:"id"`
	Label string `json:"label,omitempty"`
	Lang  string `json:"lang,omitempty"`
}

// speechRequest is the OpenAI /v1/audio/speech payload.
type speechRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float64 `json:"speed"`
	Stream         bool    `json:"stream"`
}

// Client talks to an OpenAI-compatible TTS endpoint (bundled Kokoro
// sidecar, Kokoro-FastAPI, Speaches, or a hosted API).
type Client struct {
	cfg models.TTSSettings
	hc  *http.Client
}

// NewClient builds a client bound to the given settings snapshot.
func NewClient(cfg models.TTSSettings) *Client {
	return &Client{cfg: cfg, hc: http.DefaultClient}
}

func (c *Client) newReq(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+path, body)
	if err != nil {
		return nil, err
	}
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	return req, nil
}

// Ping checks sidecar health.
func (c *Client) Ping(ctx context.Context) error {
	req, err := c.newReq(ctx, http.MethodGet, "/health", nil)
	if err != nil {
		return err
	}
	res, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("sidecar health returned %d", res.StatusCode)
	}
	return nil
}

// Voices lists voices. Tries the bundled sidecar shape first, then the
// Kokoro-FastAPI/Speaches /v1/audio/voices listing.
func (c *Client) Voices(ctx context.Context) ([]Voice, error) {
	voices, err := c.fetchVoices(ctx, "/voices")
	if err == nil && len(voices) > 0 {
		return voices, nil
	}
	return c.fetchVoices(ctx, "/v1/audio/voices")
}

func (c *Client) fetchVoices(ctx context.Context, path string) ([]Voice, error) {
	req, err := c.newReq(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("voices %s returned %d: %s", path, res.StatusCode, truncateErr(body))
	}
	return parseVoices(body)
}

// parseVoices accepts both {"voices":[...]} and OpenAI-list {"data":[...]}
// shapes; items may carry id, voice_id, or name.
func parseVoices(body []byte) ([]Voice, error) {
	var parsed struct {
		Voices []json.RawMessage `json:"voices"`
		Data   []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	raw := parsed.Voices
	if len(raw) == 0 {
		raw = parsed.Data
	}
	out := make([]Voice, 0, len(raw))
	for _, r := range raw {
		var v struct {
			ID      string `json:"id"`
			VoiceID string `json:"voice_id"`
			Name    string `json:"name"`
			Label   string `json:"label"`
			Lang    string `json:"lang"`
		}
		if err := json.Unmarshal(r, &v); err != nil {
			continue
		}
		id := v.ID
		if id == "" {
			id = v.VoiceID
		}
		if id == "" {
			id = v.Name
		}
		if id == "" {
			continue
		}
		label := v.Label
		if label == "" {
			label = v.Name
		}
		out = append(out, Voice{ID: id, Label: label, Lang: v.Lang})
	}
	return out, nil
}

// Synthesize renders text to audio and returns the bytes plus the
// sidecar's Content-Type. Format comes from the configured response_format.
func (c *Client) Synthesize(ctx context.Context, text, voice string, speed float64) ([]byte, string, error) {
	return c.SynthesizeFormat(ctx, text, voice, speed, c.cfg.ResponseFormat)
}

// SynthesizeFormat is Synthesize with an explicit response_format.
func (c *Client) SynthesizeFormat(ctx context.Context, text, voice string, speed float64, format string) ([]byte, string, error) {
	if strings.TrimSpace(text) == "" {
		return nil, "", errors.New("text is required")
	}
	if len(text) > MaxTextBytes {
		return nil, "", fmt.Errorf("text exceeds %d bytes", MaxTextBytes)
	}
	if voice == "" {
		voice = c.cfg.DefaultVoice
	}
	if format == "" {
		format = "mp3"
	}
	payload, err := json.Marshal(speechRequest{
		Model:          c.cfg.Model,
		Input:          text,
		Voice:          voice,
		ResponseFormat: format,
		Speed:          speed,
	})
	if err != nil {
		return nil, "", err
	}
	req, err := c.newReq(ctx, http.MethodPost, "/v1/audio/speech", bytes.NewReader(payload))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/*, application/octet-stream")
	res, err := c.hc.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, MaxAudioBytes+1))
	if err != nil {
		return nil, "", err
	}
	if len(body) > MaxAudioBytes {
		return nil, "", errors.New("audio response too large")
	}
	if res.StatusCode >= 300 {
		return nil, "", fmt.Errorf("sidecar synthesize returned %d: %s", res.StatusCode, truncateErr(body))
	}
	if len(body) == 0 {
		return nil, "", errors.New("empty audio response")
	}
	return body, res.Header.Get("Content-Type"), nil
}

func truncateErr(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 200 {
		return s[:200] + "…"
	}
	return s
}
