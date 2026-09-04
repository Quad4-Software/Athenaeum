package storage

import (
	"context"
)

// migrateV25 adds scholarly citation fields and rebuilds books_fts.
func (s *Store) migrateV25(ctx context.Context) error {
	cols := []struct {
		name string
		ddl  string
	}{
		{"doi", `ALTER TABLE books ADD COLUMN doi TEXT NOT NULL DEFAULT ''`},
		{"arxiv_id", `ALTER TABLE books ADD COLUMN arxiv_id TEXT NOT NULL DEFAULT ''`},
		{"pubmed_id", `ALTER TABLE books ADD COLUMN pubmed_id TEXT NOT NULL DEFAULT ''`},
		{"journal", `ALTER TABLE books ADD COLUMN journal TEXT NOT NULL DEFAULT ''`},
		{"volume", `ALTER TABLE books ADD COLUMN volume TEXT NOT NULL DEFAULT ''`},
		{"issue", `ALTER TABLE books ADD COLUMN issue TEXT NOT NULL DEFAULT ''`},
		{"pages", `ALTER TABLE books ADD COLUMN pages TEXT NOT NULL DEFAULT ''`},
		{"published_year", `ALTER TABLE books ADD COLUMN published_year INTEGER NOT NULL DEFAULT 0`},
	}
	for _, c := range cols {
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
	if err := s.exec(ctx, `
CREATE INDEX IF NOT EXISTS idx_books_doi ON books(doi);
CREATE INDEX IF NOT EXISTS idx_books_arxiv_id ON books(arxiv_id);
CREATE INDEX IF NOT EXISTS idx_books_pubmed_id ON books(pubmed_id);
`); err != nil {
		return err
	}
	// Rebuild FTS so doi/arxiv_id/pubmed_id/journal are searchable.
	for _, q := range []string{
		`DROP TRIGGER IF EXISTS books_fts_ai`,
		`DROP TRIGGER IF EXISTS books_fts_ad`,
		`DROP TRIGGER IF EXISTS books_fts_au`,
		`DROP TABLE IF EXISTS books_fts`,
	} {
		if err := s.exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

// migratePostgresV25 adds citation columns and refreshes search_tsv for existing PG DBs on v24.
func (s *Store) migratePostgresV25(ctx context.Context) error {
	cols := []struct {
		name string
		ddl  string
	}{
		{"doi", `ALTER TABLE books ADD COLUMN doi TEXT NOT NULL DEFAULT ''`},
		{"arxiv_id", `ALTER TABLE books ADD COLUMN arxiv_id TEXT NOT NULL DEFAULT ''`},
		{"pubmed_id", `ALTER TABLE books ADD COLUMN pubmed_id TEXT NOT NULL DEFAULT ''`},
		{"journal", `ALTER TABLE books ADD COLUMN journal TEXT NOT NULL DEFAULT ''`},
		{"volume", `ALTER TABLE books ADD COLUMN volume TEXT NOT NULL DEFAULT ''`},
		{"issue", `ALTER TABLE books ADD COLUMN issue TEXT NOT NULL DEFAULT ''`},
		{"pages", `ALTER TABLE books ADD COLUMN pages TEXT NOT NULL DEFAULT ''`},
		{"published_year", `ALTER TABLE books ADD COLUMN published_year INTEGER NOT NULL DEFAULT 0`},
	}
	for _, c := range cols {
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
	if err := s.exec(ctx, `
CREATE INDEX IF NOT EXISTS idx_books_doi ON books(doi);
CREATE INDEX IF NOT EXISTS idx_books_arxiv_id ON books(arxiv_id);
CREATE INDEX IF NOT EXISTS idx_books_pubmed_id ON books(pubmed_id);
`); err != nil {
		return err
	}
	// Recreate generated FTS column to include citation fields.
	if err := s.exec(ctx, `ALTER TABLE books DROP COLUMN IF EXISTS search_tsv`); err != nil {
		return err
	}
	return s.exec(ctx, `
ALTER TABLE books ADD COLUMN search_tsv tsvector GENERATED ALWAYS AS (
	setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
	setweight(to_tsvector('simple', coalesce(author, '')), 'B') ||
	setweight(to_tsvector('simple', coalesce(series, '')), 'C') ||
	setweight(to_tsvector('simple', coalesce(description, '')), 'D') ||
	setweight(to_tsvector('simple', coalesce(doi, '')), 'A') ||
	setweight(to_tsvector('simple', coalesce(arxiv_id, '')), 'A') ||
	setweight(to_tsvector('simple', coalesce(pubmed_id, '')), 'A') ||
	setweight(to_tsvector('simple', coalesce(journal, '')), 'C')
) STORED;
CREATE INDEX IF NOT EXISTS idx_books_search_tsv ON books USING GIN (search_tsv);
`)
}
