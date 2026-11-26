package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"pidiekab/internal/modules/staticpage/application"
	"pidiekab/internal/modules/staticpage/domain"
	"pidiekab/internal/platform/httpserver/middleware"
)

var slugUnsafeRegex = regexp.MustCompile(`[^a-z0-9]+`)

func sanitizeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugUnsafeRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

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

	// Standard /api/v1/admin/pages routes
	mux.Handle("GET /api/v1/admin/pages", wrapGet(h.HandleListPages))
	mux.Handle("POST /api/v1/admin/pages", wrapProtected(h.HandleCreatePage))
	mux.Handle("GET /api/v1/admin/pages/{id}", wrapGet(h.HandleGetPage))
	mux.Handle("PUT /api/v1/admin/pages/{id}", wrapProtected(h.HandleUpdatePage))
	mux.Handle("PUT /api/v1/admin/pages/{id}/key", wrapProtected(h.HandleUpdatePageKey))
	mux.Handle("DELETE /api/v1/admin/pages/{id}", wrapProtected(h.HandleDeletePage))
	mux.Handle("PUT /api/v1/admin/pages/{id}/translations/{locale}", wrapProtected(h.HandleSaveTranslation))
	mux.Handle("POST /api/v1/admin/pages/{id}/translations/{locale}/publish", wrapProtected(h.HandlePublishTranslation))
	mux.Handle("POST /api/v1/admin/pages/{id}/translations/{locale}/archive", wrapProtected(h.HandleArchiveTranslation))
	mux.Handle("POST /api/v1/admin/pages/{id}/translations/{locale}/draft", wrapProtected(h.HandleDraftTranslation))

	// Static-pages alias routes matching frontend /api/v1/admin/static-pages
	mux.Handle("GET /api/v1/admin/static-pages", wrapGet(h.HandleListPages))
	mux.Handle("POST /api/v1/admin/static-pages", wrapProtected(h.HandleCreatePage))
	mux.Handle("GET /api/v1/admin/static-pages/{id}", wrapGet(h.HandleGetPage))
	mux.Handle("PUT /api/v1/admin/static-pages/{id}", wrapProtected(h.HandleUpdatePage))
	mux.Handle("PUT /api/v1/admin/static-pages/{id}/key", wrapProtected(h.HandleUpdatePageKey))
	mux.Handle("DELETE /api/v1/admin/static-pages/{id}", wrapProtected(h.HandleDeletePage))
	mux.Handle("PUT /api/v1/admin/static-pages/{id}/translations/{locale}", wrapProtected(h.HandleSaveTranslation))
	mux.Handle("POST /api/v1/admin/static-pages/{id}/translations/{locale}/publish", wrapProtected(h.HandlePublishTranslation))
	mux.Handle("POST /api/v1/admin/static-pages/{id}/translations/{locale}/archive", wrapProtected(h.HandleArchiveTranslation))
	mux.Handle("POST /api/v1/admin/static-pages/{id}/translations/{locale}/draft", wrapProtected(h.HandleDraftTranslation))
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

	// Populate BodyHTML helper field for each translation
	for i := range pages {
		for loc, trans := range pages[i].Translations {
			if trans != nil && len(trans.Body) > 0 {
				var bodyMap map[string]any
				if err := json.Unmarshal(trans.Body, &bodyMap); err == nil {
					if htmlVal, ok := bodyMap["html"].(string); ok && htmlVal != "" {
						trans.BodyHTML = htmlVal
					} else if mdVal, ok := bodyMap["markdown"].(string); ok && mdVal != "" {
						trans.BodyHTML = mdVal
					}
				}
				pages[i].Translations[loc] = trans
			}
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data":  pages,
		"pages": pages,
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

	// Populate BodyHTML helper field
	for loc, trans := range page.Translations {
		if trans != nil && len(trans.Body) > 0 {
			var bodyMap map[string]any
			if err := json.Unmarshal(trans.Body, &bodyMap); err == nil {
				if htmlVal, ok := bodyMap["html"].(string); ok && htmlVal != "" {
					trans.BodyHTML = htmlVal
				} else if mdVal, ok := bodyMap["markdown"].(string); ok && mdVal != "" {
					trans.BodyHTML = mdVal
				}
			}
			page.Translations[loc] = trans
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": page,
		"page": page,
	})
}

type updatePageRequest struct {
	PageKey      string                             `json:"page_key"`
	Translations map[string]updatePageTransRequest `json:"translations"`
}

type updatePageTransRequest struct {
	Title    string `json:"title"`
	Slug     string `json:"slug"`
	BodyHTML string `json:"body_html"`
	Body     any    `json:"body"`
	Status   string `json:"status"`
}

func (h *PageHandler) HandleUpdatePage(w http.ResponseWriter, r *http.Request) {
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

	var req updatePageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON body")
		return
	}

	cleanKey := strings.TrimSpace(req.PageKey)
	if cleanKey != "" {
		if err := h.service.UpdatePageKey(r.Context(), admin.ID, id, cleanKey); err != nil {
			if errors.Is(err, domain.ErrPageKeyEmpty) || errors.Is(err, domain.ErrPageKeyInvalid) {
				writeJSONError(w, http.StatusBadRequest, "INVALID_PAGE_KEY", err.Error())
				return
			}
			writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update page key: "+err.Error())
			return
		}
	}

	for locale, tReq := range req.Translations {
		locale = strings.ToLower(strings.TrimSpace(locale))
		if locale != domain.LocaleID && locale != domain.LocaleEN {
			continue
		}

		bodyContent := strings.TrimSpace(tReq.BodyHTML)
		if bodyContent == "" && tReq.Body != nil {
			if s, ok := tReq.Body.(string); ok {
				bodyContent = s
			} else if m, ok := tReq.Body.(map[string]any); ok {
				if hVal, ok := m["html"].(string); ok {
					bodyContent = hVal
				} else if mdVal, ok := m["markdown"].(string); ok {
					bodyContent = mdVal
				}
			}
		}

		title := strings.TrimSpace(tReq.Title)
		if title == "" {
			if bodyContent != "" || strings.TrimSpace(tReq.Slug) != "" {
				words := strings.Split(cleanKey, "-")
				for i, w := range words {
					if len(w) > 0 {
						words[i] = strings.ToUpper(w[:1]) + w[1:]
					}
				}
				title = strings.Join(words, " ")
			} else {
				continue
			}
		}

		bodyJSON, _ := json.Marshal(map[string]string{
			"html":     bodyContent,
			"markdown": bodyContent,
		})

		slug := sanitizeSlug(tReq.Slug)
		if slug == "" {
			slug = sanitizeSlug(title)
		}
		if slug == "" {
			slug = sanitizeSlug(cleanKey)
		}
		if slug == "" {
			slug = "page-" + locale
		}

		saveInput := application.SaveTranslationInput{
			Title: title,
			Slug:  slug,
			Body:  bodyJSON,
		}

		_, err := h.service.SaveTranslation(r.Context(), admin.ID, id, locale, saveInput)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}

		status := strings.ToLower(strings.TrimSpace(tReq.Status))
		if status == domain.StatusPublished {
			_ = h.service.PublishTranslation(r.Context(), admin.ID, id, locale)
		} else if status == domain.StatusArchived {
			_ = h.service.ArchiveTranslation(r.Context(), admin.ID, id, locale)
		} else if status == domain.StatusDraft {
			_ = h.service.DraftTranslation(r.Context(), admin.ID, id, locale)
		}
	}

	// Fetch updated page
	page, err := h.service.GetPage(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve updated page")
		return
	}

	// Populate BodyHTML helper field
	for loc, trans := range page.Translations {
		if trans != nil && len(trans.Body) > 0 {
			var bodyMap map[string]any
			if err := json.Unmarshal(trans.Body, &bodyMap); err == nil {
				if htmlVal, ok := bodyMap["html"].(string); ok && htmlVal != "" {
					trans.BodyHTML = htmlVal
				} else if mdVal, ok := bodyMap["markdown"].(string); ok && mdVal != "" {
					trans.BodyHTML = mdVal
				}
			}
			page.Translations[loc] = trans
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": page,
		"page": page,
	})
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

	input.Slug = sanitizeSlug(input.Slug)
	if input.Slug == "" {
		input.Slug = sanitizeSlug(input.Title)
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

func (h *PageHandler) HandleDraftTranslation(w http.ResponseWriter, r *http.Request) {
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

	err = h.service.DraftTranslation(r.Context(), admin.ID, id, locale)
	if err != nil {
		if errors.Is(err, domain.ErrTranslationNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Translation not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to draft static page translation")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "Static page translation set to draft successfully",
	})
}

type updatePageKeyRequest struct {
	PageKey string `json:"page_key"`
}

func (h *PageHandler) HandleUpdatePageKey(w http.ResponseWriter, r *http.Request) {
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

	var req updatePageKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON body")
		return
	}

	cleanKey := strings.TrimSpace(req.PageKey)
	if err := h.service.UpdatePageKey(r.Context(), admin.ID, id, cleanKey); err != nil {
		if errors.Is(err, domain.ErrPageNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Static page not found")
			return
		}
		if errors.Is(err, domain.ErrPageKeyEmpty) || errors.Is(err, domain.ErrPageKeyInvalid) {
			writeJSONError(w, http.StatusBadRequest, "INVALID_PAGE_KEY", err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update static page key")
		return
	}

	page, err := h.service.GetPage(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get updated static page")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data":    page,
		"page":    page,
		"message": "Nama kunci halaman berhasil diperbarui",
	})
}

func (h *PageHandler) HandleDeletePage(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.DeletePage(r.Context(), admin.ID, id); err != nil {
		if errors.Is(err, domain.ErrPageNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Static page not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete static page")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "Halaman statis berhasil dihapus",
		"id":      id,
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
