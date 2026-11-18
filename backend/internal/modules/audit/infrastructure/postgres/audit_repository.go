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

// List retrieves a paginated slice of audit logs with optional filters.
func (r *Repository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.LogItem, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	var actionPattern *string
	if filter.Action != nil && *filter.Action != "" {
		p := "%" + *filter.Action + "%"
		actionPattern = &p
	}

	query := `
		SELECT
			al.id,
			al.admin_id,
			a.email,
			a.display_name,
			al.action,
			al.entity_type,
			al.entity_id,
			al.metadata,
			al.ip_address::text,
			al.created_at,
			COUNT(*) OVER() AS full_count
		FROM audit_logs al
		LEFT JOIN admins a ON a.id = al.admin_id
		WHERE ($1::text IS NULL OR al.action ILIKE $1)
		  AND ($2::text IS NULL OR al.entity_type = $2)
		  AND ($3::bigint IS NULL OR al.admin_id = $3)
		ORDER BY al.created_at DESC, al.id DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.pool.Query(ctx, query, actionPattern, filter.EntityType, filter.AdminID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var items []*domain.LogItem
	total := 0

	for rows.Next() {
		var item domain.LogItem
		var metaRaw []byte
		var ipStr *string
		var fullCount int64

		if err := rows.Scan(
			&item.ID,
			&item.AdminID,
			&item.AdminEmail,
			&item.AdminDisplayName,
			&item.Action,
			&item.EntityType,
			&item.EntityID,
			&metaRaw,
			&ipStr,
			&item.CreatedAt,
			&fullCount,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log row: %w", err)
		}

		item.IPAddress = ipStr
		if len(metaRaw) > 0 {
			var m map[string]any
			if err := json.Unmarshal(metaRaw, &m); err == nil {
				item.Metadata = m
			}
		}

		total = int(fullCount)
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed reading audit log rows: %w", err)
	}

	if items == nil {
		items = []*domain.LogItem{}
	}

	return items, total, nil
}
