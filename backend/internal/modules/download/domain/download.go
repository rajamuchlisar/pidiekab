package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrDownloadNotFound = errors.New("download item not found")
	ErrTitleRequired    = errors.New("title_id is required")
	ErrFileURLRequired  = errors.New("file_url is required")
)

type DownloadItem struct {
	ID            int64     `json:"id"`
	TitleID       string    `json:"title_id"`
	TitleEN       string    `json:"title_en"`
	Category      string    `json:"category"`
	DescriptionID string    `json:"description_id"`
	DescriptionEN string    `json:"description_en"`
	FileURL       string    `json:"file_url"`
	FileName      string    `json:"file_name"`
	FileSize      int64     `json:"file_size"`
	FileType      string    `json:"file_type"`
	DownloadCount int64     `json:"download_count"`
	IsActive      bool      `json:"is_active"`
	SortOrder     int       `json:"sort_order"`
	CreatedBy     *int64    `json:"created_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (d *DownloadItem) Validate() error {
	d.TitleID = strings.TrimSpace(d.TitleID)
	d.TitleEN = strings.TrimSpace(d.TitleEN)
	d.Category = strings.TrimSpace(d.Category)
	d.FileURL = strings.TrimSpace(d.FileURL)
	d.FileName = strings.TrimSpace(d.FileName)
	d.FileType = strings.ToUpper(strings.TrimSpace(d.FileType))

	if d.TitleID == "" {
		return ErrTitleRequired
	}
	if d.FileURL == "" {
		return ErrFileURLRequired
	}
	if d.Category == "" {
		d.Category = "Umum"
	}
	if d.FileType == "" {
		d.FileType = "PDF"
	}
	return nil
}
