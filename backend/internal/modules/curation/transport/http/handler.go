package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"pidiekab/internal/modules/curation/application"
	"pidiekab/internal/modules/curation/domain"
	"pidiekab/internal/platform/httpserver/middleware"
)

type CurationHandler struct {
	service     *application.CurationService
	authService middleware.SessionAuthenticator
}

func NewCurationHandler(service *application.CurationService, authService middleware.SessionAuthenticator) *CurationHandler {
	return &CurationHandler{
		service:     service,
		authService: authService,
	}
}

func (h *CurationHandler) RegisterRoutes(mux *http.ServeMux) {
	authMiddleware := middleware.RequireAdminSession(h.authService)
	csrfMiddleware := middleware.CSRFProtection

	wrapProtected := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(csrfMiddleware(handler))
	}

	wrapGet := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(handler)
	}

	mux.Handle("GET /api/v1/admin/curation/{locale}", wrapGet(h.HandleGetCuratedHomepage))
	mux.Handle("PUT /api/v1/admin/curation/featured", wrapProtected(h.HandleSetFeatured))
	mux.Handle("PUT /api/v1/admin/curation/{locale}/lead", wrapProtected(h.HandleSetLeadStory))
	mux.Handle("PUT /api/v1/admin/curation/{locale}/supporting", wrapProtected(h.HandleSetSupportingStories))
}

func (h *CurationHandler) HandleGetCuratedHomepage(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	locale := strings.ToLower(r.PathValue("locale"))
	curated, err := h.service.GetCuratedHomepage(r.Context(), locale)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidLocale) {
			writeJSONError(w, http.StatusBadRequest, "INVALID_LOCALE", "Locale must be 'id' or 'en'")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get curated homepage")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": curated})
}

type setFeaturedRequest struct {
	ContentIDs []int64 `json:"content_ids"`
}

func (h *CurationHandler) HandleSetFeatured(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	var req setFeaturedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON body")
		return
	}

	if err := h.service.SetFeatured(r.Context(), admin.ID, req.ContentIDs); err != nil {
		if errors.Is(err, domain.ErrDuplicateContent) || errors.Is(err, domain.ErrContentNotPublished) {
			writeJSONError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update featured contents")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "Featured contents updated successfully"})
}

type setLeadStoryRequest struct {
	ContentID int64 `json:"content_id"`
}

func (h *CurationHandler) HandleSetLeadStory(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	locale := strings.ToLower(r.PathValue("locale"))

	var req setLeadStoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON body")
		return
	}

	if err := h.service.SetLeadStory(r.Context(), admin.ID, locale, req.ContentID); err != nil {
		if errors.Is(err, domain.ErrInvalidLocale) || errors.Is(err, domain.ErrContentNotPublished) {
			writeJSONError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to set lead story")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "Lead story updated successfully"})
}

type setSupportingRequest struct {
	ContentIDs []int64 `json:"content_ids"`
}

func (h *CurationHandler) HandleSetSupportingStories(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	locale := strings.ToLower(r.PathValue("locale"))

	var req setSupportingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON body")
		return
	}

	if err := h.service.SetSupportingStories(r.Context(), admin.ID, locale, req.ContentIDs); err != nil {
		if errors.Is(err, domain.ErrInvalidLocale) || errors.Is(err, domain.ErrDuplicateContent) || errors.Is(err, domain.ErrContentNotPublished) {
			writeJSONError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to set supporting stories")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "Supporting stories updated successfully"})
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
