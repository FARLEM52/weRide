package database

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(ctx context.Context, cfg Config) error {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	migrationsPath := filepath.Join(basepath, "migrations")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	migrationsURL := fmt.Sprintf("file://%s", migrationsPath)

	const maxAttempts = 15
	const retryDelay = 2 * time.Second

	var (
		m   *migrate.Migrate
		err error
	)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		m, err = migrate.New(migrationsURL, connString)
		if err == nil {
			defer m.Close()
			if err = m.Up(); err == nil || err == migrate.ErrNoChange {
				return nil
			}
		}

		if attempt == maxAttempts {
			break
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("migration canceled: %w", ctx.Err())
		case <-time.After(retryDelay):
		}
	}

	return fmt.Errorf("failed to run migrations after %d attempts: %w", maxAttempts, err)
}
