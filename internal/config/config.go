package config

import (
	"flag"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	CookieSecret    string `env:"COOKIE_SECRET"`
}

func Load() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "Path to file for URL storage (JSON)")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "PostgreSQL connection string (DATABASE_DSN)")
	flag.StringVar(&cfg.CookieSecret, "s", "", "Secret for signing user cookie (COOKIE_SECRET)")
	flag.Parse()

	// Env overrides flag/default (cleanenv reads only from env when set)
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
