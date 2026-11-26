package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pidiekab/internal/modules/media/domain"
)

type MediaRepository struct {
	pool *pgxpool.Pool
}

func NewMediaRepository(pool *pgxpool.Pool) *MediaRepository {
	return &MediaRepository{pool: pool}
}

type UpdateMetadataParams struct {
	AltTextID    string
	AltTextEN    string
	SourceURL    string
	RightsStatus string
	Attribution  string
}

func (r *MediaRepository) Create(ctx context.Context, asset *domain.MediaAsset) (*domain.MediaAsset, error) {
	query := `
		INSERT INTO media_assets (
			storage_key, original_filename, mime_type, file_size, width, height,
			alt_text_id, alt_text_en, source_url, rights_status, attribution,
			uploaded_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW()
		)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query,
		asset.StorageKey,
		asset.OriginalFilename,
		asset.MimeType,
		asset.FileSize,
		asset.Width,
		asset.Height,
		asset.AltTextID,
		asset.AltTextEN,
		asset.SourceURL,
		asset.RightsStatus,
		asset.Attribution,
		asset.UploadedBy,
	).Scan(&asset.ID, &asset.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert media asset: %w", err)
	}

	return asset, nil
}

func (r *MediaRepository) GetByID(ctx context.Context, id int64) (*domain.MediaAsset, error) {
	query := `
		SELECT
			id, storage_key, original_filename, mime_type, file_size, width, height,
			alt_text_id, alt_text_en, source_url, rights_status, attribution,
			uploaded_by, created_at
		FROM media_assets
		WHERE id = $1
	`
	var m domain.MediaAsset
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.StorageKey, &m.OriginalFilename, &m.MimeType, &m.FileSize,
		&m.Width, &m.Height, &m.AltTextID, &m.AltTextEN, &m.SourceURL,
		&m.RightsStatus, &m.Attribution, &m.UploadedBy, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMediaNotFound
		}
		return nil, fmt.Errorf("failed to get media by id: %w", err)
	}
	return &m, nil
}

func (r *MediaRepository) GetByStorageKey(ctx context.Context, key string) (*domain.MediaAsset, error) {
	query := `
		SELECT
			id, storage_key, original_filename, mime_type, file_size, width, height,
			alt_text_id, alt_text_en, source_url, rights_status, attribution,
			uploaded_by, created_at
		FROM media_assets
		WHERE storage_key = $1
	`
	var m domain.MediaAsset
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&m.ID, &m.StorageKey, &m.OriginalFilename, &m.MimeType, &m.FileSize,
		&m.Width, &m.Height, &m.AltTextID, &m.AltTextEN, &m.SourceURL,
		&m.RightsStatus, &m.Attribution, &m.UploadedBy, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMediaNotFound
		}
		return nil, fmt.Errorf("failed to get media by storage key: %w", err)
	}
	return &m, nil
}

func (r *MediaRepository) UpdateMetadata(ctx context.Context, id int64, params UpdateMetadataParams) (*domain.MediaAsset, error) {
	query := `
		UPDATE media_assets
		SET
			alt_text_id = $1,
			alt_text_en = $2,
			source_url = $3,
			rights_status = $4,
			attribution = $5
		WHERE id = $6
		RETURNING
			id, storage_key, original_filename, mime_type, file_size, width, height,
			alt_text_id, alt_text_en, source_url, rights_status, attribution,
			uploaded_by, created_at
	`
	var m domain.MediaAsset
	err := r.pool.QueryRow(ctx, query,
		params.AltTextID, params.AltTextEN, params.SourceURL,
		params.RightsStatus, params.Attribution, id,
	).Scan(
		&m.ID, &m.StorageKey, &m.OriginalFilename, &m.MimeType, &m.FileSize,
		&m.Width, &m.Height, &m.AltTextID, &m.AltTextEN, &m.SourceURL,
		&m.RightsStatus, &m.Attribution, &m.UploadedBy, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMediaNotFound
		}
		return nil, fmt.Errorf("failed to update media metadata: %w", err)
	}
	return &m, nil
}

func (r *MediaRepository) GetUsage(ctx context.Context, id int64) ([]string, error) {
	var references []string

	rows, err := r.pool.Query(ctx, `
		SELECT COALESCE(ct.title, 'Artikel #' || c.id::text)
		FROM contents c
		LEFT JOIN content_translations ct ON ct.content_id = c.id AND ct.locale = 'id'
		WHERE c.main_media_id = $1
	`, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var title string
			if err := rows.Scan(&title); err == nil && title != "" {
				references = append(references, title)
			}
		}
	}

	var logoCount, faviconCount int
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM site_settings WHERE logo_media_id = $1`, id).Scan(&logoCount)
	if logoCount > 0 {
		references = append(references, "Logo Website (Pengaturan)")
	}
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM site_settings WHERE favicon_media_id = $1`, id).Scan(&faviconCount)
	if faviconCount > 0 {
		references = append(references, "Favicon Website (Pengaturan)")
	}

	return references, nil
}

func (r *MediaRepository) Delete(ctx context.Context, id int64, force bool) error {
	references, err := r.GetUsage(ctx, id)
	if err != nil {
		return err
	}

	if len(references) > 0 && !force {
		return fmt.Errorf("%w: %s", domain.ErrMediaInUse, strings.Join(references, ", "))
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if force && len(references) > 0 {
		if _, err := tx.Exec(ctx, `UPDATE contents SET main_media_id = NULL WHERE main_media_id = $1`, id); err != nil {
			return fmt.Errorf("failed to unlink contents main_media_id: %w", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE site_settings SET logo_media_id = NULL WHERE logo_media_id = $1`, id); err != nil {
			return fmt.Errorf("failed to unlink site_settings logo_media_id: %w", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE site_settings SET favicon_media_id = NULL WHERE favicon_media_id = $1`, id); err != nil {
			return fmt.Errorf("failed to unlink site_settings favicon_media_id: %w", err)
		}
	}

	tag, err := tx.Exec(ctx, `DELETE FROM media_assets WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete media asset: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrMediaNotFound
	}

	return tx.Commit(ctx)
}

