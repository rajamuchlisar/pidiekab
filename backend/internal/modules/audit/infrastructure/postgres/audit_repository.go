package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"pidiekab/internal/modules/audit/domain"
)

// Repository provides PostgreSQL persistence for audit logs.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new audit log repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Record inserts a new security or operational audit event into audit_logs.
func (r *Repository) Record(ctx context.Context, entry domain.Log) error {
	query := `
		INSERT INTO audit_logs (admin_id, action, entity_type, entity_id, metadata, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, COALESCE(NULLIF($7, '0001-01-01T00:00:00Z'::timestamptz), NOW()))
		RETURNING id, created_at
	`

	var metaJSON []byte
	if entry.Metadata != nil {
		var err error
		metaJSON, err = json.Marshal(entry.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal audit metadata: %w", err)
		}
	} else {
		metaJSON = []byte("{}")
	}

	var ipParam any = nil
	if entry.IPAddress != nil && *entry.IPAddress != "" {
		parsedIP := net.ParseIP(*entry.IPAddress)
		if parsedIP != nil {
			ipParam = parsedIP.String()
		}
	}

	row := r.pool.QueryRow(ctx, query,
		entry.AdminID,
		entry.Action,
		entry.EntityType,
		entry.EntityID,
		metaJSON,
		ipParam,
		entry.CreatedAt,
	)

	var id int64
	var createdAt = entry.CreatedAt
	if err := row.Scan(&id, &createdAt); err != nil {
		return fmt.Errorf("failed to record audit log: %w", err)
	}

	return nil
}
