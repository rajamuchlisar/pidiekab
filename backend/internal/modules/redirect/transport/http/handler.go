package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"pidiekab/internal/modules/identity/application"
	"pidiekab/internal/modules/redirect/domain"
	redirectPostgres "pidiekab/internal/modules/redirect/infrastructure/postgres"
	"pidiekab/internal/platform/httpserver/middleware"
)

type RedirectHandler struct {
	repo        *redirectPostgres.RedirectRepository
	authService *application.AuthService
}

func NewRedirectHandler(repo *redirectPostgres.RedirectRepository, authService *application.AuthService) *RedirectHandler {
	return &RedirectHandler{
		repo:        repo,
		authService: authService,
	}
}

// FindTargetByOldPath queries the repository for an active redirect.
func (h *RedirectHandler) FindTargetByOldPath(ctx context.Context, oldPath string) (*domain.Redirect, error) {
	return h.repo.FindTargetByOldPath(ctx, oldPath)
}

// FindRedirect satisfies the httpserver.RedirectFinder interface.
func (h *RedirectHandler) FindRedirect(ctx context.Context, path string) (string, int, error) {
	red, err := h.repo.FindTargetByOldPath(ctx, path)
	if err != nil || red == nil {
		return "", 0, err
	}
	return red.NewPath, red.StatusCode, nil
}

func (h *RedirectHandler) RegisterRoutes(mux *http.ServeMux) {
	authMW := middleware.RequireAdminSession(h.authService)
	csrfMW := middleware.CSRFProtection

	// Admin API endpoints
	mux.Handle("GET /api/v1/admin/redirects", authMW(http.HandlerFunc(h.handleList)))
	mux.Handle("POST /api/v1/admin/redirects", authMW(csrfMW(http.HandlerFunc(h.handleCreate))))
	mux.Handle("DELETE /api/v1/admin/redirects/{id}", authMW(csrfMW(http.HandlerFunc(h.handleDelete))))
}

func (h *RedirectHandler) handleList(w http.ResponseWriter, r *http.Request) {
	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 1 {
			offset = (parsed - 1) * limit
		}
	}

	items, total, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to list redirects"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": items,
		"meta": map[string]any{
			"total": total,
			"limit": limit,
		},
	})
}

func (h *RedirectHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}

	var req struct {
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json payload"})
		return
	}

	red := &domain.Redirect{
		OldPath:   req.OldPath,
		NewPath:   req.NewPath,
		CreatedBy: &admin.ID,
	}

	if err := h.repo.Upsert(r.Context(), red); err != nil {
		if err == domain.ErrEmptyPath || err == domain.ErrSelfRedirect {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to save redirect"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": red})
}

func (h *RedirectHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid redirect id"})
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "redirect not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