func (r *MediaRepository) List(ctx context.Context, page, perPage int, q string) ([]domain.MediaAsset, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var countQuery, listQuery string
	var args []any

	if q != "" {
		filter := "%" + q + "%"
		countQuery = `
			SELECT COUNT(*) FROM media_assets
			WHERE original_filename ILIKE $1 OR attribution ILIKE $1 OR alt_text_id ILIKE $1
		`
		listQuery = `
			SELECT
				id, storage_key, original_filename, mime_type, file_size, width, height,
				alt_text_id, alt_text_en, source_url, rights_status, attribution,
				uploaded_by, created_at,
				(
					(SELECT COUNT(*) FROM contents WHERE main_media_id = media_assets.id) +
					(SELECT COUNT(*) FROM site_settings WHERE logo_media_id = media_assets.id OR favicon_media_id = media_assets.id)
				) AS usage_count
			FROM media_assets
			WHERE original_filename ILIKE $1 OR attribution ILIKE $1 OR alt_text_id ILIKE $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []any{filter}
	} else {
		countQuery = `SELECT COUNT(*) FROM media_assets`
		listQuery = `
			SELECT
				id, storage_key, original_filename, mime_type, file_size, width, height,
				alt_text_id, alt_text_en, source_url, rights_status, attribution,
				uploaded_by, created_at,
				(
					(SELECT COUNT(*) FROM contents WHERE main_media_id = media_assets.id) +
					(SELECT COUNT(*) FROM site_settings WHERE logo_media_id = media_assets.id OR favicon_media_id = media_assets.id)
				) AS usage_count
			FROM media_assets
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count media assets: %w", err)
	}

	var rows pgx.Rows
	var err error
	if q != "" {
		rows, err = r.pool.Query(ctx, listQuery, args[0], perPage, offset)
	} else {
		rows, err = r.pool.Query(ctx, listQuery, perPage, offset)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list media assets: %w", err)
	}
	defer rows.Close()

	var assets []domain.MediaAsset
	for rows.Next() {
		var m domain.MediaAsset
		if err := rows.Scan(
			&m.ID, &m.StorageKey, &m.OriginalFilename, &m.MimeType, &m.FileSize,
			&m.Width, &m.Height, &m.AltTextID, &m.AltTextEN, &m.SourceURL,
			&m.RightsStatus, &m.Attribution, &m.UploadedBy, &m.CreatedAt,
			&m.UsageCount,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan media row: %w", err)
		}
		m.InUse = m.UsageCount > 0
		assets = append(assets, m)
	}

	return assets, total, nil
}
