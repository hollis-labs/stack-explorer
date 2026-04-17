package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// migrate runs all pending SQL migrations in order.
func migrate(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL
	)`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	type migration struct {
		version  int
		filename string
	}
	var migrations []migration
	for _, e := range entries {
		name := e.Name()
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}
		ver, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		migrations = append(migrations, migration{version: ver, filename: name})
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	for _, m := range migrations {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.version).Scan(&count); err != nil {
			return fmt.Errorf("check migration %d: %w", m.version, err)
		}
		if count > 0 {
			continue
		}

		data, err := migrationsFS.ReadFile("migrations/" + m.filename)
		if err != nil {
			return fmt.Errorf("read migration %d: %w", m.version, err)
		}

		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("apply migration %d (%s): %w", m.version, m.filename, err)
		}

		if _, err := db.Exec(
			"INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
			m.version, time.Now().UTC().Format(time.RFC3339),
		); err != nil {
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}
	}

	return nil
}
