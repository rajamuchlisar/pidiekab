package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"pidiekab/internal/modules/media/application"
	mediaDomain "pidiekab/internal/modules/media/domain"
	"pidiekab/internal/modules/media/infrastructure/postgres"
	"pidiekab/internal/modules/media/infrastructure/storage"
	"pidiekab/internal/platform/httpserver/middleware"
)

type MediaHandler struct {
	service     *application.MediaService
	authService middleware.SessionAuthenticator
}

func NewMediaHandler(service *application.MediaService, authService middleware.SessionAuthenticator) *MediaHandler {
	return &MediaHandler{
		service:     service,
		authService: authService,
	}
}

func (h *MediaHandler) RegisterRoutes(mux *http.ServeMux) {
	// Public media serving
	mux.HandleFunc("GET /media/", h.HandleServeMedia)

	authMiddleware := middleware.RequireAdminSession(h.authService)
	csrfMiddleware := middleware.CSRFProtection

	wrapProtected := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(csrfMiddleware(handler))
	}

	wrapGet := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(handler)
	}

	// Admin media library endpoints (behind RequireAdminSession)
	mux.Handle("GET /api/v1/admin/media", wrapGet(h.HandleListMedia))
	mux.Handle("POST /api/v1/admin/media", wrapProtected(h.HandleUploadMedia))
	mux.Handle("POST /api/v1/admin/media/upload", wrapProtected(h.HandleUploadMedia))
	mux.Handle("GET /api/v1/admin/media/{id}", wrapGet(h.HandleGetMedia))
	mux.Handle("PUT /api/v1/admin/media/{id}", wrapProtected(h.HandleUpdateMetadata))
	mux.Handle("DELETE /api/v1/admin/media/{id}", wrapProtected(h.HandleDeleteMedia))
}

func (h *MediaHandler) HandleUploadMedia(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	// Limit multipart request size to MaxMediaSizeBytes + 1MB for headers
	r.Body = http.MaxBytesReader(w, r.Body, mediaDomain.MaxMediaSizeBytes+1024*1024)
	if err := r.ParseMultipartForm(mediaDomain.MaxMediaSizeBytes); err != nil {
		writeJSONError(w, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "Upload exceeds maximum allowed size")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Missing file field in multipart form")
		return
	}
	defer file.Close()

	altID := r.FormValue("alt_id")
	if altID == "" {
		altID = r.FormValue("alt_text_id")
	}
	altEN := r.FormValue("alt_en")
	if altEN == "" {
		altEN = r.FormValue("alt_text_en")
	}
	sourceURL := r.FormValue("source")
	if sourceURL == "" {
		sourceURL = r.FormValue("source_origin")
	}
	if sourceURL == "" {
		sourceURL = r.FormValue("source_url")
	}
	rightsStatus := r.FormValue("license")
	if rightsStatus == "" {
		rightsStatus = r.FormValue("rights_status")
	}
	if rightsStatus == "" {
		rightsStatus = "rights_cleared"
	}
	attribution := r.FormValue("attribution")
	if attribution == "" {
		attribution = r.FormValue("title")
	}

	meta := application.UploadMetadata{
		AltTextID:    altID,
		AltTextEN:    altEN,
		SourceURL:    sourceURL,
		RightsStatus: rightsStatus,
		Attribution:  attribution,
	}

	asset, err := h.service.Upload(r.Context(), admin.ID, header.Filename, file, meta)
	if err != nil {
		if errors.Is(err, mediaDomain.ErrUnsupportedMediaType) {
			writeJSONError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", err.Error())
			return
		}
		if errors.Is(err, mediaDomain.ErrFileTooLarge) {
			writeJSONError(w, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", err.Error())
			return
		}
		if errors.Is(err, mediaDomain.ErrInvalidImage) {
			writeJSONError(w, http.StatusBadRequest, "INVALID_IMAGE", err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process upload")
		return
	}

	mediaURL := "/media/" + asset.StorageKey
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data":  asset,
		"asset": asset,
		"id":    asset.ID,
		"url":   mediaURL,
	})
}

func (h *MediaHandler) HandleListMedia(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}
	q := r.URL.Query().Get("q")

	assets, total, err := h.service.ListMedia(r.Context(), page, perPage, q)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list media")
		return
	}

	type mediaItemDTO struct {
		mediaDomain.MediaAsset
		URL string `json:"url"`
	}

	itemsDTO := make([]mediaItemDTO, 0, len(assets))
	for _, a := range assets {
		itemsDTO = append(itemsDTO, mediaItemDTO{
			MediaAsset: a,
			URL:        "/media/" + a.StorageKey,
		})
	}

	totalPages := (total + perPage - 1) / perPage
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data":        itemsDTO,
		"items":       itemsDTO,
		"total":       total,
		"total_items": total,
		"meta": map[string]any{
			"page":        page,
			"per_page":    perPage,
			"total_items": total,
			"total_pages": totalPages,
		},
	})
}

