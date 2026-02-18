package database

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(ctx context.Context, cfg Config) error {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	migrationsPath := filepath.Join(basepath, "migrations")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	migrationsURL := fmt.Sprintf("file://%s", migrationsPath)

	m, err := migrate.New(migrationsURL, connString)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
