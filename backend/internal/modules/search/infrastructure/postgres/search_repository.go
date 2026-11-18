package postgres

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"pidiekab/internal/modules/search/domain"
)

type SearchRepository struct {
	pool *pgxpool.Pool
}

func NewSearchRepository(pool *pgxpool.Pool) *SearchRepository {
	return &SearchRepository{pool: pool}
}

type SearchParams struct {
	Locale       string
	Query        string
	CategorySlug *string
	Page         int
	PerPage      int
}

func (r *SearchRepository) Search(ctx context.Context, p SearchParams) ([]domain.SearchResultItem, int, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 50 {
		p.PerPage = 12
	}
	offset := (p.Page - 1) * p.PerPage

	tsConfig := "simple"
	if p.Locale == "en" {
		tsConfig = "english"
	}
	likePattern := "%" + p.Query + "%"

	countQuery := `
		SELECT COUNT(*)
		FROM contents c
		JOIN content_translations ct ON ct.content_id = c.id
		JOIN categories cat ON cat.id = c.category_id AND cat.is_active = TRUE
		JOIN category_translations cat_t ON cat_t.category_id = cat.id AND cat_t.locale = ct.locale
		WHERE ct.locale = $1
		  AND ct.status = 'published'
		  AND (ct.search_vector @@ plainto_tsquery($2, $3) OR ct.title ILIKE $4)
		  AND ($5::text IS NULL OR cat_t.slug = $5)
	`

	var totalItems int
	err := r.pool.QueryRow(ctx, countQuery, p.Locale, tsConfig, p.Query, likePattern, p.CategorySlug).Scan(&totalItems)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	if totalItems == 0 {
		return []domain.SearchResultItem{}, 0, nil
	}

	searchQuery := `
		SELECT
			c.id, ct.title, ct.slug,
			cat_t.slug, cat_t.name,
			d.name, v.name,
			ts_rank_cd(ct.search_vector, plainto_tsquery($2, $3)) AS rank
		FROM contents c
		JOIN content_translations ct ON ct.content_id = c.id
		JOIN categories cat ON cat.id = c.category_id AND cat.is_active = TRUE
		JOIN category_translations cat_t ON cat_t.category_id = cat.id AND cat_t.locale = ct.locale
		LEFT JOIN villages v ON v.id = c.village_id
		LEFT JOIN districts d ON d.id = v.district_id
		WHERE ct.locale = $1
		  AND ct.status = 'published'
		  AND (ct.search_vector @@ plainto_tsquery($2, $3) OR ct.title ILIKE $4)
		  AND ($5::text IS NULL OR cat_t.slug = $5)
		ORDER BY rank DESC, c.id DESC
		LIMIT $6 OFFSET $7
	`

	rows, err := r.pool.Query(ctx, searchQuery, p.Locale, tsConfig, p.Query, likePattern, p.CategorySlug, p.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var results []domain.SearchResultItem
	for rows.Next() {
		var id int64
		var title, contentSlug, catSlug, catName string
		var districtName, villageName *string
		var rank float64

		if err := rows.Scan(
			&id, &title, &contentSlug,
			&catSlug, &catName,
			&districtName, &villageName,
			&rank,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan search row: %w", err)
		}

		url := fmt.Sprintf("/%s/%s/%s/", p.Locale, catSlug, contentSlug)

		var loc *domain.LocationSummary
		if districtName != nil || villageName != nil {
			loc = &domain.LocationSummary{
				District: districtName,
				Village:  villageName,
			}
		}

		results = append(results, domain.SearchResultItem{
			ID:    strconv.FormatInt(id, 10),
			Title: title,
			Category: domain.CategorySummary{
				Slug: catSlug,
				Name: catName,
			},
			Location: loc,
			URL:      url,
		})
	}

	return results, totalItems, nil
}
