package storage

import (
	"context"
)

func (s *Store) migrateV8(ctx context.Context) error {
	return s.exec(ctx, `
CREATE TABLE IF NOT EXISTS user_favorites (
	user_id    INTEGER NOT NULL,
	book_id    INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	created_at INTEGER NOT NULL,
	PRIMARY KEY (user_id, book_id)
);
CREATE INDEX IF NOT EXISTS idx_user_favorites_user ON user_favorites(user_id);
`)
}
