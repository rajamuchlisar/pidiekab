package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"pidiekab/internal/modules/search/application"
	"pidiekab/internal/modules/search/domain"
)

type SearchHandler struct {
	service *application.SearchService
}

func NewSearchHandler(service *application.SearchService) *SearchHandler {
	return &SearchHandler{service: service}
}

func (h *SearchHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/public/{locale}/search", h.HandleSearch)
}

func (h *SearchHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	locale := r.PathValue("locale")
	q := r.URL.Query().Get("q")
	category := strings.TrimSpace(r.URL.Query().Get("category"))

	var categorySlug *string
	if category != "" {
		categorySlug = &category
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		} else {
			writeProblem(w, r, http.StatusBadRequest, "INVALID_PAGINATION", "Invalid page number.")
			return
		}
	}

	perPage := 12
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 50 {
			perPage = pp
		} else {
			writeProblem(w, r, http.StatusBadRequest, "INVALID_PAGINATION", "per_page must be between 1 and 50.")
			return
		}
	}

	input := application.SearchInput{
		Locale:       locale,
		Query:        q,
		CategorySlug: categorySlug,
		Page:         page,
		PerPage:      perPage,
	}

	res, err := h.service.Search(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidLocale) {
			writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", "Locale must be 'id' or 'en'.")
			return
		}
		if errors.Is(err, domain.ErrQueryTooShort) || errors.Is(err, domain.ErrQueryTooLong) {
			writeProblem(w, r, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Search query failed.")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=60")
	_ = json.NewEncoder(w).Encode(res)
}

func writeProblem(w http.ResponseWriter, r *http.Request, status int, code, detail string) {
	reqID := r.Header.Get("X-Request-Id")
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(status)

	body := map[string]any{
		"type":       "about:blank",
		"title":      http.StatusText(status),
		"status":     status,
		"code":       code,
		"detail":     detail,
		"request_id": reqID,
	}

	_ = json.NewEncoder(w).Encode(body)
}

