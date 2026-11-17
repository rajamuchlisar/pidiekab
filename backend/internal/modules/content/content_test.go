package content_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	auditPostgres "pidiekab/internal/modules/audit/infrastructure/postgres"
	categoryApp "pidiekab/internal/modules/category/application"
	categoryDomain "pidiekab/internal/modules/category/domain"
	categoryPostgres "pidiekab/internal/modules/category/infrastructure/postgres"
	"pidiekab/internal/modules/content/application"
	"pidiekab/internal/modules/content/domain"
	contentPostgres "pidiekab/internal/modules/content/infrastructure/postgres"
	identityDomain "pidiekab/internal/modules/identity/domain"
	identityPostgres "pidiekab/internal/modules/identity/infrastructure/postgres"
	"pidiekab/internal/platform/database"
)

func TestContentLifecycleIntegration(t *testing.T) {
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
		t.Skipf("skipping content integration test: database unreachable: %v", err)
		return
	}
	defer pool.Close()

	// 1. Admin
	adminRepo := identityPostgres.NewAdminRepository(pool)
	testEmail := "content_admin_" + time.Now().Format("20060102150405.000000") + "@pidiekab.id"
	admin := &identityDomain.Admin{
		Email:        testEmail,
		PasswordHash: "any",
		DisplayName:  "Content Lead",
		IsActive:     true,
	}
	if err := adminRepo.Create(ctx, admin); err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}

	// 2. Active Category & Inactive Category
	auditRepo := auditPostgres.NewRepository(pool)
	catRepo := categoryPostgres.NewRepository(pool)
	catSvc := categoryApp.NewService(catRepo, auditRepo)

	ts := time.Now().Format("150405")
	activeCat, err := catSvc.CreateCategory(ctx, true, admin.ID, []*categoryDomain.CategoryTranslation{
		{Locale: "id", Name: "Wisata & Destinasi", Slug: "wisata-destinasi-" + ts},
		{Locale: "en", Name: "Tourism & Destinations", Slug: "tourism-destinations-" + ts},
	}, nil)
	if err != nil {
		t.Fatalf("failed to create active category: %v", err)
	}

	inactiveCat, err := catSvc.CreateCategory(ctx, false, admin.ID, []*categoryDomain.CategoryTranslation{
		{Locale: "id", Name: "Kategori Nonaktif", Slug: "kategori-nonaktif-" + ts},
	}, nil)
	if err != nil {
		t.Fatalf("failed to create inactive category: %v", err)
	}

	// 3. Content Service
	contentRepo := contentPostgres.NewContentRepository(pool)
	transRepo := contentPostgres.NewTranslationRepository(pool)
	sourceRepo := contentPostgres.NewSourceRepository(pool)
	contentSvc := application.NewContentService(contentRepo, transRepo, sourceRepo, catSvc, auditRepo)

	ip := "127.0.0.1"

	// Test A: Creating content under inactive category should fail
	_, err = contentSvc.CreateContent(ctx, inactiveCat.ID, nil, nil, nil, nil, nil, nil, admin.ID, &ip)
	if err != domain.ErrCategoryInactive {
		t.Errorf("expected ErrCategoryInactive, got %v", err)
	}

	// Test B: Invalid coordinates (latitude without longitude)
	badLat := 5.38
	_, err = contentSvc.CreateContent(ctx, activeCat.ID, nil, nil, &badLat, nil, nil, nil, admin.ID, &ip)
	if err != domain.ErrInvalidCoordinates {
		t.Errorf("expected ErrInvalidCoordinates, got %v", err)
	}

	// Test C: Valid coordinates (Mantak Tari beach coordinates)
	validLat := 5.382100
	validLng := 95.961200
	acc := "accurate"
	addr := "Gampong Mantak Tari, Kecamatan Simpang Tiga"
	content, err := contentSvc.CreateContent(ctx, activeCat.ID, nil, &addr, &validLat, &validLng, &acc, nil, admin.ID, &ip)
	if err != nil {
		t.Fatalf("failed to create content: %v", err)
	}
	if content.ID <= 0 {
		t.Errorf("expected positive content ID, got %d", content.ID)
	}

	// Test D: Save ID Translation (Draft)
	bodyID := map[string]any{"type": "doc", "content": "Pantai pasir putih dengan pemandangan Selat Malaka."}
	summaryID := "Destinasi bahari unggulan di Simpang Tiga Pidie."
	slugID := "pantai-mantak-tari-" + ts
	transID, err := contentSvc.SaveTranslation(ctx, content.ID, "id", "Pantai Mantak Tari", slugID, &summaryID, bodyID, nil, nil, admin.ID, &ip)
	if err != nil {
		t.Fatalf("failed to save ID translation: %v", err)
	}
	if transID.Status != domain.StatusDraft {
		t.Errorf("expected draft status, got %s", transID.Status)
	}

	// Test E: Save EN Translation (Draft)
	bodyEN := map[string]any{"type": "doc", "content": "White sand beach overlooking Malacca Strait."}
	slugEN := "mantak-tari-beach-" + ts
	transEN, err := contentSvc.SaveTranslation(ctx, content.ID, "en", "Mantak Tari Beach", slugEN, nil, bodyEN, nil, nil, admin.ID, &ip)
	if err != nil {
		t.Fatalf("failed to save EN translation: %v", err)
	}
	if transEN.Status != domain.StatusDraft {
		t.Errorf("expected draft status, got %s", transEN.Status)
	}

	// Test F: Duplicate slug in same locale rejected
	_, err = contentSvc.SaveTranslation(ctx, content.ID, "id", "Pantai Duplikat", slugID, nil, bodyID, nil, nil, admin.ID, &ip)
	// Upsert on same content updates it, but duplicate on different content should fail
	otherContent, _ := contentSvc.CreateContent(ctx, activeCat.ID, nil, nil, nil, nil, nil, nil, admin.ID, &ip)
	_, err = contentSvc.SaveTranslation(ctx, otherContent.ID, "id", "Pantai Duplikat", slugID, nil, bodyID, nil, nil, admin.ID, &ip)
	if err != domain.ErrSlugAlreadyExists {
		t.Errorf("expected ErrSlugAlreadyExists on different content, got %v", err)
	}

	// Test G: Attempting to publish without sources must fail
	err = contentSvc.PublishLocale(ctx, content.ID, "id", admin.ID, &ip)
	if err != domain.ErrNoSourcesRecorded {
		t.Errorf("expected ErrNoSourcesRecorded when publishing without citations, got %v", err)
	}

	// Test H: Add source citation
	srcURL := "https://disporapar.pidiekab.go.id"
	source, err := contentSvc.AddSource(ctx, content.ID, "official", "Dinas Pariwisata Pidie", &srcURL, nil, nil, nil, admin.ID, &ip)
	if err != nil {
		t.Fatalf("failed to add source citation: %v", err)
	}
	if source.ID <= 0 {
		t.Errorf("expected positive source ID, got %d", source.ID)
	}

	// Test I: Publish ID Translation succeeds
	err = contentSvc.PublishLocale(ctx, content.ID, "id", admin.ID, &ip)
	if err != nil {
		t.Fatalf("failed to publish ID locale: %v", err)
	}

	// Verify ID is published with timestamp and publisher
	publishedID, err := contentSvc.GetTranslation(ctx, content.ID, "id")
	if err != nil {
		t.Fatalf("failed to get published ID translation: %v", err)
	}
	if publishedID.Status != domain.StatusPublished {
		t.Errorf("expected published status, got %s", publishedID.Status)
	}
	if publishedID.PublishedAt == nil || publishedID.PublishedBy == nil {
		t.Error("expected published_at and published_by to be set")
	}

	// Test J: Verify EN translation remains Draft (independent locale statuses)
	publishedEN, err := contentSvc.GetTranslation(ctx, content.ID, "en")
	if err != nil {
		t.Fatalf("failed to get EN translation: %v", err)
	}
	if publishedEN.Status != domain.StatusDraft {
		t.Errorf("expected EN locale to remain draft, got %s", publishedEN.Status)
	}

	// Test K: Archive ID Translation
	err = contentSvc.ArchiveLocale(ctx, content.ID, "id", admin.ID, &ip)
	if err != nil {
		t.Fatalf("failed to archive ID locale: %v", err)
	}
	archivedID, _ := contentSvc.GetTranslation(ctx, content.ID, "id")
	if archivedID.Status != domain.StatusArchived {
		t.Errorf("expected archived status, got %s", archivedID.Status)
	}

	// Test L: Delete Content
	err = contentSvc.DeleteContent(ctx, content.ID, admin.ID, &ip)
	if err != nil {
		t.Fatalf("failed to delete content: %v", err)
	}
	_, err = contentSvc.GetContent(ctx, content.ID)
	if err != domain.ErrContentNotFound {
		t.Errorf("expected ErrContentNotFound after delete, got %v", err)
	}
}

