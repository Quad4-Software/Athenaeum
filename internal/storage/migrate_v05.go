package storage

import (
	"context"
)

func (s *Store) migrateV5(ctx context.Context) error {
	return s.exec(ctx, `
CREATE TABLE IF NOT EXISTS refresh_tokens (
	token       TEXT PRIMARY KEY,
	user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	expires_at  INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at);
`)
}
