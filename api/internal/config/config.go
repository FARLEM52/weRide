package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"path/filepath"
)

type Config struct {
	JWTSecret       string `yaml:"JWT_SECRET" env:"JWT_SECRET" envDefault:"your_secret_key_here"`
	UserServiceAddr string `yaml:"USER_SERVICE_ADDR" env:"USER_SERVICE_ADDR" envDefault:"localhost:50051"`
	RestPort        string `yaml:"REST_PORT" env:"REST_PORT" envDefault:"8080"`
}

func New() (*Config, error) {
	var cfg Config

	wd, err := os.Getwd() // получаем рабочую директорию
	if err != nil {
		return nil, fmt.Errorf("cannot get working directory: %w", err)
	}

	// Собираем абсолютный путь до config.yaml
	path := filepath.Join(wd, "api", "config", "config.yaml")

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &cfg, nil
}
