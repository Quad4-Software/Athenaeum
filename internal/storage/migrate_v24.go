package storage

import (
	"context"
)

// migrateV24 adds invites, webhooks, and Pocket ID settings.
func (s *Store) migrateV24(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS invites (
			id                INTEGER PRIMARY KEY AUTOINCREMENT,
			token             TEXT    NOT NULL UNIQUE,
			kind              TEXT    NOT NULL,
			email             TEXT    NOT NULL DEFAULT '',
			permissions       INTEGER NOT NULL DEFAULT 0,
			created_by        INTEGER NOT NULL,
			expires_at        INTEGER,
			guest_expires_at  INTEGER,
			pocket_id_user_id TEXT    NOT NULL DEFAULT '',
			accepted_at       INTEGER,
			accepted_user_id  INTEGER,
			revoked_at        INTEGER,
			created_at        INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_invites_token ON invites(token)`,
		`CREATE TABLE IF NOT EXISTS webhooks (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			url        TEXT    NOT NULL,
			secret     TEXT    NOT NULL DEFAULT '',
			events     TEXT    NOT NULL DEFAULT '[]',
			enabled    INTEGER NOT NULL DEFAULT 1,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS webhook_deliveries (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			webhook_id   INTEGER NOT NULL,
			event        TEXT    NOT NULL,
			payload      TEXT    NOT NULL,
			status_code  INTEGER NOT NULL DEFAULT 0,
			success      INTEGER NOT NULL DEFAULT 0,
			attempts     INTEGER NOT NULL DEFAULT 0,
			last_error   TEXT    NOT NULL DEFAULT '',
			created_at   INTEGER NOT NULL,
			delivered_at INTEGER
		)`,
		`CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id, created_at DESC)`,
		`CREATE TABLE IF NOT EXISTS pocketid_settings (
			id                INTEGER PRIMARY KEY CHECK (id = 1),
			enabled           INTEGER NOT NULL DEFAULT 0,
			base_url          TEXT    NOT NULL DEFAULT '',
			api_key           TEXT    NOT NULL DEFAULT '',
			default_group_ids TEXT    NOT NULL DEFAULT '[]',
			updated_at        INTEGER NOT NULL
		)`,
		`INSERT OR IGNORE INTO pocketid_settings (id, enabled, base_url, api_key, default_group_ids, updated_at)
		 VALUES (1, 0, '', '', '[]', 0)`,
	}
	for _, q := range stmts {
		if err := s.exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}
