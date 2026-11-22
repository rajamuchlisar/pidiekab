package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"pidiekab/internal/modules/identity/domain"
)

// AdminRepository manages persistence for administrators in PostgreSQL.
type AdminRepository struct {
	pool *pgxpool.Pool
}

// NewAdminRepository creates a new AdminRepository and ensures schema compatibility.
func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	repo := &AdminRepository{pool: pool}
	_ = repo.EnsureSchema(context.Background())
	return repo
}

// EnsureSchema ensures role column exists in admins table.
func (r *AdminRepository) EnsureSchema(ctx context.Context) error {
	query := `
		ALTER TABLE admins ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'super_admin';
		UPDATE admins SET role = 'super_admin' WHERE role IS NULL OR role = '';
	`
	_, err := r.pool.Exec(ctx, query)
	return err
}

// Create inserts a new administrator record into the admins table.
func (r *AdminRepository) Create(ctx context.Context, a *domain.Admin) error {
	normalizedEmail := domain.NormalizeEmail(a.Email)
	role := a.Role
	if !domain.IsValidRole(role) {
		role = domain.RoleEditor
	}

	query := `
		INSERT INTO admins (email, password_hash, display_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, normalizedEmail, a.PasswordHash, a.DisplayName, role, a.IsActive)
	if err := row.Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return domain.ErrEmailAlreadyExists
		}
		return fmt.Errorf("failed to insert admin: %w", err)
	}

	a.Email = normalizedEmail
	a.Role = role
	return nil
}

// GetByID finds an administrator by ID.
func (r *AdminRepository) GetByID(ctx context.Context, id int64) (*domain.Admin, error) {
	query := `
		SELECT id, email, password_hash, display_name, COALESCE(role, 'super_admin'), is_active, last_login_at, created_at, updated_at
		FROM admins
		WHERE id = $1
	`

	var a domain.Admin
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&a.ID,
		&a.Email,
		&a.PasswordHash,
		&a.DisplayName,
		&a.Role,
		&a.IsActive,
		&a.LastLoginAt,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAdminNotFound
		}
		return nil, fmt.Errorf("failed to query admin by id: %w", err)
	}

	return &a, nil
}

// GetByEmail finds an administrator by email (case-insensitive normalized).
func (r *AdminRepository) GetByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	normalized := domain.NormalizeEmail(email)
	query := `
		SELECT id, email, password_hash, display_name, COALESCE(role, 'super_admin'), is_active, last_login_at, created_at, updated_at
		FROM admins
		WHERE email = $1
	`

	var a domain.Admin
	err := r.pool.QueryRow(ctx, query, normalized).Scan(
		&a.ID,
		&a.Email,
		&a.PasswordHash,
		&a.DisplayName,
		&a.Role,
		&a.IsActive,
		&a.LastLoginAt,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAdminNotFound
		}
		return nil, fmt.Errorf("failed to query admin by email: %w", err)
	}

	return &a, nil
}

// UpdatePassword updates an administrator's password hash and updated_at timestamp.
func (r *AdminRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	query := `
		UPDATE admins
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`

	cmdTag, err := r.pool.Exec(ctx, query, passwordHash, id)
	if err != nil {
		return fmt.Errorf("failed to update admin password: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrAdminNotFound
	}
	return nil
}

// UpdateRole updates an administrator's role.
func (r *AdminRepository) UpdateRole(ctx context.Context, id int64, role string) error {
	query := `
		UPDATE admins
		SET role = $1, updated_at = NOW()
		WHERE id = $2
	`

	cmdTag, err := r.pool.Exec(ctx, query, role, id)
	if err != nil {
		return fmt.Errorf("failed to update admin role: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrAdminNotFound
	}
	return nil
}

// UpdateLastLogin updates the last_login_at timestamp for an admin.
func (r *AdminRepository) UpdateLastLogin(ctx context.Context, id int64, loginTime time.Time) error {
	query := `
		UPDATE admins
		SET last_login_at = $1
		WHERE id = $2
	`

	cmdTag, err := r.pool.Exec(ctx, query, loginTime, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrAdminNotFound
	}
	return nil
}

// CountActive returns the number of currently active administrators.
func (r *AdminRepository) CountActive(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM admins WHERE is_active = TRUE`
	var count int
	if err := r.pool.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count active admins: %w", err)
	}
	return count, nil
}

// CountActiveSuperAdmins returns the number of active super administrators.
func (r *AdminRepository) CountActiveSuperAdmins(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM admins WHERE is_active = TRUE AND role = 'super_admin'`
	var count int
	if err := r.pool.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count active super admins: %w", err)
	}
	return count, nil
}

// SetIsActive toggles active status of an admin.
func (r *AdminRepository) SetIsActive(ctx context.Context, id int64, isActive bool) error {
	query := `
		UPDATE admins
		SET is_active = $1, updated_at = NOW()
		WHERE id = $2
	`

	cmdTag, err := r.pool.Exec(ctx, query, isActive, id)
	if err != nil {
		return fmt.Errorf("failed to set admin active status: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrAdminNotFound
	}
	return nil
}

// List returns all administrators ordered by ID.
func (r *AdminRepository) List(ctx context.Context) ([]*domain.Admin, error) {
	query := `
		SELECT id, email, password_hash, display_name, COALESCE(role, 'super_admin'), is_active, last_login_at, created_at, updated_at
		FROM admins
		ORDER BY id ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query admins: %w", err)
	}
	defer rows.Close()

	var admins []*domain.Admin
	for rows.Next() {
		var a domain.Admin
		if err := rows.Scan(
			&a.ID,
			&a.Email,
			&a.PasswordHash,
			&a.DisplayName,
			&a.Role,
			&a.IsActive,
			&a.LastLoginAt,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan admin row: %w", err)
		}
		admins = append(admins, &a)
	}

	return admins, rows.Err()
}
