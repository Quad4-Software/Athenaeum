package storage

import (
	"context"
)

func (s *Store) migrateV2(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS users (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	username      TEXT    NOT NULL UNIQUE COLLATE NOCASE,
	password_hash TEXT    NOT NULL,
	is_admin      INTEGER NOT NULL DEFAULT 0,
	created_at    INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
	token      TEXT PRIMARY KEY,
	user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	expires_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS collections (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id     INTEGER NOT NULL DEFAULT 0,
	name        TEXT    NOT NULL,
	description TEXT    NOT NULL DEFAULT '',
	created_at  INTEGER NOT NULL,
	UNIQUE(user_id, name)
);
CREATE INDEX IF NOT EXISTS idx_collections_user ON collections(user_id);

CREATE TABLE IF NOT EXISTS collection_items (
	collection_id INTEGER NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
	book_id       INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	sort_order    INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (collection_id, book_id)
);
CREATE INDEX IF NOT EXISTS idx_collection_items_book ON collection_items(book_id);
`
	if err := s.exec(ctx, ddl); err != nil {
		return err
	}

	hasUserID, err := s.tableHasColumn(ctx, "progress", "user_id")
	if err != nil {
		return err
	}
	if hasUserID {
		return nil
	}

	const progressV2 = `
CREATE TABLE progress_v2 (
	user_id    INTEGER NOT NULL DEFAULT 0,
	book_id    INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	location   TEXT    NOT NULL DEFAULT '',
	percent    REAL    NOT NULL DEFAULT 0,
	updated_at INTEGER NOT NULL,
	PRIMARY KEY (user_id, book_id)
);
INSERT INTO progress_v2 (user_id, book_id, location, percent, updated_at)
SELECT 0, book_id, location, percent, updated_at FROM progress;
DROP TABLE progress;
ALTER TABLE progress_v2 RENAME TO progress;
`
	return s.exec(ctx, progressV2)
}
