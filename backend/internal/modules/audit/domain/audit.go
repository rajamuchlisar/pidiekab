package domain

import (
	"time"
)

// Log represents an entry in the system audit trail.
type Log struct {
	ID         int64          `json:"id"`
	AdminID    *int64         `json:"admin_id"`
	Action     string         `json:"action"`
	EntityType string         `json:"entity_type"`
	EntityID   *int64         `json:"entity_id"`
	Metadata   map[string]any `json:"metadata"`
	IPAddress  *string        `json:"ip_address"`
	CreatedAt  time.Time      `json:"created_at"`
}

// LogItem represents an enriched audit log record joined with admin details.
type LogItem struct {
	ID               int64          `json:"id"`
	AdminID          *int64         `json:"admin_id"`
	AdminEmail       *string        `json:"admin_email"`
	AdminDisplayName *string        `json:"admin_display_name"`
	Action           string         `json:"action"`
	EntityType       string         `json:"entity_type"`
	EntityID         *int64         `json:"entity_id"`
	Metadata         map[string]any `json:"metadata"`
	IPAddress        *string        `json:"ip_address"`
	CreatedAt        time.Time      `json:"created_at"`
}

// ListFilter provides query parameters for audit log exploration.
type ListFilter struct {
	Limit      int
	Offset     int
	Action     *string
	EntityType *string
	AdminID    *int64
}
