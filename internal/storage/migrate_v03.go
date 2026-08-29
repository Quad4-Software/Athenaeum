package storage

import (
	"context"
)

func (s *Store) migrateV3(ctx context.Context) error {
	hasKind, err := s.tableHasColumn(ctx, "collections", "kind")
	if err != nil {
		return err
	}
	if hasKind {
		return nil
	}
	return s.exec(ctx, `
ALTER TABLE collections ADD COLUMN kind TEXT NOT NULL DEFAULT 'manual';
ALTER TABLE collections ADD COLUMN query_json TEXT NOT NULL DEFAULT '';
`)
}
