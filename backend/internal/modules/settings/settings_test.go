package settings_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"pidiekab/internal/modules/identity/domain"
	settingsApp "pidiekab/internal/modules/settings/application"
	settingsDomain "pidiekab/internal/modules/settings/domain"
	settingsPG "pidiekab/internal/modules/settings/infrastructure/postgres"
	"pidiekab/internal/platform/database"
)

func TestSettingsValidation(t *testing.T) {
	if err := settingsDomain.ValidateLocale("id"); err != nil {
		t.Errorf("expected id valid, got: %v", err)
	}
	if err := settingsDomain.ValidateLocale("fr"); err == nil {
		t.Error("expected error for non id/en locale")
	}
	if err := settingsDomain.ValidateSiteName("   "); err == nil {
		t.Error("expected error for empty site name")
	}

	// OPEN-021 Contact and Social Link Policy
	if err := settingsDomain.ValidateContactEmail("invalid-email"); err == nil {
		t.Error("expected error for invalid contact email")
	}
	if err := settingsDomain.ValidateContactEmail("info@pidiekab.web.id"); err != nil {
		t.Errorf("expected valid email, got: %v", err)
	}

	if err := settingsDomain.ValidateContactPhone("+62 812-3456-7890"); err != nil {
		t.Errorf("expected valid phone, got: %v", err)
	}
	if err := settingsDomain.ValidateContactPhone("phone-with-letters-abc"); err == nil {
		t.Error("expected error for phone with letters")
	}

	validSocial := json.RawMessage(`{"instagram": "@pesonapidie", "website": "https://pidiekab.web.id"}`)
	if err := settingsDomain.ValidateSocialLinks(validSocial); err != nil {
		t.Errorf("expected valid social links, got: %v", err)
	}

	insecureSocial := json.RawMessage(`{"site": "http://insecure-site.com"}`)
	if err := settingsDomain.ValidateSocialLinks(insecureSocial); err == nil {
		t.Error("expected error for insecure http:// scheme")
	}

	maliciousSocial := json.RawMessage(`{"site": "javascript:alert(1)"}`)
	if err := settingsDomain.ValidateSocialLinks(maliciousSocial); err == nil {
		t.Error("expected error for javascript: scheme")
	}
}

func TestSettingsLifecycle(t *testing.T) {
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
		t.Skipf("skipping settings postgres test: %v", err)
	}
	defer pool.Close()

	repo := settingsPG.NewSettingsRepository(pool)
	service := settingsApp.NewSettingsService(repo, nil)

	// 1. Get initial settings (singleton ensure)
	settings, err := service.GetSettings(ctx)
	if err != nil {
		t.Fatalf("failed to get settings: %v", err)
	}
	if settings.ID != 1 {
		t.Fatalf("expected singleton id 1, got %d", settings.ID)
	}

	// 2. Update general settings
	adminID := int64(1)
	updateGen := settingsApp.UpdateGeneralInput{
		ContactEmail:   "kontak@pidiekab.id",
		ContactPhone:   "+62 653 21234",
		ContactAddress: "Kota Sigli, Kabupaten Pidie, Aceh",
		SocialLinks:    json.RawMessage(`{"instagram":"@pesonapidie"}`),
	}
	res, err := service.UpdateGeneral(ctx, adminID, updateGen)
	if err != nil {
		t.Fatalf("failed to update general settings: %v", err)
	}
	if res.ContactEmail != "kontak@pidiekab.id" {
		t.Errorf("expected email updated, got %s", res.ContactEmail)
	}

	// 3. Upsert Indonesian translation
	updateTransID := settingsApp.UpdateTranslationInput{
		SiteName:              "PidieKab",
		Tagline:               "Pesona Alam dan Budaya Pidie",
		Description:           "Platform independen promosi potensi daerah Kabupaten Pidie.",
		DefaultSeoTitle:       "Wisata Pidie — Keindahan Alam & Budaya",
		DefaultSeoDescription: "Temukan destinasi wisata terbaik di Pidie.",
	}
	transID, err := service.UpdateTranslation(ctx, adminID, "id", updateTransID)
	if err != nil {
		t.Fatalf("failed to update translation ID: %v", err)
	}
	if transID.SiteName != "PidieKab" {
		t.Errorf("expected site name updated, got %s", transID.SiteName)
	}

	// 4. Verify in full settings object
	refreshed, err := service.GetSettings(ctx)
	if err != nil {
		t.Fatalf("failed to get refreshed settings: %v", err)
	}
	if refreshed.Translations["id"] == nil || refreshed.Translations["id"].Tagline != "Pesona Alam dan Budaya Pidie" {
		t.Errorf("expected translation ID to be persisted in settings")
	}
}

// Suppress unused domain import
var _ = domain.Admin{}
