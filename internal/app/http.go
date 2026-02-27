package app

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/Irongoshan-ux/url-shortener/internal/config"
	"github.com/Irongoshan-ux/url-shortener/internal/handler"
	"github.com/Irongoshan-ux/url-shortener/internal/service"
	"github.com/Irongoshan-ux/url-shortener/internal/validation"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func PingHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if pool == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := pool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func NewHTTPHandler(cfg *config.Config, svc *service.Service, pool *pgxpool.Pool) (http.Handler, error) {
	baseURL, err := validation.NormalizeBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	r := chi.NewRouter()
	r.Use(GzipMiddleware)
	r.Use(LoggingMiddleware(log))
	r.Use(middleware.Recoverer)

	r.Get("/ping", PingHandler(pool))

	h := handler.NewHandler(svc, baseURL, log)
	r.Mount("/", h.Router())

	return r, nil
}
