package storage

import (
	"context"
)

// migrateV22 adds optional Kokoro TTS sidecar settings.
func (s *Store) migrateV22(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS tts_settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			enabled INTEGER NOT NULL DEFAULT 0,
			base_url TEXT NOT NULL DEFAULT '',
			api_key TEXT NOT NULL DEFAULT '',
			default_voice TEXT NOT NULL DEFAULT 'af_heart',
			timeout_sec INTEGER NOT NULL DEFAULT 60,
			updated_at INTEGER NOT NULL
		)`,
		`INSERT OR IGNORE INTO tts_settings (id, enabled, base_url, api_key, default_voice, timeout_sec, updated_at)
			VALUES (1, 0, '', '', 'af_heart', 60, 0)`,
	}
	for _, q := range stmts {
		if err := s.exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}
