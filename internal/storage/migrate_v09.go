package storage

import (
	"context"
)

func (s *Store) migrateV9(ctx context.Context) error {
	return s.exec(ctx, `
CREATE TABLE IF NOT EXISTS audit_log (
	id              INTEGER PRIMARY KEY AUTOINCREMENT,
	actor_id        INTEGER NOT NULL,
	actor_name      TEXT    NOT NULL DEFAULT '',
	target_user_id  INTEGER NOT NULL DEFAULT 0,
	target_name     TEXT    NOT NULL DEFAULT '',
	action          TEXT    NOT NULL,
	details         TEXT    NOT NULL DEFAULT '',
	ip              TEXT    NOT NULL DEFAULT '',
	created_at      INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_audit_log_created ON audit_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_actor ON audit_log(actor_id);
`)
}
