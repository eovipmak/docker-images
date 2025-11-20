package entities

import (
	"database/sql"
	"time"
)

// MonitorCheck represents a single monitoring check result
type MonitorCheck struct {
	ID             string         `db:"id" json:"id"`
	MonitorID      string         `db:"monitor_id" json:"monitor_id"`
	CheckedAt      time.Time      `db:"checked_at" json:"checked_at"`
	StatusCode     sql.NullInt64  `db:"status_code" json:"status_code,omitempty"`
	ResponseTimeMs sql.NullInt64  `db:"response_time_ms" json:"response_time_ms,omitempty"`
	SSLValid       sql.NullBool   `db:"ssl_valid" json:"ssl_valid,omitempty"`
	SSLExpiresAt   sql.NullTime   `db:"ssl_expires_at" json:"ssl_expires_at,omitempty"`
	ErrorMessage   sql.NullString `db:"error_message" json:"error_message,omitempty"`
	Success        bool           `db:"success" json:"success"`
}
