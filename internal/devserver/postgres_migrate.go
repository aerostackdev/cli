package devserver

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

const migrationsTable = `CREATE TABLE IF NOT EXISTS _aerostack_migrations (
	name TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ DEFAULT NOW()
);`

// ApplyPostgresMigrations applies migrations from migrations_postgres/ to the given Postgres connection.
// Uses _aerostack_migrations table to track applied migrations.
func ApplyPostgresMigrations(connStr string, binding string) (int, error) {
	if strings.Contains(connStr, "$") {
		return 0, fmt.Errorf("connection string has unresolved env vars for binding %q", binding)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return 0, fmt.Errorf("failed to connect: %w", err)
	}
	defer db.Close()

	// Ensure migrations table exists
	if _, err := db.Exec(migrationsTable); err != nil {
		return 0, fmt.Errorf("failed to create migrations table: %w", err)
	}

	entries, err := os.ReadDir("migrations_postgres")
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // No Postgres migrations folder
		}
		return 0, fmt.Errorf("failed to read migrations_postgres: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	applied := 0
	for _, f := range files {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM _aerostack_migrations WHERE name = $1)", f).Scan(&exists)
		if err != nil {
			return applied, fmt.Errorf("failed to check migration %s: %w", f, err)
		}
		if exists {
			continue
		}

		content, err := os.ReadFile("migrations_postgres/" + f)
		if err != nil {
			return applied, fmt.Errorf("failed to read %s: %w", f, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return applied, fmt.Errorf("failed to begin tx for %s: %w", f, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return applied, fmt.Errorf("migration %s failed: %w", f, err)
		}

		if _, err := tx.Exec("INSERT INTO _aerostack_migrations (name) VALUES ($1)", f); err != nil {
			tx.Rollback()
			return applied, fmt.Errorf("failed to record migration %s: %w", f, err)
		}

		if err := tx.Commit(); err != nil {
			return applied, fmt.Errorf("failed to commit %s: %w", f, err)
		}
		applied++
	}

	return applied, nil
}
