package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"pidiekab/internal/modules/content/domain"
)

// TranslationRepository manages PostgreSQL persistence for content translations.
type TranslationRepository struct {
	pool *pgxpool.Pool
}

// NewTranslationRepository creates a new TranslationRepository.
func NewTranslationRepository(pool *pgxpool.Pool) *TranslationRepository {
	return &TranslationRepository{pool: pool}
}

// Upsert creates or updates a content translation for a specific locale.
func (r *TranslationRepository) Upsert(ctx context.Context, t *domain.ContentTranslation) error {
	bodyJSON, err := json.Marshal(t.Body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	var extraJSON []byte
	if t.ExtraData != nil {
		extraJSON, err = json.Marshal(t.ExtraData)
		if err != nil {
			return fmt.Errorf("failed to marshal extra_data: %w", err)
		}
	}

	// Update tsvector: weights A (title), B (summary)
	query := `
		INSERT INTO content_translations (
			content_id, locale, title, slug, summary, body, extra_data,
			seo_title, seo_description, status, scheduled_at, search_vector,
			created_by, updated_by, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			setweight(to_tsvector('simple', COALESCE($3, '')), 'A') || setweight(to_tsvector('simple', COALESCE($5, '')), 'B'),
			$12, $12, NOW(), NOW()
		)
		ON CONFLICT (content_id, locale) DO UPDATE
		SET title = EXCLUDED.title,
		    slug = EXCLUDED.slug,
		    summary = EXCLUDED.summary,
		    body = EXCLUDED.body,
		    extra_data = EXCLUDED.extra_data,
		    seo_title = EXCLUDED.seo_title,
		    seo_description = EXCLUDED.seo_description,
		    search_vector = EXCLUDED.search_vector,
		    status = CASE WHEN EXCLUDED.status = 'scheduled' THEN EXCLUDED.status ELSE content_translations.status END,
		    scheduled_at = CASE WHEN EXCLUDED.status = 'scheduled' THEN EXCLUDED.scheduled_at ELSE content_translations.scheduled_at END,
		    updated_by = EXCLUDED.updated_by,
		    updated_at = NOW()
		RETURNING id, status, published_at, scheduled_at, created_at, updated_at
	`

	err = r.pool.QueryRow(ctx, query,
		t.ContentID,
		t.Locale,
		t.Title,
		t.Slug,
		t.Summary,
		bodyJSON,
		extraJSON,
		t.SEOTitle,
		t.SEODescription,
		t.Status,
		t.ScheduledAt,
		t.CreatedBy,
	).Scan(&t.ID, &t.Status, &t.PublishedAt, &t.ScheduledAt, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrSlugAlreadyExists
		}
		return fmt.Errorf("failed to upsert content translation: %w", err)
	}

	return nil
}

// GetByContentAndLocale retrieves a localized translation.
func (r *TranslationRepository) GetByContentAndLocale(ctx context.Context, contentID int64, locale string) (*domain.ContentTranslation, error) {
	query := `
		SELECT id, content_id, locale, title, slug, summary, body, extra_data,
		       seo_title, seo_description, status, published_at, scheduled_at, created_by, updated_by, published_by, created_at, updated_at
		FROM content_translations
		WHERE content_id = $1 AND locale = $2
	`

	var t domain.ContentTranslation
	var bodyJSON, extraJSON []byte
	var statusStr string

	err := r.pool.QueryRow(ctx, query, contentID, locale).Scan(
		&t.ID,
		&t.ContentID,
		&t.Locale,
		&t.Title,
		&t.Slug,
		&t.Summary,
		&bodyJSON,
		&extraJSON,
		&t.SEOTitle,
		&t.SEODescription,
		&statusStr,
		&t.PublishedAt,
		&t.ScheduledAt,
		&t.CreatedBy,
		&t.UpdatedBy,
		&t.PublishedBy,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTranslationNotFound
		}
		return nil, fmt.Errorf("failed to query translation: %w", err)
	}

	t.Status = domain.TranslationStatus(statusStr)
	if len(bodyJSON) > 0 {
		_ = json.Unmarshal(bodyJSON, &t.Body)
	}
	if len(extraJSON) > 0 {
		_ = json.Unmarshal(extraJSON, &t.ExtraData)
	}

	return &t, nil
}

