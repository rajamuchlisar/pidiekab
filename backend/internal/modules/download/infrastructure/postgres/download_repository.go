package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pidiekab/internal/modules/download/domain"
)

type DownloadRepository struct {
	pool *pgxpool.Pool
}

func NewDownloadRepository(pool *pgxpool.Pool) *DownloadRepository {
	repo := &DownloadRepository{pool: pool}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = repo.EnsureSchema(ctx)
	return repo
}

func (r *DownloadRepository) EnsureSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS downloads (
		id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		title_id TEXT NOT NULL,
		title_en TEXT NOT NULL DEFAULT '',
		category TEXT NOT NULL DEFAULT 'Umum',
		description_id TEXT NOT NULL DEFAULT '',
		description_en TEXT NOT NULL DEFAULT '',
		file_url TEXT NOT NULL,
		file_name TEXT NOT NULL DEFAULT '',
		file_size BIGINT NOT NULL DEFAULT 0,
		file_type TEXT NOT NULL DEFAULT 'PDF',
		download_count BIGINT NOT NULL DEFAULT 0,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		sort_order INT NOT NULL DEFAULT 0,
		created_by BIGINT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS downloads_active_sort_idx ON downloads (is_active, sort_order ASC, id DESC);
	CREATE INDEX IF NOT EXISTS downloads_category_idx ON downloads (category);
	`
	if _, err := r.pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("failed to ensure downloads table: %w", err)
	}

	// Clean up any legacy or dead asset URLs
	_, _ = r.pool.Exec(ctx, `
		UPDATE downloads 
		SET file_url = '/api/v1/public/downloads/' || id || '/file'
		WHERE file_url LIKE '%/assets/%';
	`)

	// Seed baseline download if empty
	var count int
	_ = r.pool.QueryRow(ctx, "SELECT count(*) FROM downloads").Scan(&count)
	if count == 0 {
		sampleQuery := `
		INSERT INTO downloads (title_id, title_en, category, description_id, description_en, file_url, file_name, file_size, file_type, sort_order)
		VALUES 
		('Panduan Resmi Wisata & Kebudayaan Kabupaten Pidie', 'Official Travel & Cultural Guide of Pidie Regency', 'Panduan Wisata', 'Panduan lengkap destinasi wisata alam, sejarah, dan sentra kuliner khas di Kabupaten Pidie.', 'Comprehensive guide to nature, heritage, and culinary destinations across Pidie Regency.', '/api/v1/public/downloads/1/file', 'panduan-wisata-pidie.pdf', 2450000, 'PDF', 1),
		('Peta Wisata & Direktori Kecamatan Pidie', 'Pidie Tourism Map & Sub-district Directory', 'Peta & Brosur', 'Peta persebaran objek wisata, rute transportasi, dan informasi fasilitas publik di seluruh kecamatan.', 'Tourist map showing destinations, transport routes, and public facilities across Pidie.', '/api/v1/public/downloads/2/file', 'peta-wisata-pidie.pdf', 3890000, 'PDF', 2),
		('Profil Potensi Investasi Daerah Kabupaten Pidie', 'Pidie Regional Investment & Economic Profile', 'Informasi Publik', 'Dokumen informasi komoditas unggulan pertanian, perkebunan, kelautan, dan peluang investasi.', 'Overview of leading agricultural, marine, and commercial investment opportunities in Pidie.', '/api/v1/public/downloads/3/file', 'profil-investasi-pidie.pdf', 1780000, 'PDF', 3)
		ON CONFLICT DO NOTHING;
		`
		_, _ = r.pool.Exec(ctx, sampleQuery)
	}

	return nil
}

type ListFilter struct {
	Category string
	IsActive *bool
	Search   string
	Limit    int
	Offset   int
}

func (r *DownloadRepository) List(ctx context.Context, f ListFilter) ([]domain.DownloadItem, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if f.Category != "" && f.Category != "Semua" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, f.Category)
		argIdx++
	}

	if f.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *f.IsActive)
		argIdx++
	}

	if f.Search != "" {
		likeTerm := "%" + strings.ToLower(f.Search) + "%"
		conditions = append(conditions, fmt.Sprintf("(LOWER(title_id) LIKE $%d OR LOWER(title_en) LIKE $%d OR LOWER(description_id) LIKE $%d OR LOWER(file_name) LIKE $%d)", argIdx, argIdx, argIdx, argIdx))
		args = append(args, likeTerm)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM downloads %s", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count downloads: %w", err)
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	queryArgs := append(args, limit, offset)
	dataQuery := fmt.Sprintf(`
		SELECT id, title_id, title_en, category, description_id, description_en,
		       file_url, file_name, file_size, file_type, download_count,
		       is_active, sort_order, created_by, created_at, updated_at
		FROM downloads
		%s
		ORDER BY sort_order ASC, id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, dataQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query downloads: %w", err)
	}
	defer rows.Close()

	var items []domain.DownloadItem
	for rows.Next() {
		var it domain.DownloadItem
		if err := rows.Scan(
			&it.ID, &it.TitleID, &it.TitleEN, &it.Category, &it.DescriptionID, &it.DescriptionEN,
			&it.FileURL, &it.FileName, &it.FileSize, &it.FileType, &it.DownloadCount,
			&it.IsActive, &it.SortOrder, &it.CreatedBy, &it.CreatedAt, &it.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, it)
	}

	return items, total, nil
}

