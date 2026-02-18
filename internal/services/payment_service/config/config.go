package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"we_ride/internal/services/payment_service/database"
)

type Config struct {
	Postgres database.Config `yaml:"POSTGRES"`

	GRPCPort string `env:"GRPC_PORT" env-default:"50053" yaml:"GRPC_PORT"`
	GRPCHost string `env:"GRPC_HOST" env-default:"0.0.0.0" yaml:"GRPC_HOST"`

	YookassaShopID    string `env:"YOOKASSA_SHOP_ID"    yaml:"YOOKASSA_SHOP_ID"`
	YookassaSecretKey string `env:"YOOKASSA_SECRET_KEY" yaml:"YOOKASSA_SECRET_KEY"`
}

func New() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig("internal/services/payment_service/config/config.yaml", &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("config.New: %w", err)
		}
	}
	return &cfg, nil
}
