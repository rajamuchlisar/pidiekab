package domain

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrPageNotFound        = errors.New("static page not found")
	ErrPageKeyEmpty        = errors.New("static page key cannot be empty")
	ErrPageKeyInvalid      = errors.New("static page key must only contain lowercase alphanumeric characters and hyphens")
	ErrTranslationNotFound = errors.New("static page translation not found")
	ErrInvalidStatus       = errors.New("invalid static page status; must be draft, published, or archived")
	ErrInvalidLocale       = errors.New("invalid locale; must be 'id' or 'en'")
	ErrInvalidSlug         = errors.New("invalid slug format; must be lowercase alphanumeric with hyphens")
	ErrInvalidBody         = errors.New("body must be a valid JSON object")
	ErrTitleEmpty          = errors.New("title cannot be empty")
	ErrCannotPublishNoBody = errors.New("cannot publish page with empty body or title")
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"

	LocaleID = "id"
	LocaleEN = "en"
)

var (
	pageKeyRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	slugRegex    = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

type StaticPage struct {
	ID        int64     `json:"id"`
	PageKey   string    `json:"page_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StaticPageTranslation struct {
	ID             int64           `json:"id"`
	StaticPageID   int64           `json:"static_page_id"`
	Locale         string          `json:"locale"`
	Title          string          `json:"title"`
	Slug           string          `json:"slug"`
	Body           json.RawMessage `json:"body"`
	BodyHTML       string          `json:"body_html,omitempty"`
	SeoTitle       string          `json:"seo_title"`
	SeoDescription string          `json:"seo_description"`
	Status         string          `json:"status"`
	PublishedAt    *time.Time      `json:"published_at,omitempty"`
	CreatedBy      *int64          `json:"created_by,omitempty"`
	UpdatedBy      *int64          `json:"updated_by,omitempty"`
	PublishedBy    *int64          `json:"published_by,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type StaticPageDetail struct {
	StaticPage
	Translations map[string]*StaticPageTranslation `json:"translations"`
}

func ValidatePageKey(key string) error {
	k := strings.TrimSpace(key)
	if k == "" {
		return ErrPageKeyEmpty
	}
	if !pageKeyRegex.MatchString(k) {
		return ErrPageKeyInvalid
	}
	return nil
}

func ValidateSlug(slug string) error {
	s := strings.TrimSpace(slug)
	if s == "" || !slugRegex.MatchString(s) {
		return ErrInvalidSlug
	}
	return nil
}

func ValidateLocale(locale string) error {
	l := strings.ToLower(strings.TrimSpace(locale))
	if l != LocaleID && l != LocaleEN {
		return ErrInvalidLocale
	}
	return nil
}

func ValidateBodyJSON(body json.RawMessage) error {
	if len(body) == 0 {
		return ErrInvalidBody
	}
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err != nil {
		return ErrInvalidBody
	}
	return nil
}
