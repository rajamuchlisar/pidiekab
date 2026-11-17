package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	auditDomain "pidiekab/internal/modules/audit/domain"
	"pidiekab/internal/modules/identity/application"
	"pidiekab/internal/modules/identity/domain"
	identityHttp "pidiekab/internal/modules/identity/transport/http"
	"pidiekab/internal/platform/authn"
	"pidiekab/internal/platform/httpserver/middleware"
)

type memAdminRepo struct {
	admins map[int64]*domain.Admin
	byMail map[string]*domain.Admin
	nextID int64
}

func newMemAdminRepo() *memAdminRepo {
	return &memAdminRepo{
		admins: make(map[int64]*domain.Admin),
		byMail: make(map[string]*domain.Admin),
		nextID: 1,
	}
}

func (m *memAdminRepo) Create(ctx context.Context, a *domain.Admin) error {
	norm := domain.NormalizeEmail(a.Email)
	a.ID = m.nextID
	m.nextID++
	a.Email = norm
	m.admins[a.ID] = a
	m.byMail[norm] = a
	return nil
}

func (m *memAdminRepo) GetByID(ctx context.Context, id int64) (*domain.Admin, error) {
	a, ok := m.admins[id]
	if !ok {
		return nil, domain.ErrAdminNotFound
	}
	return a, nil
}

func (m *memAdminRepo) GetByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	a, ok := m.byMail[domain.NormalizeEmail(email)]
	if !ok {
		return nil, domain.ErrAdminNotFound
	}
	return a, nil
}

func (m *memAdminRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	a, ok := m.admins[id]
	if !ok {
		return domain.ErrAdminNotFound
	}
	a.PasswordHash = passwordHash
	return nil
}

func (m *memAdminRepo) UpdateLastLogin(ctx context.Context, id int64, loginTime time.Time) error {
	a, ok := m.admins[id]
	if !ok {
		return domain.ErrAdminNotFound
	}
	a.LastLoginAt = &loginTime
	return nil
}

func (m *memAdminRepo) CountActive(ctx context.Context) (int, error) { return len(m.admins), nil }
func (m *memAdminRepo) SetIsActive(ctx context.Context, id int64, isActive bool) error {
	return nil
}
func (m *memAdminRepo) List(ctx context.Context) ([]*domain.Admin, error) { return nil, nil }

type memSessionRepo struct {
	sessions map[string]*domain.Session
	nextID   int64
}

func newMemSessionRepo() *memSessionRepo {
	return &memSessionRepo{
		sessions: make(map[string]*domain.Session),
		nextID:   1,
	}
}

func (m *memSessionRepo) Create(ctx context.Context, s *domain.Session) error {
	s.ID = m.nextID
	m.nextID++
	m.sessions[s.TokenHash] = s
	return nil
}

func (m *memSessionRepo) GetActiveByTokenHash(ctx context.Context, tokenHash string, now time.Time) (*domain.Session, error) {
	s, ok := m.sessions[tokenHash]
	if !ok || s.RevokedAt != nil || now.After(s.ExpiresAt) {
		return nil, domain.ErrSessionNotFound
	}
	return s, nil
}

func (m *memSessionRepo) RevokeByTokenHash(ctx context.Context, tokenHash string, revokeTime time.Time) error {
	s, ok := m.sessions[tokenHash]
	if !ok {
		return domain.ErrSessionNotFound
	}
	s.RevokedAt = &revokeTime
	return nil
}

func (m *memSessionRepo) RevokeAllForAdmin(ctx context.Context, adminID int64, revokeTime time.Time) error {
	for _, s := range m.sessions {
		if s.AdminID == adminID {
			s.RevokedAt = &revokeTime
		}
	}
	return nil
}

type memAudit struct{}

func (m *memAudit) Record(ctx context.Context, entry auditDomain.Log) error { return nil }

