package storage

import (
	"context"
	"strings"
	"unicode"
)

// Cap FTS id lists so broad searches stay bounded for IN clauses and memory.
const maxFTSResults = 500

// searchFTS runs a full-text query and returns matching book ids ordered by rank.
func (s *Store) searchFTS(ctx context.Context, query string) ([]int64, error) {
	if s.driver == DriverPostgres {
		return s.searchBooksPostgres(ctx, query)
	}
	q := buildFTSQuery(query, " AND ")
	if q == "" {
		return nil, nil
	}

	ids, err := s.runFTS(ctx, q)
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
	return s.runFTS(ctx, orQ)
}

func (s *Store) runFTS(ctx context.Context, q string) ([]int64, error) {
	rows, err := s.queryContext(ctx, `
SELECT books.id
FROM books_fts
JOIN books ON books.id = books_fts.rowid
WHERE books_fts MATCH ?
ORDER BY rank
LIMIT ?`, q, maxFTSResults)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int64, 0, 64)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Store) searchBooksPostgres(ctx context.Context, query string) ([]int64, error) {
	q := buildPostgresTSQuery(query, " & ")
	if q == "" {
		return nil, nil
	}
	ids, err := s.runBooksTSQuery(ctx, q)
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
	return s.runBooksTSQuery(ctx, orQ)
}

func (s *Store) runBooksTSQuery(ctx context.Context, q string) ([]int64, error) {
	rows, err := s.queryContext(ctx, `
SELECT id
FROM books
WHERE search_tsv @@ to_tsquery('simple', ?)
ORDER BY ts_rank(search_tsv, to_tsquery('simple', ?)) DESC
LIMIT ?`, q, q, maxFTSResults)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]int64, 0, 64)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// buildFTSQuery turns user input into a safe FTS5 prefix query.
func buildFTSQuery(input, join string) string {
	var terms []string
	for raw := range strings.FieldsSeq(input) {
		t := sanitizeSearchToken(raw)
		if t == "" {
			continue
		}
		t = strings.ReplaceAll(t, `"`, `""`)
		terms = append(terms, `"`+t+`"*`)
	}
	return strings.Join(terms, join)
}

// buildPostgresTSQuery turns user input into a to_tsquery prefix expression.
func buildPostgresTSQuery(input, join string) string {
	var terms []string
	for raw := range strings.FieldsSeq(input) {
		t := sanitizeSearchToken(raw)
		if t == "" {
			continue
		}
		terms = append(terms, t+":*")
	}
	return strings.Join(terms, join)
}

func sanitizeSearchToken(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
