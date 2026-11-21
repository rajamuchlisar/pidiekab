package http

import (
	"encoding/json"
	"net/http"

	"pidiekab/internal/modules/analytics/application"
	"pidiekab/internal/platform/httpserver/middleware"
)

type AnalyticsHandler struct {
	service     *application.AnalyticsService
	authService middleware.SessionAuthenticator
}

func NewAnalyticsHandler(service *application.AnalyticsService, authService middleware.SessionAuthenticator) *AnalyticsHandler {
	return &AnalyticsHandler{
		service:     service,
		authService: authService,
	}
}

func (h *AnalyticsHandler) TrackRequest(r *http.Request) {
	h.service.TrackRequest(r)
}

func (h *AnalyticsHandler) RegisterRoutes(mux *http.ServeMux) {
	authMiddleware := middleware.RequireAdminSession(h.authService)

	// Admin stats endpoint
	mux.Handle("GET /api/v1/admin/dashboard/stats", authMiddleware(http.HandlerFunc(h.HandleGetDashboardStats)))

	// Public tracking beacon endpoint
	mux.HandleFunc("POST /api/v1/public/analytics/track", h.HandleTrackPublic)
}

func (h *AnalyticsHandler) HandleGetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetDashboardStats(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to retrieve dashboard analytics",
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": stats,
	})
}

func (h *AnalyticsHandler) HandleTrackPublic(w http.ResponseWriter, r *http.Request) {
	h.service.TrackRequest(r)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
