package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"pidiekab/internal/modules/menu/application"
	"pidiekab/internal/modules/menu/domain"
	"pidiekab/internal/platform/httpserver/middleware"
)

type MenuHandler struct {
	service     *application.MenuService
	authService middleware.SessionAuthenticator
}

func NewMenuHandler(service *application.MenuService, authService middleware.SessionAuthenticator) *MenuHandler {
	return &MenuHandler{
		service:     service,
		authService: authService,
	}
}

func (h *MenuHandler) RegisterRoutes(mux *http.ServeMux) {
	authMiddleware := middleware.RequireAdminSession(h.authService)
	csrfMiddleware := middleware.CSRFProtection

	wrapProtected := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(csrfMiddleware(handler))
	}

	wrapGet := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(handler)
	}

	// Public active menus
	mux.HandleFunc("GET /api/v1/menus", h.HandlePublicListMenus)

	// Admin protected routes
	mux.Handle("GET /api/v1/admin/menus", wrapGet(h.HandleListMenus))
	mux.Handle("POST /api/v1/admin/menus", wrapProtected(h.HandleCreateMenu))
	mux.Handle("GET /api/v1/admin/menus/{id}", wrapGet(h.HandleGetMenu))
	mux.Handle("PUT /api/v1/admin/menus/{id}", wrapProtected(h.HandleUpdateMenu))
	mux.Handle("DELETE /api/v1/admin/menus/{id}", wrapProtected(h.HandleDeleteMenu))
	mux.Handle("POST /api/v1/admin/menus/reorder", wrapProtected(h.HandleReorderMenus))
}

type menuPayload struct {
	TitleID      string `json:"title_id"`
	TitleEN      string `json:"title_en"`
	ItemType     string `json:"item_type"`
	TargetID     *int64 `json:"target_id"`
	ExternalURL  string `json:"external_url"`
	OpenInNewTab bool   `json:"open_in_new_tab"`
	SortOrder    int    `json:"sort_order"`
	IsActive     bool   `json:"is_active"`
}

type reorderPayload struct {
	IDs []int64 `json:"ids"`
}

func (h *MenuHandler) HandlePublicListMenus(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListMenus(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat daftar menu")
		return
	}
	var active []domain.MenuItem
	for _, it := range items {
		if it.IsActive {
			active = append(active, it)
		}
	}
	if active == nil {
		active = []domain.MenuItem{}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": active})
}

func (h *MenuHandler) HandleListMenus(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi admin diperlukan")
		return
	}

	items, err := h.service.ListMenus(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil daftar menu")
		return
	}

	if items == nil {
		items = []domain.MenuItem{}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": items})
}

func (h *MenuHandler) HandleGetMenu(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi admin diperlukan")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "ID menu tidak valid")
		return
	}

	item, err := h.service.GetMenuByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrMenuNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Menu tidak ditemukan")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil detail menu")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": item})
}

func (h *MenuHandler) HandleCreateMenu(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi admin diperlukan")
		return
	}

	var req menuPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Payload JSON tidak valid")
		return
	}

	item := &domain.MenuItem{
		TitleID:      req.TitleID,
		TitleEN:      req.TitleEN,
		ItemType:     req.ItemType,
		TargetID:     req.TargetID,
		ExternalURL:  req.ExternalURL,
		OpenInNewTab: req.OpenInNewTab,
		SortOrder:    req.SortOrder,
		IsActive:     req.IsActive,
	}

	created, err := h.service.CreateMenu(r.Context(), admin.ID, item)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": created})
}

func (h *MenuHandler) HandleUpdateMenu(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi admin diperlukan")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "ID menu tidak valid")
		return
	}

	var req menuPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Payload JSON tidak valid")
		return
	}

	item := &domain.MenuItem{
		ID:           id,
		TitleID:      req.TitleID,
		TitleEN:      req.TitleEN,
		ItemType:     req.ItemType,
		TargetID:     req.TargetID,
		ExternalURL:  req.ExternalURL,
		OpenInNewTab: req.OpenInNewTab,
		SortOrder:    req.SortOrder,
		IsActive:     req.IsActive,
	}

	updated, err := h.service.UpdateMenu(r.Context(), admin.ID, item)
	if err != nil {
		if errors.Is(err, domain.ErrMenuNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Menu tidak ditemukan")
			return
		}
		writeJSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": updated})
}

func (h *MenuHandler) HandleDeleteMenu(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi admin diperlukan")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "ID menu tidak valid")
		return
	}

	if err := h.service.DeleteMenu(r.Context(), admin.ID, id); err != nil {
		if errors.Is(err, domain.ErrMenuNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Menu tidak ditemukan")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal menghapus menu")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "Menu berhasil dihapus"})
}

func (h *MenuHandler) HandleReorderMenus(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi admin diperlukan")
		return
	}

	var req reorderPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Payload JSON tidak valid")
		return
	}

	if err := h.service.ReorderMenus(r.Context(), admin.ID, req.IDs); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengubah urutan menu")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "Urutan menu berhasil diperbarui"})
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
