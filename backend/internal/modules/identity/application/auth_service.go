package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
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
	UpdateProfile(ctx context.Context, id int64, displayName string) error
	UpdateRole(ctx context.Context, id int64, role string) error
	UpdatePermissions(ctx context.Context, id int64, permissions []string) error
	UpdateLastLogin(ctx context.Context, id int64, loginTime time.Time) error
	CountActive(ctx context.Context) (int, error)
	CountActiveSuperAdmins(ctx context.Context) (int, error)
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

// LoginResult contains the session token and authenticated administrator details.
type LoginResult struct {
	RawToken  string
	ExpiresAt time.Time
	Admin     *domain.Admin
}

// Login authenticates an administrator with email and password, establishing a new session.
func (s *AuthService) Login(ctx context.Context, email, password string, ipAddress *string) (*LoginResult, error) {
	normalizedEmail := domain.NormalizeEmail(email)
	if normalizedEmail == "" || password == "" {
		s.recordAudit(ctx, nil, "admin.login.failed", "admin", nil, map[string]any{
			"email":  email,
			"reason": "empty_credentials",
		}, ipAddress)
		return nil, domain.ErrInvalidCredentials
	}

	admin, err := s.admins.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, domain.ErrAdminNotFound) {
			s.recordAudit(ctx, nil, "admin.login.failed", "admin", nil, map[string]any{
				"email":  normalizedEmail,
				"reason": "not_found",
			}, ipAddress)
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to query admin: %w", err)
	}

	if !admin.IsActive {
		s.recordAudit(ctx, &admin.ID, "admin.login.failed", "admin", &admin.ID, map[string]any{
			"email":  normalizedEmail,
			"reason": "inactive_account",
		}, ipAddress)
		return nil, domain.ErrInactiveAdmin
	}

	match, err := authn.VerifyPassword(password, admin.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}
	if !match {
		s.recordAudit(ctx, &admin.ID, "admin.login.failed", "admin", &admin.ID, map[string]any{
			"email":  normalizedEmail,
			"reason": "invalid_password",
		}, ipAddress)
		return nil, domain.ErrInvalidCredentials
	}

	// Generate high-entropy 256-bit session token
	rawToken, err := authn.GenerateRandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	tokenHash := authn.HashToken(rawToken)
	now := time.Now().UTC()
	expiresAt := now.Add(s.sessionLifetime)

	session := &domain.Session{
		AdminID:   admin.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	_ = s.admins.UpdateLastLogin(ctx, admin.ID, now)
	s.recordAudit(ctx, &admin.ID, "admin.login.success", "admin", &admin.ID, nil, ipAddress)

	return &LoginResult{
		RawToken:  rawToken,
		ExpiresAt: expiresAt,
		Admin:     admin,
	}, nil
}

// AuthenticateSession validates a raw bearer/cookie session token and returns the administrator.
func (s *AuthService) AuthenticateSession(ctx context.Context, rawToken string) (*domain.Admin, *domain.Session, error) {
	if rawToken == "" {
		return nil, nil, domain.ErrSessionNotFound
	}

	tokenHash := authn.HashToken(rawToken)
	now := time.Now().UTC()

	session, err := s.sessions.GetActiveByTokenHash(ctx, tokenHash, now)
	if err != nil {
		return nil, nil, err
	}

	admin, err := s.admins.GetByID(ctx, session.AdminID)
	if err != nil {
		return nil, nil, err
	}

	if !admin.IsActive {
		return nil, nil, domain.ErrInactiveAdmin
	}

	return admin, session, nil
}

// Logout revokes the session corresponding to the provided raw token.
func (s *AuthService) Logout(ctx context.Context, rawToken string, ipAddress *string) error {
	if rawToken == "" {
		return nil
	}

	tokenHash := authn.HashToken(rawToken)
	now := time.Now().UTC()

	session, err := s.sessions.GetActiveByTokenHash(ctx, tokenHash, now)
	if err != nil {
		return nil
	}

	if err := s.sessions.RevokeByTokenHash(ctx, tokenHash, now); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	s.recordAudit(ctx, &session.AdminID, "admin.logout", "session", &session.ID, nil, ipAddress)
	return nil
}

