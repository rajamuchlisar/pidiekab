package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"pidiekab/internal/platform/config"
	"pidiekab/internal/platform/database"
	"pidiekab/internal/platform/httpserver"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("fatal: invalid configuration on startup", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var dbPool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		dbPool, err = database.NewPool(ctx, database.Config{
			URL:               cfg.DatabaseURL,
			MaxConns:          cfg.DatabaseMaxConns,
			MinConns:          cfg.DatabaseMinConns,
			MaxConnLifetime:   cfg.DatabaseMaxConnLifetime,
			MaxConnIdleTime:   cfg.DatabaseMaxConnIdleTime,
			HealthCheckPeriod: cfg.DatabaseHealthCheckPeriod,
		}, logger)
		if err != nil {
			logger.Error("fatal: database connection failed on startup", "error", err)
			os.Exit(1)
		}
		defer dbPool.Close()
	} else {
		logger.Warn("PIDIEKAB_DATABASE_URL is not configured; running without database connection pool")
	}

	server := httpserver.New(cfg, logger, dbPool)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting pidiekab server",
			"addr", cfg.Address,
			"environment", cfg.Environment,
			"log_level", cfg.LogLevel.String(),
		)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received, commencing graceful shutdown", "timeout", cfg.ShutdownTimeout.String())

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
		logger.Info("server shutdown completed successfully")
	}
}
