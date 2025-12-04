package entities

import (
	"time"

	"github.com/lib/pq"
)

// MaintenanceWindow represents a scheduled maintenance period
type MaintenanceWindow struct {
	ID        string    `db:"id" json:"id"`
	TenantID  int       `db:"tenant_id" json:"tenant_id"`
	Name      string    `db:"name" json:"name"`
	StartTime time.Time `db:"start_time" json:"start_time"`
	EndTime   time.Time `db:"end_time" json:"end_time"`
	// Repeat interval in seconds (0 for one-time)
	RepeatInterval int `db:"repeat_interval" json:"repeat_interval"`
	// Which monitors does this apply to? Empty means ALL.
	MonitorIDs pq.StringArray `db:"monitor_ids" json:"monitor_ids,omitempty"`
	// Or specific tags?
	Tags pq.StringArray `db:"tags" json:"tags,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
