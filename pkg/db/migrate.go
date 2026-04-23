package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
)

// Migrate runs all pending migrations
func Migrate(db *sql.DB, migrationsDir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if migrationsDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		migrationsDir = filepath.Join(wd, "migrations")
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// MigrateDown rolls back the last migration
func MigrateDown(db *sql.DB, migrationsDir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if migrationsDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		migrationsDir = filepath.Join(wd, "migrations")
	}

	if err := goose.Down(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// MigrateTo rolls back/forward to a specific version
func MigrateTo(db *sql.DB, migrationsDir string, version int) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if migrationsDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		migrationsDir = filepath.Join(wd, "migrations")
	}

	if err := goose.UpTo(db, migrationsDir, int64(version)); err != nil {
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}

	return nil
}

// CurrentVersion returns the current migration version
func CurrentVersion(db *sql.DB) (int, error) {
	version, err := goose.GetDBVersion(db)
	if err != nil {
		return 0, err
	}
	return int(version), nil
}
