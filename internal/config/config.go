package config

import (
	"flag"
	"os"
)

const (
	defaultServerAddress   = "localhost:8080"
	defaultBaseURL         = "http://localhost:8080"
	defaultFileStoragePath = "urls.json"
)

const (
	envServerAddress   = "SERVER_ADDRESS"
	envBaseURL         = "BASE_URL"
	envFileStoragePath = "FILE_STORAGE_PATH"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
}

func Load() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.ServerAddress, "a", defaultServerAddress, "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base URL for shortened URLs")
	flag.StringVar(&cfg.FileStoragePath, "f", defaultFileStoragePath, "Path to file for URL storage (JSON)")
	flag.Parse()

	if v := os.Getenv(envServerAddress); v != "" {
		cfg.ServerAddress = v
	}
	if v := os.Getenv(envBaseURL); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv(envFileStoragePath); v != "" {
		cfg.FileStoragePath = v
	}

	return &cfg, nil
}
