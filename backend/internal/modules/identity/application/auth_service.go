package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	auditDomain "pidiekab/internal/modules/audit/domain"
	"pidiekab/internal/modules/identity/domain"
	"pidiekab/internal/platform/authn"
)

// AdminRepo defines the data access contract required for admin identity.
type AdminRepo interface {
	Create(ctx context.Context, a *domain.Admin) error
	GetByID(ctx context.Context, id int64) (*domain.Admin, error)
	GetByEmail(ctx context.Context, email string) (*domain.Admin, error)
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	UpdateLastLogin(ctx context.Context, id int64, loginTime time.Time) error
	CountActive(ctx context.Context) (int, error)
	SetIsActive(ctx context.Context, id int64, isActive bool) error
	List(ctx context.Context) ([]*domain.Admin, error)
}

// SessionRepo defines the data access contract required for session management.
type SessionRepo interface {
	Create(ctx context.Context, s *domain.Session) error
	GetActiveByTokenHash(ctx context.Context, tokenHash string, now time.Time) (*domain.Session, error)
	RevokeByTokenHash(ctx context.Context, tokenHash string, revokeTime time.Time) error
	RevokeAllForAdmin(ctx context.Context, adminID int64, revokeTime time.Time) error
}

// AuditRecorder defines the contract for recording security events.
type AuditRecorder interface {
	Record(ctx context.Context, entry auditDomain.Log) error
}

// AuthService orchestrates administrator authentication, sessions, and credentials.
type AuthService struct {
	admins          AdminRepo
	sessions        SessionRepo
	audit           AuditRecorder
	argonParams     authn.Argon2idParams
	sessionLifetime time.Duration
	logger          *slog.Logger
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	admins AdminRepo,
	sessions SessionRepo,
	audit AuditRecorder,
	argonParams authn.Argon2idParams,
	sessionLifetime time.Duration,
	logger *slog.Logger,
) *AuthService {
	if sessionLifetime <= 0 {
		sessionLifetime = 12 * time.Hour // default absolute lifetime
	}
	return &AuthService{
		admins:          admins,
		sessions:        sessions,
		audit:           audit,
		argonParams:     argonParams,
		sessionLifetime: sessionLifetime,
		logger:          logger,
	}
}

// LoginResult contains the generated raw token and authenticated admin identity.
type LoginResult struct {
	RawToken  string
	ExpiresAt time.Time
	Admin     *domain.Admin
}

// Login authenticates an admin by email and password, creating a secure session.
func (s *AuthService) Login(ctx context.Context, email, password string, ipAddress *string) (*LoginResult, error) {
	normalizedEmail := domain.NormalizeEmail(email)
	if normalizedEmail == "" || password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	admin, err := s.admins.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, domain.ErrAdminNotFound) {
			s.recordAudit(ctx, nil, "admin.login.failure", "admin", nil, map[string]any{"reason": "not_found", "email": normalizedEmail}, ipAddress)
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to lookup admin: %w", err)
	}

	if !admin.IsActive {
		s.recordAudit(ctx, &admin.ID, "admin.login.failure", "admin", &admin.ID, map[string]any{"reason": "inactive_account"}, ipAddress)
		return nil, domain.ErrInvalidCredentials
	}

	match, err := authn.VerifyPassword(password, admin.PasswordHash)
	if err != nil {
		s.logger.Error("password verification error", "admin_id", admin.ID, "error", err)
		return nil, domain.ErrInvalidCredentials
	}
	if !match {
		s.recordAudit(ctx, &admin.ID, "admin.login.failure", "admin", &admin.ID, map[string]any{"reason": "password_mismatch"}, ipAddress)
		return nil, domain.ErrInvalidCredentials
	}

	// Password matched; generate secure opaque token
	rawToken, err := authn.GenerateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	tokenHash := authn.HashToken(rawToken)
	now := time.Now().UTC()
	expiresAt := now.Add(s.sessionLifetime)

	sess := &domain.Session{
		AdminID:   admin.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	if err := s.sessions.Create(ctx, sess); err != nil {
		return nil, fmt.Errorf("failed to persist admin session: %w", err)
	}

	if err := s.admins.UpdateLastLogin(ctx, admin.ID, now); err != nil {
		s.logger.Warn("failed to update admin last login", "admin_id", admin.ID, "error", err)
	}

	admin.LastLoginAt = &now
	s.recordAudit(ctx, &admin.ID, "admin.login.success", "admin", &admin.ID, nil, ipAddress)

	return &LoginResult{
		RawToken:  rawToken,
		ExpiresAt: expiresAt,
		Admin:     admin,
	}, nil
}

