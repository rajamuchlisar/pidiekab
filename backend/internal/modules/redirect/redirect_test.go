package redirect_test

import (
	"testing"

	"pidiekab/internal/modules/redirect/domain"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"   ", ""},
		{"id", "/id/"},
		{"/id", "/id/"},
		{"/id/", "/id/"},
		{"id/wisata-alam", "/id/wisata-alam/"},
		{"/id/wisata-alam/pantai.jpg", "/id/wisata-alam/pantai.jpg"},
		{"id/wisata-alam/pantai.png", "/id/wisata-alam/pantai.png"},
		{"/id/wisata-alam/artikel-lama", "/id/wisata-alam/artikel-lama/"},
	}

	for _, tc := range tests {
		actual := domain.NormalizePath(tc.input)
		if actual != tc.expected {
			t.Errorf("NormalizePath(%q) = %q, want %q", tc.input, actual, tc.expected)
		}
	}
}
