package location_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"pidiekab/internal/modules/location/infrastructure/postgres"
	"pidiekab/internal/platform/database"
)

func TestLocationRepositoryIntegration(t *testing.T) {
	dbURL := os.Getenv("PIDIEKAB_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:admin@127.0.0.1:5432/pidiekab_test?sslmode=disable"
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()

	pool, err := database.NewPool(ctx, database.Config{
		URL:             dbURL,
		MaxConns:        5,
		MinConns:        1,
		MaxConnLifetime: 5 * time.Minute,
	}, logger)
	if err != nil {
		t.Skipf("skipping location test: database unreachable: %v", err)
		return
	}
	defer pool.Close()

	repo := postgres.NewRepository(pool)

	// Seed baseline
	if err := repo.SeedPidieBaselineIfEmpty(ctx); err != nil {
		t.Fatalf("failed to seed Pidie baseline: %v", err)
	}

	// Re-run seed (idempotent)
	if err := repo.SeedPidieBaselineIfEmpty(ctx); err != nil {
		t.Fatalf("seed must be idempotent, failed on second run: %v", err)
	}

	// List districts
	districts, err := repo.ListDistricts(ctx)
	if err != nil {
		t.Fatalf("failed to list districts: %v", err)
	}

	if len(districts) == 0 {
		t.Fatal("expected at least 1 district, got 0")
	}

	foundKotaSigli := false
	for _, d := range districts {
		if d.Name == "Kota Sigli" {
			foundKotaSigli = true
			break
		}
	}
	if !foundKotaSigli {
		t.Error("expected to find Kota Sigli district in seed data")
	}
}

