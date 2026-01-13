package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/Irongoshan-ux/url-shortener/internal/handler"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
	"github.com/Irongoshan-ux/url-shortener/internal/service"
)

func main() {
	addr := ":8080"

	repo := repository.NewMemoryRepository()

	svc := service.NewService(repo)

	h := handler.NewHandler(svc)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	h.RegisterRoutes(r)

	// Start server
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
