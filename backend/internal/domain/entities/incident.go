package entities

import (
	"database/sql"
	"time"
)

// Incident represents an incident triggered by an alert rule
type Incident struct {
	ID           string       `db:"id" json:"id"`
	MonitorID    string       `db:"monitor_id" json:"monitor_id"`
	AlertRuleID  string       `db:"alert_rule_id" json:"alert_rule_id"`
	StartedAt    time.Time    `db:"started_at" json:"started_at"`
	ResolvedAt   sql.NullTime `db:"resolved_at" json:"resolved_at,omitempty"`
	Status       string       `db:"status" json:"status"` // 'open', 'resolved'
	TriggerValue string       `db:"trigger_value" json:"trigger_value,omitempty"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
}
