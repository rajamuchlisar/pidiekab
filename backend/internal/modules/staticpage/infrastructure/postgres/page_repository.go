package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pidiekab/internal/modules/staticpage/domain"
)

type PageRepository struct {
	pool *pgxpool.Pool
}

func NewPageRepository(pool *pgxpool.Pool) *PageRepository {
	return &PageRepository{pool: pool}
}

func (r *PageRepository) CreatePage(ctx context.Context, key string) (*domain.StaticPage, error) {
	query := `
		INSERT INTO static_pages (page_key, created_at, updated_at)
		VALUES ($1, NOW(), NOW())
		RETURNING id, page_key, created_at, updated_at
	`
	var p domain.StaticPage
	err := r.pool.QueryRow(ctx, query, key).Scan(&p.ID, &p.PageKey, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create static page: %w", err)
	}
	return &p, nil
}

func (r *PageRepository) GetPageByID(ctx context.Context, id int64) (*domain.StaticPageDetail, error) {
	queryPage := `
		SELECT id, page_key, created_at, updated_at
		FROM static_pages
		WHERE id = $1
	`
	var detail domain.StaticPageDetail
	err := r.pool.QueryRow(ctx, queryPage, id).Scan(
		&detail.ID, &detail.PageKey, &detail.CreatedAt, &detail.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPageNotFound
		}
		return nil, fmt.Errorf("failed to get static page: %w", err)
	}

	detail.Translations = make(map[string]*domain.StaticPageTranslation)
	queryTrans := `
		SELECT
			id, static_page_id, locale, title, slug, body,
			seo_title, seo_description, status, published_at,
			created_by, updated_by, published_by, created_at, updated_at
		FROM static_page_translations
		WHERE static_page_id = $1
	`
	rows, err := r.pool.Query(ctx, queryTrans, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get page translations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t domain.StaticPageTranslation
		var bodyBytes []byte
		if err := rows.Scan(
			&t.ID, &t.StaticPageID, &t.Locale, &t.Title, &t.Slug, &bodyBytes,
			&t.SeoTitle, &t.SeoDescription, &t.Status, &t.PublishedAt,
			&t.CreatedBy, &t.UpdatedBy, &t.PublishedBy, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan translation: %w", err)
		}
		t.Body = json.RawMessage(bodyBytes)
		detail.Translations[t.Locale] = &t
	}

	return &detail, nil
}

func (r *PageRepository) GetPageByKey(ctx context.Context, pageKey string) (*domain.StaticPageDetail, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `SELECT id FROM static_pages WHERE page_key = $1`, pageKey).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPageNotFound
		}
		return nil, fmt.Errorf("failed to get page id by key: %w", err)
	}
	return r.GetPageByID(ctx, id)
}

func (r *PageRepository) ListPages(ctx context.Context) ([]domain.StaticPageDetail, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, page_key, created_at, updated_at FROM static_pages ORDER BY page_key ASC`)
	if err != nil {
		return nil, fmt.Errorf("failed to query static pages: %w", err)
	}
	defer rows.Close()

	var pages []domain.StaticPageDetail
	for rows.Next() {
		var p domain.StaticPageDetail
		if err := rows.Scan(&p.ID, &p.PageKey, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan page: %w", err)
		}
		p.Translations = make(map[string]*domain.StaticPageTranslation)
		pages = append(pages, p)
	}

	// Fetch translations for all pages
	tRows, err := r.pool.Query(ctx, `
		SELECT
			id, static_page_id, locale, title, slug, body,
			seo_title, seo_description, status, published_at,
			created_by, updated_by, published_by, created_at, updated_at
		FROM static_page_translations
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query translations: %w", err)
	}
	defer tRows.Close()

	transMap := make(map[int64]map[string]*domain.StaticPageTranslation)
	for tRows.Next() {
		var t domain.StaticPageTranslation
		var bodyBytes []byte
		if err := tRows.Scan(
			&t.ID, &t.StaticPageID, &t.Locale, &t.Title, &t.Slug, &bodyBytes,
			&t.SeoTitle, &t.SeoDescription, &t.Status, &t.PublishedAt,
			&t.CreatedBy, &t.UpdatedBy, &t.PublishedBy, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan translation row: %w", err)
		}
		t.Body = json.RawMessage(bodyBytes)
		if transMap[t.StaticPageID] == nil {
			transMap[t.StaticPageID] = make(map[string]*domain.StaticPageTranslation)
		}
		transMap[t.StaticPageID][t.Locale] = &t
	}

	for i := range pages {
		if trs, ok := transMap[pages[i].ID]; ok {
			pages[i].Translations = trs
		}
	}

	return pages, nil
}

type UpsertTranslationParams struct {
	StaticPageID   int64
	Locale         string
	Title          string
	Slug           string
	Body           json.RawMessage
	SeoTitle       string
	SeoDescription string
	AdminID        int64
}

func (r *PageRepository) UpsertTranslation(ctx context.Context, params UpsertTranslationParams) (*domain.StaticPageTranslation, error) {
	query := `
		INSERT INTO static_page_translations (
			static_page_id, locale, title, slug, body,
			seo_title, seo_description, status,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, 'draft', $8, $8, NOW(), NOW()
		)
		ON CONFLICT (static_page_id, locale) DO UPDATE
		SET
			title = EXCLUDED.title,
			slug = EXCLUDED.slug,
			body = EXCLUDED.body,
			seo_title = EXCLUDED.seo_title,
			seo_description = EXCLUDED.seo_description,
			updated_by = EXCLUDED.updated_by,
			updated_at = NOW()
		RETURNING
			id, static_page_id, locale, title, slug, body,
			seo_title, seo_description, status, published_at,
			created_by, updated_by, published_by, created_at, updated_at
	`
	var t domain.StaticPageTranslation
	var bodyBytes []byte
	err := r.pool.QueryRow(ctx, query,
		params.StaticPageID, params.Locale, params.Title, params.Slug, []byte(params.Body),
		params.SeoTitle, params.SeoDescription, params.AdminID,
	).Scan(
		&t.ID, &t.StaticPageID, &t.Locale, &t.Title, &t.Slug, &bodyBytes,
		&t.SeoTitle, &t.SeoDescription, &t.Status, &t.PublishedAt,
		&t.CreatedBy, &t.UpdatedBy, &t.PublishedBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert static page translation: %w", err)
	}
	t.Body = json.RawMessage(bodyBytes)
	return &t, nil
}

func (r *PageRepository) UpdateStatus(ctx context.Context, pageID int64, locale, status string, adminID int64) error {
	var query string
	var args []any

	if status == domain.StatusPublished {
		query = `
			UPDATE static_page_translations
			SET
				status = $1,
				published_at = NOW(),
				published_by = $2,
				updated_by = $2,
				updated_at = NOW()
			WHERE static_page_id = $3 AND locale = $4
		`
		args = []any{status, adminID, pageID, locale}
	} else {
		query = `
			UPDATE static_page_translations
			SET
				status = $1,
				updated_by = $2,
				updated_at = NOW()
			WHERE static_page_id = $3 AND locale = $4
		`
		args = []any{status, adminID, pageID, locale}
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update static page translation status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTranslationNotFound
	}
	return nil
}

