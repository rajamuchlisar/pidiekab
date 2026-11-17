package http

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pidiekab/internal/modules/content/application"
	"pidiekab/internal/modules/content/domain"
	"pidiekab/internal/modules/content/infrastructure/postgres"
	identityApp "pidiekab/internal/modules/identity/application"
	"pidiekab/internal/platform/httpserver/middleware"
)

// Handler manages HTTP transport for content and translations.
type Handler struct {
	service     *application.ContentService
	authService *identityApp.AuthService
}

// NewHandler creates a new content transport Handler.
func NewHandler(service *application.ContentService, authService *identityApp.AuthService) *Handler {
	return &Handler{
		service:     service,
		authService: authService,
	}
}

// RegisterRoutes registers content endpoints onto the ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	authMiddleware := middleware.RequireAdminSession(h.authService)
	csrfMiddleware := middleware.CSRFProtection

	// Content Identities
	mux.Handle("GET /api/v1/admin/contents", authMiddleware(http.HandlerFunc(h.handleListContents)))
	mux.Handle("POST /api/v1/admin/contents", authMiddleware(csrfMiddleware(http.HandlerFunc(h.handleCreateContent))))
	mux.Handle("GET /api/v1/admin/contents/{content_id}", authMiddleware(http.HandlerFunc(h.handleGetContent)))
	mux.Handle("PUT /api/v1/admin/contents/{content_id}", authMiddleware(csrfMiddleware(http.HandlerFunc(h.handleUpdateShared))))
	mux.Handle("DELETE /api/v1/admin/contents/{content_id}", authMiddleware(csrfMiddleware(http.HandlerFunc(h.handleDeleteContent))))

	// Translations
	mux.Handle("GET /api/v1/admin/contents/{content_id}/translations/{locale}", authMiddleware(http.HandlerFunc(h.handleGetTranslation)))
	mux.Handle("PUT /api/v1/admin/contents/{content_id}/translations/{locale}", authMiddleware(csrfMiddleware(http.HandlerFunc(h.handleSaveTranslation))))
	mux.Handle("POST /api/v1/admin/contents/{content_id}/translations/{locale}/publish", authMiddleware(csrfMiddleware(http.HandlerFunc(h.handlePublishLocale))))
	mux.Handle("POST /api/v1/admin/contents/{content_id}/translations/{locale}/archive", authMiddleware(csrfMiddleware(http.HandlerFunc(h.handleArchiveLocale))))

	// Preview
	mux.Handle("GET /api/v1/admin/preview/contents/{content_id}", authMiddleware(http.HandlerFunc(h.handlePreviewContent)))

	// Sources
	mux.Handle("GET /api/v1/admin/contents/{content_id}/sources", authMiddleware(http.HandlerFunc(h.handleListSources)))
	mux.Handle("POST /api/v1/admin/contents/{content_id}/sources", authMiddleware(csrfMiddleware(http.HandlerFunc(h.handleAddSource))))
	mux.Handle("DELETE /api/v1/admin/contents/{content_id}/sources/{source_id}", authMiddleware(csrfMiddleware(http.HandlerFunc(h.handleDeleteSource))))
}

func (h *Handler) handleListContents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var categoryID *int64
	if catStr := q.Get("category_id"); catStr != "" {
		if id, err := strconv.ParseInt(catStr, 10, 64); err == nil {
			categoryID = &id
		}
	}

	var status *string
	if s := q.Get("status"); s != "" {
		status = &s
	}

	var locale *string
	if l := q.Get("locale"); l != "" {
		locale = &l
	}

	var query *string
	if search := q.Get("q"); search != "" {
		query = &search
	}

	page := 1
	if pStr := q.Get("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 12
	if ppStr := q.Get("per_page"); ppStr != "" {
		if pp, err := strconv.Atoi(ppStr); err == nil && pp > 0 && pp <= 50 {
			perPage = pp
		}
	}

	items, total, err := h.service.ListContents(r.Context(), postgres.FilterParams{
		CategoryID: categoryID,
		Status:     status,
		Locale:     locale,
		Query:      query,
		Page:       page,
		PerPage:    perPage,
	})
	if err != nil {
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list contents.")
		return
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": items,
		"meta": map[string]any{
			"page":        page,
			"per_page":    perPage,
			"total_items": total,
			"total_pages": totalPages,
		},
	})
}

