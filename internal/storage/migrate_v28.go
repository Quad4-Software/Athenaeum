package storage

import (
	"context"
)

// migrateV28 adds the reading_sessions table that groups heartbeat writes
// into per-book reading sessions.
func (s *Store) migrateV28(ctx context.Context) error {
	return s.exec(ctx, `
CREATE TABLE IF NOT EXISTS reading_sessions (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id    INTEGER NOT NULL,
	book_id    INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	started_at INTEGER NOT NULL,
	ended_at   INTEGER NOT NULL,
	seconds    INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_reading_sessions_user_ended ON reading_sessions(user_id, ended_at);
CREATE INDEX IF NOT EXISTS idx_reading_sessions_book ON reading_sessions(book_id);
`)
}

// migratePostgresV28 applies the same change on Postgres dialect.
func (s *Store) migratePostgresV28(ctx context.Context) error {
	return s.exec(ctx, `
CREATE TABLE IF NOT EXISTS reading_sessions (
	id         BIGSERIAL PRIMARY KEY,
	user_id    BIGINT NOT NULL,
	book_id    BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	started_at BIGINT NOT NULL,
	ended_at   BIGINT NOT NULL,
	seconds    BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_reading_sessions_user_ended ON reading_sessions(user_id, ended_at);
CREATE INDEX IF NOT EXISTS idx_reading_sessions_book ON reading_sessions(book_id);
`)
}
