package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	auditDomain "pidiekab/internal/modules/audit/domain"
	categoryDomain "pidiekab/internal/modules/category/domain"
	"pidiekab/internal/modules/content/domain"
	"pidiekab/internal/modules/content/infrastructure/postgres"
)

// CategoryChecker checks whether a category exists and is active.
type CategoryChecker interface {
	GetCategory(ctx context.Context, id int64) (*categoryDomain.CategoryWithTranslations, error)
}

// AuditRecorder logs security and content actions.
type AuditRecorder interface {
	Record(ctx context.Context, entry auditDomain.Log) error
}

// ContentService orchestrates content operations.
type ContentService struct {
	contentRepo     *postgres.ContentRepository
	translationRepo *postgres.TranslationRepository
	sourceRepo      *postgres.SourceRepository
	categoryChecker CategoryChecker
	audit           AuditRecorder
}

// NewContentService creates a new ContentService.
func NewContentService(
	contentRepo *postgres.ContentRepository,
	translationRepo *postgres.TranslationRepository,
	sourceRepo *postgres.SourceRepository,
	categoryChecker CategoryChecker,
	audit AuditRecorder,
) *ContentService {
	return &ContentService{
		contentRepo:     contentRepo,
		translationRepo: translationRepo,
		sourceRepo:      sourceRepo,
		categoryChecker: categoryChecker,
		audit:           audit,
	}
}

// CreateContent validates category and coordinates, then persists a new content record.
func (s *ContentService) CreateContent(
	ctx context.Context,
	categoryID int64,
	villageID *int64,
	address *string,
	lat, lng *float64,
	accuracy *string,
	mainMediaID *int64,
	adminID int64,
	ip *string,
) (*domain.Content, error) {
	// Verify category exists and is active
	cat, err := s.categoryChecker.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category: %w", err)
	}
	if !cat.IsActive {
		return nil, domain.ErrCategoryInactive
	}

	// Validate coordinates
	if err := domain.ValidateCoordinates(lat, lng, accuracy); err != nil {
		return nil, err
	}

	c := &domain.Content{
		CategoryID:       categoryID,
		VillageID:        villageID,
		Address:          address,
		Latitude:         lat,
		Longitude:        lng,
		LocationAccuracy: accuracy,
		MainMediaID:      mainMediaID,
		CreatedBy:        &adminID,
	}

	if err := s.contentRepo.Create(ctx, c); err != nil {
		return nil, err
	}

	s.recordAudit(ctx, &adminID, "content.create", "content", &c.ID, map[string]any{"category_id": categoryID}, ip)
	return c, nil
}

// GetContent retrieves content with all its translations and sources.
func (s *ContentService) GetContent(ctx context.Context, id int64) (*domain.ContentDetail, error) {
	c, err := s.contentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	translations, err := s.translationRepo.ListByContentID(ctx, id)
	if err != nil {
		return nil, err
	}

	sources, err := s.sourceRepo.ListByContentID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &domain.ContentDetail{
		Content:      *c,
		Translations: translations,
		Sources:      sources,
	}, nil
}

// UpdateShared updates language-neutral content fields.
func (s *ContentService) UpdateShared(
	ctx context.Context,
	id int64,
	categoryID int64,
	villageID *int64,
	address *string,
	lat, lng *float64,
	accuracy *string,
	mainMediaID *int64,
	adminID int64,
	ip *string,
) error {
	cat, err := s.categoryChecker.GetCategory(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("invalid category: %w", err)
	}
	if !cat.IsActive {
		return domain.ErrCategoryInactive
	}

	if err := domain.ValidateCoordinates(lat, lng, accuracy); err != nil {
		return err
	}

	c := &domain.Content{
		ID:               id,
		CategoryID:       categoryID,
		VillageID:        villageID,
		Address:          address,
		Latitude:         lat,
		Longitude:        lng,
		LocationAccuracy: accuracy,
		MainMediaID:      mainMediaID,
		UpdatedBy:        &adminID,
	}

	if err := s.contentRepo.UpdateShared(ctx, c); err != nil {
		return err
	}

	s.recordAudit(ctx, &adminID, "content.update", "content", &id, map[string]any{"category_id": categoryID}, ip)
	return nil
}

// DeleteContent removes content and its cascading records.
func (s *ContentService) DeleteContent(ctx context.Context, id int64, adminID int64, ip *string) error {
	if err := s.contentRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.recordAudit(ctx, &adminID, "content.delete", "content", &id, nil, ip)
	return nil
}

// ListContents returns paginated content summaries.
func (s *ContentService) ListContents(ctx context.Context, params postgres.FilterParams) ([]*domain.ContentListItem, int, error) {
	return s.contentRepo.List(ctx, params)
}

// SaveTranslation creates or updates a localized version of content.
func (s *ContentService) SaveTranslation(
	ctx context.Context,
	contentID int64,
	locale, title, slug string,
	summary *string,
	body map[string]any,
	seoTitle, seoDescription *string,
	adminID int64,
	ip *string,
) (*domain.ContentTranslation, error) {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale != "id" && locale != "en" {
		return nil, domain.ErrInvalidSlug
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, domain.ErrTitleEmpty
	}

	slug = strings.ToLower(strings.TrimSpace(slug))
	if !domain.IsValidSlug(slug) {
		return nil, domain.ErrInvalidSlug
	}

	if len(body) == 0 {
		return nil, domain.ErrBodyEmpty
	}

	// Verify parent content exists
	if _, err := s.contentRepo.GetByID(ctx, contentID); err != nil {
		return nil, err
	}

	t := &domain.ContentTranslation{
		ContentID:      contentID,
		Locale:         locale,
		Title:          title,
		Slug:           slug,
		Summary:        summary,
		Body:           body,
		SEOTitle:       seoTitle,
		SEODescription: seoDescription,
		Status:         domain.StatusDraft,
		CreatedBy:      &adminID,
	}

	if err := s.translationRepo.Upsert(ctx, t); err != nil {
		return nil, err
	}

	s.recordAudit(ctx, &adminID, "content.translation.save", "content_translation", &t.ID, map[string]any{
		"content_id": contentID,
		"locale":     locale,
		"slug":       slug,
	}, ip)

	return t, nil
}

