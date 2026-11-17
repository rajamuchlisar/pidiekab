package domain

import (
	"time"
)

// Log represents an entry in the system audit trail.
type Log struct {
	ID         int64
	AdminID    *int64
	Action     string
	EntityType string
	EntityID   *int64
	Metadata   map[string]any
	IPAddress  *string
	CreatedAt  time.Time
}
