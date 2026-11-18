package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidLocale       = errors.New("invalid locale; must be 'id' or 'en'")
	ErrSiteNameEmpty       = errors.New("site name cannot be empty")
	ErrSettingsNotFound    = errors.New("site settings not found")
	ErrInvalidContactEmail = errors.New("invalid contact email format")
	ErrInvalidContactPhone = errors.New("invalid contact phone format")
	ErrInvalidSocialURL    = errors.New("social link URL must start with https://")
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

// ValidateContactEmail ensures email conforms to standard syntax when provided.
func ValidateContactEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") || strings.Contains(email, " ") {
		return ErrInvalidContactEmail
	}
	return nil
}

// ValidateContactPhone ensures telephone numbers contain only valid telecom characters.
func ValidateContactPhone(phone string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil
	}
	for _, ch := range phone {
		if !((ch >= '0' && ch <= '9') || ch == '+' || ch == '-' || ch == ' ' || ch == '(' || ch == ')') {
			return ErrInvalidContactPhone
		}
	}
	return nil
}

// ValidateSocialLinks ensures external social/contact links do not use insecure or dangerous schemes (OPEN-021).
func ValidateSocialLinks(raw json.RawMessage) error {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "{}" {
		return nil
	}
	var links map[string]string
	if err := json.Unmarshal(raw, &links); err != nil {
		return errors.New("social_links must be a valid JSON object")
	}
	for _, u := range links {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		lower := strings.ToLower(u)
		if strings.Contains(lower, "://") {
			if !strings.HasPrefix(lower, "https://") {
				return ErrInvalidSocialURL
			}
		} else if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "data:") {
			return ErrInvalidSocialURL
		}
	}
	return nil
}
