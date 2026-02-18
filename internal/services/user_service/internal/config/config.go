package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"time"

	"we_ride/internal/services/user_service/db/postgres"
)

type Config struct {
	Postgres postgres.Config `yaml:"POSTGRES"`

	GRPCPort string `yaml:"GRPC_PORT" env:"GRPC_PORT" env-default:"50052"`

	JWTAccessTokenTTL time.Duration `yaml:"jwt_access_token_ttl" env:"JWT_ACCESS_TOKEN_TTL" env-default:"15m"`
	JwtSecret         string        `yaml:"JWT_SECRET"           env:"JWT_SECRET"           env-default:"secret"`
}

func New() (*Config, error) {
	var cfg Config

	// Пробуем прочитать из файла, если нет — берём из переменных окружения
	if err := cleanenv.ReadConfig("internal/services/user_service/config/local.yaml", &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}
	return &cfg, nil
}
