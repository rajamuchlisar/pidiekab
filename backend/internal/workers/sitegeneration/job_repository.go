package sitegeneration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoPendingJob = errors.New("no pending generation job available")
)

type Job struct {
	ID          int64           `json:"id"`
	JobType     string          `json:"job_type"`
	Payload     json.RawMessage `json:"payload"`
	Status      string          `json:"status"`
	Attempts    int             `json:"attempts"`
	AvailableAt time.Time       `json:"available_at"`
	LockedAt    *time.Time      `json:"locked_at,omitempty"`
	LastError   *string         `json:"last_error,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type JobRepository struct {
	pool *pgxpool.Pool
}

func NewJobRepository(pool *pgxpool.Pool) *JobRepository {
	return &JobRepository{pool: pool}
}

// Enqueue inserts a new generation job within an existing transaction or pool.
func (r *JobRepository) Enqueue(ctx context.Context, tx pgx.Tx, jobType string, payload any) (int64, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal job payload: %w", err)
	}

	query := `
		INSERT INTO site_generation_jobs (job_type, payload, status, attempts, available_at, created_at, updated_at)
		VALUES ($1, $2, 'pending', 0, NOW(), NOW(), NOW())
		RETURNING id
	`

	var id int64
	if tx != nil {
		err = tx.QueryRow(ctx, query, jobType, payloadBytes).Scan(&id)
	} else {
		err = r.pool.QueryRow(ctx, query, jobType, payloadBytes).Scan(&id)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to enqueue generation job: %w", err)
	}
	return id, nil
}

// ClaimNextJob safely locks the next available pending job using FOR UPDATE SKIP LOCKED.
func (r *JobRepository) ClaimNextJob(ctx context.Context) (*Job, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin job claim transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		SELECT id, job_type, payload, status, attempts, available_at, locked_at, last_error, created_at, updated_at
		FROM site_generation_jobs
		WHERE status = 'pending' AND available_at <= NOW()
		ORDER BY available_at ASC, created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`

	var j Job
	var payloadBytes []byte
	err = tx.QueryRow(ctx, query).Scan(
		&j.ID, &j.JobType, &payloadBytes, &j.Status, &j.Attempts,
		&j.AvailableAt, &j.LockedAt, &j.LastError, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoPendingJob
		}
		return nil, fmt.Errorf("failed to query next pending job: %w", err)
	}
	j.Payload = json.RawMessage(payloadBytes)

	// Transition status to processing
	updateQuery := `
		UPDATE site_generation_jobs
		SET status = 'processing', locked_at = NOW(), attempts = attempts + 1, updated_at = NOW()
		WHERE id = $1
	`
	if _, err := tx.Exec(ctx, updateQuery, j.ID); err != nil {
		return nil, fmt.Errorf("failed to lock job %d: %w", j.ID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit job claim: %w", err)
	}

	j.Status = "processing"
	j.Attempts++
	now := time.Now().UTC()
	j.LockedAt = &now
	return &j, nil
}

func (r *JobRepository) MarkSucceeded(ctx context.Context, id int64) error {
	query := `
		UPDATE site_generation_jobs
		SET status = 'succeeded', locked_at = NULL, last_error = NULL, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *JobRepository) MarkFailed(ctx context.Context, id int64, errDetail string, retryDelay time.Duration, maxAttempts int) error {
	// Check current attempts
	var attempts int
	err := r.pool.QueryRow(ctx, `SELECT attempts FROM site_generation_jobs WHERE id = $1`, id).Scan(&attempts)
	if err != nil {
		return err
	}

	if attempts >= maxAttempts {
		// Permanently failed
		query := `
			UPDATE site_generation_jobs
			SET status = 'failed', locked_at = NULL, last_error = $1, updated_at = NOW()
			WHERE id = $2
		`
		_, err := r.pool.Exec(ctx, query, errDetail, id)
		return err
	}

	// Schedule retry
	query := `
		UPDATE site_generation_jobs
		SET status = 'pending', locked_at = NULL, last_error = $1, available_at = NOW() + $2::interval, updated_at = NOW()
		WHERE id = $3
	`
	intervalStr := fmt.Sprintf("%d seconds", int(retryDelay.Seconds()))
	_, err = r.pool.Exec(ctx, query, errDetail, intervalStr, id)
	return err
}
