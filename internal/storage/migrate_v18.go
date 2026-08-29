package storage

import (
	"context"
)

func (s *Store) migrateV18(ctx context.Context) error {
	has, err := s.tableHasColumn(ctx, "server_config", "scan_workers")
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	return s.exec(ctx, `ALTER TABLE server_config ADD COLUMN scan_workers INTEGER NOT NULL DEFAULT 0`)
}
