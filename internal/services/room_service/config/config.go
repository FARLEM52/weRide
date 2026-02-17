package config

import (
	"fmt"

	"we_ride/internal/services/room_service/database"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres database.Config `env_prefix:"POSTGRES_" yaml:"POSTGRES"`

	GRPCPort string `env:"GRPC_PORT" env-default:"50051"   yaml:"GRPC_PORT"`
	GRPCHost string `env:"GRPC_HOST" env-default:"0.0.0.0" yaml:"GRPC_HOST"`

	RESTPort string `env:"REST_PORT" env-default:"8081"    yaml:"REST_PORT"`
	RESTHost string `env:"REST_HOST" env-default:"0.0.0.0" yaml:"REST_HOST"`
}

func New() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadConfig("app/config/config.yaml", &cfg)
	if err != nil {
		return nil, fmt.Errorf("config.New: %w", err)
	}

	return &cfg, nil
}
