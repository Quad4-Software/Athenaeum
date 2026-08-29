package storage

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"athenaeum/internal/models"
)

// GetPocketIDSettings returns the singleton Pocket ID connector config.
func (s *Store) GetPocketIDSettings(ctx context.Context) (models.PocketIDSettings, error) {
	var c models.PocketIDSettings
	var enabled int
	var groupsJSON string
	err := s.queryRowContext(ctx, `
SELECT enabled, base_url, api_key, default_group_ids FROM pocketid_settings WHERE id=1`).
		Scan(&enabled, &c.BaseURL, &c.APIKey, &groupsJSON)
	if err != nil {
		return c, err
	}
	c.Enabled = enabled != 0
	if groupsJSON != "" {
		_ = json.Unmarshal([]byte(groupsJSON), &c.DefaultGroupIDs)
	}
	if c.DefaultGroupIDs == nil {
		c.DefaultGroupIDs = []string{}
	}
	return c, nil
}

// SavePocketIDSettings updates the singleton Pocket ID connector config.
func (s *Store) SavePocketIDSettings(ctx context.Context, c models.PocketIDSettings) error {
	groupsJSON, err := json.Marshal(c.DefaultGroupIDs)
	if err != nil {
		return err
	}
	_, err = s.execContext(ctx, `
UPDATE pocketid_settings SET enabled=?, base_url=?, api_key=?, default_group_ids=?, updated_at=?
WHERE id=1`,
		boolToInt(c.Enabled), strings.TrimSpace(c.BaseURL), c.APIKey, string(groupsJSON), time.Now().Unix())
	return err
}
