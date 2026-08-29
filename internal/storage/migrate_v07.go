package storage

import (
	"context"
)

func (s *Store) migrateV7(ctx context.Context) error {
	return s.exec(ctx, `
CREATE TABLE IF NOT EXISTS book_chapters (
	book_id   INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	idx       INTEGER NOT NULL,
	title     TEXT    NOT NULL DEFAULT '',
	start_ms  INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (book_id, idx)
);
CREATE INDEX IF NOT EXISTS idx_book_chapters_book ON book_chapters(book_id);
`)
}
