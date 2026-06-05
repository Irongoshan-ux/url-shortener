package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Irongoshan-ux/url-shortener/internal/config"
)

const shutdownTimeout = 10 * time.Second

// NewHTTPServer validates cfg.ServerAddress and returns an http.Server using the given handler.
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

// Run starts the server and shuts it down gracefully when ctx is cancelled.
func Run(ctx context.Context, cfg *config.Config, srv *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- Serve(cfg, srv)
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
