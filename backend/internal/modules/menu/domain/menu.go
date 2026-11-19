package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrMenuNotFound     = errors.New("menu item not found")
	ErrTitleIDEmpty     = errors.New("judul menu bahasa indonesia tidak boleh kosong")
	ErrInvalidType      = errors.New("tipe menu tidak valid; harus static_page, content, category, atau external")
	ErrExternalURLEmpty = errors.New("url eksternal tidak boleh kosong untuk menu bertipe external")
	ErrTargetIDRequired = errors.New("target id wajib dipilih untuk menu bertipe static_page, content, atau category")
)

const (
	TypeStaticPage = "static_page"
	TypeContent    = "content"
	TypeCategory   = "category"
	TypeExternal   = "external"
)

type MenuItem struct {
	ID           int64     `json:"id"`
	TitleID      string    `json:"title_id"`
	TitleEN      string    `json:"title_en"`
	ItemType     string    `json:"item_type"` // static_page, content, category, external
	TargetID     *int64    `json:"target_id,omitempty"`
	ExternalURL  string    `json:"external_url,omitempty"`
	OpenInNewTab bool      `json:"open_in_new_tab"`
	SortOrder    int       `json:"sort_order"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Enriched fields for CMS display
	TargetTitle string `json:"target_title,omitempty"`
	ResolvedURL string `json:"resolved_url,omitempty"`
}

func (m *MenuItem) Validate() error {
	m.TitleID = strings.TrimSpace(m.TitleID)
	m.TitleEN = strings.TrimSpace(m.TitleEN)
	m.ItemType = strings.TrimSpace(m.ItemType)
	m.ExternalURL = strings.TrimSpace(m.ExternalURL)

	if m.TitleID == "" {
		return ErrTitleIDEmpty
	}

	switch m.ItemType {
	case TypeStaticPage, TypeContent, TypeCategory:
		if m.TargetID == nil || *m.TargetID <= 0 {
			return ErrTargetIDRequired
		}
	case TypeExternal:
		if m.ExternalURL == "" {
			return ErrExternalURLEmpty
		}
	default:
		return ErrInvalidType
	}

	return nil
}
