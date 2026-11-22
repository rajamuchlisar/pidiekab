package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrAdminNotFound              = errors.New("admin account not found")
	ErrInactiveAdmin              = errors.New("admin account is inactive")
	ErrInvalidCredentials         = errors.New("invalid email or password")
	ErrCannotDeactivateLastAdmin  = errors.New("cannot deactivate the last active administrator")
	ErrCannotChangeLastSuperAdmin = errors.New("cannot change role of the last active super administrator")
	ErrSelfDeactivation           = errors.New("cannot deactivate own administrator account")
	ErrEmailAlreadyExists         = errors.New("email address is already in use")
	ErrSessionNotFound            = errors.New("session not found")
	ErrSessionExpired             = errors.New("session has expired")
	ErrSessionRevoked             = errors.New("session has been revoked")
	ErrInvalidPassword            = errors.New("password does not meet security requirements")
	ErrEmptyDisplayName           = errors.New("display name cannot be empty")
	ErrInvalidRole                = errors.New("invalid administrator role")
)

// Role constants for Role-Based Access Control (RBAC).
const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleEditor     = "editor"
	RoleAuthor     = "author"
)

// ModuleInfo represents one of the real, existing menus available in PidieKab Admin CMS.
type ModuleInfo struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
}

// AvailableModules contains the exact 12 menus currently available in the PidieKab Admin CMS.
var AvailableModules = []ModuleInfo{
	{
		Key:         "dashboard",
		Name:        "Dashboard",
		Path:        "/dashboard",
		Icon:        "chart-bar",
		Description: "Melihat statistik pengunjung, total artikel, dan ringkasan operasional",
	},
	{
		Key:         "contents",
		Name:        "Kelola Konten",
		Path:        "/contents",
		Icon:        "document-text",
		Description: "Membuat, mengedit, dan mempublikasikan artikel berita, budaya, dan potensi daerah",
	},
	{
		Key:         "categories",
		Name:        "Kategori",
		Path:        "/categories",
		Icon:        "tag",
		Description: "Mengelola taksonomi kategori konten (nama, slug bilingual, status aktif)",
	},
	{
		Key:         "media",
		Name:        "Media Library",
		Path:        "/media",
		Icon:        "photo",
		Description: "Mengunggah dan mengelola berkas gambar, foto ilustrasi, dan dokumen media",
	},
	{
		Key:         "curation",
		Name:        "Kurasi Beranda",
		Path:        "/curation",
		Icon:        "sparkles",
		Description: "Mengatur penempatan berita utama (Lead Story) dan konten unggulan di beranda publik",
	},
	{
		Key:         "static-pages",
		Name:        "Halaman Statis",
		Path:        "/static-pages",
		Icon:        "document-duplicate",
		Description: "Mengelola konten halaman Profil Daerah, Kontak, Kebijakan Privasi, dan Syarat Ketentuan",
	},
	{
		Key:         "menus",
		Name:        "Menu Navigasi",
		Path:        "/menus",
		Icon:        "bars-3",
		Description: "Mengatur susunan dan tautan navigasi header portal publik",
	},
	{
		Key:         "unduhan",
		Name:        "Unduhan Dokumen",
		Path:        "/unduhan",
		Icon:        "arrow-down-tray",
		Description: "Mengunggah dan mengelola berkas panduan publik, peta wisata, dan dokumen unduhan",
	},
	{
		Key:         "settings",
		Name:        "Pengaturan Web",
		Path:        "/settings",
		Icon:        "cog-6-tooth",
		Description: "Mengatur identitas website, meta SEO, kontak resmi, dan konfigurasi portal",
	},
	{
		Key:         "redirects",
		Name:        "URL Redirects",
		Path:        "/redirects",
		Icon:        "arrows-right-left",
		Description: "Mengelola pengalihan URL (301/302) untuk pencegahan tautan rusak",
	},
	{
		Key:         "admins",
		Name:        "Administrator",
		Path:        "/admins",
		Icon:        "users",
		Description: "Mengelola akun pengguna administrator dan konfigurasi hak akses menu",
	},
	{
		Key:         "audit-logs",
		Name:        "Audit Trail",
		Path:        "/audit-logs",
		Icon:        "shield-check",
		Description: "Melihat riwayat log aktivitas keamanan dan jejak perubahan sistem",
	},
}

