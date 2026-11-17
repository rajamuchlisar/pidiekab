package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrCategoryNotFound  = errors.New("category not found")
	ErrCategoryInUse     = errors.New("category cannot be deleted because it is referenced by existing content")
	ErrSlugAlreadyExists = errors.New("slug is already in use for this locale")
	ErrInvalidSlug       = errors.New("slug must contain only lowercase letters, numbers, and single hyphens")
	ErrInvalidLocale     = errors.New("locale must be 'id' or 'en'")
	ErrCategoryNameEmpty = errors.New("category name cannot be empty")
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// IsValidSlug checks if a slug consists only of lowercase alphanumeric words separated by hyphens.
func IsValidSlug(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 100 {
		return false
	}
	return slugRegex.MatchString(s)
}

// Category represents a content classification entity.
type Category struct {
	ID        int64     `json:"id"`
	IsActive  bool      `json:"is_active"`
	CreatedBy *int64    `json:"created_by,omitempty"`
	UpdatedBy *int64    `json:"updated_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CategoryTranslation represents a localized name and description for a category.
type CategoryTranslation struct {
	ID          int64     `json:"id"`
	CategoryID  int64     `json:"category_id"`
	Locale      string    `json:"locale"` // 'id' or 'en'
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Slug        string    `json:"slug"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CategoryWithTranslations aggregates the parent category with all its translations.
type CategoryWithTranslations struct {
	Category
	Translations map[string]*CategoryTranslation `json:"translations"`
}
