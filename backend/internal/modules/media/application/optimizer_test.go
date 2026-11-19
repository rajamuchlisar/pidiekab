package application

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

func TestOptimizeImage(t *testing.T) {
	// Create a 2400x1600 image with high-frequency details (representing a real-world camera photo)
	largeImg := image.NewRGBA(image.Rect(0, 0, 2400, 1600))
	for y := 0; y < 1600; y++ {
		for x := 0; x < 2400; x++ {
			largeImg.Set(x, y, color.RGBA{
				R: uint8((x*37 + y*19) % 256),
				G: uint8((x*13 + y*43) % 256),
				B: uint8((x*7 + y*61) % 256),
				A: 255,
			})
		}
	}

	var buf bytes.Buffer
	// Encode as high-quality unoptimized JPEG (Quality 98, typical for camera export)
	if err := jpeg.Encode(&buf, largeImg, &jpeg.Options{Quality: 98}); err != nil {
		t.Fatalf("failed to encode test jpeg: %v", err)
	}

	origBytes := buf.Bytes()
	origSize := len(origBytes)

	optimized, newW, newH, newMime := OptimizeImage(origBytes, "image/jpeg", 2400, 1600)

	t.Logf("Original: %d bytes (2400x1600), Optimized: %d bytes (%dx%d, %s)",
		origSize, len(optimized), newW, newH, newMime)

	if len(optimized) >= origSize {
		t.Errorf("expected optimized size to be smaller than %d, got %d", origSize, len(optimized))
	}
	if newW > MaxDimension || newH > MaxDimension {
		t.Errorf("expected dimensions within %d, got %dx%d", MaxDimension, newW, newH)
	}
	if newW != 1920 || newH != 1280 {
		t.Errorf("expected 1920x1280, got %dx%d", newW, newH)
	}
}
