package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"pidiekab/internal/modules/curation/domain"
)

type CurationRepository struct {
	pool *pgxpool.Pool
}

func NewCurationRepository(pool *pgxpool.Pool) *CurationRepository {
	return &CurationRepository{pool: pool}
}

func (r *CurationRepository) IsContentPublished(ctx context.Context, contentID int64, locale string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM content_translations ct
		JOIN contents c ON c.id = ct.content_id
		JOIN categories cat ON cat.id = c.category_id AND cat.is_active = TRUE
		WHERE ct.content_id = $1 AND ct.locale = $2 AND ct.status = 'published'
	`
	var count int
	err := r.pool.QueryRow(ctx, query, contentID, locale).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CurationRepository) SetFeatured(ctx context.Context, adminID int64, contentIDs []int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Clear current featured
	if _, err := tx.Exec(ctx, `DELETE FROM featured_contents`); err != nil {
		return fmt.Errorf("failed to clear featured contents: %w", err)
	}

	for idx, cid := range contentIDs {
		pos := idx + 1
		query := `
			INSERT INTO featured_contents (content_id, position, created_by, created_at)
			VALUES ($1, $2, $3, NOW())
		`
		if _, err := tx.Exec(ctx, query, cid, pos, adminID); err != nil {
			return fmt.Errorf("failed to insert featured item position %d: %w", pos, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *CurationRepository) GetFeatured(ctx context.Context, locale string) ([]domain.FeaturedItem, error) {
	query := `
		SELECT
			fc.id, fc.content_id, fc.position,
			COALESCE(ct.title, ''), COALESCE(ct.slug, ''), COALESCE(cat_t.name, ''),
			fc.created_at
		FROM featured_contents fc
		JOIN contents c ON c.id = fc.content_id
		LEFT JOIN content_translations ct ON ct.content_id = c.id AND ct.locale = $1 AND ct.status = 'published'
		LEFT JOIN categories cat ON cat.id = c.category_id AND cat.is_active = TRUE
		LEFT JOIN category_translations cat_t ON cat_t.category_id = cat.id AND cat_t.locale = $1
		ORDER BY fc.position ASC
	`
	rows, err := r.pool.Query(ctx, query, locale)
	if err != nil {
		return nil, fmt.Errorf("failed to query featured contents: %w", err)
	}
	defer rows.Close()

	var items []domain.FeaturedItem
	for rows.Next() {
		var item domain.FeaturedItem
		if err := rows.Scan(
			&item.ID, &item.ContentID, &item.Position,
			&item.Title, &item.Slug, &item.Category,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *CurationRepository) SetLeadStory(ctx context.Context, adminID int64, locale string, contentID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Remove existing lead story for this locale
	if _, err := tx.Exec(ctx, `DELETE FROM homepage_story_items WHERE locale = $1 AND role = 'lead'`, locale); err != nil {
		return err
	}

	// Insert new lead story (position 1)
	query := `
		INSERT INTO homepage_story_items (locale, content_id, role, position, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, 'lead', 1, $3, $3, NOW(), NOW())
		ON CONFLICT (locale, content_id) DO UPDATE
		SET role = 'lead', position = 1, updated_by = EXCLUDED.updated_by, updated_at = NOW()
	`
	if _, err := tx.Exec(ctx, query, locale, contentID, adminID); err != nil {
		return fmt.Errorf("failed to set lead story: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *CurationRepository) SetSupportingStories(ctx context.Context, adminID int64, locale string, contentIDs []int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Remove existing supporting stories for this locale
	if _, err := tx.Exec(ctx, `DELETE FROM homepage_story_items WHERE locale = $1 AND role = 'supporting'`, locale); err != nil {
		return err
	}

	for idx, cid := range contentIDs {
		pos := idx + 1
		query := `
			INSERT INTO homepage_story_items (locale, content_id, role, position, created_by, updated_by, created_at, updated_at)
			VALUES ($1, $2, 'supporting', $3, $4, $4, NOW(), NOW())
		`
		if _, err := tx.Exec(ctx, query, locale, cid, pos, adminID); err != nil {
			return fmt.Errorf("failed to insert supporting story position %d: %w", pos, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *CurationRepository) GetStories(ctx context.Context, locale string) (*domain.StoryItem, []domain.StoryItem, error) {
	query := `
		SELECT
			h.id, h.locale, h.content_id, h.role, h.position,
			COALESCE(ct.title, ''), COALESCE(ct.slug, ''), COALESCE(cat_t.name, ''),
			h.created_at
		FROM homepage_story_items h
		JOIN contents c ON c.id = h.content_id
		LEFT JOIN content_translations ct ON ct.content_id = c.id AND ct.locale = h.locale AND ct.status = 'published'
		LEFT JOIN categories cat ON cat.id = c.category_id AND cat.is_active = TRUE
		LEFT JOIN category_translations cat_t ON cat_t.category_id = cat.id AND cat_t.locale = h.locale
		WHERE h.locale = $1
		ORDER BY h.role ASC, h.position ASC
	`
	rows, err := r.pool.Query(ctx, query, locale)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query story items: %w", err)
	}
	defer rows.Close()

	var lead *domain.StoryItem
	var supporting []domain.StoryItem

	for rows.Next() {
		var item domain.StoryItem
		if err := rows.Scan(
			&item.ID, &item.Locale, &item.ContentID, &item.Role, &item.Position,
			&item.Title, &item.Slug, &item.Category,
			&item.CreatedAt,
		); err != nil {
			return nil, nil, err
		}

		if item.Role == domain.RoleLead {
			lead = &item
		} else {
			supporting = append(supporting, item)
		}
	}

	return lead, supporting, nil
}

func (r *CurationRepository) GetLatest(ctx context.Context, locale string, excludeIDs []int64, limit int) ([]domain.CuratedItem, error) {
	if limit <= 0 {
		limit = 6
	}

	query := `
		SELECT
			c.id, ct.title, ct.slug, cat_t.slug, cat_t.name,
			COALESCE(ct.summary, ''), COALESCE(ct.published_at, ct.created_at)
		FROM contents c
		JOIN content_translations ct ON ct.content_id = c.id
		JOIN categories cat ON cat.id = c.category_id AND cat.is_active = TRUE
		JOIN category_translations cat_t ON cat_t.category_id = cat.id AND cat_t.locale = ct.locale
		WHERE ct.locale = $1
		  AND ct.status = 'published'
		  AND NOT (c.id = ANY($2))
		ORDER BY ct.published_at DESC NULLS LAST, c.id DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, locale, excludeIDs, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest items: %w", err)
	}
	defer rows.Close()

	var items []domain.CuratedItem
	for rows.Next() {
		var it domain.CuratedItem
		var pubAt time.Time
		if err := rows.Scan(
			&it.ContentID, &it.Title, &it.Slug, &it.CategorySlug, &it.CategoryName,
			&it.Summary, &pubAt,
		); err != nil {
			return nil, err
		}
		it.PublishedAt = pubAt
		items = append(items, it)
	}

	return items, nil
}