func (h *MediaHandler) HandleGetMedia(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AdminFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "Invalid media ID")
		return
	}

	asset, err := h.service.GetMedia(r.Context(), id)
	if err != nil {
		if errors.Is(err, mediaDomain.ErrMediaNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Media asset not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get media")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data":  asset,
		"asset": asset,
		"url":   "/media/" + asset.StorageKey,
	})
}

type updateMetadataRequest struct {
	AltTextID    string `json:"alt_text_id"`
	AltTextEN    string `json:"alt_text_en"`
	SourceURL    string `json:"source_url"`
	RightsStatus string `json:"rights_status"`
	Attribution  string `json:"attribution"`
}

func (h *MediaHandler) HandleUpdateMetadata(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "Invalid media ID")
		return
	}

	var req updateMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON body")
		return
	}

	params := postgres.UpdateMetadataParams{
		AltTextID:    strings.TrimSpace(req.AltTextID),
		AltTextEN:    strings.TrimSpace(req.AltTextEN),
		SourceURL:    strings.TrimSpace(req.SourceURL),
		RightsStatus: strings.TrimSpace(req.RightsStatus),
		Attribution:  strings.TrimSpace(req.Attribution),
	}

	updated, err := h.service.UpdateMetadata(r.Context(), admin.ID, id, params)
	if err != nil {
		if errors.Is(err, mediaDomain.ErrMediaNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Media asset not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update metadata")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": updated,
	})
}

func (h *MediaHandler) HandleDeleteMedia(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.AdminFromContext(r.Context())
	if !ok || admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session required")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ID", "Invalid media ID")
		return
	}

	err = h.service.DeleteMedia(r.Context(), admin.ID, id)
	if err != nil {
		if errors.Is(err, mediaDomain.ErrMediaNotFound) {
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Media asset not found")
			return
		}
		if errors.Is(err, mediaDomain.ErrMediaInUse) {
			writeJSONError(w, http.StatusConflict, "MEDIA_IN_USE", "Cannot delete media asset that is currently referenced by content or site settings")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete media asset")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MediaHandler) HandleServeMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Extract path after "/media/"
	rawKey := strings.TrimPrefix(r.URL.Path, "/media/")
	if rawKey == "" || strings.Contains(rawKey, "..") {
		http.NotFound(w, r)
		return
	}

	stream, asset, err := h.service.OpenMediaStream(r.Context(), rawKey)
	if err != nil {
		if errors.Is(err, storage.ErrPathTraversal) || errors.Is(err, mediaDomain.ErrMediaNotFound) || errors.Is(err, storage.ErrFileNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Error reading media", http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	// Content Type
	mimeType := "application/octet-stream"
	if asset != nil && asset.MimeType != "" {
		mimeType = asset.MimeType
	} else {
		ext := strings.ToLower(filepath.Ext(rawKey))
		switch ext {
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".png":
			mimeType = "image/png"
		case ".webp":
			mimeType = "image/webp"
		}
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	if r.Method == http.MethodHead {
		return
	}

	_, _ = io.Copy(w, stream)
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
