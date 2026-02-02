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
	"github.com/rs/zerolog"
)

func NewHTTPHandler(cfg *config.Config, svc *service.Service) (http.Handler, error) {
	baseURL, err := validation.NormalizeBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	r := chi.NewRouter()
	r.Use(LoggingMiddleware(log))
	r.Use(middleware.Recoverer)

	h := handler.NewHandler(svc, baseURL)
	r.Mount("/", h.Router())

	return r, nil
}
