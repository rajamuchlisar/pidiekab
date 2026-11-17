package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidRole         = errors.New("story role must be 'lead' or 'supporting'")
	ErrInvalidLocale       = errors.New("locale must be 'id' or 'en'")
	ErrContentNotPublished = errors.New("content must be published in the requested locale before it can be curated")
	ErrDuplicateContent    = errors.New("duplicate content ID provided in curation list")
	ErrInvalidPosition     = errors.New("position must be greater than zero")
)

const (
	RoleLead       = "lead"
	RoleSupporting = "supporting"

	LocaleID = "id"
	LocaleEN = "en"
)

type FeaturedItem struct {
	ID        int64     `json:"id"`
	ContentID int64     `json:"content_id"`
	Position  int       `json:"position"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

type StoryItem struct {
	ID        int64     `json:"id"`
	Locale    string    `json:"locale"`
	ContentID int64     `json:"content_id"`
	Role      string    `json:"role"`
	Position  int       `json:"position"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

type HomepageCuration struct {
	Locale            string         `json:"locale"`
	LeadStory         *StoryItem     `json:"lead_story"`
	SupportingStories []StoryItem    `json:"supporting_stories"`
	Featured          []FeaturedItem `json:"featured"`
	Latest            []CuratedItem  `json:"latest"`
	Random            []CuratedItem  `json:"random"`
}

type CuratedItem struct {
	ContentID    int64     `json:"content_id"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	CategorySlug string    `json:"category_slug"`
	CategoryName string    `json:"category_name"`
	Summary      string    `json:"summary"`
	PublishedAt  time.Time `json:"published_at"`
}

func ValidateLocale(locale string) error {
	l := strings.ToLower(strings.TrimSpace(locale))
	if l != LocaleID && l != LocaleEN {
		return ErrInvalidLocale
	}
	return nil
}
