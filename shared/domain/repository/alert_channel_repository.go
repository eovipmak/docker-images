package repository

import "github.com/eovipmak/v-insight/shared/domain/entities"

// AlertChannelRepository defines the interface for alert channel data operations
type AlertChannelRepository interface {
	// Create creates a new alert channel
	Create(channel *entities.AlertChannel) error

	// GetByID retrieves an alert channel by its ID
	GetByID(id string) (*entities.AlertChannel, error)

	// GetByTenantID retrieves all alert channels for a specific tenant
	GetByTenantID(tenantID int) ([]*entities.AlertChannel, error)

	// GetByAlertRuleID retrieves all alert channels associated with a specific alert rule
	GetByAlertRuleID(tenantID int, alertRuleID string) ([]*entities.AlertChannel, error)

	// Update updates an existing alert channel
	Update(channel *entities.AlertChannel) error

	// Delete deletes an alert channel by its ID
	Delete(id string) error
}
