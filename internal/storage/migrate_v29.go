package storage

import (
	"context"
)

// feedTokensV29SQLiteDDL creates the feed table on SQLite; it is reused by
// the fresh schema in schema_v1.go.
const feedTokensV29SQLiteDDL = `
CREATE TABLE IF NOT EXISTS feed_tokens (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	token         TEXT    NOT NULL UNIQUE,
	user_id       INTEGER NOT NULL DEFAULT 0,
	name          TEXT    NOT NULL DEFAULT '',
	library_id    INTEGER NOT NULL DEFAULT 0,
	collection_id INTEGER NOT NULL DEFAULT 0,
	created_at    INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_feed_tokens_user ON feed_tokens(user_id);
`

// migrateV29 adds the feed_tokens table for token-scoped audio RSS feeds.
func (s *Store) migrateV29(ctx context.Context) error {
	return s.exec(ctx, feedTokensV29SQLiteDDL)
}

// migratePostgresV29 applies the same change on Postgres dialect.
func (s *Store) migratePostgresV29(ctx context.Context) error {
	return s.exec(ctx, `
CREATE TABLE IF NOT EXISTS feed_tokens (
	id            BIGSERIAL PRIMARY KEY,
	token         TEXT   NOT NULL UNIQUE,
	user_id       BIGINT NOT NULL DEFAULT 0,
	name          TEXT   NOT NULL DEFAULT '',
	library_id    BIGINT NOT NULL DEFAULT 0,
	collection_id BIGINT NOT NULL DEFAULT 0,
	created_at    BIGINT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_feed_tokens_user ON feed_tokens(user_id);
`)
}
