package storage

import (
	"context"
)

func (s *Store) migrateV4(ctx context.Context) error {
	hasLib, err := s.tableHasColumn(ctx, "books", "library_id")
	if err != nil {
		return err
	}
	if hasLib {
		return nil
	}

	if err := s.exec(ctx, `
CREATE TABLE IF NOT EXISTS libraries (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	name        TEXT    NOT NULL,
	mount_path  TEXT    NOT NULL UNIQUE,
	sort_order  INTEGER NOT NULL DEFAULT 0,
	created_at  INTEGER NOT NULL
);
INSERT OR IGNORE INTO libraries (id, name, mount_path, sort_order, created_at)
VALUES (1, 'Main Library', '', 0, strftime('%s','now'));
`); err != nil {
		return err
	}

	const rebuild = `
CREATE TABLE books_v4 (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	library_id    INTEGER NOT NULL DEFAULT 1 REFERENCES libraries(id) ON DELETE CASCADE,
	title         TEXT    NOT NULL,
	author        TEXT    NOT NULL DEFAULT '',
	series        TEXT    NOT NULL DEFAULT '',
	series_index  REAL    NOT NULL DEFAULT 0,
	format        TEXT    NOT NULL,
	rel_path      TEXT    NOT NULL,
	abs_path      TEXT    NOT NULL,
	file_size     INTEGER NOT NULL DEFAULT 0,
	has_cover     INTEGER NOT NULL DEFAULT 0,
	language      TEXT    NOT NULL DEFAULT '',
	description   TEXT    NOT NULL DEFAULT '',
	mtime         INTEGER NOT NULL DEFAULT 0,
	added_at      INTEGER NOT NULL,
	modified_at   INTEGER NOT NULL,
	UNIQUE(library_id, rel_path)
);
INSERT INTO books_v4 (
	id, library_id, title, author, series, series_index, format, rel_path, abs_path,
	file_size, has_cover, language, description, mtime, added_at, modified_at
)
SELECT
	id, 1, title, author, series, series_index, format, rel_path, abs_path,
	file_size, has_cover, language, description, mtime, added_at, modified_at
FROM books;
DROP TABLE books;
ALTER TABLE books_v4 RENAME TO books;
CREATE INDEX IF NOT EXISTS idx_books_library ON books(library_id);
CREATE INDEX IF NOT EXISTS idx_books_title  ON books(title);
CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);
CREATE INDEX IF NOT EXISTS idx_books_format ON books(format);
CREATE INDEX IF NOT EXISTS idx_books_added  ON books(added_at);
CREATE INDEX IF NOT EXISTS idx_books_series ON books(series);
`
	return s.exec(ctx, rebuild)
}
