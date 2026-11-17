package database_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"pidiekab/internal/platform/database"
)

func TestNewPoolValidation(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()

	// 1. Empty URL
	_, err := database.NewPool(ctx, database.Config{URL: ""}, logger)
	if err == nil {
		t.Error("expected error for empty database URL, got nil")
	}

	// 2. Unreachable database URL with explicit short timeout
	shortCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	badCfg := database.Config{
		URL: "postgres://invalid_user:invalid_pass@127.0.0.1:54321/nonexistent?sslmode=disable",
	}
	_, err = database.NewPool(shortCtx, badCfg, logger)
	if err == nil {
		t.Error("expected error for unreachable database, got nil")
	}

	// 3. Ping helper on nil pool
	if err := database.Ping(ctx, nil); err == nil {
		t.Error("expected error for nil pool ping, got nil")
	}

	// 4. WithTransaction on nil pool
	err = database.WithTransaction(ctx, nil, func(tx pgx.Tx) error {
		return nil
	})
	if err == nil {
		t.Error("expected error for nil pool transaction, got nil")
	}
}

func TestNewPoolIntegrationWithLocalPostgres(t *testing.T) {
	dbURL := os.Getenv("PIDIEKAB_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:admin@127.0.0.1:5432/pidiekab_test?sslmode=disable"
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()

	cfg := database.Config{
		URL:               dbURL,
		MaxConns:          5,
		MinConns:          1,
		MaxConnLifetime:   30 * time.Minute,
		MaxConnIdleTime:   10 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}

	pool, err := database.NewPool(ctx, cfg, logger)
	if err != nil {
		t.Skipf("skipping integration test: local PostgreSQL not reachable: %v", err)
		return
	}
	defer pool.Close()

	// Verify Ping
	if err := database.Ping(ctx, pool); err != nil {
		t.Fatalf("database ping failed: %v", err)
	}

	// Verify basic query
	var result int
	err = pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("failed to query row: %v", err)
	}
	if result != 1 {
		t.Errorf("expected result 1, got %d", result)
	}

	// Verify WithTransaction Commit
	err = database.WithTransaction(ctx, pool, func(tx pgx.Tx) error {
		var val string
		return tx.QueryRow(ctx, "SELECT 'committed'").Scan(&val)
	})
	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}

	// Verify WithTransaction Rollback on error
	expectedErr := errors.New("simulated error to trigger rollback")
	err = database.WithTransaction(ctx, pool, func(tx pgx.Tx) error {
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected transaction to return %v, got %v", expectedErr, err)
	}
}
