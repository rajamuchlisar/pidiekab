package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrContentNotFound          = errors.New("content record not found")
	ErrTranslationNotFound      = errors.New("content translation not found for locale")
	ErrSourceNotFound           = errors.New("content source not found")
	ErrNoSourcesRecorded        = errors.New("cannot publish content: at least one verifiable source is required")
	ErrCategoryInactive         = errors.New("cannot publish content: assigned category is inactive")
	ErrInvalidCoordinates       = errors.New("coordinates must have both latitude and longitude or neither")
	ErrLatitudeOutOfRange       = errors.New("latitude must be between -90 and 90")
	ErrLongitudeOutOfRange      = errors.New("longitude must be between -180 and 180")
	ErrInvalidAccuracy          = errors.New("location accuracy must be either 'accurate' or 'approximate'")
	ErrAccuracyRequiresCoords   = errors.New("location accuracy requires valid coordinates")
	ErrInvalidSlug              = errors.New("slug must contain only lowercase alphanumeric characters and hyphens")
	ErrSlugAlreadyExists        = errors.New("slug is already in use for this locale")
	ErrTitleEmpty               = errors.New("content title cannot be empty")
	ErrBodyEmpty                = errors.New("content body cannot be empty")
	ErrInvalidSourceType        = errors.New("source type must be one of: research, visit, official, social")
	ErrSourceNameEmpty          = errors.New("source name cannot be empty")
	ErrInvalidStatus            = errors.New("status must be draft, scheduled, published, or archived")
	ErrCannotArchiveUnpublished = errors.New("cannot archive content translation that has never been published")
	ErrScheduledAtRequired      = errors.New("scheduled_at must be provided when status is scheduled")
	ErrScheduledAtMustBeFuture  = errors.New("scheduled time must be in the future")
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// IsValidSlug validates that a slug complies with URL safety.
func IsValidSlug(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 120 {
		return false
	}
	return slugRegex.MatchString(s)
}

// ValidateCoordinates ensures coordinates obey database check constraints.
func ValidateCoordinates(lat, lng *float64, accuracy *string) error {
	if (lat == nil) != (lng == nil) {
		return ErrInvalidCoordinates
	}

	if lat != nil {
		if *lat < -90 || *lat > 90 {
			return ErrLatitudeOutOfRange
		}
		if *lng < -180 || *lng > 180 {
			return ErrLongitudeOutOfRange
		}
	}

	if accuracy != nil && *accuracy != "" {
		acc := strings.ToLower(strings.TrimSpace(*accuracy))
		if acc != "accurate" && acc != "approximate" {
			return ErrInvalidAccuracy
		}
		if lat == nil {
			return ErrAccuracyRequiresCoords
		}
	}

	return nil
}

// Content represents the shared language-neutral identity of a discovery destination or story.
type Content struct {
	ID               int64     `json:"id"`
	CategoryID       int64     `json:"category_id"`
	VillageID        *int64    `json:"village_id,omitempty"`
	Address          *string   `json:"address,omitempty"`
	Latitude         *float64  `json:"latitude,omitempty"`
	Longitude        *float64  `json:"longitude,omitempty"`
	LocationAccuracy *string   `json:"location_accuracy,omitempty"`
	MainMediaID      *int64    `json:"main_media_id,omitempty"`
	CreatedBy        *int64    `json:"created_by,omitempty"`
	UpdatedBy        *int64    `json:"updated_by,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TranslationStatus represents the publication lifecycle of a locale translation.
type TranslationStatus string

const (
	LocaleID = "id"
	LocaleEN = "en"

	StatusDraft     TranslationStatus = "draft"
	StatusScheduled TranslationStatus = "scheduled"
	StatusPublished TranslationStatus = "published"
	StatusArchived  TranslationStatus = "archived"
)

// ContentTranslation represents a localized version of a content identity.
type ContentTranslation struct {
	ID             int64             `json:"id"`
	ContentID      int64             `json:"content_id"`
	Locale         string            `json:"locale"` // 'id' or 'en'
	Title          string            `json:"title"`
	Slug           string            `json:"slug"`
	Summary        *string           `json:"summary,omitempty"`
	Body           map[string]any    `json:"body"`
	ExtraData      map[string]any    `json:"extra_data,omitempty"`
	SEOTitle       *string           `json:"seo_title,omitempty"`
	SEODescription *string           `json:"seo_description,omitempty"`
	Status         TranslationStatus `json:"status"`
	PublishedAt    *time.Time        `json:"published_at,omitempty"`
	ScheduledAt    *time.Time        `json:"scheduled_at,omitempty"`
	CreatedBy      *int64            `json:"created_by,omitempty"`
	UpdatedBy      *int64            `json:"updated_by,omitempty"`
	PublishedBy    *int64            `json:"published_by,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// ContentSource represents a verifiable reference or attribution for content.
type ContentSource struct {
	ID                int64      `json:"id"`
	ContentID         int64      `json:"content_id"`
	SourceType        string     `json:"source_type"` // 'research', 'visit', 'official', 'social'
	SourceName        string     `json:"source_name"`
	SourceURL         *string    `json:"source_url,omitempty"`
	SourcePublishedAt *time.Time `json:"source_published_at,omitempty"`
	AccessedAt        *time.Time `json:"accessed_at,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
	CreatedBy         *int64     `json:"created_by,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// ContentDetail aggregates shared fields, translations, and sources.
type ContentDetail struct {
	Content
	Translations map[string]*ContentTranslation `json:"translations"`
	Sources      []*ContentSource               `json:"sources"`
}

// ContentListItem represents a summarized item in the Admin content list.
type ContentListItem struct {
	ID           int64                                     `json:"id"`
	CategoryID   int64                                     `json:"category_id"`
	Address      *string                                   `json:"address,omitempty"`
	Translations map[string]*ContentTranslationListSummary `json:"translations"`
	CreatedAt    time.Time                                 `json:"created_at"`
	UpdatedAt    time.Time                                 `json:"updated_at"`
}

// ContentTranslationListSummary summarizes translation title and status for list views.
type ContentTranslationListSummary struct {
	Title       *string           `json:"title"`
	Status      TranslationStatus `json:"status"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
}

// ValidatePublishEligibility verifies all strict business invariants before publishing a locale.
func ValidatePublishEligibility(c *Content, t *ContentTranslation, sources []*ContentSource, categoryActive bool) error {
	if !categoryActive {
		return ErrCategoryInactive
	}
	if strings.TrimSpace(t.Title) == "" {
		return ErrTitleEmpty
	}
	if !IsValidSlug(t.Slug) {
		return ErrInvalidSlug
	}
	if len(t.Body) == 0 {
		return ErrBodyEmpty
	}
	if len(sources) == 0 {
		return ErrNoSourcesRecorded
	}
	return nil
}
