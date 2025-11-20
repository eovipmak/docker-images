package entities

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// AlertChannel represents an alert channel configuration
type AlertChannel struct {
	ID        string         `db:"id" json:"id"`
	TenantID  int            `db:"tenant_id" json:"tenant_id"`
	Type      string         `db:"type" json:"type"` // 'webhook', 'discord', 'email'
	Name      string         `db:"name" json:"name"`
	Config    ChannelConfig  `db:"config" json:"config"`
	Enabled   bool           `db:"enabled" json:"enabled"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
}

// ChannelConfig represents the JSONB configuration for an alert channel
type ChannelConfig map[string]interface{}

// Value implements the driver.Valuer interface for database serialization
func (c ChannelConfig) Value() (driver.Value, error) {
	if c == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface for database deserialization
func (c *ChannelConfig) Scan(value interface{}) error {
	if value == nil {
		*c = make(ChannelConfig)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan ChannelConfig: value is not []byte")
	}

	return json.Unmarshal(bytes, c)
}
