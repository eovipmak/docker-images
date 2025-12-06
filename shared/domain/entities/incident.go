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
	UserID       int          `db:"user_id" json:"user_id"`
	StartedAt    time.Time    `db:"started_at" json:"started_at"`
	ResolvedAt   sql.NullTime `db:"resolved_at" json:"resolved_at,omitempty" swaggertype:"string"`
	Status       string       `db:"status" json:"status"` // 'open', 'resolved'
	TriggerValue string       `db:"trigger_value" json:"trigger_value,omitempty"`
	NotifiedAt   sql.NullTime `db:"notified_at" json:"notified_at,omitempty" swaggertype:"string"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
	
	// Joined fields
	MonitorName   string `db:"monitor_name" json:"monitor_name,omitempty"`
	MonitorURL    string `db:"monitor_url" json:"monitor_url,omitempty"`
	AlertRuleName string `db:"alert_rule_name" json:"alert_rule_name,omitempty"`
	TriggerType   string `db:"trigger_type" json:"trigger_type,omitempty"`
}
