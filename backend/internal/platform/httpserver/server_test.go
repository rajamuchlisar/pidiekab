package httpserver_test

import (
	"context"
	"encoding/json"
	"errors"
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

type mockPinger struct {
	pingErr error
}

func (m *mockPinger) Ping(ctx context.Context) error {
	return m.pingErr
}

func newTestServerWithStaticAndDB(publicDir, cmsDir string, db httpserver.DatabasePinger) http.Handler {
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
	return httpserver.BuildHandler(cfg, logger, db)
}

func TestHealthLiveEndpoint(t *testing.T) {
	handler := newTestServerWithStaticAndDB("", "", nil)

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

func TestHealthReadyEndpointUnconfigured(t *testing.T) {
	handler := newTestServerWithStaticAndDB("", "", nil)

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
	if res["reason"] != "database_not_configured" {
		t.Errorf("expected reason 'database_not_configured', got %v", res["reason"])
	}
}

func TestHealthReadyEndpointUnreachable(t *testing.T) {
	pinger := &mockPinger{pingErr: errors.New("connection refused")}
	handler := newTestServerWithStaticAndDB("", "", pinger)

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
	if res["reason"] != "database_unreachable" {
		t.Errorf("expected reason 'database_unreachable', got %v", res["reason"])
	}
}

func TestHealthReadyEndpointHealthy(t *testing.T) {
	pinger := &mockPinger{pingErr: nil}
	handler := newTestServerWithStaticAndDB("", "", pinger)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if res["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", res["status"])
	}
	if res["database"] != "connected" {
		t.Errorf("expected database 'connected', got %v", res["database"])
	}
}

func TestRootRedirectsToIndonesian(t *testing.T) {
	handler := newTestServerWithStaticAndDB("", "", nil)

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

	handler := newTestServerWithStaticAndDB(tempDir, "", nil)

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

	handler := newTestServerWithStaticAndDB("", tempDir, nil)

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

func TestSecurityHeadersHardening(t *testing.T) {
	handler := newTestServerWithStaticAndDB("", "", nil)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	headers := rec.Header()
	if csp := headers.Get("Content-Security-Policy"); csp == "" {
		t.Error("expected Content-Security-Policy header, got empty")
	}
	if perm := headers.Get("Permissions-Policy"); perm == "" {
		t.Error("expected Permissions-Policy header, got empty")
	}
	if coop := headers.Get("Cross-Origin-Opener-Policy"); coop != "same-origin" {
		t.Errorf("expected Cross-Origin-Opener-Policy 'same-origin', got %q", coop)
	}
	if corp := headers.Get("Cross-Origin-Resource-Policy"); corp != "same-origin" {
		t.Errorf("expected Cross-Origin-Resource-Policy 'same-origin', got %q", corp)
	}

	// Test HTTPS request gets HSTS header
	reqHTTPS := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	reqHTTPS.Header.Set("X-Forwarded-Proto", "https")
	recHTTPS := httptest.NewRecorder()
	handler.ServeHTTP(recHTTPS, reqHTTPS)
	if hsts := recHTTPS.Header().Get("Strict-Transport-Security"); hsts == "" {
		t.Error("expected Strict-Transport-Security header for HTTPS request, got empty")
	}
}

func TestHTTPToHTTPSRedirectServer(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	redirectSrv := httpserver.NewHTTPToHTTPSRedirectServer(":80", logger)

	req := httptest.NewRequest(http.MethodGet, "/id/wisata-destinasi?page=1", nil)
	req.Host = "pidiekab.web.id:80"
	rec := httptest.NewRecorder()
	redirectSrv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301 Moved Permanently, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	expected := "https://pidiekab.web.id/id/wisata-destinasi?page=1"
	if location != expected {
		t.Errorf("expected redirect location %q, got %q", expected, location)
	}
}

func TestTLSConfiguration(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	cfg := config.Config{
		Address:           ":443",
		TLSCertFile:       "dummy_cert.pem",
		TLSKeyFile:        "dummy_key.pem",
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	srv := httpserver.New(cfg, logger, nil)
	if srv.TLSConfig == nil {
		t.Fatal("expected TLSConfig to be configured, got nil")
	}
	if srv.TLSConfig.MinVersion < 0x0303 { // TLS 1.2
		t.Errorf("expected min TLS version 1.2, got %x", srv.TLSConfig.MinVersion)
	}
	if len(srv.TLSConfig.CipherSuites) == 0 {
		t.Error("expected explicit secure cipher suites, got empty")
	}
}

func TestHealthMetricsEndpoint(t *testing.T) {
	handler := newTestServerWithStaticAndDB("", "", nil)

	req := httptest.NewRequest(http.MethodGet, "/health/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if res["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", res["status"])
	}
	if _, ok := res["uptime_seconds"]; !ok {
		t.Error("missing uptime_seconds in metrics response")
	}
	if _, ok := res["mem_alloc_mb"]; !ok {
		t.Error("missing mem_alloc_mb in metrics response")
	}
}

type mockRedirectFinder struct{}

func (m *mockRedirectFinder) RegisterRoutes(mux *http.ServeMux) {}
func (m *mockRedirectFinder) FindRedirect(ctx context.Context, path string) (string, int, error) {
	if path == "/legacy-path/" || path == "/legacy-path" {
		return "/id/wisata-alam/pantai-pelangi/", http.StatusMovedPermanently, nil
	}
	return "", 0, nil
}

func TestRedirectFinderIntegration(t *testing.T) {
	cfg := config.Config{
		Environment:     "test",
		PublicStaticDir: t.TempDir(),
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := httpserver.BuildHandler(cfg, logger, nil, &mockRedirectFinder{})

	req := httptest.NewRequest(http.MethodGet, "/legacy-path/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301 Moved Permanently, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "/id/wisata-alam/pantai-pelangi/" {
		t.Errorf("expected redirect location '/id/wisata-alam/pantai-pelangi/', got %q", loc)
	}
}

func TestOpenAPIEndpoint(t *testing.T) {
	handler := newTestServerWithStaticAndDB("", "", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Since api/openapi.yaml exists at project root, it should resolve and return 200 OK
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /api/v1/openapi.yaml, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/yaml; charset=utf-8" {
		t.Errorf("expected application/yaml; charset=utf-8, got %q", contentType)
	}
}

func TestCustomErrorPagesServing(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create root 404.html, 401.html, 403.html
	err404Content := "<html><body>Custom Root 404</body></html>"
	err401Content := "<html><body>Custom 401 Unauthorized</body></html>"
	err403Content := "<html><body>Custom 403 Forbidden</body></html>"
	id404Content := "<html><body>Custom ID 404 Halaman Tidak Ditemukan</body></html>"

	_ = os.WriteFile(filepath.Join(tempDir, "404.html"), []byte(err404Content), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "401.html"), []byte(err401Content), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "403.html"), []byte(err403Content), 0644)

	_ = os.MkdirAll(filepath.Join(tempDir, "id", "404"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "id", "404", "index.html"), []byte(id404Content), 0644)

	handler := newTestServerWithStaticAndDB(tempDir, "", nil)

	// 1. Missing page at root -> custom root 404
	req := httptest.NewRequest(http.MethodGet, "/unknown-page", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
	if rec.Body.String() != err404Content {
		t.Errorf("expected %q, got %q", err404Content, rec.Body.String())
	}

	// 2. Missing page under /id/ -> custom ID 404
	req = httptest.NewRequest(http.MethodGet, "/id/unknown-page/", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
	if rec.Body.String() != id404Content {
		t.Errorf("expected %q, got %q", id404Content, rec.Body.String())
	}

	// 3. Direct access to /401 without extension
	req = httptest.NewRequest(http.MethodGet, "/401", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != err401Content {
		t.Errorf("expected %q, got %q", err401Content, rec.Body.String())
	}
}
