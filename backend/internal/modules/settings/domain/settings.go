package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidLocale    = errors.New("invalid locale; must be 'id' or 'en'")
	ErrSiteNameEmpty    = errors.New("site name cannot be empty")
	ErrSettingsNotFound = errors.New("site settings not found")
)

const (
	LocaleID = "id"
	LocaleEN = "en"
)

type SiteSettings struct {
	ID             int16                              `json:"id"`
	LogoMediaID    *int64                             `json:"logo_media_id,omitempty"`
	FaviconMediaID *int64                             `json:"favicon_media_id,omitempty"`
	ContactEmail   string                             `json:"contact_email"`
	ContactPhone   string                             `json:"contact_phone"`
	ContactAddress string                             `json:"contact_address"`
	SocialLinks    json.RawMessage                    `json:"social_links"`
	UpdatedBy      *int64                             `json:"updated_by,omitempty"`
	UpdatedAt      time.Time                          `json:"updated_at"`
	Translations   map[string]*SiteSettingTranslation `json:"translations"`
}

type SiteSettingTranslation struct {
	ID                    int64     `json:"id"`
	SiteSettingsID        int16     `json:"site_settings_id"`
	Locale                string    `json:"locale"`
	SiteName              string    `json:"site_name"`
	Tagline               string    `json:"tagline"`
	Description           string    `json:"description"`
	DefaultSeoTitle       string    `json:"default_seo_title"`
	DefaultSeoDescription string    `json:"default_seo_description"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func ValidateLocale(locale string) error {
	l := strings.ToLower(strings.TrimSpace(locale))
	if l != LocaleID && l != LocaleEN {
		return ErrInvalidLocale
	}
	return nil
}

func ValidateSiteName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrSiteNameEmpty
	}
	return nil
}
