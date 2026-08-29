package storage

import (
	"context"
)

// migrateV21 adds tags, ratings, reader prefs, auth/TOTP, sharing, SMTP,
// content search, KOSync, and offline grants.
func (s *Store) migrateV21(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL COLLATE NOCASE UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS book_tags (
			book_id INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
			tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
			PRIMARY KEY (book_id, tag_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_book_tags_tag ON book_tags(tag_id)`,
		`CREATE TABLE IF NOT EXISTS book_ratings (
			user_id INTEGER NOT NULL,
			book_id INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
			rating INTEGER NOT NULL CHECK(rating BETWEEN 1 AND 5),
			updated_at INTEGER NOT NULL,
			PRIMARY KEY (user_id, book_id)
		)`,
		`CREATE TABLE IF NOT EXISTS reader_prefs (
			user_id INTEGER NOT NULL PRIMARY KEY,
			prefs_json TEXT NOT NULL DEFAULT '{}',
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS auth_settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			allow_registration INTEGER NOT NULL DEFAULT 0,
			require_totp INTEGER NOT NULL DEFAULT 0,
			updated_at INTEGER NOT NULL
		)`,
		`INSERT OR IGNORE INTO auth_settings (id, allow_registration, require_totp, updated_at) VALUES (1, 0, 0, 0)`,
		`CREATE TABLE IF NOT EXISTS share_links (
			id INTEGER PRIMARY KEY,
			token TEXT NOT NULL UNIQUE,
			book_id INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
			created_by INTEGER NOT NULL,
			expires_at INTEGER,
			created_at INTEGER NOT NULL,
			download_count INTEGER NOT NULL DEFAULT 0,
			max_downloads INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_share_links_book ON share_links(book_id)`,
		`CREATE TABLE IF NOT EXISTS smtp_settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			enabled INTEGER NOT NULL DEFAULT 0,
			host TEXT NOT NULL DEFAULT '',
			port INTEGER NOT NULL DEFAULT 587,
			username TEXT NOT NULL DEFAULT '',
			password TEXT NOT NULL DEFAULT '',
			from_addr TEXT NOT NULL DEFAULT '',
			use_tls INTEGER NOT NULL DEFAULT 1,
			updated_at INTEGER NOT NULL
		)`,
		`INSERT OR IGNORE INTO smtp_settings (id, enabled, host, port, username, password, from_addr, use_tls, updated_at)
			VALUES (1, 0, '', 587, '', '', '', 1, 0)`,
		`CREATE TABLE IF NOT EXISTS user_kindle_email (
			user_id INTEGER NOT NULL PRIMARY KEY,
			email TEXT NOT NULL DEFAULT '',
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS book_content (
			book_id INTEGER NOT NULL,
			chunk_index INTEGER NOT NULL,
			content TEXT NOT NULL,
			PRIMARY KEY (book_id, chunk_index)
		)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS book_content_fts USING fts5(
			content,
			book_id UNINDEXED,
			chunk_index UNINDEXED,
			tokenize='unicode61'
		)`,
		`CREATE TABLE IF NOT EXISTS kosync_documents (
			user_id INTEGER NOT NULL,
			document TEXT NOT NULL,
			progress TEXT NOT NULL DEFAULT '',
			percentage REAL NOT NULL DEFAULT 0,
			device TEXT NOT NULL DEFAULT '',
			device_id TEXT NOT NULL DEFAULT '',
			timestamp INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (user_id, document)
		)`,
		`CREATE TABLE IF NOT EXISTS offline_grants (
			user_id INTEGER NOT NULL,
			book_id INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
			created_at INTEGER NOT NULL,
			PRIMARY KEY (user_id, book_id)
		)`,
	}
	for _, q := range stmts {
		if err := s.exec(ctx, q); err != nil {
			return err
		}
	}

	userCols := []struct{ col, ddl string }{
		{"totp_secret", `ALTER TABLE users ADD COLUMN totp_secret TEXT`},
		{"totp_enabled", `ALTER TABLE users ADD COLUMN totp_enabled INTEGER NOT NULL DEFAULT 0`},
	}
	for _, c := range userCols {
		has, err := s.tableHasColumn(ctx, "users", c.col)
		if err != nil {
			return err
		}
		if !has {
			if err := s.exec(ctx, c.ddl); err != nil {
				return err
			}
		}
	}

	oidcCols := []struct{ col, ddl string }{
		{"admin_groups", `ALTER TABLE oidc_config ADD COLUMN admin_groups TEXT NOT NULL DEFAULT ''`},
		{"group_claim", `ALTER TABLE oidc_config ADD COLUMN group_claim TEXT NOT NULL DEFAULT 'groups'`},
	}
	for _, c := range oidcCols {
		has, err := s.tableHasColumn(ctx, "oidc_config", c.col)
		if err != nil {
			return err
		}
		if !has {
			if err := s.exec(ctx, c.ddl); err != nil {
				return err
			}
		}
	}

	return nil
}
