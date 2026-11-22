package http

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pidiekab/internal/modules/identity/application"
	"pidiekab/internal/modules/identity/domain"
	"pidiekab/internal/platform/authn"
	"pidiekab/internal/platform/httpserver/middleware"
)

// Handler manages HTTP transport for administrator authentication and identity.
type Handler struct {
	authService  *application.AuthService
	isProduction bool
}

// NewHandler creates a new identity transport Handler.
func NewHandler(authService *application.AuthService, isProduction bool) *Handler {
	return &Handler{
		authService:  authService,
		isProduction: isProduction,
	}
}

// RegisterRoutes registers identity routes onto the provided ServeMux using a default in-memory rate limiter.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	h.RegisterRoutesWithLimiter(mux, middleware.NewMemoryRateLimiter(10, 1*time.Minute))
}

// RegisterRoutesWithLimiter registers identity routes with a custom rate limiter (or nil to disable).
func (h *Handler) RegisterRoutesWithLimiter(mux *http.ServeMux, rateLimiter *middleware.MemoryRateLimiter) {
	// Public auth route with rate limiting
	var loginHandler http.Handler = http.HandlerFunc(h.handleLogin)
	if rateLimiter != nil {
		loginHandler = rateLimiter.LimitHandler(loginHandler)
	}
	mux.Handle("POST /api/v1/admin/auth/login", loginHandler)

	// Protected auth routes (require active session)
	authMiddleware := middleware.RequireAdminSession(h.authService)
	csrfMiddleware := middleware.CSRFProtection

	mux.Handle("POST /api/v1/admin/auth/logout", authMiddleware(csrfMiddleware(http.HandlerFunc(h.handleLogout))))
	mux.Handle("GET /api/v1/admin/auth/me", authMiddleware(http.HandlerFunc(h.handleMe)))
	mux.Handle("POST /api/v1/admin/auth/password/change", authMiddleware(csrfMiddleware(http.HandlerFunc(h.handleChangePassword))))

	// RBAC Roles and Available Modules query routes
	mux.Handle("GET /api/v1/admin/roles", authMiddleware(http.HandlerFunc(h.handleListRoles)))
	mux.Handle("GET /api/v1/admin/modules", authMiddleware(http.HandlerFunc(h.handleListModules)))

	// Admin management routes with RBAC protection:
	// Only super_admin can create, update role, update menu permissions, or toggle status of admins
	superAdminOnly := middleware.RequireRoles(domain.RoleSuperAdmin)
	adminOrSuper := middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin)

	mux.Handle("GET /api/v1/admin/admins", authMiddleware(adminOrSuper(http.HandlerFunc(h.handleListAdmins))))
	mux.Handle("POST /api/v1/admin/admins", authMiddleware(superAdminOnly(csrfMiddleware(http.HandlerFunc(h.handleCreateAdmin)))))
	mux.Handle("PUT /api/v1/admin/admins/{id}/role", authMiddleware(superAdminOnly(csrfMiddleware(http.HandlerFunc(h.handleUpdateAdminRole)))))
	mux.Handle("PUT /api/v1/admin/admins/{id}/permissions", authMiddleware(superAdminOnly(csrfMiddleware(http.HandlerFunc(h.handleUpdateAdminPermissions)))))
	mux.Handle("PUT /api/v1/admin/admins/{id}/status", authMiddleware(superAdminOnly(csrfMiddleware(http.HandlerFunc(h.handleToggleAdminStatus)))))
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON request body.", nil)
		return
	}

	ip := extractClientIP(r)
	result, err := h.authService.Login(r.Context(), req.Email, req.Password, &ip)
	if err != nil {
		writeProblem(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password.", nil)
		return
	}

	maxAge := int(time.Until(result.ExpiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.AdminSessionCookieName,
		Value:    result.RawToken,
		Path:     "/",
		Expires:  result.ExpiresAt,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteLaxMode,
	})

	csrfToken, err := authn.GenerateRandomToken(16)
	if err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:     middleware.CSRFCookieName,
			Value:    csrfToken,
			Path:     "/",
			Expires:  result.ExpiresAt,
			MaxAge:   maxAge,
			HttpOnly: false,
			Secure:   h.isProduction,
			SameSite: http.SameSiteLaxMode,
		})
	}

	perms := result.Admin.Permissions
	if result.Admin.Role == domain.RoleSuperAdmin || len(perms) == 0 {
		perms = domain.AllModuleKeys()
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"admin": map[string]any{
			"id":          result.Admin.ID,
			"email":       result.Admin.Email,
			"displayName": result.Admin.DisplayName,
			"role":        result.Admin.Role,
			"permissions": perms,
		},
		"data": map[string]any{
			"id":          strconv.FormatInt(result.Admin.ID, 10),
			"email":       result.Admin.Email,
			"displayName": result.Admin.DisplayName,
			"name":        result.Admin.DisplayName,
			"role":        result.Admin.Role,
			"permissions": perms,
		},
	})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middleware.AdminSessionCookieName)
	if err == nil && cookie.Value != "" {
		ip := extractClientIP(r)
		_ = h.authService.Logout(r.Context(), cookie.Value, &ip)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.AdminSessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.CSRFCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: false,
		Secure:   h.isProduction,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.", nil)
		return
	}

	perms := admin.Permissions
	if admin.Role == domain.RoleSuperAdmin || len(perms) == 0 {
		perms = domain.AllModuleKeys()
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"admin": map[string]any{
			"id":          admin.ID,
			"email":       admin.Email,
			"displayName": admin.DisplayName,
			"role":        admin.Role,
			"permissions": perms,
		},
		"data": map[string]any{
			"id":          strconv.FormatInt(admin.ID, 10),
			"email":       admin.Email,
			"displayName": admin.DisplayName,
			"name":        admin.DisplayName,
			"role":        admin.Role,
			"permissions": perms,
		},
	})
}

