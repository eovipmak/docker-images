package entities

import (
	"database/sql"
	"time"
)

// MonitorCheck represents a single monitoring check result
type MonitorCheck struct {
	ID             string         `db:"id" json:"id"`
	MonitorID      string         `db:"monitor_id" json:"monitor_id"`
	TenantID       int            `db:"tenant_id" json:"tenant_id,omitempty"`     // Populated in some queries (joins)
	MonitorType    string         `db:"monitor_type" json:"monitor_type,omitempty"` // Populated in some queries (joins)
	CheckedAt      time.Time      `db:"checked_at" json:"checked_at"`
	StatusCode     sql.NullInt64  `db:"status_code" json:"status_code,omitempty" swaggertype:"integer"`
	ResponseTimeMs sql.NullInt64  `db:"response_time_ms" json:"response_time_ms,omitempty" swaggertype:"integer"`
	SSLValid       sql.NullBool   `db:"ssl_valid" json:"ssl_valid,omitempty" swaggertype:"boolean"`
	SSLExpiresAt   sql.NullTime   `db:"ssl_expires_at" json:"ssl_expires_at,omitempty" swaggertype:"string"`
	ErrorMessage   sql.NullString `db:"error_message" json:"error_message,omitempty" swaggertype:"string"`
	Success        bool           `db:"success" json:"success"`
}
