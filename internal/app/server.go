package app

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/Irongoshan-ux/url-shortener/internal/config"
)

func NewHTTPServer(cfg *config.Config, h http.Handler) (*http.Server, error) {
	addr := strings.TrimSpace(cfg.ServerAddress)
	if addr == "" {
		return nil, fmt.Errorf("server address is required")
	}
	if strings.Contains(addr, "://") {
		return nil, fmt.Errorf("server address must be host:port (got %q)", addr)
	}
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return nil, fmt.Errorf("invalid server address %q: %w", addr, err)
	}

	return &http.Server{
		Addr:    addr,
		Handler: h,
	}, nil
}
