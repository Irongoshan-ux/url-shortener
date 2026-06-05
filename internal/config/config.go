// Package config loads application settings: defaults, JSON file, flags, then environment (each step overrides the previous).
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config holds runtime settings for the HTTP server, storage backends, cookies, and optional audit sinks.
type Config struct {
	ServerAddress   string `json:"server_address" env:"SERVER_ADDRESS"`
	BaseURL         string `json:"base_url" env:"BASE_URL"`
	FileStoragePath string `json:"file_storage_path" env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `json:"database_dsn" env:"DATABASE_DSN"`
	CookieSecret    string `json:"cookie_secret" env:"COOKIE_SECRET"`
	EnableHTTPS     bool   `json:"enable_https" env:"ENABLE_HTTPS"`
	AuditFile       string `json:"audit_file" env:"AUDIT_FILE"`
	AuditURL        string `json:"audit_url" env:"AUDIT_URL"`
}

func defaultConfig() Config {
	return Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}
}

func resolveConfigPath(args []string) string {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var path string
	fs.StringVar(&path, "c", "", "JSON config file path")
	fs.StringVar(&path, "config", "", "JSON config file path")
	_ = fs.Parse(args)
	if path != "" {
		return path
	}
	return os.Getenv("CONFIG")
}

func loadConfigFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file not found: %s", path)
		}
		return fmt.Errorf("read config file %s: %w", path, err)
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse config file %s: %w", path, err)
	}
	return nil
}

// Load parses configuration with priority: defaults → JSON file → flags → environment.
func Load() (*Config, error) {
	return load(flag.CommandLine, os.Args[1:])
}

func load(fs *flag.FlagSet, args []string) (*Config, error) {
	cfg := defaultConfig()

	if path := resolveConfigPath(args); path != "" {
		if err := loadConfigFile(path, &cfg); err != nil {
			return nil, err
		}
	}

	var configPath string
	fs.StringVar(&configPath, "c", "", "JSON config file path (CONFIG)")
	fs.StringVar(&configPath, "config", "", "JSON config file path (CONFIG)")
	fs.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	fs.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL for shortened URLs")
	fs.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "Path to file for URL storage (JSON)")
	fs.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "PostgreSQL connection string (DATABASE_DSN)")
	fs.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "enable HTTPS")
	fs.StringVar(&cfg.AuditFile, "audit-file", cfg.AuditFile, "Append-only audit log file path (AUDIT_FILE)")
	fs.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL, "Remote audit collector POST URL (AUDIT_URL)")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
