package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"we_ride/internal/services/room_service/database"
)

type Config struct {
	Postgres database.Config `yaml:"POSTGRES"`

	GRPCPort string `env:"GRPC_PORT" env-default:"50051"   yaml:"GRPC_PORT"`
	GRPCHost string `env:"GRPC_HOST" env-default:"0.0.0.0" yaml:"GRPC_HOST"`

	RESTPort string `env:"REST_PORT" env-default:"8081" yaml:"REST_PORT"`
	RESTHost string `env:"REST_HOST" env-default:"0.0.0.0" yaml:"REST_HOST"`

	UserServiceAddr    string `env:"USER_SERVICE_ADDR"    env-default:"localhost:50052" yaml:"USER_SERVICE_ADDR"`
	PaymentServiceAddr string `env:"PAYMENT_SERVICE_ADDR" env-default:"localhost:50053" yaml:"PAYMENT_SERVICE_ADDR"`
}

func New() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig("internal/services/room_service/config/config.yaml", &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("config.New: %w", err)
		}
		return &cfg, nil
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("config.New env override: %w", err)
	}

	return &cfg, nil
}
