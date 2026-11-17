package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"pidiekab/internal/modules/category/domain"
)

// Repository manages persistence for categories in PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new category repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new category and its initial translations atomically.
func (r *Repository) Create(ctx context.Context, cat *domain.Category, translations []*domain.CategoryTranslation) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queryCat := `
		INSERT INTO categories (is_active, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	if err := tx.QueryRow(ctx, queryCat, cat.IsActive, cat.CreatedBy).Scan(&cat.ID, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
		return fmt.Errorf("failed to insert category: %w", err)
	}

	queryTrans := `
		INSERT INTO category_translations (category_id, locale, name, description, slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	for _, t := range translations {
		t.CategoryID = cat.ID
		if err := tx.QueryRow(ctx, queryTrans, t.CategoryID, t.Locale, t.Name, t.Description, t.Slug).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return domain.ErrSlugAlreadyExists
			}
			return fmt.Errorf("failed to insert category translation for locale %s: %w", t.Locale, err)
		}
	}

	return tx.Commit(ctx)
}

// GetByID finds a category and all its localized translations.
func (r *Repository) GetByID(ctx context.Context, id int64) (*domain.CategoryWithTranslations, error) {
	queryCat := `
		SELECT id, is_active, created_by, updated_by, created_at, updated_at
		FROM categories
		WHERE id = $1
	`
	var res domain.CategoryWithTranslations
	err := r.pool.QueryRow(ctx, queryCat, id).Scan(
		&res.ID,
		&res.IsActive,
		&res.CreatedBy,
		&res.UpdatedBy,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to query category: %w", err)
	}

	res.Translations = make(map[string]*domain.CategoryTranslation)

	queryTrans := `
		SELECT id, category_id, locale, name, description, slug, created_at, updated_at
		FROM category_translations
		WHERE category_id = $1
	`
	rows, err := r.pool.Query(ctx, queryTrans, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query category translations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t domain.CategoryTranslation
		if err := rows.Scan(&t.ID, &t.CategoryID, &t.Locale, &t.Name, &t.Description, &t.Slug, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan translation row: %w", err)
		}
		res.Translations[t.Locale] = &t
	}

	return &res, rows.Err()
}

// List returns all categories with their translations.
func (r *Repository) List(ctx context.Context, activeOnly bool) ([]*domain.CategoryWithTranslations, error) {
	queryCat := `
		SELECT id, is_active, created_by, updated_by, created_at, updated_at
		FROM categories
	`
	if activeOnly {
		queryCat += ` WHERE is_active = TRUE`
	}
	queryCat += ` ORDER BY id ASC`

	rows, err := r.pool.Query(ctx, queryCat)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	categoryMap := make(map[int64]*domain.CategoryWithTranslations)
	var orderedIDs []int64

	for rows.Next() {
		var c domain.CategoryWithTranslations
		c.Translations = make(map[string]*domain.CategoryTranslation)
		if err := rows.Scan(&c.ID, &c.IsActive, &c.CreatedBy, &c.UpdatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categoryMap[c.ID] = &c
		orderedIDs = append(orderedIDs, c.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(orderedIDs) == 0 {
		return []*domain.CategoryWithTranslations{}, nil
	}

	// Query all translations for these categories
	queryTrans := `
		SELECT id, category_id, locale, name, description, slug, created_at, updated_at
		FROM category_translations
		ORDER BY id ASC
	`
	transRows, err := r.pool.Query(ctx, queryTrans)
	if err != nil {
		return nil, fmt.Errorf("failed to query translations: %w", err)
	}
	defer transRows.Close()

	for transRows.Next() {
		var t domain.CategoryTranslation
		if err := transRows.Scan(&t.ID, &t.CategoryID, &t.Locale, &t.Name, &t.Description, &t.Slug, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan translation: %w", err)
		}
		if cat, exists := categoryMap[t.CategoryID]; exists {
			cat.Translations[t.Locale] = &t
		}
	}

	result := make([]*domain.CategoryWithTranslations, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		result = append(result, categoryMap[id])
	}

	return result, transRows.Err()
}

// Update modifies shared category status (is_active).
func (r *Repository) Update(ctx context.Context, id int64, isActive bool, updatedBy *int64) error {
	query := `
		UPDATE categories
		SET is_active = $1, updated_by = $2, updated_at = NOW()
		WHERE id = $3
	`
	cmdTag, err := r.pool.Exec(ctx, query, isActive, updatedBy, id)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrCategoryNotFound
	}
	return nil
}

// UpsertTranslation inserts or updates a category translation for a specific locale.
func (r *Repository) UpsertTranslation(ctx context.Context, t *domain.CategoryTranslation, updatedBy *int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
		INSERT INTO category_translations (category_id, locale, name, description, slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (category_id, locale) DO UPDATE
		SET name = EXCLUDED.name,
		    description = EXCLUDED.description,
		    slug = EXCLUDED.slug,
		    updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(ctx, query, t.CategoryID, t.Locale, t.Name, t.Description, t.Slug).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrSlugAlreadyExists
		}
		return fmt.Errorf("failed to upsert category translation: %w", err)
	}

	// Update category parent timestamp
	_, err = tx.Exec(ctx, `UPDATE categories SET updated_by = $1, updated_at = NOW() WHERE id = $2`, updatedBy, t.CategoryID)
	if err != nil {
		return fmt.Errorf("failed to update parent category: %w", err)
	}

	return tx.Commit(ctx)
}

// Delete permanently removes a category if it is not referenced by any content.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	// 1. Check if category is referenced by contents
	var contentCount int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM contents WHERE category_id = $1`, id).Scan(&contentCount)
	if err != nil {
		return fmt.Errorf("failed to check content count: %w", err)
	}
	if contentCount > 0 {
		return domain.ErrCategoryInUse
	}

	// 2. Delete category (CASCADE removes category_translations)
	cmdTag, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}
