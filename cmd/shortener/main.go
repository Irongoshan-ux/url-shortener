package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
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
	printBuildInfo()

	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	var repo service.Repository
	var pool *pgxpool.Pool

	if cfg.DatabaseDSN != "" {
		migrationsPath := "migrations"
		if p := os.Getenv("MIGRATIONS_PATH"); p != "" {
			migrationsPath = p
		}
		if err := runMigrations(cfg.DatabaseDSN, migrationsPath); err != nil {
			logger.Fatal().Err(err).Msg("Failed to run migrations")
		}
		pool, err = pgxpool.New(ctx, cfg.DatabaseDSN)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create connection pool")
		}
		defer pool.Close()
		repo = repository.NewPostgresRepository(pool)
		logger.Info().Str("storage", "postgres").Msg("Using PostgreSQL storage")
	} else if cfg.FileStoragePath != "" {
		repo, err = repository.NewFileRepository(cfg.FileStoragePath)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create file repository")
		}
		logger.Info().Str("storage", "file").Str("path", cfg.FileStoragePath).Msg("Using file storage")
	} else {
		repo = repository.NewMemoryRepository()
		logger.Info().Str("storage", "memory").Msg("Using in-memory storage")
	}

	svc := service.NewService(repo)
	httpHandler, httpCleanup, err := app.NewHTTPHandler(ctx, cfg, svc, pool, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize HTTP handler")
	}
	defer httpCleanup()

	server, err := app.NewHTTPServer(cfg, httpHandler)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize HTTP server")
	}

	logger.Info().Str("server", cfg.ServerAddress).Bool("https", cfg.EnableHTTPS).Msg("Server starting")
	logger.Info().Str("base_url", cfg.BaseURL).Msg("Base URL")
	if err := app.Serve(cfg, server); err != nil {
		logger.Fatal().Err(err).Msg("Server failed to start")
	}
}
