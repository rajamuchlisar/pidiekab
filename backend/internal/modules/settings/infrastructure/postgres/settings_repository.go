package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"pidiekab/internal/modules/settings/domain"
)

type SettingsRepository struct {
	pool *pgxpool.Pool
}

func NewSettingsRepository(pool *pgxpool.Pool) *SettingsRepository {
	return &SettingsRepository{pool: pool}
}

func (r *SettingsRepository) EnsureSingleton(ctx context.Context) error {
	query := `
		INSERT INTO site_settings (id, updated_at)
		VALUES (1, NOW())
		ON CONFLICT (id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query)
	return err
}

func (r *SettingsRepository) GetSettings(ctx context.Context) (*domain.SiteSettings, error) {
	if err := r.EnsureSingleton(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure singleton settings: %w", err)
	}

	query := `
		SELECT
			id, logo_media_id, favicon_media_id,
			COALESCE(contact_email, ''), COALESCE(contact_phone, ''), COALESCE(contact_address, ''),
			social_links, updated_by, updated_at
		FROM site_settings
		WHERE id = 1
	`
	var s domain.SiteSettings
	var socialBytes []byte
	err := r.pool.QueryRow(ctx, query).Scan(
		&s.ID, &s.LogoMediaID, &s.FaviconMediaID,
		&s.ContactEmail, &s.ContactPhone, &s.ContactAddress,
		&socialBytes, &s.UpdatedBy, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get site settings: %w", err)
	}
	if len(socialBytes) > 0 {
		s.SocialLinks = json.RawMessage(socialBytes)
	} else {
		s.SocialLinks = json.RawMessage(`{}`)
	}

	s.Translations = make(map[string]*domain.SiteSettingTranslation)
	transQuery := `
		SELECT
			id, site_settings_id, locale, site_name,
			COALESCE(tagline, ''), COALESCE(description, ''),
			COALESCE(default_seo_title, ''), COALESCE(default_seo_description, ''),
			created_at, updated_at
		FROM site_setting_translations
		WHERE site_settings_id = 1
	`
	rows, err := r.pool.Query(ctx, transQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query setting translations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t domain.SiteSettingTranslation
		if err := rows.Scan(
			&t.ID, &t.SiteSettingsID, &t.Locale, &t.SiteName,
			&t.Tagline, &t.Description,
			&t.DefaultSeoTitle, &t.DefaultSeoDescription,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan setting translation: %w", err)
		}
		s.Translations[t.Locale] = &t
	}

	return &s, nil
}

type UpdateGeneralParams struct {
	LogoMediaID    *int64
	FaviconMediaID *int64
	ContactEmail   string
	ContactPhone   string
	ContactAddress string
	SocialLinks    json.RawMessage
	AdminID        int64
}

func (r *SettingsRepository) UpdateGeneral(ctx context.Context, p UpdateGeneralParams) (*domain.SiteSettings, error) {
	if err := r.EnsureSingleton(ctx); err != nil {
		return nil, err
	}

	query := `
		UPDATE site_settings
		SET
			logo_media_id = $1,
			favicon_media_id = $2,
			contact_email = $3,
			contact_phone = $4,
			contact_address = $5,
			social_links = $6,
			updated_by = $7,
			updated_at = NOW()
		WHERE id = 1
	`
	_, err := r.pool.Exec(ctx, query,
		p.LogoMediaID, p.FaviconMediaID,
		p.ContactEmail, p.ContactPhone, p.ContactAddress,
		[]byte(p.SocialLinks), p.AdminID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update site settings: %w", err)
	}

	return r.GetSettings(ctx)
}

type UpsertTranslationParams struct {
	Locale                string
	SiteName              string
	Tagline               string
	Description           string
	DefaultSeoTitle       string
	DefaultSeoDescription string
}

func (r *SettingsRepository) UpsertTranslation(ctx context.Context, p UpsertTranslationParams) (*domain.SiteSettingTranslation, error) {
	if err := r.EnsureSingleton(ctx); err != nil {
		return nil, err
	}

	query := `
		INSERT INTO site_setting_translations (
			site_settings_id, locale, site_name, tagline, description,
			default_seo_title, default_seo_description, created_at, updated_at
		) VALUES (
			1, $1, $2, $3, $4, $5, $6, NOW(), NOW()
		)
		ON CONFLICT (site_settings_id, locale) DO UPDATE
		SET
			site_name = EXCLUDED.site_name,
			tagline = EXCLUDED.tagline,
			description = EXCLUDED.description,
			default_seo_title = EXCLUDED.default_seo_title,
			default_seo_description = EXCLUDED.default_seo_description,
			updated_at = NOW()
		RETURNING
			id, site_settings_id, locale, site_name,
			COALESCE(tagline, ''), COALESCE(description, ''),
			COALESCE(default_seo_title, ''), COALESCE(default_seo_description, ''),
			created_at, updated_at
	`
	var t domain.SiteSettingTranslation
	err := r.pool.QueryRow(ctx, query,
		p.Locale, p.SiteName, p.Tagline, p.Description,
		p.DefaultSeoTitle, p.DefaultSeoDescription,
	).Scan(
		&t.ID, &t.SiteSettingsID, &t.Locale, &t.SiteName,
		&t.Tagline, &t.Description,
		&t.DefaultSeoTitle, &t.DefaultSeoDescription,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert site setting translation: %w", err)
	}

	return &t, nil
}
