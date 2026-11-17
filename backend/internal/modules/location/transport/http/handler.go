package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"pidiekab/internal/modules/identity/application"
	"pidiekab/internal/modules/location/domain"
	"pidiekab/internal/modules/location/infrastructure/postgres"
	"pidiekab/internal/platform/httpserver/middleware"
)

// Handler manages HTTP transport for location master data.
type Handler struct {
	repo        *postgres.Repository
	authService *application.AuthService
}

// NewHandler creates a new location transport Handler.
func NewHandler(repo *postgres.Repository, authService *application.AuthService) *Handler {
	return &Handler{
		repo:        repo,
		authService: authService,
	}
}

// RegisterRoutes registers location endpoints onto the ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	authMiddleware := middleware.RequireAdminSession(h.authService)

	mux.Handle("GET /api/v1/admin/locations/districts", authMiddleware(http.HandlerFunc(h.handleListDistricts)))
	mux.Handle("GET /api/v1/admin/locations/villages", authMiddleware(http.HandlerFunc(h.handleListVillages)))
}

func (h *Handler) handleListDistricts(w http.ResponseWriter, r *http.Request) {
	districts, err := h.repo.ListDistricts(r.Context())
	if err != nil {
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve districts.")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": districts,
	})
}

func (h *Handler) handleListVillages(w http.ResponseWriter, r *http.Request) {
	districtIDStr := r.URL.Query().Get("district_id")
	if districtIDStr == "" {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "district_id query parameter is required.")
		return
	}

	districtID, err := strconv.ParseInt(districtIDStr, 10, 64)
	if err != nil || districtID <= 0 {
		writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "district_id must be a positive integer.")
		return
	}

	villages, err := h.repo.ListVillagesByDistrict(r.Context(), districtID)
	if err != nil {
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve villages.")
		return
	}

	if villages == nil {
		villages = []*domain.Village{} // empty array, not null
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": villages,
	})
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
