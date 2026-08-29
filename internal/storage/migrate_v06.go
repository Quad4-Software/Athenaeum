package storage

import (
	"context"
)

func (s *Store) migrateV6(ctx context.Context) error {
	hasMeta, err := s.tableHasColumn(ctx, "books", "meta_edited")
	if err != nil {
		return err
	}
	if hasMeta {
		return nil
	}
	return s.exec(ctx, `
ALTER TABLE books ADD COLUMN meta_edited INTEGER NOT NULL DEFAULT 0;
ALTER TABLE books ADD COLUMN cover_edited INTEGER NOT NULL DEFAULT 0;
`)
}
