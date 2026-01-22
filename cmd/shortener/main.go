package main

import (
	"log"

	"github.com/Irongoshan-ux/url-shortener/internal/app"
	"github.com/Irongoshan-ux/url-shortener/internal/config"
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
	httpHandler, err := app.NewHTTPHandler(cfg, svc)
	if err != nil {
		log.Fatalf("Failed to initialize HTTP handler: %v", err)
	}

	server, err := app.NewHTTPServer(cfg, httpHandler)
	if err != nil {
		log.Fatalf("Failed to initialize HTTP server: %v", err)
	}

	log.Printf("Server starting on %s", cfg.ServerAddress)
	log.Printf("Base URL: %s", cfg.BaseURL)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
