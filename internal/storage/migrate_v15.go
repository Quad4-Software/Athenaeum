package storage

import (
	"context"
)

func (s *Store) migrateV15(ctx context.Context) error {
	if err := s.exec(ctx, `
CREATE TABLE IF NOT EXISTS bookmarks (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id    INTEGER NOT NULL,
	book_id    INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	location   TEXT NOT NULL,
	label      TEXT NOT NULL DEFAULT '',
	created_at INTEGER NOT NULL,
	UNIQUE(user_id, book_id, location)
);
CREATE INDEX IF NOT EXISTS idx_bookmarks_user_book ON bookmarks(user_id, book_id);

CREATE TABLE IF NOT EXISTS highlights (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id    INTEGER NOT NULL,
	book_id    INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	location   TEXT NOT NULL,
	excerpt    TEXT NOT NULL DEFAULT '',
	note       TEXT NOT NULL DEFAULT '',
	color      TEXT NOT NULL DEFAULT 'yellow',
	created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_highlights_user_book ON highlights(user_id, book_id);
`); err != nil {
		return err
	}
	hasReadSec, err := s.tableHasColumn(ctx, "progress", "read_seconds")
	if err != nil {
		return err
	}
	if !hasReadSec {
		if err := s.exec(ctx, `ALTER TABLE progress ADD COLUMN read_seconds INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}
	hasAuto, err := s.tableHasColumn(ctx, "server_config", "auto_scan_enabled")
	if err != nil {
		return err
	}
	if !hasAuto {
		return s.exec(ctx, `
ALTER TABLE server_config ADD COLUMN auto_scan_enabled INTEGER NOT NULL DEFAULT 0;
ALTER TABLE server_config ADD COLUMN auto_scan_interval_sec INTEGER NOT NULL DEFAULT 300;
`)
	}
	return nil
}
