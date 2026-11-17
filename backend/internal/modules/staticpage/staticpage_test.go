package staticpage_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"pidiekab/internal/modules/identity/domain"
	pageApp "pidiekab/internal/modules/staticpage/application"
	pageDomain "pidiekab/internal/modules/staticpage/domain"
	pagePG "pidiekab/internal/modules/staticpage/infrastructure/postgres"
	pageHTTP "pidiekab/internal/modules/staticpage/transport/http"
	"pidiekab/internal/platform/database"
	"pidiekab/internal/platform/httpserver/middleware"
)

func TestStaticPageValidation(t *testing.T) {
	t.Run("Valid Page Key", func(t *testing.T) {
		if err := pageDomain.ValidatePageKey("about-us"); err != nil {
			t.Errorf("expected valid, got: %v", err)
		}
	})

	t.Run("Invalid Page Key with spaces or uppercase", func(t *testing.T) {
		if err := pageDomain.ValidatePageKey("About Us"); err == nil {
			t.Error("expected error for uppercase and spaces")
		}
	})

	t.Run("Valid Slug", func(t *testing.T) {
		if err := pageDomain.ValidateSlug("kebijakan-privasi"); err != nil {
			t.Errorf("expected valid slug, got: %v", err)
		}
	})

	t.Run("Invalid Body JSON", func(t *testing.T) {
		if err := pageDomain.ValidateBodyJSON([]byte(`"not-an-object"`)); err == nil {
			t.Error("expected error for non-object JSON")
		}
	})
}

func TestStaticPageLifecycle(t *testing.T) {
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
		t.Skipf("skipping static page postgres test: %v", err)
	}
	defer pool.Close()

	repo := pagePG.NewPageRepository(pool)
	service := pageApp.NewPageService(repo, nil)
	handler := pageHTTP.NewPageHandler(service, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/admin/pages", handler.HandleCreatePage)
	mux.HandleFunc("GET /api/v1/admin/pages", handler.HandleListPages)
	mux.HandleFunc("GET /api/v1/admin/pages/{id}", handler.HandleGetPage)
	mux.HandleFunc("PUT /api/v1/admin/pages/{id}/translations/{locale}", handler.HandleSaveTranslation)
	mux.HandleFunc("POST /api/v1/admin/pages/{id}/translations/{locale}/publish", handler.HandlePublishTranslation)

	adminCtx := context.WithValue(ctx, middleware.AdminContextKey, &domain.Admin{
		ID:          1,
		Email:       "admin@pidiekab.go.id",
		DisplayName: "Static Page Admin",
		IsActive:    true,
	})

	// 1. Create static page with unique key
	uniqueKey := fmt.Sprintf("terms-%d", time.Now().UnixNano())
	createPayload := []byte(fmt.Sprintf(`{"page_key":%q}`, uniqueKey))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/pages", bytes.NewReader(createPayload))
	req = req.WithContext(adminCtx)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 created, got %d: %s", rec.Code, rec.Body.String())
	}

	var res struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	pageID := res.Data.ID

	if pageID <= 0 {
		t.Fatalf("expected valid page id, got %d", pageID)
	}

	// 2. Save translation ID
	bodyJSON := []byte(`{"blocks":[{"type":"paragraph","data":{"text":"Syarat dan Ketentuan layanan PidieKab."}}]}`)
	input := pageApp.SaveTranslationInput{
		Title:          "Syarat dan Ketentuan",
		Slug:           uniqueKey,
		Body:           bodyJSON,
		SeoTitle:       "Syarat Ketentuan PidieKab",
		SeoDescription: "Ketentuan hukum resmi portal PidieKab",
	}
	trans, err := service.SaveTranslation(ctx, 1, pageID, "id", input)
	if err != nil {
		t.Fatalf("failed to save translation: %v", err)
	}
	if trans.Title != "Syarat dan Ketentuan" || trans.Status != pageDomain.StatusDraft {
		t.Fatalf("unexpected translation data: %+v", trans)
	}

	// 3. Publish translation ID
	err = service.PublishTranslation(ctx, 1, pageID, "id")
	if err != nil {
		t.Fatalf("failed to publish translation: %v", err)
	}

	// Verify published
	detail, err := service.GetPage(ctx, pageID)
	if err != nil {
		t.Fatalf("failed to get page detail: %v", err)
	}
	if detail.Translations["id"].Status != pageDomain.StatusPublished {
		t.Errorf("expected translation status published, got %s", detail.Translations["id"].Status)
	}
}
