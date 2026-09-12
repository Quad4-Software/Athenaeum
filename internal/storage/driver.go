package storage

import "strings"

// Driver identifies the SQL backend. It doubles as the SQL dialect:
// call sites branch on the predicate methods instead of comparing
// driver values directly.
type Driver string

const (
	// DriverSQLite is the default embedded database.
	DriverSQLite Driver = "sqlite"
	// DriverPostgres is an optional external PostgreSQL backend.
	DriverPostgres Driver = "postgres"
)

// database/sql driver registration names passed to sql.Open.
const (
	sqliteDBDriver = "sqlite" // modernc.org/sqlite
	pgxDBDriver    = "pgx"    // jackc/pgx stdlib
)

// isPostgres reports whether the driver is the PostgreSQL backend.
func (d Driver) isPostgres() bool { return d == DriverPostgres }

// isSQLite reports whether the driver is the embedded SQLite backend.
func (d Driver) isSQLite() bool { return d == DriverSQLite }

// rebind converts ? placeholders to this driver's form ($N for postgres).
func (d Driver) rebind(query string) string {
	if !d.isPostgres() {
		return query
	}
	return rebindDollar(query)
}

// unixDateExpr returns SQL that formats a unix-seconds column as
// YYYY-MM-DD UTC in this driver's dialect.
func (d Driver) unixDateExpr(column string) string {
	if d.isPostgres() {
		return `(to_timestamp(` + column + `) AT TIME ZONE 'UTC')::date::text`
	}
	return `date(` + column + `, 'unixepoch')`
}

// ParseDriver normalizes a driver name. Empty defaults to sqlite.
func ParseDriver(name string) (Driver, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", string(DriverSQLite), "sqlite3":
		return DriverSQLite, nil
	case string(DriverPostgres), "postgresql", "pg":
		return DriverPostgres, nil
	default:
		return "", errUnsupportedDriver(name)
	}
}

func errUnsupportedDriver(name string) error {
	return &unsupportedDriverError{name: name}
}

type unsupportedDriverError struct{ name string }

func (e *unsupportedDriverError) Error() string {
	return "unsupported database driver " + e.name + " (want sqlite or postgres)"
}
