package domain

import (
	"errors"
	"strings"
)

var (
	ErrQueryTooShort = errors.New("search query must be at least 2 characters")
	ErrQueryTooLong  = errors.New("search query must not exceed 120 characters")
	ErrInvalidLocale = errors.New("locale must be 'id' or 'en'")
)

const (
	LocaleID = "id"
	LocaleEN = "en"
)

type SearchResultItem struct {
	ID       string           `json:"id"`
	Title    string           `json:"title"`
	Category CategorySummary  `json:"category"`
	Location *LocationSummary `json:"location"`
	URL      string           `json:"url"`
}

type CategorySummary struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type LocationSummary struct {
	District *string `json:"district"`
	Village  *string `json:"village"`
}

type SearchMeta struct {
	Locale     string  `json:"locale"`
	Query      string  `json:"query"`
	Category   *string `json:"category"`
	Page       int     `json:"page"`
	PerPage    int     `json:"per_page"`
	TotalItems int     `json:"total_items"`
	TotalPages int     `json:"total_pages"`
}

type SearchResponse struct {
	Data []SearchResultItem `json:"data"`
	Meta SearchMeta         `json:"meta"`
}

func ValidateQuery(q string) (string, error) {
	trimmed := strings.TrimSpace(q)
	if len(trimmed) < 2 {
		return "", ErrQueryTooShort
	}
	if len(trimmed) > 120 {
		return "", ErrQueryTooLong
	}
	return trimmed, nil
}

func ValidateLocale(locale string) (string, error) {
	l := strings.ToLower(strings.TrimSpace(locale))
	if l != LocaleID && l != LocaleEN {
		return "", ErrInvalidLocale
	}
	return l, nil
}