// ListByContentID retrieves all translations for a content record.
func (r *TranslationRepository) ListByContentID(ctx context.Context, contentID int64) (map[string]*domain.ContentTranslation, error) {
	query := `
		SELECT id, content_id, locale, title, slug, summary, body, extra_data,
		       seo_title, seo_description, status, published_at, scheduled_at, created_by, updated_by, published_by, created_at, updated_at
		FROM content_translations
		WHERE content_id = $1
	`

	rows, err := r.pool.Query(ctx, query, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query translations for content %d: %w", contentID, err)
	}
	defer rows.Close()

	result := make(map[string]*domain.ContentTranslation)
	for rows.Next() {
		var t domain.ContentTranslation
		var bodyJSON, extraJSON []byte
		var statusStr string

		if err := rows.Scan(
			&t.ID,
			&t.ContentID,
			&t.Locale,
			&t.Title,
			&t.Slug,
			&t.Summary,
			&bodyJSON,
			&extraJSON,
			&t.SEOTitle,
			&t.SEODescription,
			&statusStr,
			&t.PublishedAt,
			&t.ScheduledAt,
			&t.CreatedBy,
			&t.UpdatedBy,
			&t.PublishedBy,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan translation row: %w", err)
		}

		t.Status = domain.TranslationStatus(statusStr)
		if len(bodyJSON) > 0 {
			_ = json.Unmarshal(bodyJSON, &t.Body)
		}
		if len(extraJSON) > 0 {
			_ = json.Unmarshal(extraJSON, &t.ExtraData)
		}

		result[t.Locale] = &t
	}

	return result, rows.Err()
}

// UpdateStatus changes translation status, published_by, and published_at timestamps.
func (r *TranslationRepository) UpdateStatus(
	ctx context.Context,
	contentID int64,
	locale string,
	status domain.TranslationStatus,
	publishedBy *int64,
	publishedAt *time.Time,
) error {
	query := `
		UPDATE content_translations
		SET status = $1, published_by = $2, published_at = $3, scheduled_at = NULL, updated_at = NOW()
		WHERE content_id = $4 AND locale = $5
	`
	cmdTag, err := r.pool.Exec(ctx, query, string(status), publishedBy, publishedAt, contentID, locale)
	if err != nil {
		return fmt.Errorf("failed to update translation status: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrTranslationNotFound
	}
	return nil
}

// UpdateSchedule updates translation status to 'scheduled' and sets scheduled_at.
func (r *TranslationRepository) UpdateSchedule(
	ctx context.Context,
	contentID int64,
	locale string,
	status domain.TranslationStatus,
	updatedBy *int64,
	scheduledAt *time.Time,
) error {
	query := `
		UPDATE content_translations
		SET status = $1, scheduled_at = $2, published_by = $3, updated_by = $3, updated_at = NOW()
		WHERE content_id = $4 AND locale = $5
	`
	cmdTag, err := r.pool.Exec(ctx, query, string(status), scheduledAt, updatedBy, contentID, locale)
	if err != nil {
		return fmt.Errorf("failed to update translation schedule: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrTranslationNotFound
	}
	return nil
}

// CancelSchedule cancels a scheduled translation and reverts it to draft.
func (r *TranslationRepository) CancelSchedule(
	ctx context.Context,
	contentID int64,
	locale string,
	updatedBy *int64,
) error {
	query := `
		UPDATE content_translations
		SET status = 'draft', scheduled_at = NULL, updated_by = $1, updated_at = NOW()
		WHERE content_id = $2 AND locale = $3 AND status = 'scheduled'
	`
	cmdTag, err := r.pool.Exec(ctx, query, updatedBy, contentID, locale)
	if err != nil {
		return fmt.Errorf("failed to cancel translation schedule: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrTranslationNotFound
	}
	return nil
}

// PromoteDueScheduled automatically promotes all scheduled translations whose scheduled_at <= now to published.
func (r *TranslationRepository) PromoteDueScheduled(ctx context.Context, now time.Time) (int64, error) {
	query := `
		WITH promoted AS (
			UPDATE content_translations
			SET status = 'published',
			    published_at = COALESCE(scheduled_at, NOW()),
			    published_by = COALESCE(published_by, updated_by, created_by),
			    scheduled_at = NULL,
			    updated_at = NOW()
			WHERE status = 'scheduled' AND scheduled_at <= $1
			RETURNING id
		)
		SELECT COUNT(*) FROM promoted
	`
	var count int64
	err := r.pool.QueryRow(ctx, query, now).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to promote due scheduled translations: %w", err)
	}
	return count, nil
}
