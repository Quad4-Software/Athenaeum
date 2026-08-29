package storage

import (
	"athenaeum/internal/models"
	"context"
)

func (s *Store) migrateV13(ctx context.Context) error {
	has, err := s.tableHasColumn(ctx, "users", "permissions")
	if err != nil {
		return err
	}
	if !has {
		if err := s.exec(ctx, `ALTER TABLE users ADD COLUMN permissions INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}
	_, err = s.db.ExecContext(ctx, `
UPDATE users SET permissions=? WHERE is_admin=1 AND permissions=0`,
		models.AllPermissions)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
UPDATE users SET permissions=? WHERE is_admin=0 AND permissions=0`,
		models.DefaultUserPermissions)
	return err
}
