// Package config loads application settings from flags and environment variables (env overrides defaults set by flags).
package config

import (
	"flag"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config holds runtime settings for the HTTP server, storage backends, cookies, and optional audit sinks.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	CookieSecret    string `env:"COOKIE_SECRET"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
}

// Load parses flags, then merges environment into cfg (see field env tags). Call once from main.
func Load() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "Path to file for URL storage (JSON)")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "PostgreSQL connection string (DATABASE_DSN)")
	flag.StringVar(&cfg.CookieSecret, "s", "", "Secret for signing user cookie (COOKIE_SECRET)")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "Append-only audit log file path (AUDIT_FILE)")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "Remote audit collector POST URL (AUDIT_URL)")
	flag.Parse()

	// Env overrides flag/default (cleanenv reads only from env when set)
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
