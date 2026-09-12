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
	publisher        TEXT    NOT NULL DEFAULT '',
	age_rating       TEXT    NOT NULL DEFAULT '',
	reading_direction TEXT   NOT NULL DEFAULT '',
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
