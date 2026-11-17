package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"pidiekab/internal/modules/content/domain"
)

// ContentRepository manages PostgreSQL persistence for content identities.
type ContentRepository struct {
	pool *pgxpool.Pool
}

// NewContentRepository creates a new ContentRepository.
func NewContentRepository(pool *pgxpool.Pool) *ContentRepository {
	return &ContentRepository{pool: pool}
}

// Create inserts a new content record.
func (r *ContentRepository) Create(ctx context.Context, c *domain.Content) error {
	query := `
		INSERT INTO contents (category_id, village_id, address, latitude, longitude, location_accuracy, main_media_id, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		c.CategoryID,
		c.VillageID,
		c.Address,
		c.Latitude,
		c.Longitude,
		c.LocationAccuracy,
		c.MainMediaID,
		c.CreatedBy,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

// GetByID retrieves a content identity by ID.
func (r *ContentRepository) GetByID(ctx context.Context, id int64) (*domain.Content, error) {
	query := `
		SELECT id, category_id, village_id, address, latitude, longitude, location_accuracy, main_media_id, created_by, updated_by, created_at, updated_at
		FROM contents
		WHERE id = $1
	`
	var c domain.Content
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID,
		&c.CategoryID,
		&c.VillageID,
		&c.Address,
		&c.Latitude,
		&c.Longitude,
		&c.LocationAccuracy,
		&c.MainMediaID,
		&c.CreatedBy,
		&c.UpdatedBy,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrContentNotFound
		}
		return nil, fmt.Errorf("failed to query content: %w", err)
	}
	return &c, nil
}

// UpdateShared modifies language-neutral shared fields.
func (r *ContentRepository) UpdateShared(ctx context.Context, c *domain.Content) error {
	query := `
		UPDATE contents
		SET category_id = $1, village_id = $2, address = $3, latitude = $4, longitude = $5, location_accuracy = $6, main_media_id = $7, updated_by = $8, updated_at = NOW()
		WHERE id = $9
		RETURNING updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		c.CategoryID,
		c.VillageID,
		c.Address,
		c.Latitude,
		c.Longitude,
		c.LocationAccuracy,
		c.MainMediaID,
		c.UpdatedBy,
		c.ID,
	).Scan(&c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrContentNotFound
		}
		return fmt.Errorf("failed to update content: %w", err)
	}
	return nil
}

// Delete permanently removes a content record (cascading deletes translations and sources).
func (r *ContentRepository) Delete(ctx context.Context, id int64) error {
	cmdTag, err := r.pool.Exec(ctx, `DELETE FROM contents WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete content: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrContentNotFound
	}
	return nil
}

// FilterParams defines criteria for listing contents.
type FilterParams struct {
	CategoryID *int64
	Status     *string
	Locale     *string
	Query      *string
	Page       int
	PerPage    int
}

// List returns a paginated list of content summaries with total count.
func (r *ContentRepository) List(ctx context.Context, p FilterParams) ([]*domain.ContentListItem, int, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 50 {
		p.PerPage = 12
	}
	offset := (p.Page - 1) * p.PerPage

	baseWhere := "WHERE 1=1"
	var args []any
	argIdx := 1

	if p.CategoryID != nil {
		baseWhere += " AND c.category_id = $" + strconv.Itoa(argIdx)
		args = append(args, *p.CategoryID)
		argIdx++
	}

	if p.Status != nil && *p.Status != "" && p.Locale != nil && *p.Locale != "" {
		baseWhere += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM content_translations ct WHERE ct.content_id = c.id AND ct.locale = $%d AND ct.status = $%d)", argIdx, argIdx+1)
		args = append(args, *p.Locale, *p.Status)
		argIdx += 2
	} else if p.Status != nil && *p.Status != "" {
		baseWhere += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM content_translations ct WHERE ct.content_id = c.id AND ct.status = $%d)", argIdx)
		args = append(args, *p.Status)
		argIdx++
	}

	if p.Query != nil && *p.Query != "" {
		baseWhere += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM content_translations ct WHERE ct.content_id = c.id AND ct.title ILIKE $%d)", argIdx)
		args = append(args, "%"+*p.Query+"%")
		argIdx++
	}

	countQuery := "SELECT COUNT(*) FROM contents c " + baseWhere
	var totalItems int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalItems); err != nil {
		return nil, 0, fmt.Errorf("failed to count contents: %w", err)
	}

	if totalItems == 0 {
		return []*domain.ContentListItem{}, 0, nil
	}

	dataQuery := fmt.Sprintf(`
		SELECT c.id, c.category_id, c.address, c.created_at, c.updated_at
		FROM contents c
		%s
		ORDER BY c.id DESC
		LIMIT $%d OFFSET $%d
	`, baseWhere, argIdx, argIdx+1)
	args = append(args, p.PerPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query contents: %w", err)
	}
	defer rows.Close()

	var items []*domain.ContentListItem
	var contentIDs []int64
	itemMap := make(map[int64]*domain.ContentListItem)

	for rows.Next() {
		var item domain.ContentListItem
		item.Translations = make(map[string]*domain.ContentTranslationListSummary)
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.Address, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan content item: %w", err)
		}
		items = append(items, &item)
		contentIDs = append(contentIDs, item.ID)
		itemMap[item.ID] = &item
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Fetch translations for these content IDs
	if len(contentIDs) > 0 {
		transQuery := `
			SELECT content_id, locale, title, status
			FROM content_translations
			WHERE content_id = ANY($1)
		`
		tRows, err := r.pool.Query(ctx, transQuery, contentIDs)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to query translation summaries: %w", err)
		}
		defer tRows.Close()

		for tRows.Next() {
			var contentID int64
			var locale string
			var title *string
			var status string
			if err := tRows.Scan(&contentID, &locale, &title, &status); err != nil {
				return nil, 0, fmt.Errorf("failed to scan translation summary: %w", err)
			}
			if item, exists := itemMap[contentID]; exists {
				item.Translations[locale] = &domain.ContentTranslationListSummary{
					Title:  title,
					Status: domain.TranslationStatus(status),
				}
			}
		}
	}

	return items, totalItems, nil
}