// Logout revokes the session associated with the provided raw token.
func (s *AuthService) Logout(ctx context.Context, rawToken string, ipAddress *string) error {
	if rawToken == "" {
		return domain.ErrSessionNotFound
	}

	tokenHash := authn.HashToken(rawToken)
	now := time.Now().UTC()

	sess, err := s.sessions.GetActiveByTokenHash(ctx, tokenHash, now)
	if err != nil {
		return domain.ErrSessionNotFound
	}

	if err := s.sessions.RevokeByTokenHash(ctx, tokenHash, now); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	s.recordAudit(ctx, &sess.AdminID, "admin.logout", "admin_session", &sess.ID, nil, ipAddress)
	return nil
}

// AuthenticateSession validates a raw token and returns the corresponding active admin.
func (s *AuthService) AuthenticateSession(ctx context.Context, rawToken string) (*domain.Admin, *domain.Session, error) {
	if rawToken == "" {
		return nil, nil, domain.ErrSessionNotFound
	}

	tokenHash := authn.HashToken(rawToken)
	now := time.Now().UTC()

	sess, err := s.sessions.GetActiveByTokenHash(ctx, tokenHash, now)
	if err != nil {
		return nil, nil, err
	}

	admin, err := s.admins.GetByID(ctx, sess.AdminID)
	if err != nil {
		return nil, nil, err
	}

	if !admin.IsActive {
		return nil, nil, domain.ErrInactiveAdmin
	}

	return admin, sess, nil
}

// ChangePassword updates an admin's password and revokes all active sessions for security.
func (s *AuthService) ChangePassword(ctx context.Context, adminID int64, currentPassword, newPassword string, ipAddress *string) error {
	if len(newPassword) < 8 {
		return domain.ErrInvalidPassword
	}

	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}

	if !admin.IsActive {
		return domain.ErrInactiveAdmin
	}

	match, err := authn.VerifyPassword(currentPassword, admin.PasswordHash)
	if err != nil || !match {
		return domain.ErrInvalidCredentials
	}

	newHash, err := authn.HashPassword(newPassword, s.argonParams)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.admins.UpdatePassword(ctx, adminID, newHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Revoke all existing sessions for this admin
	now := time.Now().UTC()
	if err := s.sessions.RevokeAllForAdmin(ctx, adminID, now); err != nil {
		s.logger.Warn("failed to revoke sessions on password change", "admin_id", adminID, "error", err)
	}

	s.recordAudit(ctx, &adminID, "admin.password.change", "admin", &adminID, nil, ipAddress)
	return nil
}

// BootstrapAdmin creates an initial administrator account.
func (s *AuthService) BootstrapAdmin(ctx context.Context, email, displayName, password string) (*domain.Admin, error) {
	normalizedEmail := domain.NormalizeEmail(email)
	if normalizedEmail == "" {
		return nil, errors.New("email cannot be empty")
	}
	if displayName == "" {
		return nil, domain.ErrEmptyDisplayName
	}
	if len(password) < 8 {
		return nil, domain.ErrInvalidPassword
	}

	hash, err := authn.HashPassword(password, s.argonParams)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	admin := &domain.Admin{
		Email:        normalizedEmail,
		PasswordHash: hash,
		DisplayName:  displayName,
		IsActive:     true,
	}

	if err := s.admins.Create(ctx, admin); err != nil {
		return nil, err
	}

	s.recordAudit(ctx, &admin.ID, "admin.bootstrap", "admin", &admin.ID, nil, nil)
	return admin, nil
}

func (s *AuthService) recordAudit(ctx context.Context, adminID *int64, action, entityType string, entityID *int64, meta map[string]any, ip *string) {
	if s.audit == nil {
		return
	}
	_ = s.audit.Record(ctx, auditDomain.Log{
		AdminID:    adminID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   meta,
		IPAddress:  ip,
		CreatedAt:  time.Now().UTC(),
	})
}
