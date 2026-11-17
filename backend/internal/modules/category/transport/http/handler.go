package http

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"

	"pidiekab/internal/modules/category/application"
	"pidiekab/internal/modules/category/domain"
	identityApp "pidiekab/internal/modules/identity/application"
	"pidiekab/internal/platform/httpserver/middleware"
)

// Handler manages HTTP transport for category administrative operations.
type Handler struct {
	service     *application.Service
	authService *identityApp.AuthService
}

// NewHandler creates a new category transport Handler.
func NewHandler(service *application.Service, authService *identityApp.AuthService) *Handler {
	return &Handler{
		service:     service,
		authService: authService,
	}
}

// RegisterRoutes registers category endpoints onto the ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	authMiddleware := middleware.RequireAdminSession(h.authService)
	csrfMiddleware := middleware.CSRFProtection

	mux.Handle("GET /api/v1/admin/categories", authMiddleware(http.HandlerFunc(h.handleList)))
	mux.Handle("POST /api/v1/admin/categories", authMiddleware(csrfMiddleware(http.HandlerFunc(h.handleCreate))))
	mux.Handle("GET /api/v1/admin/categories/{category_id}", authMiddleware(http.HandlerFunc(h.handleGet)))
	mux.Handle("PUT /api/v1/admin/categories/{category_id}", authMiddleware(csrfMiddleware(http.HandlerFunc(h.handleUpdate))))
	mux.Handle("PUT /api/v1/admin/categories/{category_id}/translations/{locale}", authMiddleware(csrfMiddleware(http.HandlerFunc(h.handleSaveTranslation))))
	mux.Handle("DELETE /api/v1/admin/categories/{category_id}", authMiddleware(csrfMiddleware(http.HandlerFunc(h.handleDelete))))
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.ListCategories(r.Context(), false)
	if err != nil {
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve categories.")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": categories,
	})
}

type translationInput struct {
	Locale      string  `json:"locale"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Slug        string  `json:"slug"`
}

type createCategoryRequest struct {
	IsActive     *bool              `json:"is_active"`
	Translations []translationInput `json:"translations"`
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON request body.")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	var translations []*domain.CategoryTranslation
	for _, t := range req.Translations {
		translations = append(translations, &domain.CategoryTranslation{
			Locale:      t.Locale,
			Name:        t.Name,
			Description: t.Description,
			Slug:        t.Slug,
		})
	}

	ip := extractClientIP(r)
	cat, err := h.service.CreateCategory(r.Context(), isActive, admin.ID, translations, &ip)
	if err != nil {
		if errors.Is(err, domain.ErrSlugAlreadyExists) {
			writeProblem(w, r, http.StatusConflict, "SLUG_ALREADY_EXISTS", "One or more category slugs already exist for this locale.")
			return
		}
		if errors.Is(err, domain.ErrInvalidSlug) || errors.Is(err, domain.ErrInvalidLocale) || errors.Is(err, domain.ErrCategoryNameEmpty) {
			writeProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create category.")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": cat,
	})
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("category_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid category ID.")
		return
	}

	cat, err := h.service.GetCategory(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Category not found.")
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get category.")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": cat,
	})
}

type updateCategoryRequest struct {
	IsActive bool `json:"is_active"`
}

func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	idStr := r.PathValue("category_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid category ID.")
		return
	}

	var req updateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON request body.")
		return
	}

	ip := extractClientIP(r)
	if err := h.service.UpdateCategory(r.Context(), id, req.IsActive, admin.ID, &ip); err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Category not found.")
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update category.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type saveTranslationRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Slug        string  `json:"slug"`
}

func (h *Handler) handleSaveTranslation(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	idStr := r.PathValue("category_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid category ID.")
		return
	}

	locale := r.PathValue("locale")

	var req saveTranslationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON request body.")
		return
	}

	ip := extractClientIP(r)
	t, err := h.service.SaveTranslation(r.Context(), id, locale, req.Name, req.Description, req.Slug, admin.ID, &ip)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Category not found.")
			return
		}
		if errors.Is(err, domain.ErrSlugAlreadyExists) {
			writeProblem(w, r, http.StatusConflict, "SLUG_ALREADY_EXISTS", "Category slug already exists for this locale.")
			return
		}
		if errors.Is(err, domain.ErrInvalidSlug) || errors.Is(err, domain.ErrInvalidLocale) || errors.Is(err, domain.ErrCategoryNameEmpty) {
			writeProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save category translation.")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": t,
	})
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	idStr := r.PathValue("category_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid category ID.")
		return
	}

	ip := extractClientIP(r)
	if err := h.service.DeleteCategory(r.Context(), id, admin.ID, &ip); err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Category not found.")
			return
		}
		if errors.Is(err, domain.ErrCategoryInUse) {
			writeProblem(w, r, http.StatusConflict, "CATEGORY_IN_USE", "Category cannot be deleted because it is referenced by existing content.")
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete category.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeProblem(w http.ResponseWriter, r *http.Request, status int, code, detail string) {
	reqID := r.Header.Get("X-Request-Id")
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":       "about:blank",
		"title":      http.StatusText(status),
		"status":     status,
		"code":       code,
		"detail":     detail,
		"request_id": reqID,
	})
}

func extractClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}
