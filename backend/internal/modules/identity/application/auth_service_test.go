package application_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	auditDomain "pidiekab/internal/modules/audit/domain"
	"pidiekab/internal/modules/identity/application"
	"pidiekab/internal/modules/identity/domain"
	"pidiekab/internal/platform/authn"
)

type memoryAdminRepo struct {
	admins map[int64]*domain.Admin
	byMail map[string]*domain.Admin
	nextID int64
}

func newMemoryAdminRepo() *memoryAdminRepo {
	return &memoryAdminRepo{
		admins: make(map[int64]*domain.Admin),
		byMail: make(map[string]*domain.Admin),
		nextID: 1,
	}
}

func (m *memoryAdminRepo) Create(ctx context.Context, a *domain.Admin) error {
	norm := domain.NormalizeEmail(a.Email)
	if _, exists := m.byMail[norm]; exists {
		return domain.ErrEmailAlreadyExists
	}
	a.ID = m.nextID
	m.nextID++
	a.Email = norm
	a.CreatedAt = time.Now().UTC()
	a.UpdatedAt = a.CreatedAt

	m.admins[a.ID] = a
	m.byMail[norm] = a
	return nil
}

func (m *memoryAdminRepo) GetByID(ctx context.Context, id int64) (*domain.Admin, error) {
	a, ok := m.admins[id]
	if !ok {
		return nil, domain.ErrAdminNotFound
	}
	return a, nil
}

func (m *memoryAdminRepo) GetByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	a, ok := m.byMail[domain.NormalizeEmail(email)]
	if !ok {
		return nil, domain.ErrAdminNotFound
	}
	return a, nil
}

func (m *memoryAdminRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	a, ok := m.admins[id]
	if !ok {
		return domain.ErrAdminNotFound
	}
	a.PasswordHash = passwordHash
	a.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *memoryAdminRepo) UpdateLastLogin(ctx context.Context, id int64, loginTime time.Time) error {
	a, ok := m.admins[id]
	if !ok {
		return domain.ErrAdminNotFound
	}
	a.LastLoginAt = &loginTime
	return nil
}

func (m *memoryAdminRepo) CountActive(ctx context.Context) (int, error) {
	count := 0
	for _, a := range m.admins {
		if a.IsActive {
			count++
		}
	}
	return count, nil
}

func (m *memoryAdminRepo) SetIsActive(ctx context.Context, id int64, isActive bool) error {
	a, ok := m.admins[id]
	if !ok {
		return domain.ErrAdminNotFound
	}
	a.IsActive = isActive
	a.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *memoryAdminRepo) List(ctx context.Context) ([]*domain.Admin, error) {
	list := make([]*domain.Admin, 0, len(m.admins))
	for _, a := range m.admins {
		list = append(list, a)
	}
	return list, nil
}

type memorySessionRepo struct {
	sessions map[string]*domain.Session
	nextID   int64
}

func newMemorySessionRepo() *memorySessionRepo {
	return &memorySessionRepo{
		sessions: make(map[string]*domain.Session),
		nextID:   1,
	}
}

func (m *memorySessionRepo) Create(ctx context.Context, s *domain.Session) error {
	s.ID = m.nextID
	m.nextID++
	s.CreatedAt = time.Now().UTC()
	m.sessions[s.TokenHash] = s
	return nil
}

func (m *memorySessionRepo) GetActiveByTokenHash(ctx context.Context, tokenHash string, now time.Time) (*domain.Session, error) {
	s, ok := m.sessions[tokenHash]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	if s.RevokedAt != nil || now.After(s.ExpiresAt) {
		return nil, domain.ErrSessionNotFound
	}
	return s, nil
}

func (m *memorySessionRepo) RevokeByTokenHash(ctx context.Context, tokenHash string, revokeTime time.Time) error {
	s, ok := m.sessions[tokenHash]
	if !ok {
		return domain.ErrSessionNotFound
	}
	s.RevokedAt = &revokeTime
	return nil
}

