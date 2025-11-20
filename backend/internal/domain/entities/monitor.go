package entities

import (
	"database/sql"
	"time"
)

// Monitor represents a domain monitoring configuration
type Monitor struct {
	ID            string       `db:"id" json:"id"`
	TenantID      int          `db:"tenant_id" json:"tenant_id"`
	Name          string       `db:"name" json:"name"`
	URL           string       `db:"url" json:"url"`
	CheckInterval int          `db:"check_interval" json:"check_interval"` // seconds
	Timeout       int          `db:"timeout" json:"timeout"`                // seconds
	Enabled       bool         `db:"enabled" json:"enabled"`
	LastCheckedAt sql.NullTime `db:"last_checked_at" json:"last_checked_at,omitempty"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at" json:"updated_at"`
}