type contentCreateRequest struct {
	CategoryID       string   `json:"category_id"`
	VillageID        *string  `json:"village_id"`
	Address          *string  `json:"address"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	LocationAccuracy *string  `json:"location_accuracy"`
	MainMediaID      *string  `json:"main_media_id"`
}

func (h *Handler) handleCreateContent(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	var req contentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON request body.")
		return
	}

	catID, err := strconv.ParseInt(req.CategoryID, 10, 64)
	if err != nil || catID <= 0 {
		writeProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "category_id must be a valid ID string.")
		return
	}

	var villageID *int64
	if req.VillageID != nil && *req.VillageID != "" {
		if vid, err := strconv.ParseInt(*req.VillageID, 10, 64); err == nil && vid > 0 {
			villageID = &vid
		}
	}

	var mainMediaID *int64
	if req.MainMediaID != nil && *req.MainMediaID != "" {
		if mid, err := strconv.ParseInt(*req.MainMediaID, 10, 64); err == nil && mid > 0 {
			mainMediaID = &mid
		}
	}

	ip := extractClientIP(r)
	c, err := h.service.CreateContent(r.Context(), catID, villageID, req.Address, req.Latitude, req.Longitude, req.LocationAccuracy, mainMediaID, admin.ID, &ip)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryInactive) {
			writeProblem(w, r, http.StatusUnprocessableEntity, "CATEGORY_INACTIVE", "Cannot assign content to an inactive category.")
			return
		}
		if errors.Is(err, domain.ErrInvalidCoordinates) || errors.Is(err, domain.ErrLatitudeOutOfRange) || errors.Is(err, domain.ErrLongitudeOutOfRange) || errors.Is(err, domain.ErrInvalidAccuracy) || errors.Is(err, domain.ErrAccuracyRequiresCoords) {
			writeProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create content: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]any{
			"id":          strconv.FormatInt(c.ID, 10),
			"category_id": strconv.FormatInt(c.CategoryID, 10),
		},
	})
}

func (h *Handler) handleGetContent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("content_id"), 10, 64)
	if err != nil || id <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid content ID.")
		return
	}

	detail, err := h.service.GetContent(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrContentNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Content not found.")
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get content.")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": detail,
	})
}

func (h *Handler) handlePreviewContent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("content_id"), 10, 64)
	if err != nil || id <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid content ID.")
		return
	}

	locale := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("locale")))
	if locale == "" {
		locale = domain.LocaleID
	}
	if locale != domain.LocaleID && locale != domain.LocaleEN {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid locale; must be 'id' or 'en'.")
		return
	}

	detail, err := h.service.GetContent(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrContentNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Content not found.")
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve content.")
		return
	}

	trans := detail.Translations[locale]

	// Strict preview response headers to prevent accidental caching or search engine indexing
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")

	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]any{
			"content_id":  detail.ID,
			"locale":      locale,
			"translation": trans,
			"shared": map[string]any{
				"category_id":       detail.CategoryID,
				"village_id":        detail.VillageID,
				"address":           detail.Address,
				"latitude":          detail.Latitude,
				"longitude":         detail.Longitude,
				"location_accuracy": detail.LocationAccuracy,
				"main_media_id":     detail.MainMediaID,
			},
			"sources":    detail.Sources,
			"is_preview": true,
		},
	})
}

func (h *Handler) handleUpdateShared(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("content_id"), 10, 64)
	if err != nil || id <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid content ID.")
		return
	}

	var req contentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON request body.")
		return
	}

	catID, err := strconv.ParseInt(req.CategoryID, 10, 64)
	if err != nil || catID <= 0 {
		writeProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "category_id must be a valid ID string.")
		return
	}

	var villageID *int64
	if req.VillageID != nil && *req.VillageID != "" {
		if vid, err := strconv.ParseInt(*req.VillageID, 10, 64); err == nil && vid > 0 {
			villageID = &vid
		}
	}

	var mainMediaID *int64
	if req.MainMediaID != nil && *req.MainMediaID != "" {
		if mid, err := strconv.ParseInt(*req.MainMediaID, 10, 64); err == nil && mid > 0 {
			mainMediaID = &mid
		}
	}

	ip := extractClientIP(r)
	err = h.service.UpdateShared(r.Context(), id, catID, villageID, req.Address, req.Latitude, req.Longitude, req.LocationAccuracy, mainMediaID, admin.ID, &ip)
	if err != nil {
		if errors.Is(err, domain.ErrContentNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Content not found.")
			return
		}
		if errors.Is(err, domain.ErrCategoryInactive) {
			writeProblem(w, r, http.StatusUnprocessableEntity, "CATEGORY_INACTIVE", "Cannot assign content to an inactive category.")
			return
		}
		writeProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleDeleteContent(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("content_id"), 10, 64)
	if err != nil || id <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid content ID.")
		return
	}

	ip := extractClientIP(r)
	if err := h.service.DeleteContent(r.Context(), id, admin.ID, &ip); err != nil {
		if errors.Is(err, domain.ErrContentNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Content not found.")
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete content.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleGetTranslation(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("content_id"), 10, 64)
	if err != nil || id <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid content ID.")
		return
	}
	locale := r.PathValue("locale")

	t, err := h.service.GetTranslation(r.Context(), id, locale)
	if err != nil {
		if errors.Is(err, domain.ErrTranslationNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Translation not found.")
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get translation.")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": t,
	})
}

type translationWriteRequest struct {
	Title          string         `json:"title"`
	Slug           string         `json:"slug"`
	Summary        *string        `json:"summary"`
	Body           map[string]any `json:"body"`
	SEOTitle       *string        `json:"seo_title"`
	SEODescription *string        `json:"seo_description"`
}

func (h *Handler) handleSaveTranslation(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("content_id"), 10, 64)
	if err != nil || id <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid content ID.")
		return
	}
	locale := r.PathValue("locale")

	var req translationWriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON request body.")
		return
	}

	ip := extractClientIP(r)
	t, err := h.service.SaveTranslation(r.Context(), id, locale, req.Title, req.Slug, req.Summary, req.Body, req.SEOTitle, req.SEODescription, admin.ID, &ip)
	if err != nil {
		if errors.Is(err, domain.ErrContentNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Content not found.")
			return
		}
		if errors.Is(err, domain.ErrSlugAlreadyExists) {
			writeProblem(w, r, http.StatusConflict, "SLUG_ALREADY_EXISTS", "Content slug is already in use for this locale.")
			return
		}
		if errors.Is(err, domain.ErrInvalidSlug) || errors.Is(err, domain.ErrTitleEmpty) || errors.Is(err, domain.ErrBodyEmpty) {
			writeProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save translation.")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": t,
	})
}

func (h *Handler) handlePublishLocale(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("content_id"), 10, 64)
	if err != nil || id <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid content ID.")
		return
	}
	locale := r.PathValue("locale")

	ip := extractClientIP(r)
	err = h.service.PublishLocale(r.Context(), id, locale, admin.ID, &ip)
	if err != nil {
		if errors.Is(err, domain.ErrContentNotFound) || errors.Is(err, domain.ErrTranslationNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		if errors.Is(err, domain.ErrCategoryInactive) || errors.Is(err, domain.ErrNoSourcesRecorded) || errors.Is(err, domain.ErrTitleEmpty) || errors.Is(err, domain.ErrBodyEmpty) {
			writeProblem(w, r, http.StatusUnprocessableEntity, "PUBLISH_VALIDATION_ERROR", err.Error())
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to publish locale: "+err.Error())
		return
	}

	// 202 Accepted per OpenAPI specification
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) handleArchiveLocale(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("content_id"), 10, 64)
	if err != nil || id <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid content ID.")
		return
	}
	locale := r.PathValue("locale")

	ip := extractClientIP(r)
	err = h.service.ArchiveLocale(r.Context(), id, locale, admin.ID, &ip)
	if err != nil {
		if errors.Is(err, domain.ErrTranslationNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Translation not found.")
			return
		}
		if errors.Is(err, domain.ErrCannotArchiveUnpublished) {
			writeProblem(w, r, http.StatusConflict, "CONFLICT", err.Error())
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to archive locale.")
		return
	}

	// 202 Accepted per OpenAPI specification
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) handleListSources(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("content_id"), 10, 64)
	if err != nil || id <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid content ID.")
		return
	}

	sources, err := h.service.ListSources(r.Context(), id)
	if err != nil {
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list sources.")
		return
	}

	if sources == nil {
		sources = []*domain.ContentSource{}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": sources,
	})
}

type addSourceRequest struct {
	SourceType        string     `json:"source_type"`
	SourceName        string     `json:"source_name"`
	SourceURL         *string    `json:"source_url"`
	SourcePublishedAt *time.Time `json:"source_published_at"`
	AccessedAt        *time.Time `json:"accessed_at"`
	Notes             *string    `json:"notes"`
}

func (h *Handler) handleAddSource(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("content_id"), 10, 64)
	if err != nil || id <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid content ID.")
		return
	}

	var req addSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON request body.")
		return
	}

	ip := extractClientIP(r)
	src, err := h.service.AddSource(r.Context(), id, req.SourceType, req.SourceName, req.SourceURL, req.SourcePublishedAt, req.AccessedAt, req.Notes, admin.ID, &ip)
	if err != nil {
		if errors.Is(err, domain.ErrContentNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Content not found.")
			return
		}
		if errors.Is(err, domain.ErrInvalidSourceType) || errors.Is(err, domain.ErrSourceNameEmpty) {
			writeProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add source.")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": src,
	})
}

func (h *Handler) handleDeleteSource(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	contentID, err := strconv.ParseInt(r.PathValue("content_id"), 10, 64)
	if err != nil || contentID <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid content ID.")
		return
	}

	sourceID, err := strconv.ParseInt(r.PathValue("source_id"), 10, 64)
	if err != nil || sourceID <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid source ID.")
		return
	}

	ip := extractClientIP(r)
	if err := h.service.DeleteSource(r.Context(), contentID, sourceID, admin.ID, &ip); err != nil {
		if errors.Is(err, domain.ErrSourceNotFound) {
			writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Source not found.")
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete source.")
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
