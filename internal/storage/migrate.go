package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const currentSchemaVersion = 24

func (s *Store) migrate(ctx context.Context) error {
	if s.driver == DriverPostgres {
		return s.migratePostgres(ctx)
	}
	return s.migrateSQLite(ctx)
}

func (s *Store) migrateSQLite(ctx context.Context) error {
	if err := s.exec(ctx, schemaV1); err != nil {
		return err
	}
	version, err := s.userVersion(ctx)
	if err != nil {
		return err
	}
	steps := []struct {
		to int
		fn func(context.Context) error
	}{
		{2, s.migrateV2},
		{3, s.migrateV3},
		{4, s.migrateV4},
		{5, s.migrateV5},
		{6, s.migrateV6},
		{7, s.migrateV7},
		{8, s.migrateV8},
		{9, s.migrateV9},
		{10, s.migrateV10},
		{11, s.migrateV11},
		{12, s.migrateV12},
		{13, s.migrateV13},
		{14, s.migrateV14},
		{15, s.migrateV15},
		{16, s.migrateV16},
		{17, s.migrateV17},
		{18, s.migrateV18},
		{19, s.migrateV19},
		{20, s.migrateV20},
		{21, s.migrateV21},
		{22, s.migrateV22},
		{23, s.migrateV23},
		{24, s.migrateV24},
	}
	for _, step := range steps {
		if version >= step.to {
			continue
		}
		if err := step.fn(ctx); err != nil {
			return err
		}
		if err := s.setUserVersion(ctx, step.to); err != nil {
			return err
		}
		version = step.to
	}
	return s.ensureFTS(ctx)
}

func (s *Store) migratePostgres(ctx context.Context) error {
	if err := s.exec(ctx, `CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL)`); err != nil {
		return err
	}
	version, err := s.userVersion(ctx)
	if err != nil {
		return err
	}
	if version == 0 {
		if err := s.exec(ctx, schemaPostgres); err != nil {
			return fmt.Errorf("postgres schema: %w", err)
		}
		if err := s.exec(ctx, schemaPostgresSeeds); err != nil {
			return fmt.Errorf("postgres seeds: %w", err)
		}
		return s.setUserVersion(ctx, currentSchemaVersion)
	}
	if version > currentSchemaVersion {
		return fmt.Errorf("database schema version %d is newer than this binary (%d)", version, currentSchemaVersion)
	}
	if version < currentSchemaVersion {
		return fmt.Errorf("postgres schema upgrade from %d to %d is not supported yet (recreate the database or stay on sqlite)", version, currentSchemaVersion)
	}
	return nil
}

func (s *Store) ensureFTS(ctx context.Context) error {
	if s.driver == DriverPostgres {
		return nil
	}
	const ftsDDL = `
CREATE VIRTUAL TABLE IF NOT EXISTS books_fts USING fts5(
	title,
	author,
	series,
	description,
	content='books',
	content_rowid='id',
	tokenize='unicode61'
);
`
	if err := s.exec(ctx, ftsDDL); err != nil {
		return err
	}

	triggers := []string{
		`CREATE TRIGGER IF NOT EXISTS books_fts_ai AFTER INSERT ON books BEGIN
			INSERT INTO books_fts(rowid, title, author, series, description)
			VALUES (new.id, new.title, new.author, new.series, new.description);
		END`,
		`CREATE TRIGGER IF NOT EXISTS books_fts_ad AFTER DELETE ON books BEGIN
			INSERT INTO books_fts(books_fts, rowid, title, author, series, description)
			VALUES ('delete', old.id, old.title, old.author, old.series, old.description);
		END`,
		`CREATE TRIGGER IF NOT EXISTS books_fts_au AFTER UPDATE ON books BEGIN
			INSERT INTO books_fts(books_fts, rowid, title, author, series, description)
			VALUES ('delete', old.id, old.title, old.author, old.series, old.description);
			INSERT INTO books_fts(rowid, title, author, series, description)
			VALUES (new.id, new.title, new.author, new.series, new.description);
		END`,
	}
	for _, q := range triggers {
		if err := s.exec(ctx, q); err != nil {
			return err
		}
	}

	var ftsCount int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM books_fts`).Scan(&ftsCount); err != nil {
		return err
	}
	var bookCount int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM books`).Scan(&bookCount); err != nil {
		return err
	}
	if ftsCount == 0 && bookCount > 0 {
		_, err := s.db.ExecContext(ctx, `
INSERT INTO books_fts(rowid, title, author, series, description)
SELECT id, title, author, series, description FROM books`)
		return err
	}
	return nil
}

func (s *Store) userVersion(ctx context.Context) (int, error) {
	if s.driver == DriverPostgres {
		var v sql.NullInt64
		err := s.db.QueryRowContext(ctx, `SELECT version FROM schema_version LIMIT 1`).Scan(&v)
		if err == sql.ErrNoRows {
			return 0, nil
		}
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				return 0, nil
			}
			return 0, err
		}
		if !v.Valid {
			return 0, nil
		}
		return int(v.Int64), nil
	}
	var v int
	err := s.db.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&v)
	return v, err
}

func (s *Store) setUserVersion(ctx context.Context, v int) error {
	if s.driver == DriverPostgres {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		if _, err := tx.ExecContext(ctx, `DELETE FROM schema_version`); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_version (version) VALUES ($1)`, v); err != nil {
			return err
		}
		return tx.Commit()
	}
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(`PRAGMA user_version = %d`, v))
	return err
}

func (s *Store) exec(ctx context.Context, sqlText string) error {
	_, err := s.db.ExecContext(ctx, sqlText)
	return err
}

func (s *Store) tableHasColumn(ctx context.Context, table, column string) (bool, error) {
	if s.driver == DriverPostgres {
		var n int
		err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM information_schema.columns
WHERE table_schema = current_schema() AND table_name = $1 AND column_name = $2`,
			table, column).Scan(&n)
		return n > 0, err
	}
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}
