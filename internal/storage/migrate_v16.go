package storage

import (
	"context"
)

func (s *Store) migrateV16(ctx context.Context) error {
	hasHidden, err := s.tableHasColumn(ctx, "books", "hidden")
	if err != nil {
		return err
	}
	if !hasHidden {
		if err := s.exec(ctx, `
ALTER TABLE books ADD COLUMN hidden INTEGER NOT NULL DEFAULT 0;
ALTER TABLE books ADD COLUMN audiobook_set_id INTEGER NOT NULL DEFAULT 0;
`); err != nil {
			return err
		}
	}
	return s.exec(ctx, `
CREATE TABLE IF NOT EXISTS audiobook_tracks (
	set_book_id INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	track_index INTEGER NOT NULL,
	rel_path TEXT NOT NULL,
	title TEXT NOT NULL DEFAULT '',
	format TEXT NOT NULL DEFAULT '',
	file_size INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (set_book_id, track_index)
);
CREATE INDEX IF NOT EXISTS idx_audiobook_tracks_set ON audiobook_tracks(set_book_id);
`)
}
