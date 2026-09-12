package storage

import "context"

// insertID runs an INSERT that ends with RETURNING id and scans the new id.
func (s *Store) insertID(ctx context.Context, query string, args ...any) (int64, error) {
	var id int64
	err := s.queryRowContext(ctx, query, args...).Scan(&id)
	return id, err
}

// unixDateExpr returns SQL that formats a unix-seconds column as YYYY-MM-DD UTC.
func (s *Store) unixDateExpr(column string) string {
	return s.driver.unixDateExpr(column)
}

// inListChunkSize bounds each IN() parameter list so batch queries stay
// under SQLite's older 999-variable limit.
const inListChunkSize = 500

// chunkIDs splits ids into consecutive groups of at most size elements.
func chunkIDs(ids []int64, size int) [][]int64 {
	if size <= 0 {
		return nil
	}
	var out [][]int64
	for start := 0; start < len(ids); start += size {
		out = append(out, ids[start:min(start+size, len(ids))])
	}
	return out
}
