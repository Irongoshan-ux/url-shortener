package main

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"

	"github.com/Irongoshan-ux/url-shortener/internal/app"
	"github.com/Irongoshan-ux/url-shortener/internal/config"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
	"github.com/Irongoshan-ux/url-shortener/internal/service"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	repo, err := repository.NewFileRepository(cfg.FileStoragePath)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create repository")
	}

	var db *sql.DB
	if cfg.DatabaseDSN != "" {
		db, err = sql.Open("postgres", cfg.DatabaseDSN)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to open database")
		}
		defer db.Close()
	}

	svc := service.NewService(repo)
	httpHandler, err := app.NewHTTPHandler(cfg, svc, db)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize HTTP handler")
	}

	server, err := app.NewHTTPServer(cfg, httpHandler)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize HTTP server")
	}

	logger.Info().Str("server", cfg.ServerAddress).Msg("Server starting")
	logger.Info().Str("base_url", cfg.BaseURL).Msg("Base URL")
	logger.Info().Str("storage_file", cfg.FileStoragePath).Msg("Storage file")
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal().Err(err).Msg("Server failed to start")
	}
}
