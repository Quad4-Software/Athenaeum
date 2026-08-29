package storage

import "strings"

// Driver identifies the SQL backend.
type Driver string

const (
	// DriverSQLite is the default embedded database.
	DriverSQLite Driver = "sqlite"
	// DriverPostgres is an optional external PostgreSQL backend.
	DriverPostgres Driver = "postgres"
)

// ParseDriver normalizes a driver name. Empty defaults to sqlite.
func ParseDriver(name string) (Driver, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "sqlite", "sqlite3":
		return DriverSQLite, nil
	case "postgres", "postgresql", "pg":
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