// AllModuleKeys returns an array containing all 12 valid module keys.
func AllModuleKeys() []string {
	keys := make([]string, len(AvailableModules))
	for i, m := range AvailableModules {
		keys[i] = m.Key
	}
	return keys
}

// IsValidModuleKey checks if a key corresponds to one of the 12 existing modules.
func IsValidModuleKey(key string) bool {
	for _, m := range AvailableModules {
		if m.Key == key {
			return true
		}
	}
	return false
}

// FilterValidModules sanitizes an array of module keys, keeping only recognized ones.
func FilterValidModules(modules []string) []string {
	valid := make([]string, 0, len(modules))
	seen := make(map[string]bool)
	for _, m := range modules {
		clean := strings.TrimSpace(m)
		if clean != "" && IsValidModuleKey(clean) && !seen[clean] {
			seen[clean] = true
			valid = append(valid, clean)
		}
	}
	return valid
}

// RoleDefaultModules returns the default recommended checklist of modules for a role.
func RoleDefaultModules(role string) []string {
	switch role {
	case RoleSuperAdmin:
		return AllModuleKeys()
	case RoleAdmin:
		return []string{
			"dashboard", "contents", "categories", "media", "curation",
			"static-pages", "menus", "unduhan", "settings", "redirects", "audit-logs",
		}
	case RoleEditor:
		return []string{
			"dashboard", "contents", "categories", "media", "curation", "unduhan",
		}
	case RoleAuthor:
		return []string{
			"dashboard", "contents", "media",
		}
	default:
		return []string{"dashboard", "contents", "media"}
	}
}

// RoleDefinition provides UI and system metadata for a role.
type RoleDefinition struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	BadgeColor  string   `json:"badge_color"`
	Permissions []string `json:"permissions"`
}

// GetAllRoles returns the full list of defined roles.
func GetAllRoles() []RoleDefinition {
	return []RoleDefinition{
		{
			Key:         RoleSuperAdmin,
			Name:        "Super Administrator",
			Description: "Akses penuh tanpa batasan ke seluruh menu sistem termasuk kelola akun admin dan konfigurasi inti.",
			BadgeColor:  "purple",
			Permissions: AllModuleKeys(),
		},
		{
			Key:         RoleAdmin,
			Name:        "Administrator",
			Description: "Mengelola seluruh operasional konten, kategori, media library, kurasi, halaman statis, menu, dan unduhan.",
			BadgeColor:  "green",
			Permissions: RoleDefaultModules(RoleAdmin),
		},
		{
			Key:         RoleEditor,
			Name:        "Editor Konten",
			Description: "Meninjau, mengedit, dan mempublikasikan konten artikel, kurasi beranda, serta mengelola berkas media.",
			BadgeColor:  "blue",
			Permissions: RoleDefaultModules(RoleEditor),
		},
		{
			Key:         RoleAuthor,
			Name:        "Penulis Konten",
			Description: "Membuat dan mengedit draf artikel sendiri serta mengunggah media pendukung liputan.",
			BadgeColor:  "amber",
			Permissions: RoleDefaultModules(RoleAuthor),
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

// Admin represents an internal administrator user.
type Admin struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	DisplayName  string     `json:"display_name"`
	Role         string     `json:"role"`
	Permissions  []string   `json:"permissions"`
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

// CanAccessModule checks whether the admin is explicitly authorized for a menu module.
// SuperAdmin always has access to all modules.
func (a *Admin) CanAccessModule(moduleKey string) bool {
	if a == nil {
		return false
	}
	if a.Role == RoleSuperAdmin {
		return true
	}
	for _, p := range a.Permissions {
		if p == moduleKey {
			return true
		}
	}
	return false
}

// HasPermission checks whether the admin is authorized for a specific permission or module.
func (a *Admin) HasPermission(perm string) bool {
	return a.CanAccessModule(perm)
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
