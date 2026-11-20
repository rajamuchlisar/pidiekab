package download_test

import (
	"testing"

	"pidiekab/internal/modules/download/domain"
)

func TestDownloadValidation(t *testing.T) {
	// Missing title
	d1 := &domain.DownloadItem{
		FileURL: "/media/test.pdf",
	}
	if err := d1.Validate(); err != domain.ErrTitleRequired {
		t.Errorf("expected ErrTitleRequired, got %v", err)
	}

	// Missing file_url
	d2 := &domain.DownloadItem{
		TitleID: "Panduan Wisata",
	}
	if err := d2.Validate(); err != domain.ErrFileURLRequired {
		t.Errorf("expected ErrFileURLRequired, got %v", err)
	}

	// Valid with defaults applied
	d3 := &domain.DownloadItem{
		TitleID: "Panduan Wisata",
		FileURL: "/media/guide.pdf",
	}
	if err := d3.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if d3.Category != "Umum" {
		t.Errorf("expected default category Umum, got %s", d3.Category)
	}
	if d3.FileType != "PDF" {
		t.Errorf("expected default file type PDF, got %s", d3.FileType)
	}
}