func (r *DownloadRepository) GetByID(ctx context.Context, id int64) (*domain.DownloadItem, error) {
	query := `
		SELECT id, title_id, title_en, category, description_id, description_en,
		       file_url, file_name, file_size, file_type, download_count,
		       is_active, sort_order, created_by, created_at, updated_at
		FROM downloads
		WHERE id = $1
	`
	var it domain.DownloadItem
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&it.ID, &it.TitleID, &it.TitleEN, &it.Category, &it.DescriptionID, &it.DescriptionEN,
		&it.FileURL, &it.FileName, &it.FileSize, &it.FileType, &it.DownloadCount,
		&it.IsActive, &it.SortOrder, &it.CreatedBy, &it.CreatedAt, &it.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDownloadNotFound
		}
		return nil, err
	}
	return &it, nil
}

func (r *DownloadRepository) Create(ctx context.Context, it *domain.DownloadItem) (*domain.DownloadItem, error) {
	query := `
		INSERT INTO downloads (
			title_id, title_en, category, description_id, description_en,
			file_url, file_name, file_size, file_type, download_count,
			is_active, sort_order, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, NOW(), NOW()
		)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		it.TitleID, it.TitleEN, it.Category, it.DescriptionID, it.DescriptionEN,
		it.FileURL, it.FileName, it.FileSize, it.FileType, it.DownloadCount,
		it.IsActive, it.SortOrder, it.CreatedBy,
	).Scan(&it.ID, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create download item: %w", err)
	}
	return it, nil
}

func (r *DownloadRepository) Update(ctx context.Context, it *domain.DownloadItem) (*domain.DownloadItem, error) {
	query := `
		UPDATE downloads
		SET title_id = $1, title_en = $2, category = $3,
		    description_id = $4, description_en = $5,
		    file_url = $6, file_name = $7, file_size = $8, file_type = $9,
		    is_active = $10, sort_order = $11, updated_at = NOW()
		WHERE id = $12
		RETURNING updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		it.TitleID, it.TitleEN, it.Category,
		it.DescriptionID, it.DescriptionEN,
		it.FileURL, it.FileName, it.FileSize, it.FileType,
		it.IsActive, it.SortOrder, it.ID,
	).Scan(&it.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDownloadNotFound
		}
		return nil, fmt.Errorf("failed to update download item: %w", err)
	}
	return it, nil
}

func (r *DownloadRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM downloads WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete download item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrDownloadNotFound
	}
	return nil
}

func (r *DownloadRepository) IncrementDownloadCount(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, "UPDATE downloads SET download_count = download_count + 1 WHERE id = $1", id)
	return err
}
