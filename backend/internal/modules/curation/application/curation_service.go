package application

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	auditDomain "pidiekab/internal/modules/audit/domain"
	"pidiekab/internal/modules/curation/domain"
	"pidiekab/internal/modules/curation/infrastructure/postgres"
)

type AuditRecorder interface {
	Record(ctx context.Context, entry auditDomain.Log) error
}

type JobEnqueuer interface {
	Enqueue(ctx context.Context, tx pgx.Tx, jobType string, payload any) (int64, error)
}

type CurationService struct {
	repo     *postgres.CurationRepository
	audit    AuditRecorder
	enqueuer JobEnqueuer
}

func NewCurationService(repo *postgres.CurationRepository, audit AuditRecorder, enqueuer JobEnqueuer) *CurationService {
	return &CurationService{
		repo:     repo,
		audit:    audit,
		enqueuer: enqueuer,
	}
}

func (s *CurationService) SetFeatured(ctx context.Context, adminID int64, contentIDs []int64) error {
	seen := make(map[int64]bool)
	for _, id := range contentIDs {
		if seen[id] {
			return domain.ErrDuplicateContent
		}
		seen[id] = true
	}

	for _, id := range contentIDs {
		// Content must be published in at least ID or EN
		pubID, _ := s.repo.IsContentPublished(ctx, id, "id")
		pubEN, _ := s.repo.IsContentPublished(ctx, id, "en")
		if !pubID && !pubEN {
			return domain.ErrContentNotPublished
		}
	}

	if err := s.repo.SetFeatured(ctx, adminID, contentIDs); err != nil {
		return err
	}

	s.recordAudit(ctx, &adminID, "curation.featured.set", "featured_contents", nil, map[string]any{
		"count": len(contentIDs),
	}, nil)

	if s.enqueuer != nil {
		_, _ = s.enqueuer.Enqueue(ctx, nil, "curation.featured.updated", map[string]any{"count": len(contentIDs)})
	}

	return nil
}

func (s *CurationService) SetLeadStory(ctx context.Context, adminID int64, locale string, contentID int64) error {
	if err := domain.ValidateLocale(locale); err != nil {
		return err
	}

	pub, err := s.repo.IsContentPublished(ctx, contentID, locale)
	if err != nil {
		return err
	}
	if !pub {
		return domain.ErrContentNotPublished
	}

	if err := s.repo.SetLeadStory(ctx, adminID, locale, contentID); err != nil {
		return err
	}

	s.recordAudit(ctx, &adminID, "curation.lead_story.set", "homepage_story_items", &contentID, map[string]any{
		"locale": locale,
	}, nil)

	if s.enqueuer != nil {
		_, _ = s.enqueuer.Enqueue(ctx, nil, "curation.lead_story.updated", map[string]any{
			"locale":     locale,
			"content_id": contentID,
		})
	}

	return nil
}

func (s *CurationService) SetSupportingStories(ctx context.Context, adminID int64, locale string, contentIDs []int64) error {
	if err := domain.ValidateLocale(locale); err != nil {
		return err
	}

	seen := make(map[int64]bool)
	for _, id := range contentIDs {
		if seen[id] {
			return domain.ErrDuplicateContent
		}
		seen[id] = true
	}

	for _, id := range contentIDs {
		pub, err := s.repo.IsContentPublished(ctx, id, locale)
		if err != nil {
			return err
		}
		if !pub {
			return domain.ErrContentNotPublished
		}
	}

	if err := s.repo.SetSupportingStories(ctx, adminID, locale, contentIDs); err != nil {
		return err
	}

	s.recordAudit(ctx, &adminID, "curation.supporting_stories.set", "homepage_story_items", nil, map[string]any{
		"locale": locale,
		"count":  len(contentIDs),
	}, nil)

	if s.enqueuer != nil {
		_, _ = s.enqueuer.Enqueue(ctx, nil, "curation.supporting_stories.updated", map[string]any{
			"locale": locale,
			"count":  len(contentIDs),
		})
	}

	return nil
}

func (s *CurationService) GetCuratedHomepage(ctx context.Context, locale string) (*domain.HomepageCuration, error) {
	if err := domain.ValidateLocale(locale); err != nil {
		return nil, err
	}

	lead, supporting, err := s.repo.GetStories(ctx, locale)
	if err != nil {
		return nil, err
	}

	featured, err := s.repo.GetFeatured(ctx, locale)
	if err != nil {
		return nil, err
	}

	// Exclude lead, supporting, and featured IDs
	excludedMap := make(map[int64]bool)
	if lead != nil {
		excludedMap[lead.ContentID] = true
	}
	for _, sup := range supporting {
		excludedMap[sup.ContentID] = true
	}
	for _, f := range featured {
		excludedMap[f.ContentID] = true
	}

	var excludedIDs []int64
	for id := range excludedMap {
		excludedIDs = append(excludedIDs, id)
	}

	latest, err := s.repo.GetLatest(ctx, locale, excludedIDs, 6)
	if err != nil {
		return nil, err
	}

	for _, it := range latest {
		excludedIDs = append(excludedIDs, it.ContentID)
	}

	random, err := s.repo.GetRandom(ctx, locale, excludedIDs, 6)
	if err != nil {
		return nil, err
	}

	return &domain.HomepageCuration{
		Locale:            locale,
		LeadStory:         lead,
		SupportingStories: supporting,
		Featured:          featured,
		Latest:            latest,
		Random:            random,
	}, nil
}

func (s *CurationService) recordAudit(ctx context.Context, adminID *int64, action, entityType string, entityID *int64, meta map[string]any, ip *string) {
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
