package config

import (
	"flag"
	"fmt"
	"net/url"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func Load() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	flag.Parse()

	if err := validateBaseURL(cfg.BaseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &cfg, nil
}

func validateBaseURL(baseURL string) error {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("base URL must include scheme (http:// or https://)")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("base URL must include host")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("base URL scheme must be http or https")
	}

	return nil
}
