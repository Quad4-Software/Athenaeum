package storage

import (
	"context"
	"time"

	"athenaeum/internal/models"
)

// GetTTSSettings returns the singleton Kokoro TTS sidecar configuration.
func (s *Store) GetTTSSettings(ctx context.Context) (models.TTSSettings, error) {
	var c models.TTSSettings
	var enabled int
	err := s.queryRowContext(ctx, `
SELECT enabled, base_url, api_key, default_voice, timeout_sec FROM tts_settings WHERE id=1`).
		Scan(&enabled, &c.BaseURL, &c.APIKey, &c.DefaultVoice, &c.TimeoutSec)
	if err != nil {
		return c, err
	}
	c.Enabled = enabled != 0
	if c.TimeoutSec <= 0 {
		c.TimeoutSec = 60
	}
	if c.DefaultVoice == "" {
		c.DefaultVoice = "af_heart"
	}
	return c, nil
}

// SaveTTSSettings updates the singleton Kokoro TTS sidecar configuration.
func (s *Store) SaveTTSSettings(ctx context.Context, c models.TTSSettings) error {
	if c.TimeoutSec <= 0 {
		c.TimeoutSec = 60
	}
	if c.DefaultVoice == "" {
		c.DefaultVoice = "af_heart"
	}
	_, err := s.execContext(ctx, `
UPDATE tts_settings SET enabled=?, base_url=?, api_key=?, default_voice=?, timeout_sec=?, updated_at=?
WHERE id=1`,
		boolToInt(c.Enabled), c.BaseURL, c.APIKey, c.DefaultVoice, c.TimeoutSec, time.Now().Unix())
	return err
}