func setupTestRouter(t *testing.T) (*http.ServeMux, *application.AuthService, *domain.Admin) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	params := authn.FastArgon2idParamsForTesting()

	adminRepo := newMemAdminRepo()
	sessionRepo := newMemSessionRepo()
	auditRepo := &memAudit{}

	authSvc := application.NewAuthService(adminRepo, sessionRepo, auditRepo, params, 1*time.Hour, logger)

	admin, err := authSvc.BootstrapAdmin(ctx, "admin@pidiekab.id", "Pidie Admin", "AdminPassword123!")
	if err != nil {
		t.Fatalf("failed to bootstrap admin: %v", err)
	}

	handler := identityHttp.NewHandler(authSvc, false)
	mux := http.NewServeMux()
	handler.RegisterRoutesWithLimiter(mux, nil)

	return mux, authSvc, admin
}

func TestAdminAuthHTTPFlow(t *testing.T) {
	mux, _, admin := setupTestRouter(t)

	// 1. Login with bad credentials -> 401
	badLoginBody := []byte(`{"email":"admin@pidiekab.id","password":"WrongPassword"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", bytes.NewReader(badLoginBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for bad login, got %d", rec.Code)
	}

	// 2. Successful Login -> 204 with Cookies
	validLoginBody := []byte(`{"email":"admin@pidiekab.id","password":"AdminPassword123!"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", bytes.NewReader(validLoginBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for login, got %d", rec.Code)
	}

	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == middleware.AdminSessionCookieName {
			sessionCookie = cookie
		}
		if cookie.Name == middleware.CSRFCookieName {
			csrfCookie = cookie
		}
	}

	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatal("session cookie missing from login response")
	}
	if !sessionCookie.HttpOnly {
		t.Error("session cookie must be HttpOnly")
	}
	if csrfCookie == nil || csrfCookie.Value == "" {
		t.Fatal("csrf cookie missing from login response")
	}

	// 3. GET /api/v1/admin/auth/me with session cookie -> 200 OK
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/auth/me", nil)
	req.AddCookie(sessionCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for me endpoint, got %d", rec.Code)
	}

	var meResp struct {
		Data struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &meResp); err != nil {
		t.Fatalf("failed to decode me response: %v", err)
	}
	if meResp.Data.Email != admin.Email {
		t.Errorf("expected email %q, got %q", admin.Email, meResp.Data.Email)
	}

	// 4. Change Password with CSRF protection -> 204 No Content
	changeBody := []byte(`{"current_password":"AdminPassword123!","new_password":"NewStrongPassword2026!"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/password/change", bytes.NewReader(changeBody))
	req.Host = "localhost"
	req.Header.Set("Origin", "http://localhost")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(middleware.CSRFHeaderName, csrfCookie.Value)
	req.AddCookie(sessionCookie)
	req.AddCookie(csrfCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for password change, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	// 5. Subsequent request with old session should fail (sessions were revoked)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/auth/me", nil)
	req.AddCookie(sessionCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after password change, got %d", rec.Code)
	}

	// 6. Login with new password
	newLoginBody := []byte(`{"email":"admin@pidiekab.id","password":"NewStrongPassword2026!"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", bytes.NewReader(newLoginBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for login with new password, got %d", rec.Code)
	}

	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == middleware.AdminSessionCookieName {
			sessionCookie = cookie
		}
		if cookie.Name == middleware.CSRFCookieName {
			csrfCookie = cookie
		}
	}

	// 7. Logout -> 204 with cleared cookies
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/logout", nil)
	req.Host = "localhost"
	req.Header.Set("Origin", "http://localhost")
	req.Header.Set(middleware.CSRFHeaderName, csrfCookie.Value)
	req.AddCookie(sessionCookie)
	req.AddCookie(csrfCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for logout, got %d", rec.Code)
	}

	// Verify session cookie was expired
	var clearedSessionCookie *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == middleware.AdminSessionCookieName {
			clearedSessionCookie = cookie
		}
	}
	if clearedSessionCookie == nil || clearedSessionCookie.MaxAge != -1 {
		t.Errorf("expected session cookie to be cleared with MaxAge -1, got %v", clearedSessionCookie)
	}
}
