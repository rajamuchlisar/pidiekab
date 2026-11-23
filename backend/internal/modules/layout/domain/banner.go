package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrBannerNotFound     = errors.New("banner tidak ditemukan")
	ErrBannerTitleEmpty   = errors.New("judul/label banner tidak boleh kosong")
	ErrInvalidPlacement   = errors.New("penempatan tata letak tidak valid")
	ErrInvalidBannerType  = errors.New("tipe banner tidak valid; harus 'image' atau 'adsense'")
	ErrImageURLRequired   = errors.New("url gambar wajib diisi untuk banner gambar")
	ErrHTMLCodeRequired   = errors.New("kode adsense/script wajib diisi untuk tipe adsense")
)

const (
	PlacementHeaderTop      = "header_top"       // Di atas header / banner pengumuman atas
	PlacementBelowHero      = "below_hero"       // Di bawah hero beranda utama
	PlacementContentSidebar = "content_sidebar"  // Di sidebar cerita/konten
	PlacementFooterTop      = "footer_top"       // Di atas footer situs
	PlacementPopupPromo     = "popup_promo"      // Pop-up / floating promo

	TypeImage   = "image"
	TypeAdsense = "adsense"
)

// ValidPlacements lists all accepted banner placement spots.
var ValidPlacements = []string{
	PlacementHeaderTop,
	PlacementBelowHero,
	PlacementContentSidebar,
	PlacementFooterTop,
	PlacementPopupPromo,
}

// Banner represents a promotional banner or Google AdSense ad slot.
type Banner struct {
	ID           int64      `json:"id"`
	Title        string     `json:"title"`
	Placement    string     `json:"placement"` // header_top, below_hero, content_sidebar, footer_top, popup_promo
	BannerType   string     `json:"banner_type"` // image, adsense
	ImageURL     string     `json:"image_url"`
	TargetURL    string     `json:"target_url"`
	HTMLCode     string     `json:"html_code"`
	OpenInNewTab bool       `json:"open_in_new_tab"`
	IsActive     bool       `json:"is_active"`
	SortOrder    int        `json:"sort_order"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (b *Banner) Validate() error {
	b.Title = strings.TrimSpace(b.Title)
	b.Placement = strings.TrimSpace(b.Placement)
	b.BannerType = strings.TrimSpace(b.BannerType)
	b.ImageURL = strings.TrimSpace(b.ImageURL)
	b.TargetURL = strings.TrimSpace(b.TargetURL)
	b.HTMLCode = strings.TrimSpace(b.HTMLCode)

	if b.Title == "" {
		return ErrBannerTitleEmpty
	}

	isValidPlacement := false
	for _, p := range ValidPlacements {
		if b.Placement == p {
			isValidPlacement = true
			break
		}
	}
	if !isValidPlacement {
		return ErrInvalidPlacement
	}

	if b.BannerType != TypeImage && b.BannerType != TypeAdsense {
		return ErrInvalidBannerType
	}

	if b.BannerType == TypeImage && b.ImageURL == "" {
		return ErrImageURLRequired
	}

	if b.BannerType == TypeAdsense && b.HTMLCode == "" {
		return ErrHTMLCodeRequired
	}

	return nil
}
