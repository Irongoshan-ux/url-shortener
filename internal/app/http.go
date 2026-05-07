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
	"github.com/Irongoshan-ux/url-shortener/internal/service"
	"github.com/Irongoshan-ux/url-shortener/internal/validation"
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

func buildAuditSubject(cfg *config.Config, log zerolog.Logger) (*audit.Subject, error) {
	var observers []audit.Observer
	if p := strings.TrimSpace(cfg.AuditFile); p != "" {
		observers = append(observers, audit.NewFileObserver(p, log))
	}
	if raw := strings.TrimSpace(cfg.AuditURL); raw != "" {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return nil, fmt.Errorf("invalid audit-url %q: must be absolute http(s) URL", cfg.AuditURL)
		}
		observers = append(observers, audit.NewHTTPObserver(raw, log))
	}
	if len(observers) == 0 {
		return nil, nil
	}
	return audit.NewSubject(observers...), nil
}

// NewHTTPHandler builds the root chi router: gzip, logging, recovery, auth cookie, /ping, optional pprof, and handler routes.
func NewHTTPHandler(ctx context.Context, cfg *config.Config, svc *service.Service, pool *pgxpool.Pool, log zerolog.Logger) (http.Handler, error) {
	_ = ctx
	baseURL, err := validation.NormalizeBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	auditSubject, err := buildAuditSubject(cfg, log)
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()
	MountPprof(r)
	r.Use(GzipMiddleware)
	r.Use(LoggingMiddleware(log))
	r.Use(middleware.Recoverer)
	r.Use(auth.CookieMiddleware(cfg.CookieSecret))

	r.Get("/ping", PingHandler(pool, log))

	h := handler.NewHandler(svc, baseURL, log, auditSubject)
	r.Mount("/", h.Router())

	return r, nil
}
