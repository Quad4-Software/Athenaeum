package storage

import "context"

// ReplaceBookContent replaces indexed text chunks for a book.
func (s *Store) ReplaceBookContent(ctx context.Context, bookID int64, chunks []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, s.rebind(`DELETE FROM book_content WHERE book_id=?`), bookID); err != nil {
		return err
	}
	if s.driver == DriverSQLite {
		if _, err := tx.ExecContext(ctx, s.rebind(`DELETE FROM book_content_fts WHERE book_id=?`), bookID); err != nil {
			return err
		}
	}
	for i, chunk := range chunks {
		if _, err := tx.ExecContext(ctx, s.rebind(
			`INSERT INTO book_content (book_id, chunk_index, content) VALUES (?,?,?)`),
			bookID, i, chunk); err != nil {
			return err
		}
		if s.driver == DriverSQLite {
			if _, err := tx.ExecContext(ctx, s.rebind(
				`INSERT INTO book_content_fts (content, book_id, chunk_index) VALUES (?,?,?)`),
				chunk, bookID, i); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

// SearchBookContentIDs runs a full-text query against indexed book text and
// returns matching book ids ordered by rank, deduplicated.
func (s *Store) SearchBookContentIDs(ctx context.Context, query string) ([]int64, error) {
	if s.driver == DriverPostgres {
		return s.searchContentPostgres(ctx, query)
	}
	q := buildFTSQuery(query, " AND ")
	if q == "" {
		return nil, nil
	}

	ids, err := s.runContentFTS(ctx, q)
	if err != nil {
		return nil, err
	}
	if len(ids) > 0 {
		return ids, nil
	}
	orQ := buildFTSQuery(query, " OR ")
	if orQ == "" || orQ == q {
		return ids, nil
	}
	return s.runContentFTS(ctx, orQ)
}

func (s *Store) runContentFTS(ctx context.Context, q string) ([]int64, error) {
	rows, err := s.queryContext(ctx, `
SELECT book_id FROM book_content_fts
WHERE book_content_fts MATCH ?
ORDER BY rank
LIMIT ?`, q, maxFTSResults)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectUniqueIDs(rows)
}

func (s *Store) searchContentPostgres(ctx context.Context, query string) ([]int64, error) {
	q := buildPostgresTSQuery(query, " & ")
	if q == "" {
		return nil, nil
	}
	ids, err := s.runContentTSQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	if len(ids) > 0 {
		return ids, nil
	}
	orQ := buildPostgresTSQuery(query, " | ")
	if orQ == "" || orQ == q {
		return ids, nil
	}
	return s.runContentTSQuery(ctx, orQ)
}

func (s *Store) runContentTSQuery(ctx context.Context, q string) ([]int64, error) {
	rows, err := s.queryContext(ctx, `
SELECT book_id FROM book_content
WHERE content_tsv @@ to_tsquery('simple', ?)
ORDER BY ts_rank(content_tsv, to_tsquery('simple', ?)) DESC
LIMIT ?`, q, q, maxFTSResults)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectUniqueIDs(rows)
}

func collectUniqueIDs(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]int64, error) {
	seen := make(map[int64]struct{})
	ids := make([]int64, 0, 64)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
