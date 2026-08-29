package storage

const schemaV1 = `
CREATE TABLE IF NOT EXISTS books (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	title         TEXT    NOT NULL,
	author        TEXT    NOT NULL DEFAULT '',
	series        TEXT    NOT NULL DEFAULT '',
	series_index  REAL    NOT NULL DEFAULT 0,
	format        TEXT    NOT NULL,
	rel_path      TEXT    NOT NULL UNIQUE,
	abs_path      TEXT    NOT NULL,
	file_size     INTEGER NOT NULL DEFAULT 0,
	has_cover     INTEGER NOT NULL DEFAULT 0,
	language      TEXT    NOT NULL DEFAULT '',
	description   TEXT    NOT NULL DEFAULT '',
	mtime         INTEGER NOT NULL DEFAULT 0,
	added_at      INTEGER NOT NULL,
	modified_at   INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_books_title  ON books(title);
CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);
CREATE INDEX IF NOT EXISTS idx_books_format ON books(format);
CREATE INDEX IF NOT EXISTS idx_books_added  ON books(added_at);
CREATE INDEX IF NOT EXISTS idx_books_series ON books(series);

CREATE TABLE IF NOT EXISTS progress (
	book_id    INTEGER PRIMARY KEY REFERENCES books(id) ON DELETE CASCADE,
	location   TEXT    NOT NULL DEFAULT '',
	percent    REAL    NOT NULL DEFAULT 0,
	updated_at INTEGER NOT NULL
);
`
