package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ApplyMigrations(database *sql.DB, migrationsDir string) error {
	if err := ensureMigrationsTable(database); err != nil {
		return err
	}

	files, err := migrationFiles(migrationsDir)
	if err != nil {
		return err
	}

	for _, fileName := range files {
		alreadyApplied, err := migrationAlreadyApplied(database, fileName)
		if err != nil {
			return err
		}
		if alreadyApplied {
			continue
		}

		migrationPath := filepath.Join(migrationsDir, fileName)
		if err := applyMigrationFile(database, fileName, migrationPath); err != nil {
			return err
		}
	}

	return nil
}

func ensureMigrationsTable(database *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		name TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := database.Exec(query); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	return nil
}

func migrationFiles(migrationsDir string) ([]string, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}

	sort.Strings(files)
	return files, nil
}

func migrationAlreadyApplied(database *sql.DB, name string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = ?);`

	var exists bool
	if err := database.QueryRow(query, name).Scan(&exists); err != nil {
		return false, fmt.Errorf("check migration state for %s: %w", name, err)
	}

	return exists, nil
}

func applyMigrationFile(database *sql.DB, name, path string) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("migration file %s missing: %w", name, err)
		}
		return fmt.Errorf("read migration %s: %w", name, err)
	}

	transaction, err := database.Begin()
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", name, err)
	}

	if _, err := transaction.Exec(string(contents)); err != nil {
		_ = transaction.Rollback()
		return fmt.Errorf("execute migration %s: %w", name, err)
	}

	if _, err := transaction.Exec(`INSERT INTO schema_migrations (name) VALUES (?);`, name); err != nil {
		_ = transaction.Rollback()
		return fmt.Errorf("record migration %s: %w", name, err)
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}

	return nil
}