func (m *memorySessionRepo) RevokeAllForAdmin(ctx context.Context, adminID int64, revokeTime time.Time) error {
	for _, s := range m.sessions {
		if s.AdminID == adminID && s.RevokedAt == nil {
			s.RevokedAt = &revokeTime
		}
	}
	return nil
}

type memoryAuditRecorder struct {
	logs []auditDomain.Log
}

func (m *memoryAuditRecorder) Record(ctx context.Context, entry auditDomain.Log) error {
	m.logs = append(m.logs, entry)
	return nil
}

func TestAuthServiceLifecycle(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	params := authn.FastArgon2idParamsForTesting()

	adminRepo := newMemoryAdminRepo()
	sessionRepo := newMemorySessionRepo()
	auditRepo := &memoryAuditRecorder{}

	svc := application.NewAuthService(adminRepo, sessionRepo, auditRepo, params, 1*time.Hour, logger)

	// 1. Bootstrap Admin
	admin, err := svc.BootstrapAdmin(ctx, "Admin@PidieKab.id ", "Super Admin", "SecureAdminPass123!")
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}

	if admin.Email != "admin@pidiekab.id" {
		t.Errorf("expected normalized email admin@pidiekab.id, got %q", admin.Email)
	}

	// 2. Login with wrong password
	ip := "127.0.0.1"
	_, err = svc.Login(ctx, "admin@pidiekab.id", "WrongPass123", &ip)
	if err != domain.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials for bad password, got %v", err)
	}

	// 3. Login with nonexistent email
	_, err = svc.Login(ctx, "nonexistent@pidiekab.id", "SecureAdminPass123!", &ip)
	if err != domain.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials for unknown user, got %v", err)
	}

	// 4. Successful login
	loginRes, err := svc.Login(ctx, "admin@pidiekab.id", "SecureAdminPass123!", &ip)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if loginRes.RawToken == "" {
		t.Error("expected non-empty raw token")
	}
	if loginRes.Admin.LastLoginAt == nil {
		t.Error("expected last login to be updated")
	}

	// 5. Authenticate session with raw token
	authAdmin, sess, err := svc.AuthenticateSession(ctx, loginRes.RawToken)
	if err != nil {
		t.Fatalf("failed to authenticate valid session: %v", err)
	}
	if authAdmin.ID != admin.ID {
		t.Errorf("expected admin ID %d, got %d", admin.ID, authAdmin.ID)
	}
	if sess.TokenHash != authn.HashToken(loginRes.RawToken) {
		t.Error("token hash mismatch on authenticated session")
	}

	// 6. Change password
	err = svc.ChangePassword(ctx, admin.ID, "SecureAdminPass123!", "NewStrongPassword2026!", &ip)
	if err != nil {
		t.Fatalf("password change failed: %v", err)
	}

	// Verify old session revoked after password change
	_, _, err = svc.AuthenticateSession(ctx, loginRes.RawToken)
	if err == nil {
		t.Error("expected session to be invalid after password change, got nil error")
	}

	// 7. Login with new password
	newLogin, err := svc.Login(ctx, "admin@pidiekab.id", "NewStrongPassword2026!", &ip)
	if err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}

	// 8. Logout
	err = svc.Logout(ctx, newLogin.RawToken, &ip)
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	// Session should no longer be active
	_, _, err = svc.AuthenticateSession(ctx, newLogin.RawToken)
	if err == nil {
		t.Error("expected session to be revoked after logout, got nil error")
	}

	// 9. Reset password by operator
	err = svc.ResetPasswordByOperator(ctx, "admin@pidiekab.id", "OperatorResetPass999!")
	if err != nil {
		t.Fatalf("operator password reset failed: %v", err)
	}

	// Login with operator reset password
	resetLogin, err := svc.Login(ctx, "admin@pidiekab.id", "OperatorResetPass999!", &ip)
	if err != nil {
		t.Fatalf("login with operator reset password failed: %v", err)
	}
	if resetLogin.RawToken == "" {
		t.Error("expected non-empty token after login with reset password")
	}

	// Verify audit logs were produced
	if len(auditRepo.logs) == 0 {
		t.Error("expected audit logs to be recorded")
	}
}