// ChangePassword updates an admin's password after validating current password.
func (s *AuthService) ChangePassword(ctx context.Context, adminID int64, currentPassword, newPassword string, ipAddress *string) error {
	if len(newPassword) < 8 {
		return domain.ErrInvalidPassword
	}

	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}

	match, err := authn.VerifyPassword(currentPassword, admin.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to verify current password: %w", err)
	}
	if !match {
		s.recordAudit(ctx, &adminID, "admin.password.change.failed", "admin", &adminID, map[string]any{
			"reason": "invalid_current_password",
		}, ipAddress)
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

// UpdateProfile updates the display name of an administrator.
func (s *AuthService) UpdateProfile(ctx context.Context, adminID int64, displayName string, ipAddress *string) (*domain.Admin, error) {
	cleanName := strings.TrimSpace(displayName)
	if cleanName == "" {
		return nil, domain.ErrEmptyDisplayName
	}

	if err := s.admins.UpdateProfile(ctx, adminID, cleanName); err != nil {
		return nil, err
	}

	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, &adminID, "UPDATE_PROFILE", "admin", &adminID, map[string]any{
		"display_name": cleanName,
	}, ipAddress)

	return admin, nil
}

// BootstrapAdmin creates an initial administrator account with super_admin role.
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
		Role:         domain.RoleSuperAdmin,
		Permissions:  domain.AllModuleKeys(),
		IsActive:     true,
	}

	if err := s.admins.Create(ctx, admin); err != nil {
		return nil, err
	}

	s.recordAudit(ctx, &admin.ID, "admin.bootstrap", "admin", &admin.ID, nil, nil)
	return admin, nil
}

// ResetPasswordByOperator resets an administrator's password directly from CLI/operator,
// immediately revoking all active sessions for security.
func (s *AuthService) ResetPasswordByOperator(ctx context.Context, email, newPassword string) error {
	normalizedEmail := domain.NormalizeEmail(email)
	if normalizedEmail == "" {
		return errors.New("email cannot be empty")
	}
	if len(newPassword) < 8 {
		return domain.ErrInvalidPassword
	}

	admin, err := s.admins.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		return fmt.Errorf("failed to find admin by email: %w", err)
	}
	if admin == nil {
		return domain.ErrAdminNotFound
	}

	hash, err := authn.HashPassword(newPassword, s.argonParams)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.admins.UpdatePassword(ctx, admin.ID, hash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Revoke all active sessions
	_ = s.sessions.RevokeAllForAdmin(ctx, admin.ID, time.Now())

	s.recordAudit(ctx, nil, "admin.operator_password_reset", "admin", &admin.ID, map[string]any{
		"email": admin.Email,
	}, nil)

	return nil
}

// ListAdmins returns all administrators.
func (s *AuthService) ListAdmins(ctx context.Context) ([]*domain.Admin, error) {
	return s.admins.List(ctx)
}

// CreateAdmin creates a new administrator with a designated role and checklist of authorized menu modules.
func (s *AuthService) CreateAdmin(ctx context.Context, creatorID int64, email, displayName, role string, permissions []string, password string, ipAddress *string) (*domain.Admin, error) {
	normalizedEmail := domain.NormalizeEmail(email)
	if normalizedEmail == "" {
		return nil, errors.New("email cannot be empty")
	}
	if strings.TrimSpace(displayName) == "" {
		return nil, domain.ErrEmptyDisplayName
	}
	if len(password) < 8 {
		return nil, domain.ErrInvalidPassword
	}

	if role == "" || !domain.IsValidRole(role) {
		role = domain.RoleEditor
	}

	if len(permissions) == 0 {
		permissions = domain.RoleDefaultModules(role)
	}
	permissions = domain.FilterValidModules(permissions)

	hash, err := authn.HashPassword(password, s.argonParams)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	admin := &domain.Admin{
		Email:        normalizedEmail,
		PasswordHash: hash,
		DisplayName:  strings.TrimSpace(displayName),
		Role:         role,
		Permissions:  permissions,
		IsActive:     true,
	}

	if err := s.admins.Create(ctx, admin); err != nil {
		return nil, err
	}

	s.recordAudit(ctx, &creatorID, "admin.create", "admin", &admin.ID, map[string]any{
		"email":       admin.Email,
		"role":        admin.Role,
		"permissions": admin.Permissions,
	}, ipAddress)

	return admin, nil
}