func (h *Handler) handleListRoles(w http.ResponseWriter, r *http.Request) {
	roles := domain.GetAllRoles()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"roles": roles,
		"data":  roles,
	})
}

func (h *Handler) handleListModules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"modules": domain.AvailableModules,
		"data":    domain.AvailableModules,
	})
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *Handler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.", nil)
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON request body.", nil)
		return
	}

	if len(req.NewPassword) < 8 {
		writeProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Password must be at least 8 characters.", map[string][]string{
			"new_password": {"Password must be at least 8 characters long."},
		})
		return
	}

	ip := extractClientIP(r)
	err := h.authService.ChangePassword(r.Context(), admin.ID, req.CurrentPassword, req.NewPassword, &ip)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			writeProblem(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Current password does not match.", nil)
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update password.", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleListAdmins(w http.ResponseWriter, r *http.Request) {
	admins, err := h.authService.ListAdmins(r.Context())
	if err != nil {
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list administrators.", nil)
		return
	}

	type adminDTO struct {
		ID          int64      `json:"id"`
		Email       string     `json:"email"`
		DisplayName string     `json:"display_name"`
		Role        string     `json:"role"`
		Permissions []string   `json:"permissions"`
		IsActive    bool       `json:"is_active"`
		LastLoginAt *time.Time `json:"last_login_at,omitempty"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   time.Time  `json:"updated_at"`
	}

	dtos := make([]adminDTO, 0, len(admins))
	for _, a := range admins {
		perms := a.Permissions
		if a.Role == domain.RoleSuperAdmin || len(perms) == 0 {
			perms = domain.AllModuleKeys()
		}
		dtos = append(dtos, adminDTO{
			ID:          a.ID,
			Email:       a.Email,
			DisplayName: a.DisplayName,
			Role:        a.Role,
			Permissions: perms,
			IsActive:    a.IsActive,
			LastLoginAt: a.LastLoginAt,
			CreatedAt:   a.CreatedAt,
			UpdatedAt:   a.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"admins": dtos,
		"data":   dtos,
	})
}

type createAdminRequest struct {
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	Password    string   `json:"password"`
}

func (h *Handler) handleCreateAdmin(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Active admin session required.", nil)
		return
	}

	var req createAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON request body.", nil)
		return
	}

	ip := extractClientIP(r)
	created, err := h.authService.CreateAdmin(r.Context(), admin.ID, req.Email, req.DisplayName, req.Role, req.Permissions, req.Password, &ip)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			writeProblem(w, r, http.StatusConflict, "EMAIL_EXISTS", "Email address is already in use.", nil)
			return
		}
		if errors.Is(err, domain.ErrInvalidPassword) || errors.Is(err, domain.ErrEmptyDisplayName) || errors.Is(err, domain.ErrInvalidRole) {
			writeProblem(w, r, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create administrator.", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": created,
	})
}

type updateRoleRequest struct {
	Role string `json:"role"`
}

func (h *Handler) handleUpdateAdminRole(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Active admin session required.", nil)
		return
	}

	targetID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || targetID <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid administrator ID.", nil)
		return
	}

	var req updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON request body.", nil)
		return
	}

	if !domain.IsValidRole(req.Role) {
		writeProblem(w, r, http.StatusBadRequest, "INVALID_ROLE", "Peran (role) yang dipilih tidak valid.", nil)
		return
	}

	ip := extractClientIP(r)
	err = h.authService.UpdateAdminRole(r.Context(), admin.ID, targetID, req.Role, &ip)
	if err != nil {
		if errors.Is(err, domain.ErrCannotChangeLastSuperAdmin) {
			writeProblem(w, r, http.StatusConflict, "LAST_SUPER_ADMIN_PROTECTED", "Tidak dapat mengubah peran Super Administrator aktif terakhir.", nil)
			return
		}
		if errors.Is(err, domain.ErrAdminNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Administrator tidak ditemukan.", nil)
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memperbarui peran administrator.", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type updatePermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

func (h *Handler) handleUpdateAdminPermissions(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Active admin session required.", nil)
		return
	}

	targetID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || targetID <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid administrator ID.", nil)
		return
	}

	var req updatePermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON request body.", nil)
		return
	}

	ip := extractClientIP(r)
	err = h.authService.UpdateAdminPermissions(r.Context(), admin.ID, targetID, req.Permissions, &ip)
	if err != nil {
		if errors.Is(err, domain.ErrAdminNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Administrator tidak ditemukan.", nil)
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memperbarui izin menu administrator.", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type toggleStatusRequest struct {
	IsActive bool `json:"is_active"`
}

func (h *Handler) handleToggleAdminStatus(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Active admin session required.", nil)
		return
	}

	targetID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || targetID <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid administrator ID.", nil)
		return
	}

	var req toggleStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON request body.", nil)
		return
	}

	ip := extractClientIP(r)
	err = h.authService.ToggleAdminStatus(r.Context(), admin.ID, targetID, req.IsActive, &ip)
	if err != nil {
		if errors.Is(err, domain.ErrSelfDeactivation) {
			writeProblem(w, r, http.StatusBadRequest, "SELF_DEACTIVATION_FORBIDDEN", "You cannot deactivate your own account.", nil)
			return
		}
		if errors.Is(err, domain.ErrCannotDeactivateLastAdmin) {
			writeProblem(w, r, http.StatusConflict, "LAST_ADMIN_PROTECTED", "Cannot deactivate the last active administrator or super administrator.", nil)
			return
		}
		if errors.Is(err, domain.ErrAdminNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Administrator tidak ditemukan.", nil)
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update administrator status.", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeProblem(w http.ResponseWriter, r *http.Request, status int, code, detail string, validationErrors map[string][]string) {
	reqID := r.Header.Get("X-Request-Id")
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(status)

	body := map[string]any{
		"type":       "about:blank",
		"title":      http.StatusText(status),
		"status":     status,
		"code":       code,
		"detail":     detail,
		"request_id": reqID,
	}
	if validationErrors != nil {
		body["errors"] = validationErrors
	}

	_ = json.NewEncoder(w).Encode(body)
}

func extractClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}
