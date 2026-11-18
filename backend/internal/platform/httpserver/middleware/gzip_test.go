package middleware_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pidiekab/internal/platform/httpserver/middleware"
)

func TestGzipMiddleware(t *testing.T) {
	samplePayload := strings.Repeat("PidieKab Regional Promotion Platform Aceh ", 50)

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(samplePayload))
	})

	handler := middleware.Gzip(baseHandler)

	t.Run("without gzip acceptance", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") != "" {
			t.Errorf("expected no Content-Encoding, got %q", rec.Header().Get("Content-Encoding"))
		}
		if rec.Body.String() != samplePayload {
			t.Errorf("expected body to match samplePayload")
		}
	})

	t.Run("with gzip acceptance", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip, deflate")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") != "gzip" {
			t.Errorf("expected Content-Encoding 'gzip', got %q", rec.Header().Get("Content-Encoding"))
		}
		if !strings.Contains(rec.Header().Get("Vary"), "Accept-Encoding") {
			t.Errorf("expected Vary to include Accept-Encoding, got %q", rec.Header().Get("Vary"))
		}

		// Decompress and verify
		reader, err := gzip.NewReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to decompress gzip stream: %v", err)
		}

		if string(decompressed) != samplePayload {
			t.Errorf("decompressed content does not match original payload")
		}

		// Ensure compressed size is significantly smaller than original
		if rec.Body.Len() >= len(samplePayload) {
			t.Errorf("expected compressed size (%d) to be smaller than original (%d)", rec.Body.Len(), len(samplePayload))
		}
	})

	t.Run("skip compression for binary images", func(t *testing.T) {
		imagePayload := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG magic bytes
		imageHandler := middleware.Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(imagePayload)
		}))

		req := httptest.NewRequest(http.MethodGet, "/logo.png", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		imageHandler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") != "" {
			t.Errorf("expected no compression for image/png, got %q", rec.Header().Get("Content-Encoding"))
		}
		if !bytes.Equal(rec.Body.Bytes(), imagePayload) {
			t.Errorf("expected raw image bytes to remain unmodified")
		}
	})

	t.Run("skip compression for HEAD requests", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodHead, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") != "" {
			t.Errorf("expected no Content-Encoding on HEAD, got %q", rec.Header().Get("Content-Encoding"))
		}
	})
}
