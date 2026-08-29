package storage

import (
	"context"
)

func (s *Store) migrateV19(ctx context.Context) error {
	return s.exec(ctx, `
DROP TABLE IF EXISTS p2p_shared_libraries;
DROP TABLE IF EXISTS p2p_peers;
DROP TABLE IF EXISTS p2p_interfaces;
DROP TABLE IF EXISTS p2p_config;
`)
}
