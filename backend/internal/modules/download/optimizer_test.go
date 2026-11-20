package download

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestOptimizeDownloadFile_Limits(t *testing.T) {
	// 1. Empty file
	_, _, _, err := OptimizeDownloadFile("empty.pdf", []byte{})
	if err == nil {
		t.Error("expected error for empty file, got nil")
	}

	// 2. Disallowed extension
	_, _, _, err = OptimizeDownloadFile("malicious.exe", []byte("MZ12345"))
	if err == nil || !strings.Contains(err.Error(), "tidak diizinkan") {
		t.Errorf("expected disallowed extension error, got %v", err)
	}

	// 3. File too large (> 10MB)
	largeData := make([]byte, MaxDownloadFileSizeBytes+100)
	_, _, _, err = OptimizeDownloadFile("too-large.pdf", largeData)
	if err == nil || !strings.Contains(err.Error(), "melebihi batas maksimum 10 MB") {
		t.Errorf("expected file too large error, got %v", err)
	}
}

func TestOptimizeDownloadFile_ImageCompression(t *testing.T) {
	// Create an uncompressed PNG image
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.NoCompression}
	_ = enc.Encode(&buf, img)
	originalData := buf.Bytes()

	optimized, fileType, size, err := OptimizeDownloadFile("sample-map.png", originalData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fileType != "PNG" {
		t.Errorf("expected fileType PNG, got %s", fileType)
	}
	if size >= int64(len(originalData)) {
		t.Errorf("expected compressed size (%d) to be smaller than uncompressed (%d)", size, len(originalData))
	}
	if len(optimized) != int(size) {
		t.Errorf("mismatch between len(optimized) and size")
	}
}

func TestOptimizeDownloadFile_TextTrim(t *testing.T) {
	textData := []byte("Header1,Header2\r\nValue1,Value2\r\n\r\n\n  \n")
	optimized, fileType, size, err := OptimizeDownloadFile("data.csv", textData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fileType != "CSV" {
		t.Errorf("expected CSV fileType, got %s", fileType)
	}
	if size >= int64(len(textData)) {
		t.Errorf("expected trimmed size (%d) to be smaller than original (%d)", size, len(textData))
	}
	if bytes.HasSuffix(optimized, []byte("\n")) {
		t.Errorf("expected trailing newlines to be trimmed")
	}
}
