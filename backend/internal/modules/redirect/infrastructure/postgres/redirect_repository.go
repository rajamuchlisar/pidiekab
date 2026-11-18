package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pidiekab/internal/modules/redirect/domain"
)

type RedirectRepository struct {
	pool *pgxpool.Pool
}

func NewRedirectRepository(pool *pgxpool.Pool) *RedirectRepository {
	return &RedirectRepository{pool: pool}
}

// FindTargetByOldPath queries for an existing active redirect for the specified old_path.
func (r *RedirectRepository) FindTargetByOldPath(ctx context.Context, oldPath string) (*domain.Redirect, error) {
	norm := domain.NormalizePath(oldPath)
	if norm == "" {
		return nil, nil
	}

	query := `
		SELECT id, old_path, new_path, status_code, created_by, created_at
		FROM url_redirects
		WHERE old_path = $1
	`

	var red domain.Redirect
	err := r.pool.QueryRow(ctx, query, norm).Scan(
		&red.ID,
		&red.OldPath,
		&red.NewPath,
		&red.StatusCode,
		&red.CreatedBy,
		&red.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query redirect: %w", err)
	}

	return &red, nil
}

// Upsert creates or updates a permanent 301 redirect entry.
func (r *RedirectRepository) Upsert(ctx context.Context, red *domain.Redirect) error {
	normOld := domain.NormalizePath(red.OldPath)
	normNew := domain.NormalizePath(red.NewPath)

	if normOld == "" || normNew == "" {
		return domain.ErrEmptyPath
	}
	if normOld == normNew {
		return domain.ErrSelfRedirect
	}

	query := `
		INSERT INTO url_redirects (old_path, new_path, status_code, created_by, created_at)
		VALUES ($1, $2, 301, $3, NOW())
		ON CONFLICT (old_path) DO UPDATE
		SET new_path = EXCLUDED.new_path,
		    status_code = 301,
		    created_by = EXCLUDED.created_by
		RETURNING id, created_at
	`

	err := r.pool.QueryRow(ctx, query, normOld, normNew, red.CreatedBy).Scan(&red.ID, &red.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to upsert redirect: %w", err)
	}

	red.OldPath = normOld
	red.NewPath = normNew
	red.StatusCode = 301
	return nil
}

// List returns a paginated list of redirects.
func (r *RedirectRepository) List(ctx context.Context, limit, offset int) ([]domain.Redirect, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM url_redirects`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count redirects: %w", err)
	}

	query := `
		SELECT id, old_path, new_path, status_code, created_by, created_at
		FROM url_redirects
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query redirects list: %w", err)
	}
	defer rows.Close()

	var list []domain.Redirect
	for rows.Next() {
		var red domain.Redirect
		if err := rows.Scan(
			&red.ID,
			&red.OldPath,
			&red.NewPath,
			&red.StatusCode,
			&red.CreatedBy,
			&red.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan redirect row: %w", err)
		}
		list = append(list, red)
	}

	return list, total, nil
}

// Delete removes a redirect entry by its ID.
func (r *RedirectRepository) Delete(ctx context.Context, id int64) error {
	cmdTag, err := r.pool.Exec(ctx, `DELETE FROM url_redirects WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete redirect: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("redirect not found")
	}
	return nil
}

// UpsertRedirect is a convenience helper for modules recording redirects.
func (r *RedirectRepository) UpsertRedirect(ctx context.Context, oldPath, newPath string, statusCode int, adminID *int64, _ *string) error {
	red := &domain.Redirect{
		OldPath:    oldPath,
		NewPath:    newPath,
		StatusCode: statusCode,
		CreatedBy:  adminID,
	}
	return r.Upsert(ctx, red)
}
