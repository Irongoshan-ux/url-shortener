package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"

	"github.com/Irongoshan-ux/url-shortener/internal/app"
	"github.com/Irongoshan-ux/url-shortener/internal/config"
	"github.com/Irongoshan-ux/url-shortener/internal/handler"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
	"github.com/Irongoshan-ux/url-shortener/internal/service"
	"github.com/Irongoshan-ux/url-shortener/internal/validation"
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

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

	if closer, ok := repo.(interface{ Close() error }); ok {
		defer func() {
			if err := closer.Close(); err != nil {
				logger.Error().Err(err).Msg("Failed to close storage")
			}
		}()
	}

	svc := service.NewService(repo)

	baseURL, err := validation.NormalizeBaseURL(cfg.BaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to normalize base URL")
	}
	facade := handler.NewShortenerFacade(svc, baseURL, logger)

	httpHandler, httpCleanup, err := app.NewHTTPHandler(ctx, cfg, facade, pool, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize HTTP handler")
	}
	defer httpCleanup()

	server, err := app.NewHTTPServer(cfg, httpHandler)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize HTTP server")
	}

	grpcSrv, err := app.NewGRPCServer(cfg, facade, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize gRPC server")
	}

	logger.Info().Str("server", cfg.ServerAddress).Bool("https", cfg.EnableHTTPS).Msg("HTTP server starting")
	logger.Info().Str("grpc_server", cfg.GRPCServer).Bool("https", cfg.EnableHTTPS).Msg("gRPC server starting")
	logger.Info().Str("base_url", cfg.BaseURL).Msg("Base URL")

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return app.Run(gctx, cfg, server, nil)
	})
	g.Go(func() error {
		return app.RunGRPC(gctx, cfg, grpcSrv, cfg.GRPCServer)
	})

	if err := g.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal().Err(err).Msg("Server failed")
	}
	logger.Info().Msg("Server stopped gracefully")
}
