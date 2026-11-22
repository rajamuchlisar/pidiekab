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
	ErrCannotChangeLastSuperAdmin = errors.New("cannot change role of the last active super administrator")
	ErrSelfDeactivation          = errors.New("cannot deactivate own administrator account")
	ErrEmailAlreadyExists        = errors.New("email address is already in use")
	ErrSessionNotFound           = errors.New("session not found")
	ErrSessionExpired            = errors.New("session has expired")
	ErrSessionRevoked            = errors.New("session has been revoked")
	ErrInvalidPassword           = errors.New("password does not meet security requirements")
	ErrEmptyDisplayName          = errors.New("display name cannot be empty")
	ErrInvalidRole               = errors.New("invalid administrator role")
)

// Role constants for Role-Based Access Control (RBAC).
const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleEditor     = "editor"
	RoleAuthor     = "author"
)

// RoleDefinition provides UI and system metadata for a role.
type RoleDefinition struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	BadgeColor  string   `json:"badge_color"`
	Permissions []string `json:"permissions"`
}

// GetAllRoles returns the full list of defined roles and their permissions.
func GetAllRoles() []RoleDefinition {
	return []RoleDefinition{
		{
			Key:         RoleSuperAdmin,
			Name:        "Super Administrator",
			Description: "Akses penuh tanpa batasan ke seluruh sistem termasuk kelola akun admin, role, dan konfigurasi inti.",
			BadgeColor:  "purple",
			Permissions: RolePermissions(RoleSuperAdmin),
		},
		{
			Key:         RoleAdmin,
			Name:        "Administrator",
			Description: "Mengelola seluruh operasional konten, kategori, media library, kurasi, halaman statis, menu, dan unduhan.",
			BadgeColor:  "green",
			Permissions: RolePermissions(RoleAdmin),
		},
		{
			Key:         RoleEditor,
			Name:        "Editor Konten",
			Description: "Meninjau, mengedit, dan mempublikasikan konten artikel, kurasi beranda, serta mengelola berkas media.",
			BadgeColor:  "blue",
			Permissions: RolePermissions(RoleEditor),
		},
		{
			Key:         RoleAuthor,
			Name:        "Penulis Konten",
			Description: "Membuat dan mengedit draf artikel sendiri serta mengunggah media pendukung liputan.",
			BadgeColor:  "amber",
			Permissions: RolePermissions(RoleAuthor),
		},
	}
}

// IsValidRole validates if a given string matches an allowed role.
func IsValidRole(role string) bool {
	switch role {
	case RoleSuperAdmin, RoleAdmin, RoleEditor, RoleAuthor:
		return true
	default:
		return false
	}
}

// RolePermissions returns the list of permissions granted to a role.
func RolePermissions(role string) []string {
	switch role {
	case RoleSuperAdmin:
		return []string{
			"admins:manage", "settings:manage", "audit:view",
			"content:read", "content:write", "content:publish", "content:delete",
			"category:manage", "media:manage", "curation:manage",
			"pages:manage", "menus:manage", "downloads:manage",
			"redirects:manage", "analytics:view",
		}
	case RoleAdmin:
		return []string{
			"settings:manage", "audit:view",
			"content:read", "content:write", "content:publish", "content:delete",
			"category:manage", "media:manage", "curation:manage",
			"pages:manage", "menus:manage", "downloads:manage",
			"redirects:manage", "analytics:view",
		}
	case RoleEditor:
		return []string{
			"content:read", "content:write", "content:publish",
			"category:manage", "media:manage", "curation:manage",
			"downloads:manage", "analytics:view",
		}
	case RoleAuthor:
		return []string{
			"content:read", "content:write",
			"media:manage", "analytics:view",
		}
	default:
		return nil
	}
}

// Admin represents an internal administrator user.
type Admin struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	DisplayName  string     `json:"display_name"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// HasRole checks if the admin possesses any of the required roles.
// SuperAdmin always satisfies any role requirement.
func (a *Admin) HasRole(roles ...string) bool {
	if a == nil {
		return false
	}
	if a.Role == RoleSuperAdmin {
		return true
	}
	for _, r := range roles {
		if a.Role == r {
			return true
		}
	}
	return false
}

// HasPermission checks if the admin has a specific permission.
func (a *Admin) HasPermission(perm string) bool {
	if a == nil {
		return false
	}
	if a.Role == RoleSuperAdmin {
		return true
	}
	for _, p := range RolePermissions(a.Role) {
		if p == perm {
			return true
		}
	}
	return false
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
