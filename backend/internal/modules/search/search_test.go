package search_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	searchApp "pidiekab/internal/modules/search/application"
	searchDomain "pidiekab/internal/modules/search/domain"
	searchPG "pidiekab/internal/modules/search/infrastructure/postgres"
	searchHTTP "pidiekab/internal/modules/search/transport/http"
	"pidiekab/internal/platform/database"
)

func TestSearchValidation(t *testing.T) {
	t.Run("Query too short", func(t *testing.T) {
		_, err := searchDomain.ValidateQuery("a")
		if err != searchDomain.ErrQueryTooShort {
			t.Errorf("expected ErrQueryTooShort, got: %v", err)
		}
	})

	t.Run("Valid query", func(t *testing.T) {
		q, err := searchDomain.ValidateQuery("pantai pelangi")
		if err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
		if q != "pantai pelangi" {
			t.Errorf("expected trimmed query, got %s", q)
		}
	})

	t.Run("Invalid locale", func(t *testing.T) {
		_, err := searchDomain.ValidateLocale("es")
		if err != searchDomain.ErrInvalidLocale {
			t.Errorf("expected ErrInvalidLocale, got: %v", err)
		}
	})
}

func TestSearchHTTPIntegration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:admin@127.0.0.1:5432/pidiekab_test?sslmode=disable"
	}
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	pool, err := database.NewPool(ctx, database.Config{
		URL:      dbURL,
		MaxConns: 5,
		MinConns: 1,
	}, logger)
	if err != nil {
		t.Skipf("skipping search postgres test: %v", err)
	}
	defer pool.Close()

	repo := searchPG.NewSearchRepository(pool)
	service := searchApp.NewSearchService(repo)
	handler := searchHTTP.NewSearchHandler(service)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// 1. Invalid query length: 400 Bad Request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/id/search?q=x", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for short query, got %d", rec.Code)
	}

	// 2. Valid query
	reqValid := httptest.NewRequest(http.MethodGet, "/api/v1/public/id/search?q=sigli&page=1&per_page=10", nil)
	recValid := httptest.NewRecorder()
	mux.ServeHTTP(recValid, reqValid)
	if recValid.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", recValid.Code, recValid.Body.String())
	}

	var res searchDomain.SearchResponse
	if err := json.Unmarshal(recValid.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode search response: %v", err)
	}
	if res.Meta.Query != "sigli" || res.Meta.Locale != "id" {
		t.Errorf("unexpected meta in response: %+v", res.Meta)
	}
}

