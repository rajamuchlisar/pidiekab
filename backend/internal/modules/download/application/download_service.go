package application

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	auditDomain "pidiekab/internal/modules/audit/domain"
	"pidiekab/internal/modules/download/domain"
	"pidiekab/internal/modules/download/infrastructure/postgres"
)

type AuditRecorder interface {
	Record(ctx context.Context, entry auditDomain.Log) error
}

type JobEnqueuer interface {
	Enqueue(ctx context.Context, tx pgx.Tx, jobType string, payload any) (int64, error)
}

type DownloadService struct {
	repo     *postgres.DownloadRepository
	audit    AuditRecorder
	enqueuer JobEnqueuer
}

func NewDownloadService(repo *postgres.DownloadRepository, audit AuditRecorder) *DownloadService {
	return &DownloadService{
		repo:  repo,
		audit: audit,
	}
}

func (s *DownloadService) SetJobEnqueuer(enqueuer JobEnqueuer) {
	s.enqueuer = enqueuer
}

func (s *DownloadService) triggerSSGBuild(ctx context.Context, action string, downloadID int64) {
	if s.enqueuer != nil {
		_, _ = s.enqueuer.Enqueue(ctx, nil, "site.downloads_updated", map[string]any{
			"action":      action,
			"download_id": downloadID,
			"timestamp":   time.Now().Format(time.RFC3339),
		})
	}
}

func (s *DownloadService) ListDownloads(ctx context.Context, f postgres.ListFilter) ([]domain.DownloadItem, int, error) {
	return s.repo.List(ctx, f)
}

func (s *DownloadService) GetDownload(ctx context.Context, id int64) (*domain.DownloadItem, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DownloadService) CreateDownload(ctx context.Context, adminID int64, it *domain.DownloadItem) (*domain.DownloadItem, error) {
	if err := it.Validate(); err != nil {
		return nil, err
	}
	it.CreatedBy = &adminID

	created, err := s.repo.Create(ctx, it)
	if err != nil {
		return nil, err
	}

	if s.audit != nil {
		_ = s.audit.Record(ctx, auditDomain.Log{
			AdminID:    &adminID,
			Action:     "download.create",
			EntityType: "downloads",
			EntityID:   &created.ID,
			Metadata: map[string]any{
				"title_id": created.TitleID,
				"category": created.Category,
				"file_url": created.FileURL,
			},
		})
	}

	s.triggerSSGBuild(ctx, "create", created.ID)
	return created, nil
}

func (s *DownloadService) UpdateDownload(ctx context.Context, adminID int64, it *domain.DownloadItem) (*domain.DownloadItem, error) {
	if err := it.Validate(); err != nil {
		return nil, err
	}

	updated, err := s.repo.Update(ctx, it)
	if err != nil {
		return nil, err
	}

	if s.audit != nil {
		_ = s.audit.Record(ctx, auditDomain.Log{
			AdminID:    &adminID,
			Action:     "download.update",
			EntityType: "downloads",
			EntityID:   &updated.ID,
			Metadata: map[string]any{
				"title_id": updated.TitleID,
				"category": updated.Category,
				"file_url": updated.FileURL,
			},
		})
	}

	s.triggerSSGBuild(ctx, "update", updated.ID)
	return updated, nil
}

func (s *DownloadService) DeleteDownload(ctx context.Context, adminID int64, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if s.audit != nil {
		_ = s.audit.Record(ctx, auditDomain.Log{
			AdminID:    &adminID,
			Action:     "download.delete",
			EntityType: "downloads",
			EntityID:   &id,
			Metadata:   map[string]any{"id": id},
		})
	}

	s.triggerSSGBuild(ctx, "delete", id)
	return nil
}

func (s *DownloadService) UpdateDownloadInternal(ctx context.Context, it *domain.DownloadItem) (*domain.DownloadItem, error) {
	return s.repo.Update(ctx, it)
}

func (s *DownloadService) TrackDownload(ctx context.Context, id int64) (*domain.DownloadItem, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = s.repo.IncrementDownloadCount(ctx, id)
	return item, nil
}
