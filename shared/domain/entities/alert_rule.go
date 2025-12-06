package entities

import (
	"database/sql"
	"time"
)

// AlertRule represents an alert rule configuration
type AlertRule struct {
	ID             string         `db:"id" json:"id"`
	UserID         int            `db:"user_id" json:"user_id"`
	MonitorID      sql.NullString `db:"monitor_id" json:"monitor_id,omitempty" swaggertype:"string"`
	Name           string         `db:"name" json:"name"`
	TriggerType    string         `db:"trigger_type" json:"trigger_type"` // 'down', 'ssl_expiry', 'slow_response'
	ThresholdValue int            `db:"threshold_value" json:"threshold_value"`
	Enabled        bool           `db:"enabled" json:"enabled"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at" json:"updated_at"`
}

// AlertRuleWithChannels represents an alert rule with its associated channels
type AlertRuleWithChannels struct {
	AlertRule
	ChannelIDs []string `json:"channel_ids,omitempty"`
}
