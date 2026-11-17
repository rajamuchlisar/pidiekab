package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"pidiekab/internal/modules/identity/domain"
)

// SessionRepository manages persistence for administrator sessions in PostgreSQL.
type SessionRepository struct {
	pool *pgxpool.Pool
}

// NewSessionRepository creates a new SessionRepository.
func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

// Create inserts a new session record into admin_sessions.
func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	query := `
		INSERT INTO admin_sessions (admin_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`

	row := r.pool.QueryRow(ctx, query, s.AdminID, s.TokenHash, s.ExpiresAt)
	if err := row.Scan(&s.ID, &s.CreatedAt); err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}

	return nil
}

// GetActiveByTokenHash finds an active, unrevoked, unexpired session by token hash.
func (r *SessionRepository) GetActiveByTokenHash(ctx context.Context, tokenHash string, now time.Time) (*domain.Session, error) {
	query := `
		SELECT id, admin_id, token_hash, expires_at, revoked_at, created_at
		FROM admin_sessions
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > $2
	`

	var s domain.Session
	err := r.pool.QueryRow(ctx, query, tokenHash, now).Scan(
		&s.ID,
		&s.AdminID,
		&s.TokenHash,
		&s.ExpiresAt,
		&s.RevokedAt,
		&s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to query session by token hash: %w", err)
	}

	return &s, nil
}

// RevokeByTokenHash revokes a single session.
func (r *SessionRepository) RevokeByTokenHash(ctx context.Context, tokenHash string, revokeTime time.Time) error {
	query := `
		UPDATE admin_sessions
		SET revoked_at = $1
		WHERE token_hash = $2
		  AND revoked_at IS NULL
	`

	cmdTag, err := r.pool.Exec(ctx, query, revokeTime, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

// RevokeAllForAdmin revokes all active sessions for a specific admin (used on password change or deactivation).
func (r *SessionRepository) RevokeAllForAdmin(ctx context.Context, adminID int64, revokeTime time.Time) error {
	query := `
		UPDATE admin_sessions
		SET revoked_at = $1
		WHERE admin_id = $2
		  AND revoked_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, revokeTime, adminID)
	if err != nil {
		return fmt.Errorf("failed to revoke all sessions for admin: %w", err)
	}

	return nil
}
