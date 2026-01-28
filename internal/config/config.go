package config

import (
	"flag"
	"os"
)

const (
	defaultServerAddress = "localhost:8080"
	defaultBaseURL       = "http://localhost:8080"
)

const (
	envServerAddress = "SERVER_ADDRESS"
	envBaseURL       = "BASE_URL"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func Load() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.ServerAddress, "a", defaultServerAddress, "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base URL for shortened URLs")
	flag.Parse()

	if v := os.Getenv(envServerAddress); v != "" {
		cfg.ServerAddress = v
	}
	if v := os.Getenv(envBaseURL); v != "" {
		cfg.BaseURL = v
	}

	return &cfg, nil
}
