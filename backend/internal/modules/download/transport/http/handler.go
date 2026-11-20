package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	identityApp "pidiekab/internal/modules/identity/application"
	"pidiekab/internal/modules/download/application"
	"pidiekab/internal/modules/download/domain"
	"pidiekab/internal/modules/download/infrastructure/postgres"
	"pidiekab/internal/platform/httpserver/middleware"
)

type FileStorage interface {
	Save(ctx context.Context, ext string, r io.Reader) (storageKey string, bytesWritten int64, err error)
}

type DownloadHandler struct {
	service     *application.DownloadService
	authService *identityApp.AuthService
	storage     FileStorage
}

func NewDownloadHandler(service *application.DownloadService, authService *identityApp.AuthService, storage FileStorage) *DownloadHandler {
	return &DownloadHandler{
		service:     service,
		authService: authService,
		storage:     storage,
	}
}

func (h *DownloadHandler) RegisterRoutes(mux *http.ServeMux) {
	// Public routes
	mux.HandleFunc("GET /api/v1/public/downloads", h.HandlePublicListDownloads)
	mux.HandleFunc("GET /api/v1/public/downloads/{id}/file", h.HandlePublicDownloadFile)

	authMiddleware := middleware.RequireAdminSession(h.authService)
	csrfMiddleware := middleware.CSRFProtection

	wrapProtected := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(csrfMiddleware(handler))
	}

	wrapGet := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(handler)
	}

	// Admin routes
	mux.Handle("GET /api/v1/admin/downloads", wrapGet(h.HandleListDownloads))
	mux.Handle("GET /api/v1/admin/downloads/{id}", wrapGet(h.HandleGetDownload))
	mux.Handle("POST /api/v1/admin/downloads", wrapProtected(h.HandleCreateDownload))
	mux.Handle("PUT /api/v1/admin/downloads/{id}", wrapProtected(h.HandleUpdateDownload))
	mux.Handle("DELETE /api/v1/admin/downloads/{id}", wrapProtected(h.HandleDeleteDownload))
	mux.Handle("POST /api/v1/admin/downloads/upload", wrapProtected(h.HandleUploadDownloadFile))
}

func (h *DownloadHandler) HandlePublicListDownloads(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	active := true
	filter := postgres.ListFilter{
		Category: q.Get("category"),
		IsActive: &active,
		Search:   q.Get("q"),
		Limit:    100,
	}

	items, total, err := h.service.ListDownloads(r.Context(), filter)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve downloads")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"items": items,
		"total": total,
	})
}

func (h *DownloadHandler) HandlePublicDownloadFile(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}

	item, err := h.service.TrackDownload(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrDownloadNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Download error", http.StatusInternalServerError)
		return
	}

	if !item.IsActive {
		http.NotFound(w, r)
		return
	}

	// Redirect to target file_url
	http.Redirect(w, r, item.FileURL, http.StatusFound)
}

func (h *DownloadHandler) HandleListDownloads(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 50
	if l := q.Get("per_page"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	offset := 0
	if p := q.Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			offset = (val - 1) * limit
		}
	}

	filter := postgres.ListFilter{
		Category: q.Get("category"),
		Search:   q.Get("q"),
		Limit:    limit,
		Offset:   offset,
	}

	if act := q.Get("is_active"); act != "" {
		b := act == "true" || act == "1"
		filter.IsActive = &b
	}

	items, total, err := h.service.ListDownloads(r.Context(), filter)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list downloads")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"items": items,
		"total": total,
	})
}

func (h *DownloadHandler) HandleGetDownload(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "Invalid download ID")
		return
	}

	item, err := h.service.GetDownload(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrDownloadNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Download item not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get download item")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": item,
	})
}

type downloadRequest struct {
	TitleID       string `json:"title_id"`
	TitleEN       string `json:"title_en"`
	Category      string `json:"category"`
	DescriptionID string `json:"description_id"`
	DescriptionEN string `json:"description_en"`
	FileURL       string `json:"file_url"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	FileType      string `json:"file_type"`
	IsActive      bool   `json:"is_active"`
	SortOrder     int    `json:"sort_order"`
}

func (h *DownloadHandler) HandleCreateDownload(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	var req downloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON body")
		return
	}

	item := &domain.DownloadItem{
		TitleID:       req.TitleID,
		TitleEN:       req.TitleEN,
		Category:      req.Category,
		DescriptionID: req.DescriptionID,
		DescriptionEN: req.DescriptionEN,
		FileURL:       req.FileURL,
		FileName:      req.FileName,
		FileSize:      req.FileSize,
		FileType:      req.FileType,
		IsActive:      req.IsActive,
		SortOrder:     req.SortOrder,
	}

	created, err := h.service.CreateDownload(r.Context(), admin.ID, item)
	if err != nil {
		if errors.Is(err, domain.ErrTitleRequired) || errors.Is(err, domain.ErrFileURLRequired) {
			writeJSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create download item")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": created,
	})
}

func (h *DownloadHandler) HandleUpdateDownload(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "Invalid download ID")
		return
	}

	var req downloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON body")
		return
	}

	item := &domain.DownloadItem{
		ID:            id,
		TitleID:       req.TitleID,
		TitleEN:       req.TitleEN,
		Category:      req.Category,
		DescriptionID: req.DescriptionID,
		DescriptionEN: req.DescriptionEN,
		FileURL:       req.FileURL,
		FileName:      req.FileName,
		FileSize:      req.FileSize,
		FileType:      req.FileType,
		IsActive:      req.IsActive,
		SortOrder:     req.SortOrder,
	}

	updated, err := h.service.UpdateDownload(r.Context(), admin.ID, item)
	if err != nil {
		if errors.Is(err, domain.ErrDownloadNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Download item not found")
			return
		}
		if errors.Is(err, domain.ErrTitleRequired) || errors.Is(err, domain.ErrFileURLRequired) {
			writeJSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update download item")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": updated,
	})
}

func (h *DownloadHandler) HandleDeleteDownload(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "Invalid download ID")
		return
	}

	if err := h.service.DeleteDownload(r.Context(), admin.ID, id); err != nil {
		if errors.Is(err, domain.ErrDownloadNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Download item not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete download item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DownloadHandler) HandleUploadDownloadFile(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	if h.storage == nil {
		writeJSONError(w, http.StatusInternalServerError, "STORAGE_NOT_CONFIGURED", "Storage is not configured")
		return
	}

	// 50 MB max limit for documents & downloads
	_ = r.ParseMultipartForm(50 * 1024 * 1024)
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "FILE_REQUIRED", "Form field 'file' is required")
		return
	}
	defer file.Close()

	ext := strings.TrimPrefix(filepath.Ext(header.Filename), ".")
	storageKey, size, err := h.storage.Save(r.Context(), ext, file)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "STORAGE_ERROR", "Failed to save uploaded file")
		return
	}

	fileType := strings.ToUpper(ext)
	if fileType == "" {
		fileType = "FILE"
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"file_url":  "/media/" + storageKey,
		"file_name": header.Filename,
		"file_size": size,
		"file_type": fileType,
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
