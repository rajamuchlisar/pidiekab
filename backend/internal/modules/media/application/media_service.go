package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	auditDomain "pidiekab/internal/modules/audit/domain"
	"pidiekab/internal/modules/media/domain"
	"pidiekab/internal/modules/media/infrastructure/postgres"
	"pidiekab/internal/modules/media/infrastructure/storage"
)

type AuditRecorder interface {
	Record(ctx context.Context, entry auditDomain.Log) error
}

type MediaService struct {
	repo    *postgres.MediaRepository
	storage storage.Storage
	audit   AuditRecorder
}

func NewMediaService(repo *postgres.MediaRepository, st storage.Storage, audit AuditRecorder) *MediaService {
	return &MediaService{
		repo:    repo,
		storage: st,
		audit:   audit,
	}
}

type UploadMetadata struct {
	AltTextID    string
	AltTextEN    string
	SourceURL    string
	RightsStatus string
	Attribution  string
}

func (s *MediaService) Upload(ctx context.Context, adminID int64, filename string, r io.Reader, meta UploadMetadata) (*domain.MediaAsset, error) {
	// Read with limit (MaxMediaSizeBytes + 1 to detect overflow)
	lr := io.LimitReader(r, domain.MaxMediaSizeBytes+1)
	buf, err := io.ReadAll(lr)
	if err != nil {
		return nil, fmt.Errorf("failed to read upload stream: %w", err)
	}

	if int64(len(buf)) > domain.MaxMediaSizeBytes {
		return nil, domain.ErrFileTooLarge
	}

	mimeType, width, height, err := domain.SniffAndValidateImage(buf, filename)
	if err != nil {
		return nil, err
	}

	// Compress and optimize image to web standards (max 1920px Full HD, quality 82)
	buf, width, height, mimeType = OptimizeImage(buf, mimeType, width, height)

	// Compute SHA-256 for integrity check
	hasher := sha256.New()
	hasher.Write(buf)
	checksum := hex.EncodeToString(hasher.Sum(nil))

	ext := filepath.Ext(filename)
	if ext == "" || (mimeType == domain.MimeJPEG && ext != ".jpg" && ext != ".jpeg") {
		switch mimeType {
		case domain.MimeJPEG:
			ext = ".jpg"
		case domain.MimePNG:
			ext = ".png"
		case domain.MimeWebP:
			ext = ".webp"
		}
	}

	storageKey, size, err := s.storage.Save(ctx, ext, bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("failed to store media file: %w", err)
	}

	asset := &domain.MediaAsset{
		StorageKey:       storageKey,
		OriginalFilename: filename,
		MimeType:         mimeType,
		FileSize:         size,
		Width:            &width,
		Height:           &height,
		AltTextID:        meta.AltTextID,
		AltTextEN:        meta.AltTextEN,
		SourceURL:        meta.SourceURL,
		RightsStatus:     meta.RightsStatus,
		Attribution:      meta.Attribution,
		UploadedBy:       &adminID,
	}

	created, err := s.repo.Create(ctx, asset)
	if err != nil {
		// Rollback storage file on DB error
		_ = s.storage.Delete(ctx, storageKey)
		return nil, err
	}

	s.recordAudit(ctx, &adminID, "media.upload", "media_asset", &created.ID, map[string]any{
		"storage_key":       storageKey,
		"original_filename": filename,
		"mime_type":         mimeType,
		"file_size":         size,
		"sha256":            checksum,
	}, nil)

	return created, nil
}

func (s *MediaService) GetMedia(ctx context.Context, id int64) (*domain.MediaAsset, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MediaService) GetMediaByStorageKey(ctx context.Context, key string) (*domain.MediaAsset, error) {
	return s.repo.GetByStorageKey(ctx, key)
}

func (s *MediaService) UpdateMetadata(ctx context.Context, adminID int64, id int64, params postgres.UpdateMetadataParams) (*domain.MediaAsset, error) {
	updated, err := s.repo.UpdateMetadata(ctx, id, params)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, &adminID, "media.update_metadata", "media_asset", &id, map[string]any{
		"alt_text_id":   params.AltTextID,
		"alt_text_en":   params.AltTextEN,
		"rights_status": params.RightsStatus,
		"attribution":   params.Attribution,
	}, nil)

	return updated, nil
}

func (s *MediaService) DeleteMedia(ctx context.Context, adminID int64, id int64) error {
	asset, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from DB first (handles foreign key in-use constraints)
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Delete storage file
	_ = s.storage.Delete(ctx, asset.StorageKey)

	s.recordAudit(ctx, &adminID, "media.delete", "media_asset", &id, map[string]any{
		"storage_key": asset.StorageKey,
	}, nil)

	return nil
}

func (s *MediaService) recordAudit(ctx context.Context, adminID *int64, action, entityType string, entityID *int64, meta map[string]any, ip *string) {
	if s.audit == nil {
		return
	}
	_ = s.audit.Record(ctx, auditDomain.Log{
		AdminID:    adminID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   meta,
		IPAddress:  ip,
		CreatedAt:  time.Now().UTC(),
	})
}

func (s *MediaService) ListMedia(ctx context.Context, page, perPage int, q string) ([]domain.MediaAsset, int, error) {
	return s.repo.List(ctx, page, perPage, strings.TrimSpace(q))
}

func (s *MediaService) OpenMediaStream(ctx context.Context, storageKey string) (io.ReadCloser, *domain.MediaAsset, error) {
	asset, err := s.repo.GetByStorageKey(ctx, storageKey)
	if err != nil {
		if errors.Is(err, domain.ErrMediaNotFound) {
			// Check if file exists directly
			stream, openErr := s.storage.Open(ctx, storageKey)
			if openErr != nil {
				return nil, nil, domain.ErrMediaNotFound
			}
			return stream, nil, nil
		}
		return nil, nil, err
	}

	stream, err := s.storage.Open(ctx, storageKey)
	if err != nil {
		return nil, nil, err
	}

	return stream, asset, nil
}
