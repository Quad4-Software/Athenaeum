package storage

import (
	"context"
)

func (s *Store) migrateV14(ctx context.Context) error {
	hasGuest, err := s.tableHasColumn(ctx, "users", "is_guest")
	if err != nil {
		return err
	}
	if !hasGuest {
		if err := s.exec(ctx, `
ALTER TABLE users ADD COLUMN is_guest INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN expires_at INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_users_expires ON users(expires_at) WHERE expires_at > 0;
`); err != nil {
			return err
		}
	}
	return s.exec(ctx, `
CREATE TABLE IF NOT EXISTS server_config (
	id                    INTEGER PRIMARY KEY CHECK (id = 1),
	metrics_enabled       INTEGER NOT NULL DEFAULT 0,
	metrics_auth          INTEGER NOT NULL DEFAULT 1,
	metrics_username      TEXT NOT NULL DEFAULT '',
	metrics_password_hash TEXT NOT NULL DEFAULT '',
	trusted_proxies       TEXT NOT NULL DEFAULT '',
	cors_enabled          INTEGER NOT NULL DEFAULT 0,
	cors_origins          TEXT NOT NULL DEFAULT '',
	csp_enabled           INTEGER NOT NULL DEFAULT 1,
	csp_policy            TEXT NOT NULL DEFAULT ''
);
INSERT OR IGNORE INTO server_config (id) VALUES (1);
`)
}
