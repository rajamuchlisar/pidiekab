package domain

import (
	"time"
)

// Regency represents a kabupaten/kota administrative division.
type Regency struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      *string   `json:"code,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// District represents a kecamatan administrative division.
type District struct {
	ID        int64     `json:"id"`
	RegencyID int64     `json:"regency_id"`
	Name      string    `json:"name"`
	Code      *string   `json:"code,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Village represents a gampong/desa administrative division.
type Village struct {
	ID         int64     `json:"id"`
	DistrictID int64     `json:"district_id"`
	Name       string    `json:"name"`
	Code       *string   `json:"code,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

