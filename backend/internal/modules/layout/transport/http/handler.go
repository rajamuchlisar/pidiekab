package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"pidiekab/internal/modules/layout/application"
	"pidiekab/internal/modules/layout/domain"
	"pidiekab/internal/platform/httpserver/middleware"
)

type LayoutHandler struct {
	service     *application.LayoutService
	authService middleware.SessionAuthenticator
}

func NewLayoutHandler(service *application.LayoutService, authService middleware.SessionAuthenticator) *LayoutHandler {
	return &LayoutHandler{
		service:     service,
		authService: authService,
	}
}

func (h *LayoutHandler) RegisterRoutes(mux *http.ServeMux) {
	authMiddleware := middleware.RequireAdminSession(h.authService)
	csrfMiddleware := middleware.CSRFProtection

	wrapProtected := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(csrfMiddleware(handler))
	}

	wrapGet := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(handler)
	}

	// Public active banners
	mux.HandleFunc("GET /api/v1/public/banners", h.HandlePublicListBanners)

	// Admin management endpoints
	mux.Handle("GET /api/v1/admin/banners", wrapGet(h.HandleListBanners))
	mux.Handle("POST /api/v1/admin/banners", wrapProtected(h.HandleCreateBanner))
	mux.Handle("GET /api/v1/admin/banners/{id}", wrapGet(h.HandleGetBanner))
	mux.Handle("PUT /api/v1/admin/banners/{id}", wrapProtected(h.HandleUpdateBanner))
	mux.Handle("DELETE /api/v1/admin/banners/{id}", wrapProtected(h.HandleDeleteBanner))
	mux.Handle("PATCH /api/v1/admin/banners/{id}/toggle", wrapProtected(h.HandleToggleActive))
}

type bannerPayload struct {
	Title        string  `json:"title"`
	Placement    string  `json:"placement"`
	BannerType   string  `json:"banner_type"`
	ImageURL     string  `json:"image_url"`
	TargetURL    string  `json:"target_url"`
	HTMLCode     string  `json:"html_code"`
	OpenInNewTab bool    `json:"open_in_new_tab"`
	IsActive     bool    `json:"is_active"`
	SortOrder    int     `json:"sort_order"`
	StartDate    *string `json:"start_date,omitempty"`
	EndDate      *string `json:"end_date,omitempty"`
}

func (h *LayoutHandler) HandlePublicListBanners(w http.ResponseWriter, r *http.Request) {
	placement := r.URL.Query().Get("placement")
	banners, err := h.service.ListActiveBanners(r.Context(), placement)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat daftar banner aktif")
		return
	}
	if banners == nil {
		banners = []domain.Banner{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": banners})
}

func (h *LayoutHandler) HandleListBanners(w http.ResponseWriter, r *http.Request) {
	placement := r.URL.Query().Get("placement")
	banners, err := h.service.ListBanners(r.Context(), placement)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat daftar banner")
		return
	}
	if banners == nil {
		banners = []domain.Banner{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": banners})
}

func (h *LayoutHandler) HandleGetBanner(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "ID banner tidak valid")
		return
	}

	banner, err := h.service.GetBannerByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrBannerNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Banner tidak ditemukan")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat banner")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": banner})
}

func (h *LayoutHandler) HandleCreateBanner(w http.ResponseWriter, r *http.Request) {
	adminID := getAdminIDFromContext(r)

	var p bannerPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Format data JSON tidak valid")
		return
	}

	banner := domain.Banner{
		Title:        p.Title,
		Placement:    p.Placement,
		BannerType:   p.BannerType,
		ImageURL:     p.ImageURL,
		TargetURL:    p.TargetURL,
		HTMLCode:     p.HTMLCode,
		OpenInNewTab: p.OpenInNewTab,
		IsActive:     p.IsActive,
		SortOrder:    p.SortOrder,
	}

	if p.StartDate != nil && *p.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, *p.StartDate); err == nil {
			banner.StartDate = &t
		}
	}
	if p.EndDate != nil && *p.EndDate != "" {
		if t, err := time.Parse(time.RFC3339, *p.EndDate); err == nil {
			banner.EndDate = &t
		}
	}

	created, err := h.service.CreateBanner(r.Context(), adminID, &banner)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

func (h *LayoutHandler) HandleUpdateBanner(w http.ResponseWriter, r *http.Request) {
	adminID := getAdminIDFromContext(r)
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "ID banner tidak valid")
		return
	}

	var p bannerPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Format data JSON tidak valid")
		return
	}

	banner := domain.Banner{
		ID:           id,
		Title:        p.Title,
		Placement:    p.Placement,
		BannerType:   p.BannerType,
		ImageURL:     p.ImageURL,
		TargetURL:    p.TargetURL,
		HTMLCode:     p.HTMLCode,
		OpenInNewTab: p.OpenInNewTab,
		IsActive:     p.IsActive,
		SortOrder:    p.SortOrder,
	}

	if p.StartDate != nil && *p.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, *p.StartDate); err == nil {
			banner.StartDate = &t
		}
	}
	if p.EndDate != nil && *p.EndDate != "" {
		if t, err := time.Parse(time.RFC3339, *p.EndDate); err == nil {
			banner.EndDate = &t
		}
	}

	updated, err := h.service.UpdateBanner(r.Context(), adminID, &banner)
	if err != nil {
		if errors.Is(err, domain.ErrBannerNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Banner tidak ditemukan")
			return
		}
		writeJSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": updated})
}

func (h *LayoutHandler) HandleDeleteBanner(w http.ResponseWriter, r *http.Request) {
	adminID := getAdminIDFromContext(r)
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "ID banner tidak valid")
		return
	}

	if err := h.service.DeleteBanner(r.Context(), adminID, id); err != nil {
		if errors.Is(err, domain.ErrBannerNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Banner tidak ditemukan")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal menghapus banner")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Banner berhasil dihapus"})
}

func (h *LayoutHandler) HandleToggleActive(w http.ResponseWriter, r *http.Request) {
	adminID := getAdminIDFromContext(r)
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "ID banner tidak valid")
		return
	}

	var p struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Format data JSON tidak valid")
		return
	}

	if err := h.service.ToggleActive(r.Context(), adminID, id, p.IsActive); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memperbarui status banner")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Status banner berhasil diperbarui", "is_active": p.IsActive})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func getAdminIDFromContext(r *http.Request) int64 {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		return 0
	}
	return admin.ID
}
