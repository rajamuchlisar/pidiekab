package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"pidiekab/internal/modules/content/domain"
)

// SourceRepository manages PostgreSQL persistence for content sources.
type SourceRepository struct {
	pool *pgxpool.Pool
}

// NewSourceRepository creates a new SourceRepository.
func NewSourceRepository(pool *pgxpool.Pool) *SourceRepository {
	return &SourceRepository{pool: pool}
}

// Create inserts a new content source record.
func (r *SourceRepository) Create(ctx context.Context, s *domain.ContentSource) error {
	query := `
		INSERT INTO content_sources (
			content_id, source_type, source_name, source_url,
			source_published_at, accessed_at, notes, created_by, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		s.ContentID,
		s.SourceType,
		s.SourceName,
		s.SourceURL,
		s.SourcePublishedAt,
		s.AccessedAt,
		s.Notes,
		s.CreatedBy,
	).Scan(&s.ID, &s.CreatedAt)
}

// GetByID finds a source by ID.
func (r *SourceRepository) GetByID(ctx context.Context, id int64) (*domain.ContentSource, error) {
	query := `
		SELECT id, content_id, source_type, source_name, source_url,
		       source_published_at, accessed_at, notes, created_by, created_at
		FROM content_sources
		WHERE id = $1
	`
	var s domain.ContentSource
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID,
		&s.ContentID,
		&s.SourceType,
		&s.SourceName,
		&s.SourceURL,
		&s.SourcePublishedAt,
		&s.AccessedAt,
		&s.Notes,
		&s.CreatedBy,
		&s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSourceNotFound
		}
		return nil, fmt.Errorf("failed to query source: %w", err)
	}
	return &s, nil
}

// Update modifies source information.
func (r *SourceRepository) Update(ctx context.Context, s *domain.ContentSource) error {
	query := `
		UPDATE content_sources
		SET source_type = $1, source_name = $2, source_url = $3,
		    source_published_at = $4, accessed_at = $5, notes = $6
		WHERE id = $7 AND content_id = $8
	`
	cmdTag, err := r.pool.Exec(ctx, query,
		s.SourceType,
		s.SourceName,
		s.SourceURL,
		s.SourcePublishedAt,
		s.AccessedAt,
		s.Notes,
		s.ID,
		s.ContentID,
	)
	if err != nil {
		return fmt.Errorf("failed to update source: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrSourceNotFound
	}
	return nil
}

// Delete removes a source record.
func (r *SourceRepository) Delete(ctx context.Context, id int64) error {
	cmdTag, err := r.pool.Exec(ctx, `DELETE FROM content_sources WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete source: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrSourceNotFound
	}
	return nil
}

// ListByContentID returns all sources associated with a content record.
func (r *SourceRepository) ListByContentID(ctx context.Context, contentID int64) ([]*domain.ContentSource, error) {
	query := `
		SELECT id, content_id, source_type, source_name, source_url,
		       source_published_at, accessed_at, notes, created_by, created_at
		FROM content_sources
		WHERE content_id = $1
		ORDER BY id ASC
	`
	rows, err := r.pool.Query(ctx, query, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sources: %w", err)
	}
	defer rows.Close()

	var sources []*domain.ContentSource
	for rows.Next() {
		var s domain.ContentSource
		if err := rows.Scan(
			&s.ID,
			&s.ContentID,
			&s.SourceType,
			&s.SourceName,
			&s.SourceURL,
			&s.SourcePublishedAt,
			&s.AccessedAt,
			&s.Notes,
			&s.CreatedBy,
			&s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan source row: %w", err)
		}
		sources = append(sources, &s)
	}
	return sources, rows.Err()
}

