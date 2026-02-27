package main

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/Irongoshan-ux/url-shortener/internal/app"
	"github.com/Irongoshan-ux/url-shortener/internal/config"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
	"github.com/Irongoshan-ux/url-shortener/internal/service"
	"github.com/rs/zerolog"
)

func runMigrations(dsn string, migrationsPath string) error {
	m, err := migrate.New("file://"+filepath.ToSlash(migrationsPath), dsn)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	var repo service.Repository
	var db *sql.DB

	if cfg.DatabaseDSN != "" {
		migrationsPath := "migrations"
		if p := os.Getenv("MIGRATIONS_PATH"); p != "" {
			migrationsPath = p
		}
		if err := runMigrations(cfg.DatabaseDSN, migrationsPath); err != nil {
			logger.Fatal().Err(err).Msg("Failed to run migrations")
		}
		db, err = sql.Open("postgres", cfg.DatabaseDSN)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to open database")
		}
		defer db.Close()
		repo = repository.NewPostgresRepository(db)
	} else if cfg.FileStoragePath != "" {
		repo, err = repository.NewFileRepository(cfg.FileStoragePath)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create file repository")
		}
	} else {
		repo = repository.NewMemoryRepository()
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
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal().Err(err).Msg("Server failed to start")
	}
}
