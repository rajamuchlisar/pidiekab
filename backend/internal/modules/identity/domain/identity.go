package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrAdminNotFound             = errors.New("admin account not found")
	ErrInactiveAdmin             = errors.New("admin account is inactive")
	ErrInvalidCredentials        = errors.New("invalid email or password")
	ErrCannotDeactivateLastAdmin = errors.New("cannot deactivate the last active administrator")
	ErrSelfDeactivation          = errors.New("cannot deactivate own administrator account")
	ErrEmailAlreadyExists        = errors.New("email address is already in use")
	ErrSessionNotFound           = errors.New("session not found")
	ErrSessionExpired            = errors.New("session has expired")
	ErrSessionRevoked            = errors.New("session has been revoked")
	ErrInvalidPassword           = errors.New("password does not meet security requirements")
	ErrEmptyDisplayName          = errors.New("display name cannot be empty")
)

// Admin represents an internal administrator user.
type Admin struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	DisplayName  string     `json:"display_name"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// NormalizeEmail returns an email with whitespace stripped and converted to lowercase.
// Matches database constraint: email = LOWER(BTRIM(email)).
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Session represents an authenticated administrator session in admin_sessions.
type Session struct {
	ID        int64      `json:"id"`
	AdminID   int64      `json:"admin_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// IsActive returns true if the session is not revoked and not past its expiration date.
func (s *Session) IsActive(now time.Time) bool {
	if s.RevokedAt != nil {
		return false
	}
	return now.Before(s.ExpiresAt)
}
