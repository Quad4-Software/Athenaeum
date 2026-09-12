package storage

import (
	"context"
)

// booksV27Columns is shared by the SQLite and Postgres migrations; the DDL is
// dialect-neutral.
var booksV27Columns = []struct {
	name string
	ddl  string
}{
	{"publisher", `ALTER TABLE books ADD COLUMN publisher TEXT NOT NULL DEFAULT ''`},
	{"age_rating", `ALTER TABLE books ADD COLUMN age_rating TEXT NOT NULL DEFAULT ''`},
	{"reading_direction", `ALTER TABLE books ADD COLUMN reading_direction TEXT NOT NULL DEFAULT ''`},
}

// migrateV27 adds ComicInfo.xml metadata columns to books.
func (s *Store) migrateV27(ctx context.Context) error {
	for _, c := range booksV27Columns {
		has, err := s.tableHasColumn(ctx, "books", c.name)
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

// migratePostgresV27 applies the same changes on Postgres dialect.
func (s *Store) migratePostgresV27(ctx context.Context) error {
	return s.migrateV27(ctx)
}
