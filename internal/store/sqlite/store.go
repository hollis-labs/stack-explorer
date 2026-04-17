package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Store implements the stack-explorer data layer backed by SQLite.
type Store struct {
	db   *sql.DB
	path string
}

// Open creates a new SQLite store at the given path, enables WAL mode and
// foreign keys, and runs any pending migrations.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &Store{db: db, path: path}, nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// DB returns the underlying *sql.DB for direct queries when needed.
func (s *Store) DB() *sql.DB {
	return s.db
}

// Path returns the filesystem path to the SQLite database file.
func (s *Store) Path() string {
	return s.path
}
