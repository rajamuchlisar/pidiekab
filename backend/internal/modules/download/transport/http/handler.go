package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	identityApp "pidiekab/internal/modules/identity/application"
	"pidiekab/internal/modules/download"
	"pidiekab/internal/modules/download/application"
	"pidiekab/internal/modules/download/domain"
	"pidiekab/internal/modules/download/infrastructure/postgres"
	"pidiekab/internal/platform/httpserver/middleware"
)

type FileStorage interface {
	Save(ctx context.Context, ext string, r io.Reader) (storageKey string, bytesWritten int64, err error)
	Open(ctx context.Context, storageKey string) (io.ReadCloser, error)
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

func sanitizeFilename(name string) string {
	clean := filepath.Base(name)
	clean = strings.ReplaceAll(clean, `"`, "")
	clean = strings.ReplaceAll(clean, "\r", "")
	clean = strings.ReplaceAll(clean, "\n", "")
	if clean == "" || clean == "." {
		return "dokumen-pidie.pdf"
	}
	return clean
}

func detectMimeType(filename, fileType string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".zip":
		return "application/zip"
	case ".csv":
		return "text/csv; charset=utf-8"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	}
	if mt := mime.TypeByExtension(ext); mt != "" {
		return mt
	}
	switch strings.ToUpper(fileType) {
	case "PDF":
		return "application/pdf"
	case "DOC", "DOCX":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "XLS", "XLSX":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "ZIP":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
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

	fileURL := strings.TrimSpace(item.FileURL)

	// 1. If external URL (pointing outside pidiekab.web.id/localhost), redirect directly
	if (strings.HasPrefix(fileURL, "http://") || strings.HasPrefix(fileURL, "https://")) &&
		!strings.Contains(fileURL, "pidiekab.web.id") &&
		!strings.Contains(fileURL, "localhost") &&
		!strings.Contains(fileURL, "127.0.0.1") {
		http.Redirect(w, r, fileURL, http.StatusFound)
		return
	}

	// 2. Prepare filename & MIME type
	filename := strings.TrimSpace(item.FileName)
	if filename == "" {
		filename = filepath.Base(fileURL)
	}
	if filename == "" || filename == "." || filename == "/" {
		filename = fmt.Sprintf("dokumen-pidie-%d.pdf", item.ID)
	}
	if filepath.Ext(filename) == "" && item.FileType != "" {
		filename = filename + "." + strings.ToLower(item.FileType)
	}

	mimeType := detectMimeType(filename, item.FileType)

	// 3. Try opening from local storage
	var stream io.ReadCloser
	if h.storage != nil {
		storageKey := ""
		if strings.Contains(fileURL, "/media/") {
			parts := strings.Split(fileURL, "/media/")
			if len(parts) > 1 {
				storageKey = parts[len(parts)-1]
			}
		} else if !strings.HasPrefix(fileURL, "http://") && !strings.HasPrefix(fileURL, "https://") && !strings.HasPrefix(fileURL, "/") {
			storageKey = fileURL
		}

		if storageKey != "" {
			if s, err := h.storage.Open(r.Context(), storageKey); err == nil {
				stream = s
			}
		}
	}

	// If file exists in storage, stream it directly to browser
	if stream != nil {
		defer stream.Close()
		w.Header().Set("Content-Type", mimeType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "public, max-age=86400")

		// Relax restrictive security headers so browser PDF viewer and media render properly
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Content-Security-Policy", "default-src 'self' blob:; object-src 'self' blob:; img-src 'self' data: https: blob:; frame-ancestors 'self'")
		w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
		w.Header().Del("Cross-Origin-Opener-Policy")

		if r.URL.Query().Get("download") == "1" {
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", sanitizeFilename(filename)))
		} else if strings.HasSuffix(strings.ToLower(filename), ".pdf") || mimeType == "application/pdf" {
			w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", sanitizeFilename(filename)))
		} else {
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", sanitizeFilename(filename)))
		}

		if seeker, ok := stream.(io.ReadSeeker); ok {
			modTime := item.UpdatedAt
			if modTime.IsZero() {
				modTime = item.CreatedAt
			}
			if modTime.IsZero() {
				modTime = time.Now()
			}
			http.ServeContent(w, r, sanitizeFilename(filename), modTime, seeker)
			return
		}

		if item.FileSize > 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(item.FileSize, 10))
		}
		_, _ = io.Copy(w, stream)
		return
	}

	// 4. Fallback: Generate authentic official Pidie PDF on demand
	pdfBytes := download.GenerateOfficialDocumentPDF(item.TitleID, item.Category, item.DescriptionID, filename)

	// Cache generated document into storage and update item
	if h.storage != nil {
		if key, size, saveErr := h.storage.Save(r.Context(), "pdf", bytes.NewReader(pdfBytes)); saveErr == nil {
			item.FileURL = "/media/" + key
			item.FileSize = size
			_, _ = h.service.UpdateDownloadInternal(r.Context(), item)
		}
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", sanitizeFilename(filename)))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write(pdfBytes)
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

	// 10 MB max limit enforced at request level
	r.Body = http.MaxBytesReader(w, r.Body, download.MaxDownloadFileSizeBytes+1024*1024)
	if err := r.ParseMultipartForm(download.MaxDownloadFileSizeBytes); err != nil {
		writeJSONError(w, http.StatusBadRequest, "FILE_TOO_LARGE", "Ukuran unggahan melebihi batas maksimum 10 MB")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "FILE_REQUIRED", "Form field 'file' is required")
		return
	}
	defer file.Close()

	if header.Size > download.MaxDownloadFileSizeBytes {
		writeJSONError(w, http.StatusBadRequest, "FILE_TOO_LARGE", "Ukuran berkas melebihi batas maksimum 10 MB")
		return
	}

	rawBytes, err := io.ReadAll(io.LimitReader(file, download.MaxDownloadFileSizeBytes+1))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "READ_ERROR", "Gagal membaca berkas unggahan")
		return
	}
	if len(rawBytes) > download.MaxDownloadFileSizeBytes {
		writeJSONError(w, http.StatusBadRequest, "FILE_TOO_LARGE", "Ukuran berkas melebihi batas maksimum 10 MB")
		return
	}

	optimizedBytes, fileType, finalSize, err := download.OptimizeDownloadFile(header.Filename, rawBytes)
	if err != nil {
		if errors.Is(err, download.ErrFileTooLarge) {
			writeJSONError(w, http.StatusBadRequest, "FILE_TOO_LARGE", err.Error())
			return
		}
		if errors.Is(err, download.ErrDisallowedExtension) {
			writeJSONError(w, http.StatusBadRequest, "DISALLOWED_FILE_TYPE", err.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, "OPTIMIZATION_ERROR", err.Error())
		return
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(header.Filename), "."))
	storageKey, sizeWritten, err := h.storage.Save(r.Context(), ext, bytes.NewReader(optimizedBytes))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "STORAGE_ERROR", "Gagal menyimpan berkas ke storage")
		return
	}

	if finalSize <= 0 {
		finalSize = sizeWritten
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"file_url":  "/media/" + storageKey,
		"file_name": header.Filename,
		"file_size": finalSize,
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