// GetTranslation retrieves a single localized translation.
func (s *ContentService) GetTranslation(ctx context.Context, contentID int64, locale string) (*domain.ContentTranslation, error) {
	return s.translationRepo.GetByContentAndLocale(ctx, contentID, strings.ToLower(strings.TrimSpace(locale)))
}

// PublishLocale performs strict validation and marks one locale translation as published.
func (s *ContentService) PublishLocale(ctx context.Context, contentID int64, locale string, adminID int64, ip *string) error {
	locale = strings.ToLower(strings.TrimSpace(locale))

	c, err := s.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		return err
	}

	cat, err := s.categoryChecker.GetCategory(ctx, c.CategoryID)
	if err != nil {
		return err
	}

	t, err := s.translationRepo.GetByContentAndLocale(ctx, contentID, locale)
	if err != nil {
		return err
	}

	sources, err := s.sourceRepo.ListByContentID(ctx, contentID)
	if err != nil {
		return err
	}

	// Validate publication eligibility
	if err := domain.ValidatePublishEligibility(c, t, sources, cat.IsActive); err != nil {
		return err
	}

	now := time.Now().UTC()
	if err := s.translationRepo.UpdateStatus(ctx, contentID, locale, domain.StatusPublished, &adminID, &now); err != nil {
		return err
	}

	s.recordAudit(ctx, &adminID, "content.locale.publish", "content_translation", &t.ID, map[string]any{
		"content_id": contentID,
		"locale":     locale,
	}, ip)

	return nil
}

// ArchiveLocale marks a previously published locale as archived.
func (s *ContentService) ArchiveLocale(ctx context.Context, contentID int64, locale string, adminID int64, ip *string) error {
	locale = strings.ToLower(strings.TrimSpace(locale))

	t, err := s.translationRepo.GetByContentAndLocale(ctx, contentID, locale)
	if err != nil {
		return err
	}

	if t.Status == domain.StatusDraft {
		return domain.ErrCannotArchiveUnpublished
	}

	if err := s.translationRepo.UpdateStatus(ctx, contentID, locale, domain.StatusArchived, nil, t.PublishedAt); err != nil {
		return err
	}

	s.recordAudit(ctx, &adminID, "content.locale.archive", "content_translation", &t.ID, map[string]any{
		"content_id": contentID,
		"locale":     locale,
	}, ip)

	return nil
}

// AddSource records a citation or reference for content.
func (s *ContentService) AddSource(
	ctx context.Context,
	contentID int64,
	sourceType, sourceName string,
	sourceURL *string,
	publishedAt, accessedAt *time.Time,
	notes *string,
	adminID int64,
	ip *string,
) (*domain.ContentSource, error) {
	sourceType = strings.ToLower(strings.TrimSpace(sourceType))
	switch sourceType {
	case "research", "visit", "official", "social":
		// valid
	default:
		return nil, domain.ErrInvalidSourceType
	}

	sourceName = strings.TrimSpace(sourceName)
	if sourceName == "" {
		return nil, domain.ErrSourceNameEmpty
	}

	// Verify content exists
	if _, err := s.contentRepo.GetByID(ctx, contentID); err != nil {
		return nil, err
	}

	src := &domain.ContentSource{
		ContentID:         contentID,
		SourceType:        sourceType,
		SourceName:        sourceName,
		SourceURL:         sourceURL,
		SourcePublishedAt: publishedAt,
		AccessedAt:        accessedAt,
		Notes:             notes,
		CreatedBy:         &adminID,
	}

	if err := s.sourceRepo.Create(ctx, src); err != nil {
		return nil, err
	}

	s.recordAudit(ctx, &adminID, "content.source.add", "content_source", &src.ID, map[string]any{
		"content_id": contentID,
		"type":       sourceType,
		"name":       sourceName,
	}, ip)

	return src, nil
}

// ListSources retrieves all sources for content.
func (s *ContentService) ListSources(ctx context.Context, contentID int64) ([]*domain.ContentSource, error) {
	return s.sourceRepo.ListByContentID(ctx, contentID)
}

// DeleteSource removes a source citation.
func (s *ContentService) DeleteSource(ctx context.Context, contentID, sourceID int64, adminID int64, ip *string) error {
	src, err := s.sourceRepo.GetByID(ctx, sourceID)
	if err != nil {
		return err
	}
	if src.ContentID != contentID {
		return domain.ErrSourceNotFound
	}

	if err := s.sourceRepo.Delete(ctx, sourceID); err != nil {
		return err
	}

	s.recordAudit(ctx, &adminID, "content.source.delete", "content_source", &sourceID, map[string]any{
		"content_id": contentID,
	}, ip)
	return nil
}

func (s *ContentService) recordAudit(ctx context.Context, adminID *int64, action, entityType string, entityID *int64, meta map[string]any, ip *string) {
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
