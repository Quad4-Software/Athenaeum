package storage

import (
	"context"
)

func (s *Store) migrateV20(ctx context.Context) error {
	return s.exec(ctx, `
CREATE INDEX IF NOT EXISTS idx_collection_items_collection ON collection_items(collection_id);
CREATE INDEX IF NOT EXISTS idx_progress_user_updated ON progress(user_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_books_visible_library_added ON books(library_id, added_at DESC) WHERE hidden = 0;
`)
}
