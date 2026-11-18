package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"pidiekab/internal/modules/staticpage/application"
	"pidiekab/internal/modules/staticpage/domain"
	"pidiekab/internal/platform/httpserver/middleware"
)

type PageHandler struct {
	service     *application.PageService
	authService middleware.SessionAuthenticator
}

func NewPageHandler(service *application.PageService, authService middleware.SessionAuthenticator) *PageHandler {
	return &PageHandler{
		service:     service,
		authService: authService,
	}
}

func (h *PageHandler) RegisterRoutes(mux *http.ServeMux) {
	authMiddleware := middleware.RequireAdminSession(h.authService)
	csrfMiddleware := middleware.CSRFProtection

	wrapProtected := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(csrfMiddleware(handler))
	}

	wrapGet := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(handler)
	}

	mux.Handle("GET /api/v1/admin/pages", wrapGet(h.HandleListPages))
	mux.Handle("POST /api/v1/admin/pages", wrapProtected(h.HandleCreatePage))
	mux.Handle("GET /api/v1/admin/pages/{id}", wrapGet(h.HandleGetPage))
	mux.Handle("PUT /api/v1/admin/pages/{id}/translations/{locale}", wrapProtected(h.HandleSaveTranslation))
	mux.Handle("POST /api/v1/admin/pages/{id}/translations/{locale}/publish", wrapProtected(h.HandlePublishTranslation))
	mux.Handle("POST /api/v1/admin/pages/{id}/translations/{locale}/archive", wrapProtected(h.HandleArchiveTranslation))
}

type createPageRequest struct {
	PageKey string `json:"page_key"`
}

func (h *PageHandler) HandleCreatePage(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	var req createPageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON body")
		return
	}

	page, err := h.service.CreatePage(r.Context(), admin.ID, req.PageKey)
	if err != nil {
		if errors.Is(err, domain.ErrPageKeyEmpty) || errors.Is(err, domain.ErrPageKeyInvalid) {
			writeJSONError(w, http.StatusBadRequest, "INVALID_PAGE_KEY", err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create static page")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": page})
}

func (h *PageHandler) HandleListPages(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	pages, err := h.service.ListPages(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list static pages")
		return
	}

	if pages == nil {
		pages = []domain.StaticPageDetail{}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": pages,
		"meta": map[string]any{
			"total_items": len(pages),
		},
	})
}

func (h *PageHandler) HandleGetPage(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "Invalid static page ID")
		return
	}

	page, err := h.service.GetPage(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPageNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Static page not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get static page")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": page})
}

func (h *PageHandler) HandleSaveTranslation(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "Invalid static page ID")
		return
	}

	locale := strings.ToLower(r.PathValue("locale"))

	var input application.SaveTranslationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON body")
		return
	}

	trans, err := h.service.SaveTranslation(r.Context(), admin.ID, id, locale, input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidLocale) || errors.Is(err, domain.ErrTitleEmpty) ||
			errors.Is(err, domain.ErrInvalidSlug) || errors.Is(err, domain.ErrInvalidBody) {
			writeJSONError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save translation")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": trans})
}

func (h *PageHandler) HandlePublishTranslation(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "Invalid static page ID")
		return
	}
	locale := strings.ToLower(r.PathValue("locale"))

	err = h.service.PublishTranslation(r.Context(), admin.ID, id, locale)
	if err != nil {
		if errors.Is(err, domain.ErrTranslationNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Translation not found")
			return
		}
		if errors.Is(err, domain.ErrCannotPublishNoBody) {
			writeJSONError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to publish static page translation")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "Static page translation published successfully",
	})
}

func (h *PageHandler) HandleArchiveTranslation(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "Invalid static page ID")
		return
	}
	locale := strings.ToLower(r.PathValue("locale"))

	err = h.service.ArchiveTranslation(r.Context(), admin.ID, id, locale)
	if err != nil {
		if errors.Is(err, domain.ErrTranslationNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Translation not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to archive static page translation")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "Static page translation archived successfully",
	})
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
