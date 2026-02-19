package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	JWTSecret          string `yaml:"JWT_SECRET"           env:"JWT_SECRET"           env-default:"your_secret_key_here"`
	UserServiceAddr    string `yaml:"USER_SERVICE_ADDR"    env:"USER_SERVICE_ADDR"    env-default:"localhost:50052"`
	RoomServiceAddr    string `yaml:"ROOM_SERVICE_ADDR"    env:"ROOM_SERVICE_ADDR"    env-default:"localhost:50051"`
	PaymentServiceAddr string `yaml:"PAYMENT_SERVICE_ADDR" env:"PAYMENT_SERVICE_ADDR" env-default:"localhost:50053"`
	RestPort           string `yaml:"REST_PORT"            env:"REST_PORT"            env-default:"8080"`
}

func New() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig("api/config/config.yaml", &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		return &cfg, nil
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read env config: %w", err)
	}

	return &cfg, nil
}
