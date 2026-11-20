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

	mux.Handle("GET /api/v1/admin/curation/featured", wrapGet(h.HandleGetFeatured))
	mux.Handle("GET /api/v1/admin/curation/stories", wrapGet(h.HandleGetStories))
	mux.Handle("GET /api/v1/admin/curation/{locale}", wrapGet(h.HandleGetCuratedHomepage))
	mux.Handle("PUT /api/v1/admin/curation/featured", wrapProtected(h.HandleSetFeatured))
	mux.Handle("PUT /api/v1/admin/curation/stories", wrapProtected(h.HandleSetStories))
	mux.Handle("PUT /api/v1/admin/curation/{locale}/lead", wrapProtected(h.HandleSetLeadStory))
	mux.Handle("PUT /api/v1/admin/curation/{locale}/supporting", wrapProtected(h.HandleSetSupportingStories))
}

func (h *CurationHandler) HandleGetFeatured(w http.ResponseWriter, r *http.Request) {
	curated, err := h.service.GetCuratedHomepage(r.Context(), "id")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get featured")
		return
	}
	items := make([]map[string]any, 0, len(curated.Featured))
	for _, f := range curated.Featured {
		items = append(items, map[string]any{
			"contentId":    f.ContentID,
			"position":     f.Position,
			"title":        f.Title,
			"categoryName": f.Category,
		})
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"items": items,
		"data":  items,
	})
}

func (h *CurationHandler) HandleGetStories(w http.ResponseWriter, r *http.Request) {
	idCurated, _ := h.service.GetCuratedHomepage(r.Context(), "id")
	enCurated, _ := h.service.GetCuratedHomepage(r.Context(), "en")
	formatStories := func(c *domain.HomepageCuration) map[string]any {
		if c == nil {
			return map[string]any{"lead": nil, "supporting": []any{}}
		}
		var lead any
		if c.LeadStory != nil {
			lead = map[string]any{
				"contentId":    c.LeadStory.ContentID,
				"title":        c.LeadStory.Title,
				"categoryName": c.LeadStory.Category,
			}
		}
		supporting := make([]map[string]any, 0, len(c.SupportingStories))
		for _, s := range c.SupportingStories {
			supporting = append(supporting, map[string]any{
				"contentId":    s.ContentID,
				"position":     s.Position,
				"title":        s.Title,
				"categoryName": s.Category,
			})
		}
		return map[string]any{
			"lead":       lead,
			"supporting": supporting,
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id": formatStories(idCurated),
		"en": formatStories(enCurated),
	})
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
	Items      []struct {
		ContentID int64 `json:"content_id"`
		Position  int   `json:"position"`
	} `json:"items"`
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

	// Extract ContentIDs if provided as Items array
	contentIDs := req.ContentIDs
	if len(contentIDs) == 0 && len(req.Items) > 0 {
		for _, it := range req.Items {
			contentIDs = append(contentIDs, it.ContentID)
		}
	}

	if err := h.service.SetFeatured(r.Context(), admin.ID, contentIDs); err != nil {
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

type setStoriesRequest struct {
	Locale        string `json:"locale"`
	LeadContentID *int64 `json:"lead_content_id"`
	Supporting    []struct {
		ContentID int64 `json:"content_id"`
		Position  int   `json:"position"`
	} `json:"supporting"`
	SupportingIDs []int64 `json:"supporting_ids"`
}

func (h *CurationHandler) HandleSetStories(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	var req setStoriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON body")
		return
	}

	locale := strings.ToLower(strings.TrimSpace(req.Locale))
	if locale == "" {
		locale = "id"
	}

	supportingIDs := req.SupportingIDs
	if len(supportingIDs) == 0 && len(req.Supporting) > 0 {
		for _, it := range req.Supporting {
			supportingIDs = append(supportingIDs, it.ContentID)
		}
	}

	if err := h.service.SetStories(r.Context(), admin.ID, locale, req.LeadContentID, supportingIDs); err != nil {
		if errors.Is(err, domain.ErrInvalidLocale) || errors.Is(err, domain.ErrDuplicateContent) || errors.Is(err, domain.ErrContentNotPublished) {
			writeJSONError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update stories")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "Stories updated successfully"})
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
