package application

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	auditDomain "pidiekab/internal/modules/audit/domain"
	"pidiekab/internal/modules/menu/domain"
	"pidiekab/internal/modules/menu/infrastructure/postgres"
)

type AuditRecorder interface {
	Record(ctx context.Context, entry auditDomain.Log) error
}

type JobEnqueuer interface {
	Enqueue(ctx context.Context, tx pgx.Tx, jobType string, payload any) (int64, error)
}

type MenuService struct {
	repo     *postgres.MenuRepository
	audit    AuditRecorder
	enqueuer JobEnqueuer
}

func NewMenuService(repo *postgres.MenuRepository, audit AuditRecorder) *MenuService {
	return &MenuService{
		repo:  repo,
		audit: audit,
	}
}

func (s *MenuService) SetJobEnqueuer(enqueuer JobEnqueuer) {
	s.enqueuer = enqueuer
}

func (s *MenuService) triggerSSGBuild(ctx context.Context, action string, menuID int64) {
	if s.enqueuer != nil {
		_, _ = s.enqueuer.Enqueue(ctx, nil, "site.menu_updated", map[string]any{
			"action":    action,
			"menu_id":   menuID,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

func (s *MenuService) ListMenus(ctx context.Context) ([]domain.MenuItem, error) {
	return s.repo.List(ctx)
}

func (s *MenuService) GetMenuByID(ctx context.Context, id int64) (*domain.MenuItem, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MenuService) CreateMenu(ctx context.Context, adminID int64, item *domain.MenuItem) (*domain.MenuItem, error) {
	if err := item.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}

	if s.audit != nil {
		_ = s.audit.Record(ctx, auditDomain.Log{
			AdminID:    &adminID,
			Action:     "CREATE",
			EntityType: "menu_item",
			EntityID:   &item.ID,
			Metadata: map[string]any{
				"title_id":  item.TitleID,
				"item_type": item.ItemType,
			},
		})
	}

	s.triggerSSGBuild(ctx, "created", item.ID)
	return item, nil
}

func (s *MenuService) UpdateMenu(ctx context.Context, adminID int64, item *domain.MenuItem) (*domain.MenuItem, error) {
	if err := item.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}

	if s.audit != nil {
		_ = s.audit.Record(ctx, auditDomain.Log{
			AdminID:    &adminID,
			Action:     "UPDATE",
			EntityType: "menu_item",
			EntityID:   &item.ID,
			Metadata: map[string]any{
				"title_id":  item.TitleID,
				"item_type": item.ItemType,
			},
		})
	}

	s.triggerSSGBuild(ctx, "updated", item.ID)
	return item, nil
}

func (s *MenuService) DeleteMenu(ctx context.Context, adminID int64, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if s.audit != nil {
		_ = s.audit.Record(ctx, auditDomain.Log{
			AdminID:    &adminID,
			Action:     "DELETE",
			EntityType: "menu_item",
			EntityID:   &id,
		})
	}

	s.triggerSSGBuild(ctx, "deleted", id)
	return nil
}

func (s *MenuService) ReorderMenus(ctx context.Context, adminID int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	if err := s.repo.Reorder(ctx, ids); err != nil {
		return err
	}

	if s.audit != nil {
		_ = s.audit.Record(ctx, auditDomain.Log{
			AdminID:    &adminID,
			Action:     "REORDER",
			EntityType: "menu_item",
			Metadata: map[string]any{
				"ids": ids,
			},
		})
	}

	s.triggerSSGBuild(ctx, "reordered", 0)
	return nil
}
