package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string `env:"POSTGRES_HOST" env-default:"localhost" yaml:"POSTGRES_HOST"`
	Port     string `env:"POSTGRES_PORT" env-default:"5432"      yaml:"POSTGRES_PORT"`
	Username string `env:"POSTGRES_USER" env-default:"root"      yaml:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASS" env-default:"1234"      yaml:"POSTGRES_PASS"`
	Name     string `env:"POSTGRES_DB"   env-default:"payments"  yaml:"POSTGRES_DB"`
}

func New(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)

	conn, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to db: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	return conn, nil
}
