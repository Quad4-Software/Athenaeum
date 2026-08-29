// Package storage provides persistence for the library. SQLite (modernc,
// pure Go, no CGO) is the default. PostgreSQL is an optional backend with
// equivalent full-text search via tsvector and GIN indexes.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("not found")

// ErrConflict is returned when an update cannot apply due to state conflict.
var ErrConflict = errors.New("conflict")

// OpenOptions configures how the store connects to a database.
type OpenOptions struct {
	Driver Driver
	// Path is the SQLite file path (ignored for postgres).
	Path string
	// URL is the PostgreSQL connection string (required for postgres).
	URL string
}

// Store wraps the database connection and exposes typed queries.
type Store struct {
	db     *sql.DB
	driver Driver
}

// Open opens and migrates a SQLite database at path. Prefer OpenWith when
// selecting a driver from configuration.
func Open(path string) (*Store, error) {
	return OpenWith(OpenOptions{Driver: DriverSQLite, Path: path})
}

// OpenWith opens and migrates a database using the given options.
func OpenWith(opts OpenOptions) (*Store, error) {
	driver := opts.Driver
	if driver == "" {
		driver = DriverSQLite
	}
	var (
		db  *sql.DB
		err error
	)
	switch driver {
	case DriverSQLite:
		if strings.TrimSpace(opts.Path) == "" {
			return nil, errors.New("sqlite database path is required")
		}
		dsn := fmt.Sprintf(
			"file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)",
			opts.Path,
		)
		db, err = sql.Open("sqlite", dsn)
		if err != nil {
			return nil, err
		}
		db.SetMaxOpenConns(1)
	case DriverPostgres:
		url := strings.TrimSpace(opts.URL)
		if url == "" {
			return nil, errors.New("postgres requires ATHENAEUM_DATABASE_URL")
		}
		db, err = sql.Open("pgx", url)
		if err != nil {
			return nil, err
		}
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
	default:
		return nil, errUnsupportedDriver(string(driver))
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database ping: %w", err)
	}

	s := &Store{db: db, driver: driver}
	if err := s.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Driver returns the active SQL backend.
func (s *Store) Driver() Driver { return s.driver }

// Close releases the database connection.
func (s *Store) Close() error { return s.db.Close() }

// DB exposes the underlying connection for tests.
func (s *Store) DB() *sql.DB { return s.db }
