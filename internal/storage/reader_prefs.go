package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"athenaeum/internal/models"
)

// GetReaderPrefs returns userID's synced reader preferences, or an empty
// set if none have been saved yet.
func (s *Store) GetReaderPrefs(ctx context.Context, userID int64) (models.ReaderPrefs, error) {
	var raw string
	var updated int64
	err := s.queryRowContext(ctx,
		`SELECT prefs_json, updated_at FROM reader_prefs WHERE user_id=?`, userID).
		Scan(&raw, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ReaderPrefs{UserID: userID, Prefs: map[string]any{}}, nil
	}
	if err != nil {
		return models.ReaderPrefs{}, err
	}
	prefs := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &prefs); err != nil {
		return models.ReaderPrefs{}, err
	}
	return models.ReaderPrefs{UserID: userID, Prefs: prefs, UpdatedAt: updated}, nil
}

// SaveReaderPrefs upserts userID's reader preferences as JSON.
func (s *Store) SaveReaderPrefs(ctx context.Context, userID int64, prefs map[string]any) (models.ReaderPrefs, error) {
	if prefs == nil {
		prefs = map[string]any{}
	}
	raw, err := json.Marshal(prefs)
	if err != nil {
		return models.ReaderPrefs{}, err
	}
	now := time.Now().Unix()
	_, err = s.execContext(ctx, `
INSERT INTO reader_prefs (user_id, prefs_json, updated_at) VALUES (?,?,?)
ON CONFLICT(user_id) DO UPDATE SET prefs_json=excluded.prefs_json, updated_at=excluded.updated_at`,
		userID, string(raw), now)
	if err != nil {
		return models.ReaderPrefs{}, err
	}
	return models.ReaderPrefs{UserID: userID, Prefs: prefs, UpdatedAt: now}, nil
}
