package curation_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"pidiekab/internal/modules/curation/application"
	"pidiekab/internal/modules/curation/domain"
	"pidiekab/internal/modules/curation/infrastructure/postgres"
	curationHTTP "pidiekab/internal/modules/curation/transport/http"
	identityDomain "pidiekab/internal/modules/identity/domain"
	"pidiekab/internal/platform/database"
	"pidiekab/internal/platform/httpserver/middleware"
)

func TestCurationValidation(t *testing.T) {
	if err := domain.ValidateLocale("id"); err != nil {
		t.Errorf("expected id valid, got: %v", err)
	}
	if err := domain.ValidateLocale("de"); err == nil {
		t.Error("expected error for non id/en locale")
	}
}

func TestCurationLifecycle(t *testing.T) {
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
		t.Skipf("skipping curation postgres test: %v", err)
	}
	defer pool.Close()

	repo := postgres.NewCurationRepository(pool)
	service := application.NewCurationService(repo, nil, nil)
	handler := curationHTTP.NewCurationHandler(service, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/admin/curation/{locale}", handler.HandleGetCuratedHomepage)
	mux.HandleFunc("PUT /api/v1/admin/curation/featured", handler.HandleSetFeatured)
	mux.HandleFunc("PUT /api/v1/admin/curation/{locale}/lead", handler.HandleSetLeadStory)
	mux.HandleFunc("PUT /api/v1/admin/curation/{locale}/supporting", handler.HandleSetSupportingStories)

	adminCtx := context.WithValue(ctx, middleware.AdminContextKey, &identityDomain.Admin{
		ID:          1,
		Email:       "admin@pidiekab.go.id",
		DisplayName: "Curation Admin",
		IsActive:    true,
	})

	// 1. Rejection of unpublished content ID (999999)
	err = service.SetLeadStory(ctx, 1, "id", 999999)
	if err != domain.ErrContentNotPublished {
		t.Errorf("expected ErrContentNotPublished for non-existent/unpublished content, got: %v", err)
	}

	// 2. Duplicate content IDs in featured
	err = service.SetFeatured(ctx, 1, []int64{10, 10})
	if err != domain.ErrDuplicateContent {
		t.Errorf("expected ErrDuplicateContent, got: %v", err)
	}

	// 3. Get Curated Homepage
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/curation/id", nil)
	req = req.WithContext(adminCtx)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"locale":"id"`) {
		t.Errorf("response does not contain locale id: %s", rec.Body.String())
	}
}

