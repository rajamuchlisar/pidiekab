package application

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	auditDomain "pidiekab/internal/modules/audit/domain"
	"pidiekab/internal/modules/settings/domain"
	"pidiekab/internal/modules/settings/infrastructure/postgres"
)

type AuditRecorder interface {
	Record(ctx context.Context, entry auditDomain.Log) error
}

type JobEnqueuer interface {
	Enqueue(ctx context.Context, tx pgx.Tx, jobType string, payload any) (int64, error)
}

type SettingsService struct {
	repo     *postgres.SettingsRepository
	audit    AuditRecorder
	enqueuer JobEnqueuer
}

func NewSettingsService(repo *postgres.SettingsRepository, audit AuditRecorder) *SettingsService {
	return &SettingsService{repo: repo, audit: audit}
}

func (s *SettingsService) SetJobEnqueuer(enqueuer JobEnqueuer) {
	s.enqueuer = enqueuer
}

func (s *SettingsService) GetSettings(ctx context.Context) (*domain.SiteSettings, error) {
	return s.repo.GetSettings(ctx)
}

type UpdateGeneralInput struct {
	LogoMediaID    *int64          `json:"logo_media_id"`
	FaviconMediaID *int64          `json:"favicon_media_id"`
	LogoURL        string          `json:"logo_url"`
	FaviconURL     string          `json:"favicon_url"`
	ContactEmail   string          `json:"contact_email"`
	ContactPhone   string          `json:"contact_phone"`
	ContactAddress string          `json:"contact_address"`
	Address        string          `json:"address"`
	SocialLinks    json.RawMessage `json:"social_links"`

	// Flat social fields
	SocialInstagram string `json:"social_instagram"`
	SocialFacebook  string `json:"social_facebook"`
	SocialYoutube   string `json:"social_youtube"`
	SocialTwitter   string `json:"social_twitter"`

	// Translation fields
	SiteName          string `json:"site_name"`
	TaglineID         string `json:"tagline_id"`
	TaglineEn         string `json:"tagline_en"`
	DefaultSeoTitleID string `json:"default_seo_title_id"`
	DefaultSeoTitleEn string `json:"default_seo_title_en"`
	DefaultSeoDescID  string `json:"default_seo_desc_id"`
	DefaultSeoDescEn  string `json:"default_seo_desc_en"`
}

func (s *SettingsService) UpdateGeneral(ctx context.Context, adminID int64, input UpdateGeneralInput) (*domain.SiteSettings, error) {
	contactAddr := strings.TrimSpace(input.ContactAddress)
	if contactAddr == "" {
		contactAddr = strings.TrimSpace(input.Address)
	}

	social := input.SocialLinks
	if len(social) == 0 || string(social) == "null" || string(social) == "{}" {
		socialMap := make(map[string]string)
		if strings.TrimSpace(input.SocialInstagram) != "" {
			socialMap["instagram"] = strings.TrimSpace(input.SocialInstagram)
		}
		if strings.TrimSpace(input.SocialFacebook) != "" {
			socialMap["facebook"] = strings.TrimSpace(input.SocialFacebook)
		}
		if strings.TrimSpace(input.SocialYoutube) != "" {
			socialMap["youtube"] = strings.TrimSpace(input.SocialYoutube)
		}
		if strings.TrimSpace(input.SocialTwitter) != "" {
			socialMap["twitter"] = strings.TrimSpace(input.SocialTwitter)
		}
		b, _ := json.Marshal(socialMap)
		social = json.RawMessage(b)
	}
	social = domain.NormalizeSocialLinks(social)

	if err := domain.ValidateContactEmail(input.ContactEmail); err != nil {
		return nil, err
	}
	if err := domain.ValidateContactPhone(input.ContactPhone); err != nil {
		return nil, err
	}
	if err := domain.ValidateSocialLinks(social); err != nil {
		return nil, err
	}

	params := postgres.UpdateGeneralParams{
		LogoMediaID:    input.LogoMediaID,
		FaviconMediaID: input.FaviconMediaID,
		LogoURL:        strings.TrimSpace(input.LogoURL),
		FaviconURL:     strings.TrimSpace(input.FaviconURL),
		ContactEmail:   strings.TrimSpace(input.ContactEmail),
		ContactPhone:   strings.TrimSpace(input.ContactPhone),
		ContactAddress: contactAddr,
		SocialLinks:    social,
		AdminID:        adminID,
	}

	res, err := s.repo.UpdateGeneral(ctx, params)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(input.SiteName) != "" {
		_, _ = s.repo.UpsertTranslation(ctx, postgres.UpsertTranslationParams{
			Locale:                domain.LocaleID,
			SiteName:              strings.TrimSpace(input.SiteName),
			Tagline:               strings.TrimSpace(input.TaglineID),
			DefaultSeoTitle:       strings.TrimSpace(input.DefaultSeoTitleID),
			DefaultSeoDescription: strings.TrimSpace(input.DefaultSeoDescID),
		})
		_, _ = s.repo.UpsertTranslation(ctx, postgres.UpsertTranslationParams{
			Locale:                domain.LocaleEN,
			SiteName:              strings.TrimSpace(input.SiteName),
			Tagline:               strings.TrimSpace(input.TaglineEn),
			DefaultSeoTitle:       strings.TrimSpace(input.DefaultSeoTitleEn),
			DefaultSeoDescription: strings.TrimSpace(input.DefaultSeoDescEn),
		})
		// Re-fetch latest populated settings
		if latest, err := s.repo.GetSettings(ctx); err == nil {
			res = latest
		}
	}

	if s.enqueuer != nil {
		_, _ = s.enqueuer.Enqueue(ctx, nil, "site.rebuild", map[string]any{"trigger": "settings_updated"})
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

	if s.enqueuer != nil {
		_, _ = s.enqueuer.Enqueue(ctx, nil, "site.rebuild", map[string]any{"trigger": "settings_translation_updated", "locale": locale})
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
