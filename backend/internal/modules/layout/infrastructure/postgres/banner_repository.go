package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"pidiekab/internal/modules/layout/domain"
)

type BannerRepository struct {
	pool *pgxpool.Pool
}

func NewBannerRepository(pool *pgxpool.Pool) *BannerRepository {
	repo := &BannerRepository{pool: pool}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = repo.EnsureSchema(ctx)
	return repo
}

func (r *BannerRepository) EnsureSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS layout_banners (
		id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		title TEXT NOT NULL,
		placement TEXT NOT NULL,
		banner_type TEXT NOT NULL DEFAULT 'image',
		image_url TEXT NOT NULL DEFAULT '',
		target_url TEXT NOT NULL DEFAULT '',
		html_code TEXT NOT NULL DEFAULT '',
		open_in_new_tab BOOLEAN NOT NULL DEFAULT TRUE,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		sort_order INT NOT NULL DEFAULT 0,
		start_date TIMESTAMPTZ,
		end_date TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS layout_banners_placement_idx ON layout_banners (placement, is_active, sort_order ASC);
	`
	if _, err := r.pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("failed to ensure layout_banners table: %w", err)
	}
	return nil
}

func (r *BannerRepository) List(ctx context.Context, placement string) ([]domain.Banner, error) {
	query := `
		SELECT id, title, placement, banner_type, image_url, target_url, html_code,
		       open_in_new_tab, is_active, sort_order, start_date, end_date, created_at, updated_at
		FROM layout_banners
	`
	var rows pgx.Rows
	var err error

	if placement != "" {
		query += " WHERE placement = $1 ORDER BY sort_order ASC, id ASC"
		rows, err = r.pool.Query(ctx, query, placement)
	} else {
		query += " ORDER BY placement ASC, sort_order ASC, id ASC"
		rows, err = r.pool.Query(ctx, query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var banners []domain.Banner
	for rows.Next() {
		var b domain.Banner
		if err := rows.Scan(
			&b.ID, &b.Title, &b.Placement, &b.BannerType, &b.ImageURL, &b.TargetURL, &b.HTMLCode,
			&b.OpenInNewTab, &b.IsActive, &b.SortOrder, &b.StartDate, &b.EndDate, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		banners = append(banners, b)
	}
	return banners, nil
}

func (r *BannerRepository) ListActive(ctx context.Context, placement string) ([]domain.Banner, error) {
	query := `
		SELECT id, title, placement, banner_type, image_url, target_url, html_code,
		       open_in_new_tab, is_active, sort_order, start_date, end_date, created_at, updated_at
		FROM layout_banners
		WHERE is_active = TRUE
		  AND (start_date IS NULL OR start_date <= NOW())
		  AND (end_date IS NULL OR end_date >= NOW())
	`
	var rows pgx.Rows
	var err error

	if placement != "" {
		query += " AND placement = $1 ORDER BY sort_order ASC, id ASC"
		rows, err = r.pool.Query(ctx, query, placement)
	} else {
		query += " ORDER BY placement ASC, sort_order ASC, id ASC"
		rows, err = r.pool.Query(ctx, query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var banners []domain.Banner
	for rows.Next() {
		var b domain.Banner
		if err := rows.Scan(
			&b.ID, &b.Title, &b.Placement, &b.BannerType, &b.ImageURL, &b.TargetURL, &b.HTMLCode,
			&b.OpenInNewTab, &b.IsActive, &b.SortOrder, &b.StartDate, &b.EndDate, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		banners = append(banners, b)
	}
	return banners, nil
}

func (r *BannerRepository) GetByID(ctx context.Context, id int64) (*domain.Banner, error) {
	query := `
		SELECT id, title, placement, banner_type, image_url, target_url, html_code,
		       open_in_new_tab, is_active, sort_order, start_date, end_date, created_at, updated_at
		FROM layout_banners
		WHERE id = $1
	`
	var b domain.Banner
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.Title, &b.Placement, &b.BannerType, &b.ImageURL, &b.TargetURL, &b.HTMLCode,
		&b.OpenInNewTab, &b.IsActive, &b.SortOrder, &b.StartDate, &b.EndDate, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBannerNotFound
		}
		return nil, err
	}
	return &b, nil
}

func (r *BannerRepository) Create(ctx context.Context, b *domain.Banner) error {
	query := `
		INSERT INTO layout_banners (
			title, placement, banner_type, image_url, target_url, html_code,
			open_in_new_tab, is_active, sort_order, start_date, end_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		b.Title, b.Placement, b.BannerType, b.ImageURL, b.TargetURL, b.HTMLCode,
		b.OpenInNewTab, b.IsActive, b.SortOrder, b.StartDate, b.EndDate,
	).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
}

func (r *BannerRepository) Update(ctx context.Context, b *domain.Banner) error {
	query := `
		UPDATE layout_banners
		SET title = $1, placement = $2, banner_type = $3, image_url = $4,
		    target_url = $5, html_code = $6, open_in_new_tab = $7, is_active = $8,
		    sort_order = $9, start_date = $10, end_date = $11, updated_at = NOW()
		WHERE id = $12
	`
	tag, err := r.pool.Exec(ctx, query,
		b.Title, b.Placement, b.BannerType, b.ImageURL,
		b.TargetURL, b.HTMLCode, b.OpenInNewTab, b.IsActive,
		b.SortOrder, b.StartDate, b.EndDate, b.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBannerNotFound
	}
	return nil
}

func (r *BannerRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM layout_banners WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBannerNotFound
	}
	return nil
}

func (r *BannerRepository) ToggleActive(ctx context.Context, id int64, isActive bool) error {
	query := `UPDATE layout_banners SET is_active = $1, updated_at = NOW() WHERE id = $2`
	tag, err := r.pool.Exec(ctx, query, isActive, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBannerNotFound
	}
	return nil
}
