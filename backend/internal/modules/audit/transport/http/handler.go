package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"pidiekab/internal/modules/audit/domain"
	auditPostgres "pidiekab/internal/modules/audit/infrastructure/postgres"
	"pidiekab/internal/modules/identity/application"
	"pidiekab/internal/platform/httpserver/middleware"
)

type AuditHandler struct {
	repo        *auditPostgres.Repository
	authService *application.AuthService
}

func NewAuditHandler(repo *auditPostgres.Repository, authService *application.AuthService) *AuditHandler {
	return &AuditHandler{
		repo:        repo,
		authService: authService,
	}
}

func (h *AuditHandler) RegisterRoutes(mux *http.ServeMux) {
	authMW := middleware.RequireAdminSession(h.authService)

	mux.Handle("GET /api/v1/admin/audit-logs", authMW(http.HandlerFunc(h.handleList)))
}

func (h *AuditHandler) handleList(w http.ResponseWriter, r *http.Request) {
	limit := 20
	page := 1

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	offset := (page - 1) * limit

	var actionFilter *string
	if a := r.URL.Query().Get("action"); a != "" {
		actionFilter = &a
	}

	var entityTypeFilter *string
	if et := r.URL.Query().Get("entity_type"); et != "" {
		entityTypeFilter = &et
	}

	var adminIDFilter *int64
	if aid := r.URL.Query().Get("admin_id"); aid != "" {
		if parsed, err := strconv.ParseInt(aid, 10, 64); err == nil {
			adminIDFilter = &parsed
		}
	}

	filter := domain.ListFilter{
		Limit:      limit,
		Offset:     offset,
		Action:     actionFilter,
		EntityType: entityTypeFilter,
		AdminID:    adminIDFilter,
	}

	items, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to list audit logs"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"logs":  items,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
