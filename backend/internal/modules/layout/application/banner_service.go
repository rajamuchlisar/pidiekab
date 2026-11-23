package application

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	auditDomain "pidiekab/internal/modules/audit/domain"
	"pidiekab/internal/modules/layout/domain"
	"pidiekab/internal/modules/layout/infrastructure/postgres"
)

type AuditRecorder interface {
	Record(ctx context.Context, entry auditDomain.Log) error
}

type JobEnqueuer interface {
	Enqueue(ctx context.Context, tx pgx.Tx, jobType string, payload any) (int64, error)
}

type LayoutService struct {
	repo     *postgres.BannerRepository
	audit    AuditRecorder
	enqueuer JobEnqueuer
}

func NewLayoutService(repo *postgres.BannerRepository, audit AuditRecorder) *LayoutService {
	return &LayoutService{
		repo:  repo,
		audit: audit,
	}
}

func (s *LayoutService) SetJobEnqueuer(enqueuer JobEnqueuer) {
	s.enqueuer = enqueuer
}

func (s *LayoutService) triggerSSGBuild(ctx context.Context, action string, bannerID int64) {
	if s.enqueuer != nil {
		_, _ = s.enqueuer.Enqueue(ctx, nil, "site.banner_updated", map[string]any{
			"action":    action,
			"banner_id": bannerID,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

func (s *LayoutService) ListBanners(ctx context.Context, placement string) ([]domain.Banner, error) {
	return s.repo.List(ctx, placement)
}

func (s *LayoutService) ListActiveBanners(ctx context.Context, placement string) ([]domain.Banner, error) {
	return s.repo.ListActive(ctx, placement)
}

func (s *LayoutService) GetBannerByID(ctx context.Context, id int64) (*domain.Banner, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *LayoutService) CreateBanner(ctx context.Context, adminID int64, b *domain.Banner) (*domain.Banner, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}

	if s.audit != nil {
		_ = s.audit.Record(ctx, auditDomain.Log{
			AdminID:    &adminID,
			Action:     "CREATE",
			EntityType: "layout_banner",
			EntityID:   &b.ID,
			Metadata: map[string]any{
				"title":       b.Title,
				"placement":   b.Placement,
				"banner_type": b.BannerType,
				"image_url":   b.ImageURL,
				"target_url":  b.TargetURL,
				"is_active":   b.IsActive,
			},
		})
	}

	s.triggerSSGBuild(ctx, "created", b.ID)
	return b, nil
}

func (s *LayoutService) UpdateBanner(ctx context.Context, adminID int64, b *domain.Banner) (*domain.Banner, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}

	old, err := s.repo.GetByID(ctx, b.ID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, b); err != nil {
		return nil, err
	}

	if s.audit != nil {
		_ = s.audit.Record(ctx, auditDomain.Log{
			AdminID:    &adminID,
			Action:     "UPDATE",
			EntityType: "layout_banner",
			EntityID:   &b.ID,
			Metadata: map[string]any{
				"title":              b.Title,
				"placement":          b.Placement,
				"banner_type":        b.BannerType,
				"is_active":          b.IsActive,
				"previous_title":     old.Title,
				"previous_placement": old.Placement,
			},
		})
	}

	s.triggerSSGBuild(ctx, "updated", b.ID)
	return b, nil
}

func (s *LayoutService) DeleteBanner(ctx context.Context, adminID int64, id int64) error {
	old, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if s.audit != nil {
		_ = s.audit.Record(ctx, auditDomain.Log{
			AdminID:    &adminID,
			Action:     "DELETE",
			EntityType: "layout_banner",
			EntityID:   &id,
			Metadata: map[string]any{
				"title":     old.Title,
				"placement": old.Placement,
			},
		})
	}

	s.triggerSSGBuild(ctx, "deleted", id)
	return nil
}

func (s *LayoutService) ToggleActive(ctx context.Context, adminID int64, id int64, isActive bool) error {
	if err := s.repo.ToggleActive(ctx, id, isActive); err != nil {
		return err
	}

	if s.audit != nil {
		_ = s.audit.Record(ctx, auditDomain.Log{
			AdminID:    &adminID,
			Action:     "UPDATE",
			EntityType: "layout_banner",
			EntityID:   &id,
			Metadata:    map[string]any{"is_active": isActive},
		})
	}

	s.triggerSSGBuild(ctx, "toggle_active", id)
	return nil
}
