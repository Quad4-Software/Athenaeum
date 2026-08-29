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
	if s.driver == DriverPostgres {
		return `(to_timestamp(` + column + `) AT TIME ZONE 'UTC')::date::text`
	}
	return `date(` + column + `, 'unixepoch')`
}
