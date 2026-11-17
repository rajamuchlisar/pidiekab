package httpserver_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pidiekab/internal/platform/config"
	"pidiekab/internal/platform/httpserver"
)

func newTestServerWithStatic(publicDir, cmsDir string) http.Handler {
	cfg := config.Config{
		Environment:       "test",
		Address:           "127.0.0.1:0",
		LogLevel:          slog.LevelDebug,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		PublicStaticDir:   publicDir,
		CMSStaticDir:      cmsDir,
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	return httpserver.BuildHandler(cfg, logger)
}

func TestHealthLiveEndpoint(t *testing.T) {
	handler := newTestServerWithStatic("", "")

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("expected application/json; charset=utf-8, got %q", contentType)
	}

	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing X-Content-Type-Options header")
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Error("missing X-Request-Id header")
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if res["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", res["status"])
	}
}

func TestHealthReadyEndpoint(t *testing.T) {
	handler := newTestServerWithStatic("", "")

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 Service Unavailable, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if res["status"] != "not_ready" {
		t.Errorf("expected status 'not_ready', got %v", res["status"])
	}
}

func TestRootRedirectsToIndonesian(t *testing.T) {
	handler := newTestServerWithStatic("", "")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301 Moved Permanently, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/id/" {
		t.Errorf("expected redirect location /id/, got %q", loc)
	}
}

func TestPublicStaticServing(t *testing.T) {
	tempDir := t.TempDir()
	idDir := filepath.Join(tempDir, "id")
	if err := os.MkdirAll(idDir, 0755); err != nil {
		t.Fatal(err)
	}
	htmlContent := "<!DOCTYPE html><html><head><title>Test</title></head><body>Hello Pidie</body></html>"
	if err := os.WriteFile(filepath.Join(idDir, "index.html"), []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}

	handler := newTestServerWithStatic(tempDir, "")

	// 1. Without trailing slash -> 301 to /id/
	req := httptest.NewRequest(http.MethodGet, "/id", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("expected 301, got %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/id/" {
		t.Errorf("expected location /id/, got %q", rec.Header().Get("Location"))
	}

	// 2. With trailing slash -> 200 and serves HTML
	req = httptest.NewRequest(http.MethodGet, "/id/", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != htmlContent {
		t.Errorf("expected %q, got %q", htmlContent, rec.Body.String())
	}

	// 3. Path traversal attack -> 403 Forbidden
	req = httptest.NewRequest(http.MethodGet, "/../etc/passwd", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden && rec.Code != http.StatusBadRequest && rec.Code != http.StatusMovedPermanently {
		t.Errorf("expected forbidden/bad request, got %d", rec.Code)
	}
}

func TestCMSStaticServingSPA(t *testing.T) {
	tempDir := t.TempDir()
	spaHtml := "<!DOCTYPE html><html><body>CMS SPA</body></html>"
	if err := os.WriteFile(filepath.Join(tempDir, "index.html"), []byte(spaHtml), 0644); err != nil {
		t.Fatal(err)
	}

	handler := newTestServerWithStatic("", tempDir)

	// Route to /cms/dashboard -> falls back to CMS index.html with no-cache
	req := httptest.NewRequest(http.MethodGet, "/cms/dashboard", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != spaHtml {
		t.Errorf("expected SPA index, got %q", rec.Body.String())
	}
	if cache := rec.Header().Get("Cache-Control"); cache != "no-store, no-cache, must-revalidate" {
		t.Errorf("expected no-cache for SPA, got %q", cache)
	}
}
