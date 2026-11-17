package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	auditDomain "pidiekab/internal/modules/audit/domain"
	"pidiekab/internal/modules/category/domain"
)

// CategoryRepository defines data access for categories.
type CategoryRepository interface {
	Create(ctx context.Context, cat *domain.Category, translations []*domain.CategoryTranslation) error
	GetByID(ctx context.Context, id int64) (*domain.CategoryWithTranslations, error)
	List(ctx context.Context, activeOnly bool) ([]*domain.CategoryWithTranslations, error)
	Update(ctx context.Context, id int64, isActive bool, updatedBy *int64) error
	UpsertTranslation(ctx context.Context, trans *domain.CategoryTranslation, updatedBy *int64) error
	Delete(ctx context.Context, id int64) error
}

// AuditRecorder defines audit logging capability.
type AuditRecorder interface {
	Record(ctx context.Context, entry auditDomain.Log) error
}

// Service orchestrates category operations and enforces business rules.
type Service struct {
	repo  CategoryRepository
	audit AuditRecorder
}

// NewService creates a new category Service.
func NewService(repo CategoryRepository, audit AuditRecorder) *Service {
	return &Service{
		repo:  repo,
		audit: audit,
	}
}

// CreateCategory validates input, inserts category and translations, and logs audit trail.
func (s *Service) CreateCategory(
	ctx context.Context,
	isActive bool,
	adminID int64,
	translations []*domain.CategoryTranslation,
	ip *string,
) (*domain.CategoryWithTranslations, error) {
	if len(translations) == 0 {
		return nil, fmt.Errorf("at least one localized translation is required")
	}

	for _, t := range translations {
		t.Locale = strings.ToLower(strings.TrimSpace(t.Locale))
		if t.Locale != "id" && t.Locale != "en" {
			return nil, domain.ErrInvalidLocale
		}
		t.Name = strings.TrimSpace(t.Name)
		if t.Name == "" {
			return nil, domain.ErrCategoryNameEmpty
		}
		t.Slug = strings.ToLower(strings.TrimSpace(t.Slug))
		if !domain.IsValidSlug(t.Slug) {
			return nil, domain.ErrInvalidSlug
		}
	}

	cat := &domain.Category{
		IsActive:  isActive,
		CreatedBy: &adminID,
	}

	if err := s.repo.Create(ctx, cat, translations); err != nil {
		return nil, err
	}

	s.recordAudit(ctx, &adminID, "category.create", "category", &cat.ID, map[string]any{"is_active": isActive}, ip)

	return s.repo.GetByID(ctx, cat.ID)
}

// GetCategory finds category by ID.
func (s *Service) GetCategory(ctx context.Context, id int64) (*domain.CategoryWithTranslations, error) {
	return s.repo.GetByID(ctx, id)
}

// ListCategories lists all categories.
func (s *Service) ListCategories(ctx context.Context, activeOnly bool) ([]*domain.CategoryWithTranslations, error) {
	return s.repo.List(ctx, activeOnly)
}

// UpdateCategory modifies category active status.
func (s *Service) UpdateCategory(ctx context.Context, id int64, isActive bool, adminID int64, ip *string) error {
	if err := s.repo.Update(ctx, id, isActive, &adminID); err != nil {
		return err
	}

	s.recordAudit(ctx, &adminID, "category.update", "category", &id, map[string]any{"is_active": isActive}, ip)
	return nil
}

// SaveTranslation creates or updates a localized version of a category.
func (s *Service) SaveTranslation(
	ctx context.Context,
	categoryID int64,
	locale, name string,
	description *string,
	slug string,
	adminID int64,
	ip *string,
) (*domain.CategoryTranslation, error) {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale != "id" && locale != "en" {
		return nil, domain.ErrInvalidLocale
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, domain.ErrCategoryNameEmpty
	}

	slug = strings.ToLower(strings.TrimSpace(slug))
	if !domain.IsValidSlug(slug) {
		return nil, domain.ErrInvalidSlug
	}

	// Verify category exists
	if _, err := s.repo.GetByID(ctx, categoryID); err != nil {
		return nil, err
	}

	t := &domain.CategoryTranslation{
		CategoryID:  categoryID,
		Locale:      locale,
		Name:        name,
		Description: description,
		Slug:        slug,
	}

	if err := s.repo.UpsertTranslation(ctx, t, &adminID); err != nil {
		return nil, err
	}

	s.recordAudit(ctx, &adminID, "category.translation.update", "category_translation", &t.ID, map[string]any{
		"category_id": categoryID,
		"locale":      locale,
		"slug":        slug,
	}, ip)

	return t, nil
}

// DeleteCategory permanently removes a category if unused.
func (s *Service) DeleteCategory(ctx context.Context, id int64, adminID int64, ip *string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.recordAudit(ctx, &adminID, "category.delete", "category", &id, nil, ip)
	return nil
}

func (s *Service) recordAudit(ctx context.Context, adminID *int64, action, entityType string, entityID *int64, meta map[string]any, ip *string) {
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

