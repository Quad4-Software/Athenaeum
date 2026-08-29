package storage

import (
	"context"
)

func (s *Store) migrateV10(ctx context.Context) error {
	return s.exec(ctx, `
ALTER TABLE books ADD COLUMN content_hash TEXT NOT NULL DEFAULT '';
ALTER TABLE books ADD COLUMN duplicate_of INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_books_content_hash ON books(content_hash);

CREATE TABLE IF NOT EXISTS user_libraries (
	user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	library_id INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
	PRIMARY KEY (user_id, library_id)
);
CREATE INDEX IF NOT EXISTS idx_user_libraries_user ON user_libraries(user_id);

CREATE TABLE IF NOT EXISTS upload_sessions (
	id          TEXT PRIMARY KEY,
	library_id  INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
	user_id     INTEGER NOT NULL,
	rel_path    TEXT NOT NULL,
	total_size  INTEGER NOT NULL DEFAULT 0,
	offset      INTEGER NOT NULL DEFAULT 0,
	done        INTEGER NOT NULL DEFAULT 0,
	book_id     INTEGER NOT NULL DEFAULT 0,
	created_at  INTEGER NOT NULL,
	updated_at  INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_upload_sessions_user ON upload_sessions(user_id);
`)
}
