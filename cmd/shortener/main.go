package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/Irongoshan-ux/url-shortener/internal/config"
	"github.com/Irongoshan-ux/url-shortener/internal/handler"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
	"github.com/Irongoshan-ux/url-shortener/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	repo := repository.NewMemoryRepository()

	svc := service.NewService(repo)

	h := handler.NewHandler(svc, cfg.BaseURL)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	h.RegisterRoutes(r)

	// Start server
	log.Printf("Server starting on %s", cfg.ServerAddress)
	log.Printf("Base URL: %s", cfg.BaseURL)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
