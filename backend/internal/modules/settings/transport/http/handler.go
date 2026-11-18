package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"pidiekab/internal/modules/settings/application"
	"pidiekab/internal/modules/settings/domain"
	"pidiekab/internal/platform/httpserver/middleware"
)

type SettingsHandler struct {
	service     *application.SettingsService
	authService middleware.SessionAuthenticator
}

func NewSettingsHandler(service *application.SettingsService, authService middleware.SessionAuthenticator) *SettingsHandler {
	return &SettingsHandler{
		service:     service,
		authService: authService,
	}
}

func (h *SettingsHandler) RegisterRoutes(mux *http.ServeMux) {
	authMiddleware := middleware.RequireAdminSession(h.authService)
	csrfMiddleware := middleware.CSRFProtection

	wrapProtected := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(csrfMiddleware(handler))
	}

	wrapGet := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(handler)
	}

	mux.Handle("GET /api/v1/admin/settings", wrapGet(h.HandleGetSettings))
	mux.Handle("PUT /api/v1/admin/settings", wrapProtected(h.HandleUpdateGeneral))
	mux.Handle("PUT /api/v1/admin/settings/translations/{locale}", wrapProtected(h.HandleUpdateTranslation))
}

func (h *SettingsHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	settings, err := h.service.GetSettings(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve site settings")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": settings})
}

func (h *SettingsHandler) HandleUpdateGeneral(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	var input application.UpdateGeneralInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON body")
		return
	}

	res, err := h.service.UpdateGeneral(r.Context(), admin.ID, input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidContactEmail) ||
			errors.Is(err, domain.ErrInvalidContactPhone) ||
			errors.Is(err, domain.ErrInvalidSocialURL) {
			writeJSONError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update general settings")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": res})
}

func (h *SettingsHandler) HandleUpdateTranslation(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	locale := strings.ToLower(r.PathValue("locale"))

	var input application.UpdateTranslationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON body")
		return
	}

	t, err := h.service.UpdateTranslation(r.Context(), admin.ID, locale, input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidLocale) || errors.Is(err, domain.ErrSiteNameEmpty) {
			writeJSONError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update setting translation")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": t})
}

func writeJSONError(w http.ResponseWriter, statusCode int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":    errCode,
			"message": message,
		},
	})
}
