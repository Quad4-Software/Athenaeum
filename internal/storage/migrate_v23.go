package storage

import (
	"context"
)

// migrateV23 adds library backend and S3 config columns.
func (s *Store) migrateV23(ctx context.Context) error {
	stmts := []string{
		`ALTER TABLE libraries ADD COLUMN backend TEXT NOT NULL DEFAULT 'local'`,
		`ALTER TABLE libraries ADD COLUMN backend_config TEXT NOT NULL DEFAULT '{}'`,
	}
	for _, q := range stmts {
		if err := s.exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}
