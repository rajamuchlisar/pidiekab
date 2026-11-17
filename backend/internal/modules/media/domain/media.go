package domain

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrUnsupportedMediaType = errors.New("unsupported media type; only JPEG, PNG, and WebP are allowed")
	ErrFileTooLarge         = errors.New("file size exceeds maximum allowed limit (10MB)")
	ErrInvalidImage         = errors.New("invalid or corrupted image data")
	ErrMediaNotFound        = errors.New("media asset not found")
	ErrMediaInUse           = errors.New("media asset is currently in use and cannot be deleted")
	ErrInvalidStorageKey    = errors.New("invalid storage key")
)

const (
	MaxMediaSizeBytes = 10 * 1024 * 1024 // 10 MB
	MimeJPEG          = "image/jpeg"
	MimePNG           = "image/png"
	MimeWebP          = "image/webp"
)

// MediaAsset represents a stored media file and its rights metadata.
type MediaAsset struct {
	ID               int64     `json:"id"`
	StorageKey       string    `json:"storage_key"`
	OriginalFilename string    `json:"original_filename"`
	MimeType         string    `json:"mime_type"`
	FileSize         int64     `json:"file_size"`
	Width            *int      `json:"width,omitempty"`
	Height           *int      `json:"height,omitempty"`
	AltTextID        string    `json:"alt_text_id"`
	AltTextEN        string    `json:"alt_text_en"`
	SourceURL        string    `json:"source_url"`
	RightsStatus     string    `json:"rights_status"`
	Attribution      string    `json:"attribution"`
	UploadedBy       *int64    `json:"uploaded_by,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// SniffAndValidateImage analyzes raw bytes to determine MIME type, dimensions, and integrity.
func SniffAndValidateImage(data []byte, filename string) (mimeType string, width, height int, err error) {
	if len(data) == 0 {
		return "", 0, 0, ErrInvalidImage
	}
	if int64(len(data)) > MaxMediaSizeBytes {
		return "", 0, 0, ErrFileTooLarge
	}

	// Extension check (defense in depth)
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
	}
	if ext != "" && !allowedExts[ext] {
		return "", 0, 0, ErrUnsupportedMediaType
	}

	// Sniff magic bytes
	if isJPEG(data) {
		cfg, _, decodeErr := image.DecodeConfig(bytes.NewReader(data))
		if decodeErr != nil {
			return "", 0, 0, fmt.Errorf("%w: failed to decode JPEG header", ErrInvalidImage)
		}
		return MimeJPEG, cfg.Width, cfg.Height, nil
	}

	if isPNG(data) {
		cfg, _, decodeErr := image.DecodeConfig(bytes.NewReader(data))
		if decodeErr != nil {
			return "", 0, 0, fmt.Errorf("%w: failed to decode PNG header", ErrInvalidImage)
		}
		return MimePNG, cfg.Width, cfg.Height, nil
	}

	if isWebP(data) {
		w, h, webpErr := parseWebPDimensions(data)
		if webpErr != nil {
			return "", 0, 0, fmt.Errorf("%w: failed to decode WebP header: %v", ErrInvalidImage, webpErr)
		}
		return MimeWebP, w, h, nil
	}

	return "", 0, 0, ErrUnsupportedMediaType
}

func isJPEG(b []byte) bool {
	return len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF
}

func isPNG(b []byte) bool {
	return len(b) >= 8 &&
		b[0] == 0x89 && b[1] == 0x50 && b[2] == 0x4E && b[3] == 0x47 &&
		b[4] == 0x0D && b[5] == 0x0A && b[6] == 0x1A && b[7] == 0x0A
}

func isWebP(b []byte) bool {
	return len(b) >= 12 &&
		string(b[0:4]) == "RIFF" &&
		string(b[8:12]) == "WEBP"
}

// parseWebPDimensions extracts width and height from standard WebP VP8, VP8L, or VP8X headers.
func parseWebPDimensions(b []byte) (int, int, error) {
	if len(b) < 16 {
		return 0, 0, io.ErrUnexpectedEOF
	}

	formatChunk := string(b[12:16])
	switch formatChunk {
	case "VP8 ": // Lossy VP8
		if len(b) < 30 {
			return 0, 0, io.ErrUnexpectedEOF
		}
		// VP8 frame starts at offset 20. Keyframe check: bytes 23,24,25 must be 0x9D, 0x01, 0x2A
		if b[23] != 0x9D || b[24] != 0x01 || b[25] != 0x2A {
			return 0, 0, errors.New("invalid VP8 keyframe sync code")
		}
		width := int(binary.LittleEndian.Uint16(b[26:28]) & 0x3FFF)
		height := int(binary.LittleEndian.Uint16(b[28:30]) & 0x3FFF)
		return width, height, nil

	case "VP8L": // Lossless VP8L
		if len(b) < 25 {
			return 0, 0, io.ErrUnexpectedEOF
		}
		// Signature byte at 20 must be 0x2F
		if b[20] != 0x2F {
			return 0, 0, errors.New("invalid VP8L signature")
		}
		val := binary.LittleEndian.Uint32(b[21:25])
		width := int(val&0x3FFF) + 1
		height := int((val>>14)&0x3FFF) + 1
		return width, height, nil

	case "VP8X": // Extended VP8X
		if len(b) < 30 {
			return 0, 0, io.ErrUnexpectedEOF
		}
		// Canvas width is 24-bit integer at bytes 24-26 (+ 1)
		width := int(b[24]) | (int(b[25]) << 8) | (int(b[26]) << 16) + 1
		// Canvas height is 24-bit integer at bytes 27-29 (+ 1)
		height := int(b[27]) | (int(b[28]) << 8) | (int(b[29]) << 16) + 1
		return width, height, nil

	default:
		return 0, 0, fmt.Errorf("unsupported WebP chunk format: %s", formatChunk)
	}
}
