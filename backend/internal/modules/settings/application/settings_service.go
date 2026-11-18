package application

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	auditDomain "pidiekab/internal/modules/audit/domain"
	"pidiekab/internal/modules/settings/domain"
	"pidiekab/internal/modules/settings/infrastructure/postgres"
)

type AuditRecorder interface {
	Record(ctx context.Context, entry auditDomain.Log) error
}

type SettingsService struct {
	repo  *postgres.SettingsRepository
	audit AuditRecorder
}

func NewSettingsService(repo *postgres.SettingsRepository, audit AuditRecorder) *SettingsService {
	return &SettingsService{repo: repo, audit: audit}
}

func (s *SettingsService) GetSettings(ctx context.Context) (*domain.SiteSettings, error) {
	return s.repo.GetSettings(ctx)
}

type UpdateGeneralInput struct {
	LogoMediaID    *int64          `json:"logo_media_id"`
	FaviconMediaID *int64          `json:"favicon_media_id"`
	ContactEmail   string          `json:"contact_email"`
	ContactPhone   string          `json:"contact_phone"`
	ContactAddress string          `json:"contact_address"`
	SocialLinks    json.RawMessage `json:"social_links"`
}

func (s *SettingsService) UpdateGeneral(ctx context.Context, adminID int64, input UpdateGeneralInput) (*domain.SiteSettings, error) {
	if err := domain.ValidateContactEmail(input.ContactEmail); err != nil {
		return nil, err
	}
	if err := domain.ValidateContactPhone(input.ContactPhone); err != nil {
		return nil, err
	}
	if err := domain.ValidateSocialLinks(input.SocialLinks); err != nil {
		return nil, err
	}

	social := input.SocialLinks
	if len(social) == 0 {
		social = json.RawMessage(`{}`)
	}

	params := postgres.UpdateGeneralParams{
		LogoMediaID:    input.LogoMediaID,
		FaviconMediaID: input.FaviconMediaID,
		ContactEmail:   strings.TrimSpace(input.ContactEmail),
		ContactPhone:   strings.TrimSpace(input.ContactPhone),
		ContactAddress: strings.TrimSpace(input.ContactAddress),
		SocialLinks:    social,
		AdminID:        adminID,
	}

	res, err := s.repo.UpdateGeneral(ctx, params)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, &adminID, "site_settings.update_general", "site_settings", nil, map[string]any{
		"contact_email": params.ContactEmail,
	}, nil)

	return res, nil
}

type UpdateTranslationInput struct {
	SiteName              string `json:"site_name"`
	Tagline               string `json:"tagline"`
	Description           string `json:"description"`
	DefaultSeoTitle       string `json:"default_seo_title"`
	DefaultSeoDescription string `json:"default_seo_description"`
}

func (s *SettingsService) UpdateTranslation(ctx context.Context, adminID int64, locale string, input UpdateTranslationInput) (*domain.SiteSettingTranslation, error) {
	if err := domain.ValidateLocale(locale); err != nil {
		return nil, err
	}
	if err := domain.ValidateSiteName(input.SiteName); err != nil {
		return nil, err
	}

	params := postgres.UpsertTranslationParams{
		Locale:                locale,
		SiteName:              strings.TrimSpace(input.SiteName),
		Tagline:               strings.TrimSpace(input.Tagline),
		Description:           strings.TrimSpace(input.Description),
		DefaultSeoTitle:       strings.TrimSpace(input.DefaultSeoTitle),
		DefaultSeoDescription: strings.TrimSpace(input.DefaultSeoDescription),
	}

	t, err := s.repo.UpsertTranslation(ctx, params)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, &adminID, "site_settings.update_translation", "site_setting_translation", &t.ID, map[string]any{
		"locale":    locale,
		"site_name": t.SiteName,
	}, nil)

	return t, nil
}

func (s *SettingsService) recordAudit(ctx context.Context, adminID *int64, action, entityType string, entityID *int64, meta map[string]any, ip *string) {
	if s.audit == nil {
		return
	}
	_ = s.audit.Record(ctx, auditDomain.Log{
		AdminID:    adminID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   meta,
		IPAddress:  ip,
		CreatedAt:  time.Now().UTC(),
	})
}
