package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	auditDomain "pidiekab/internal/modules/audit/domain"
	"pidiekab/internal/modules/staticpage/domain"
	"pidiekab/internal/modules/staticpage/infrastructure/postgres"
)

type AuditRecorder interface {
	Record(ctx context.Context, entry auditDomain.Log) error
}

type JobEnqueuer interface {
	Enqueue(ctx context.Context, tx pgx.Tx, jobType string, payload any) (int64, error)
}

// RedirectRecorder records a permanent URL redirect when published page slugs change.
type RedirectRecorder interface {
	UpsertRedirect(ctx context.Context, oldPath, newPath string, statusCode int, adminID *int64, notes *string) error
}

type PageService struct {
	repo      *postgres.PageRepository
	audit     AuditRecorder
	enqueuer  JobEnqueuer
	redirects RedirectRecorder
}

func NewPageService(repo *postgres.PageRepository, audit AuditRecorder) *PageService {
	return &PageService{repo: repo, audit: audit}
}

func (s *PageService) SetJobEnqueuer(enqueuer JobEnqueuer) {
	s.enqueuer = enqueuer
}

func (s *PageService) SetRedirectRecorder(recorder RedirectRecorder) {
	s.redirects = recorder
}

func (s *PageService) CreatePage(ctx context.Context, adminID int64, pageKey string) (*domain.StaticPage, error) {
	if err := domain.ValidatePageKey(pageKey); err != nil {
		return nil, err
	}

	page, err := s.repo.CreatePage(ctx, strings.TrimSpace(pageKey))
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, &adminID, "static_page.create", "static_page", &page.ID, map[string]any{
		"page_key": page.PageKey,
	}, nil)

	return page, nil
}

func (s *PageService) GetPage(ctx context.Context, id int64) (*domain.StaticPageDetail, error) {
	return s.repo.GetPageByID(ctx, id)
}

func (s *PageService) GetPageByKey(ctx context.Context, key string) (*domain.StaticPageDetail, error) {
	return s.repo.GetPageByKey(ctx, key)
}

func (s *PageService) ListPages(ctx context.Context) ([]domain.StaticPageDetail, error) {
	return s.repo.ListPages(ctx)
}

type SaveTranslationInput struct {
	Title          string          `json:"title"`
	Slug           string          `json:"slug"`
	Body           json.RawMessage `json:"body"`
	SeoTitle       string          `json:"seo_title"`
	SeoDescription string          `json:"seo_description"`
}

func (s *PageService) SaveTranslation(ctx context.Context, adminID int64, pageID int64, locale string, input SaveTranslationInput) (*domain.StaticPageTranslation, error) {
	if err := domain.ValidateLocale(locale); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Title) == "" {
		return nil, domain.ErrTitleEmpty
	}
	if err := domain.ValidateSlug(input.Slug); err != nil {
		return nil, err
	}
	if err := domain.ValidateBodyJSON(input.Body); err != nil {
		return nil, err
	}

	newSlug := strings.TrimSpace(input.Slug)

	// Check if existing published page is having its slug changed
	if s.redirects != nil {
		if pageDetail, err := s.repo.GetPageByID(ctx, pageID); err == nil && pageDetail != nil {
			if existing, ok := pageDetail.Translations[locale]; ok && existing != nil {
				if existing.Slug != newSlug && (existing.Status == domain.StatusPublished || existing.PublishedAt != nil) {
					oldPath := fmt.Sprintf("/%s/%s/", locale, existing.Slug)
					newPath := fmt.Sprintf("/%s/%s/", locale, newSlug)
					notes := fmt.Sprintf("Auto-generated from static page slug rename (Page #%d, %s)", pageID, locale)
					_ = s.redirects.UpsertRedirect(ctx, oldPath, newPath, 301, &adminID, &notes)
				}
			}
		}
	}

	params := postgres.UpsertTranslationParams{
		StaticPageID:   pageID,
		Locale:         locale,
		Title:          strings.TrimSpace(input.Title),
		Slug:           newSlug,
		Body:           input.Body,
		SeoTitle:       strings.TrimSpace(input.SeoTitle),
		SeoDescription: strings.TrimSpace(input.SeoDescription),
		AdminID:        adminID,
	}

	trans, err := s.repo.UpsertTranslation(ctx, params)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, &adminID, "static_page.save_translation", "static_page_translation", &trans.ID, map[string]any{
		"static_page_id": pageID,
		"locale":         locale,
		"slug":           trans.Slug,
	}, nil)

	return trans, nil
}

func (s *PageService) PublishTranslation(ctx context.Context, adminID int64, pageID int64, locale string) error {
	if err := domain.ValidateLocale(locale); err != nil {
		return err
	}

	detail, err := s.repo.GetPageByID(ctx, pageID)
	if err != nil {
		return err
	}

	trans, ok := detail.Translations[locale]
	if !ok || trans == nil {
		return domain.ErrTranslationNotFound
	}

	if strings.TrimSpace(trans.Title) == "" || len(trans.Body) == 0 {
		return domain.ErrCannotPublishNoBody
	}

	if err := s.repo.UpdateStatus(ctx, pageID, locale, domain.StatusPublished, adminID); err != nil {
		return err
	}

	s.recordAudit(ctx, &adminID, "static_page.publish", "static_page_translation", &trans.ID, map[string]any{
		"static_page_id": pageID,
		"locale":         locale,
	}, nil)

	if s.enqueuer != nil {
		_, _ = s.enqueuer.Enqueue(ctx, nil, "static_page.publish", map[string]any{
			"page_id": pageID,
			"locale":  locale,
		})
	}

	return nil
}

func (s *PageService) ArchiveTranslation(ctx context.Context, adminID int64, pageID int64, locale string) error {
	if err := domain.ValidateLocale(locale); err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, pageID, locale, domain.StatusArchived, adminID); err != nil {
		return err
	}

	s.recordAudit(ctx, &adminID, "static_page.archive", "static_page_translation", nil, map[string]any{
		"static_page_id": pageID,
		"locale":         locale,
	}, nil)

	if s.enqueuer != nil {
		_, _ = s.enqueuer.Enqueue(ctx, nil, "static_page.archive", map[string]any{
			"page_id": pageID,
			"locale":  locale,
		})
	}

	return nil
}

func (s *PageService) recordAudit(ctx context.Context, adminID *int64, action, entityType string, entityID *int64, meta map[string]any, ip *string) {
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