func (r *CurationRepository) GetRandom(ctx context.Context, locale string, excludeIDs []int64, limit int) ([]domain.CuratedItem, error) {
	if limit <= 0 {
		limit = 6
	}

	query := `
		SELECT
			c.id, ct.title, ct.slug, cat_t.slug, cat_t.name,
			COALESCE(ct.summary, ''), COALESCE(ct.published_at, ct.created_at)
		FROM contents c
		JOIN content_translations ct ON ct.content_id = c.id
		JOIN categories cat ON cat.id = c.category_id AND cat.is_active = TRUE
		JOIN category_translations cat_t ON cat_t.category_id = cat.id AND cat_t.locale = ct.locale
		WHERE ct.locale = $1
		  AND ct.status = 'published'
		  AND NOT (c.id = ANY($2))
		ORDER BY RANDOM()
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, locale, excludeIDs, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query random items: %w", err)
	}
	defer rows.Close()

	var items []domain.CuratedItem
	for rows.Next() {
		var it domain.CuratedItem
		var pubAt time.Time
		if err := rows.Scan(
			&it.ContentID, &it.Title, &it.Slug, &it.CategorySlug, &it.CategoryName,
			&it.Summary, &pubAt,
		); err != nil {
			return nil, err
		}
		it.PublishedAt = pubAt
		items = append(items, it)
	}

	return items, nil
}
