package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"time"

	"we_ride/internal/services/user_service/db/postgres"
)

type Config struct {
	Postgres postgres.Config `yaml:"POSTGRES" env:"POSTGRES"`

	GRPCPort string `yaml:"GRPC_PORT" env:"GRPC_PORT"  envDefault:"50052"`
	RESRPort string `yaml:"REST_PORT" env:"REST_PORT" envDefault:"8082"`

	JWTAccessTokenTTL time.Duration `yaml:"jwt_access_token_ttl: 15m" env:"JWT_ACCESS_TOKEN_TTL" envDefault:"15m"`
	JwtSecret         string        `yaml:"JWT_SECRET" env:"JWT_SECRET" envDefault:"secret"`
}

func New() (*Config, error) {
	var cfg Config

	path := "api/config/config.yaml"
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	return &cfg, nil
}
