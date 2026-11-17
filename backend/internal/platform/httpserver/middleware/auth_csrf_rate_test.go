package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pidiekab/internal/modules/identity/domain"
	"pidiekab/internal/platform/httpserver/middleware"
)

type mockAuthenticator struct {
	admin *domain.Admin
	sess  *domain.Session
	err   error
}

func (m *mockAuthenticator) AuthenticateSession(ctx context.Context, rawToken string) (*domain.Admin, *domain.Session, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	return m.admin, m.sess, nil
}

func TestRequireAdminSession(t *testing.T) {
	// 1. Missing cookie -> 401
	auth := &mockAuthenticator{}
	handler := middleware.RequireAdminSession(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/auth/me", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	// 2. Invalid session -> 401
	auth.err = domain.ErrSessionNotFound
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: middleware.AdminSessionCookieName, Value: "invalid-token"})
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	// 3. Valid session -> 200 with admin in context
	auth.err = nil
	auth.admin = &domain.Admin{ID: 42, Email: "admin@pidiekab.id", DisplayName: "Admin"}
	auth.sess = &domain.Session{ID: 10, AdminID: 42}

	var contextAdmin *domain.Admin
	checkHandler := middleware.RequireAdminSession(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		contextAdmin, ok = middleware.AdminFromContext(r.Context())
		if !ok {
			t.Error("admin missing from request context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: middleware.AdminSessionCookieName, Value: "valid-token"})
	rec = httptest.NewRecorder()
	checkHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if contextAdmin == nil || contextAdmin.ID != 42 {
		t.Errorf("expected admin ID 42, got %v", contextAdmin)
	}
}

func TestCSRFProtection(t *testing.T) {
	dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := middleware.CSRFProtection(dummy)

	// 1. GET requests should pass without CSRF
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/contents", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected GET to pass, got %d", rec.Code)
	}

	// 2. Login endpoint should pass without CSRF token (origin matching)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", nil)
	req.Host = "pidiekab.id"
	req.Header.Set("Origin", "https://pidiekab.id")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected login to pass with matching origin, got %d", rec.Code)
	}

	// 3. Login endpoint with mismatched origin should be rejected 403
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", nil)
	req.Host = "pidiekab.id"
	req.Header.Set("Origin", "https://evil-site.com")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected login with evil origin to return 403, got %d", rec.Code)
	}

	// 4. Mutating endpoint missing CSRF token -> 403
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/contents", nil)
	req.Host = "pidiekab.id"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected mutating request without CSRF to return 403, got %d", rec.Code)
	}

	// 5. Mutating endpoint with matching CSRF cookie and header -> 200
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/contents", nil)
	req.Host = "pidiekab.id"
	req.Header.Set("Origin", "https://pidiekab.id")
	token := "csrf-secret-token-123456"
	req.Header.Set(middleware.CSRFHeaderName, token)
	req.AddCookie(&http.Cookie{Name: middleware.CSRFCookieName, Value: token})
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected mutating request with valid CSRF to return 200, got %d", rec.Code)
	}
}

func TestMemoryRateLimiter(t *testing.T) {
	limiter := middleware.NewMemoryRateLimiter(3, 1*time.Minute)
	dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := limiter.LimitHandler(dummy)

	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", nil)
		req.RemoteAddr = "192.168.1.100:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d should have passed, got %d", i, rec.Code)
		}
	}

	// 4th request should exceed limit -> 429
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", nil)
	req.RemoteAddr = "192.168.1.100:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests, got %d", rec.Code)
	}
}
