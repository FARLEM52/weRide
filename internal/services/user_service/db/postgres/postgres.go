package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"we_ride/internal/services/room_service/database"
)

type Config struct {
	Host     string `yaml:"POSTGRES_HOST" env:"POSTGRES_HOST" env-default:"localhost"`
	Port     string `yaml:"POSTGRES_PORT" env:"POSTGRES_PORT" env-default:"5432"`
	Username string `yaml:"POSTGRES_USER" env:"POSTGRES_USER" env-default:"root"`
	Password string `yaml:"POSTGRES_PASS" env:"POSTGRES_PASS" env-default:"1234"`
	Database string `yaml:"POSTGRES_DB" env:"POSTGRES_DB" env-default:"repository"`
}

func New(ctx context.Context, config database.Config) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	conn, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		fmt.Println("Postgres ping failed:", err)
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	m, err := migrate.New(
		"file://internal/services/user_service/db/migrations",
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init migrations: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return conn, nil
}
