package storage

import (
	"context"
)

func (s *Store) migrateV17(ctx context.Context) error {
	return s.exec(ctx, `
CREATE TABLE IF NOT EXISTS p2p_config (
	id                  INTEGER PRIMARY KEY CHECK (id = 1),
	enabled             INTEGER NOT NULL DEFAULT 0,
	announce_enabled    INTEGER NOT NULL DEFAULT 0,
	instance_name       TEXT NOT NULL DEFAULT '',
	open_mode           TEXT NOT NULL DEFAULT 'closed',
	identity_created_at INTEGER NOT NULL DEFAULT 0
);
INSERT OR IGNORE INTO p2p_config (id) VALUES (1);

CREATE TABLE IF NOT EXISTS p2p_interfaces (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	name        TEXT NOT NULL UNIQUE,
	type        TEXT NOT NULL,
	enabled     INTEGER NOT NULL DEFAULT 1,
	config_json TEXT NOT NULL DEFAULT '{}',
	sort_order  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS p2p_peers (
	identity_hash TEXT PRIMARY KEY,
	display_name  TEXT NOT NULL DEFAULT '',
	status        TEXT NOT NULL DEFAULT 'pending',
	caps          INTEGER NOT NULL DEFAULT 0,
	note          TEXT NOT NULL DEFAULT '',
	first_seen    INTEGER NOT NULL,
	last_seen     INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_p2p_peers_status ON p2p_peers(status);

CREATE TABLE IF NOT EXISTS p2p_shared_libraries (
	library_id    INTEGER PRIMARY KEY REFERENCES libraries(id) ON DELETE CASCADE,
	default_caps  INTEGER NOT NULL DEFAULT 1
);
`)
}