// UpdateAdminPermissions updates the authorized checklist of modules for an administrator.
func (s *AuthService) UpdateAdminPermissions(ctx context.Context, actorID, targetID int64, permissions []string, ipAddress *string) error {
	target, err := s.admins.GetByID(ctx, targetID)
	if err != nil {
		return err
	}

	validPerms := domain.FilterValidModules(permissions)
	if len(validPerms) == 0 {
		validPerms = []string{"dashboard"}
	}

	if err := s.admins.UpdatePermissions(ctx, targetID, validPerms); err != nil {
		return err
	}

	s.recordAudit(ctx, &actorID, "admin.permissions.update", "admin", &targetID, map[string]any{
		"old_permissions": target.Permissions,
		"new_permissions": validPerms,
	}, ipAddress)

	return nil
}

// UpdateAdminRole modifies the role of an administrator.
func (s *AuthService) UpdateAdminRole(ctx context.Context, actorID, targetID int64, newRole string, ipAddress *string) error {
	if !domain.IsValidRole(newRole) {
		return domain.ErrInvalidRole
	}

	target, err := s.admins.GetByID(ctx, targetID)
	if err != nil {
		return err
	}

	// Protect against demoting the last active super_admin
	if target.Role == domain.RoleSuperAdmin && newRole != domain.RoleSuperAdmin {
		count, err := s.admins.CountActiveSuperAdmins(ctx)
		if err != nil {
			return err
		}
		if count <= 1 {
			return domain.ErrCannotChangeLastSuperAdmin
		}
	}

	if err := s.admins.UpdateRole(ctx, targetID, newRole); err != nil {
		return err
	}

	s.recordAudit(ctx, &actorID, "admin.role.update", "admin", &targetID, map[string]any{
		"old_role": target.Role,
		"new_role": newRole,
	}, ipAddress)

	return nil
}

// ToggleAdminStatus activates or deactivates an administrator with self-lockout and last-active protections.
func (s *AuthService) ToggleAdminStatus(ctx context.Context, currentAdminID, targetAdminID int64, isActive bool, ipAddress *string) error {
	if !isActive {
		if currentAdminID == targetAdminID {
			return domain.ErrSelfDeactivation
		}

		activeCount, err := s.admins.CountActive(ctx)
		if err != nil {
			return err
		}
		if activeCount <= 1 {
			return domain.ErrCannotDeactivateLastAdmin
		}

		target, err := s.admins.GetByID(ctx, targetAdminID)
		if err == nil && target.Role == domain.RoleSuperAdmin {
			superCount, err := s.admins.CountActiveSuperAdmins(ctx)
			if err != nil {
				return err
			}
			if superCount <= 1 {
				return domain.ErrCannotDeactivateLastAdmin
			}
		}
	}

	if err := s.admins.SetIsActive(ctx, targetAdminID, isActive); err != nil {
		return err
	}

	// If deactivating, immediately revoke all active sessions of target admin
	if !isActive {
		now := time.Now().UTC()
		_ = s.sessions.RevokeAllForAdmin(ctx, targetAdminID, now)
	}

	action := "admin.activate"
	if !isActive {
		action = "admin.deactivate"
	}
	s.recordAudit(ctx, &currentAdminID, action, "admin", &targetAdminID, map[string]any{
		"is_active": isActive,
	}, ipAddress)

	return nil
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
