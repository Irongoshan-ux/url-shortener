// Package app composes middleware, routes, and http.Server for the URL shortener binary.
package app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Irongoshan-ux/url-shortener/internal/audit"
	"github.com/Irongoshan-ux/url-shortener/internal/auth"
	"github.com/Irongoshan-ux/url-shortener/internal/config"
	"github.com/Irongoshan-ux/url-shortener/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// PingHandler responds 200, or pings pool when non-nil and returns 500 on database errors.
func PingHandler(pool *pgxpool.Pool, log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if pool == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
		if err := pool.Ping(r.Context()); err != nil {
			log.Error().Err(err).Msg("ping database failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func buildAuditSubject(cfg *config.Config, log zerolog.Logger) (*audit.Subject, func(), error) {
	noop := func() {}
	var observers []audit.Observer
	var closers []func()
	cleanup := func() {
		for i := len(closers) - 1; i >= 0; i-- {
			closers[i]()
		}
	}

	if p := strings.TrimSpace(cfg.AuditFile); p != "" {
		fo, err := audit.NewFileObserver(p, log)
		if err != nil {
			return nil, noop, err
		}
		observers = append(observers, fo)
		closers = append(closers, func() {
			if err := fo.Close(); err != nil {
				log.Error().Err(err).Str("path", p).Msg("audit file: close")
			}
		})
	}
	if raw := strings.TrimSpace(cfg.AuditURL); raw != "" {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme == "" || u.Host == "" {
			cleanup()
			return nil, noop, fmt.Errorf("invalid audit-url %q: must be absolute http(s) URL", cfg.AuditURL)
		}
		observers = append(observers, audit.NewHTTPObserver(raw, log))
	}
	if len(observers) == 0 {
		return nil, noop, nil
	}
	return audit.NewSubject(observers...), cleanup, nil
}

// NewHTTPHandler builds the root chi router: gzip, logging, recovery, auth cookie, /ping and handler routes.
// The returned cleanup closes audit sinks (e.g. audit file); call it on shutdown, typically with defer in main.
func NewHTTPHandler(ctx context.Context, cfg *config.Config, facade *handler.ShortenerFacade, pool *pgxpool.Pool, log zerolog.Logger) (http.Handler, func(), error) {
	_ = ctx

	auditSubject, auditCleanup, err := buildAuditSubject(cfg, log)
	if err != nil {
		return nil, func() {}, err
	}

	r := chi.NewRouter()
	r.Use(GzipMiddleware)
	r.Use(LoggingMiddleware(log))
	r.Use(middleware.Recoverer)
	r.Use(auth.CookieMiddleware(cfg.CookieSecret))

	mountPprof(r)

	r.Get("/ping", PingHandler(pool, log))

	h := handler.NewHandler(facade, log, auditSubject, cfg.TrustedSubnet)
	r.Mount("/", h.Router())

	return r, auditCleanup, nil
}
