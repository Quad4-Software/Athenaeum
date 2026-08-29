package storage

import (
	"context"
)

func (s *Store) migrateV11(ctx context.Context) error {
	hasSessionID, err := s.tableHasColumn(ctx, "sessions", "session_id")
	if err != nil {
		return err
	}
	if hasSessionID {
		return nil
	}
	return s.exec(ctx, `
ALTER TABLE sessions ADD COLUMN session_id TEXT NOT NULL DEFAULT '';
ALTER TABLE sessions ADD COLUMN ip TEXT NOT NULL DEFAULT '';
ALTER TABLE sessions ADD COLUMN user_agent TEXT NOT NULL DEFAULT '';
ALTER TABLE sessions ADD COLUMN device TEXT NOT NULL DEFAULT '';
ALTER TABLE sessions ADD COLUMN auth_method TEXT NOT NULL DEFAULT 'local';
ALTER TABLE sessions ADD COLUMN created_at INTEGER NOT NULL DEFAULT 0;
ALTER TABLE sessions ADD COLUMN last_seen_at INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);

ALTER TABLE refresh_tokens ADD COLUMN session_id TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_session_id ON refresh_tokens(session_id);

ALTER TABLE users ADD COLUMN email TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN oidc_sub TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oidc_sub ON users(oidc_sub) WHERE oidc_sub != '';

CREATE TABLE IF NOT EXISTS oidc_config (
	id                  INTEGER PRIMARY KEY CHECK (id = 1),
	enabled             INTEGER NOT NULL DEFAULT 0,
	login_local         INTEGER NOT NULL DEFAULT 1,
	issuer_url          TEXT NOT NULL DEFAULT '',
	authorize_url       TEXT NOT NULL DEFAULT '',
	token_url           TEXT NOT NULL DEFAULT '',
	userinfo_url        TEXT NOT NULL DEFAULT '',
	jwks_url            TEXT NOT NULL DEFAULT '',
	logout_url          TEXT NOT NULL DEFAULT '',
	client_id           TEXT NOT NULL DEFAULT '',
	client_secret       TEXT NOT NULL DEFAULT '',
	signing_algorithm   TEXT NOT NULL DEFAULT 'RS256',
	button_text         TEXT NOT NULL DEFAULT 'Sign in with SSO',
	match_by            TEXT NOT NULL DEFAULT 'username',
	auto_register       INTEGER NOT NULL DEFAULT 0,
	auto_launch         INTEGER NOT NULL DEFAULT 0
);
INSERT OR IGNORE INTO oidc_config (id) VALUES (1);
`)
}
