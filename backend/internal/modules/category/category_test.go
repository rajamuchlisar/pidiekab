package category_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	auditPostgres "pidiekab/internal/modules/audit/infrastructure/postgres"
	"pidiekab/internal/modules/category/application"
	"pidiekab/internal/modules/category/domain"
	categoryPostgres "pidiekab/internal/modules/category/infrastructure/postgres"
	identityDomain "pidiekab/internal/modules/identity/domain"
	identityPostgres "pidiekab/internal/modules/identity/infrastructure/postgres"
	"pidiekab/internal/platform/database"
)

func TestCategoryRepositoryAndServiceIntegration(t *testing.T) {
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
		t.Skipf("skipping category integration test: database unreachable: %v", err)
		return
	}
	defer pool.Close()

	// Ensure we have a valid admin
	adminRepo := identityPostgres.NewAdminRepository(pool)
	testEmail := "cat_admin_" + time.Now().Format("20060102150405.000000") + "@pidiekab.id"
	admin := &identityDomain.Admin{
		Email:        testEmail,
		PasswordHash: "any",
		DisplayName:  "Category Admin",
		IsActive:     true,
	}
	if err := adminRepo.Create(ctx, admin); err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}

	repo := categoryPostgres.NewRepository(pool)
	auditRepo := auditPostgres.NewRepository(pool)
	svc := application.NewService(repo, auditRepo)

	timestamp := time.Now().Format("150405")
	slugID := "wisata-alam-" + timestamp
	slugEN := "natural-tourism-" + timestamp

	ip := "127.0.0.1"

	// 1. Create Category with ID and EN translations
	descID := "Deskripsi kategori wisata alam"
	descEN := "Description of natural tourism category"
	translations := []*domain.CategoryTranslation{
		{
			Locale:      "id",
			Name:        "Wisata Alam",
			Description: &descID,
			Slug:        slugID,
		},
		{
			Locale:      "en",
			Name:        "Natural Tourism",
			Description: &descEN,
			Slug:        slugEN,
		},
	}

	created, err := svc.CreateCategory(ctx, true, admin.ID, translations, &ip)
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	if created.ID <= 0 {
		t.Errorf("expected positive category ID, got %d", created.ID)
	}
	if len(created.Translations) != 2 {
		t.Errorf("expected 2 translations, got %d", len(created.Translations))
	}
	if created.Translations["id"].Slug != slugID {
		t.Errorf("expected ID slug %q, got %q", slugID, created.Translations["id"].Slug)
	}

	// 2. Creating another category with same slug in same locale should fail
	dupTrans := []*domain.CategoryTranslation{
		{
			Locale: "id",
			Name:   "Wisata Lain",
			Slug:   slugID,
		},
	}
	_, err = svc.CreateCategory(ctx, true, admin.ID, dupTrans, &ip)
	if err != domain.ErrSlugAlreadyExists {
		t.Errorf("expected ErrSlugAlreadyExists for duplicate slug, got %v", err)
	}

	// 3. Update category status
	if err := svc.UpdateCategory(ctx, created.ID, false, admin.ID, &ip); err != nil {
		t.Fatalf("failed to update category status: %v", err)
	}
	fetched, err := svc.GetCategory(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get category: %v", err)
	}
	if fetched.IsActive != false {
		t.Errorf("expected category is_active to be false")
	}

	// 4. Update translation
	updatedDescID := "Deskripsi diperbarui"
	newSlugID := slugID + "-v2"
	updatedTrans, err := svc.SaveTranslation(ctx, created.ID, "id", "Wisata Alam Baru", &updatedDescID, newSlugID, admin.ID, &ip)
	if err != nil {
		t.Fatalf("failed to save translation: %v", err)
	}
	if updatedTrans.Slug != newSlugID {
		t.Errorf("expected updated slug %q, got %q", newSlugID, updatedTrans.Slug)
	}

	// 5. Delete category
	if err := svc.DeleteCategory(ctx, created.ID, admin.ID, &ip); err != nil {
		t.Fatalf("failed to delete category: %v", err)
	}

	// Verify not found after delete
	_, err = svc.GetCategory(ctx, created.ID)
	if err != domain.ErrCategoryNotFound {
		t.Errorf("expected ErrCategoryNotFound after delete, got %v", err)
	}
}
