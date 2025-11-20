package repository

import "github.com/eovipmak/v-insight/backend/internal/domain/entities"

// AlertRuleRepository defines the interface for alert rule data operations
type AlertRuleRepository interface {
	// Create creates a new alert rule
	Create(rule *entities.AlertRule) error

	// GetByID retrieves an alert rule by its ID
	GetByID(id string) (*entities.AlertRule, error)

	// GetByTenantID retrieves all alert rules for a specific tenant
	GetByTenantID(tenantID int) ([]*entities.AlertRule, error)

	// Update updates an existing alert rule
	Update(rule *entities.AlertRule) error

	// Delete deletes an alert rule by its ID
	Delete(id string) error

	// AttachChannels attaches channels to an alert rule
	AttachChannels(ruleID string, channelIDs []string) error

	// DetachChannels detaches channels from an alert rule
	DetachChannels(ruleID string, channelIDs []string) error

	// GetChannelsByRuleID retrieves all channel IDs associated with an alert rule
	GetChannelsByRuleID(ruleID string) ([]string, error)

	// GetWithChannels retrieves an alert rule with its associated channel IDs
	GetWithChannels(id string) (*entities.AlertRuleWithChannels, error)

	// GetAllWithChannelsByTenantID retrieves all alert rules with channels for a tenant
	GetAllWithChannelsByTenantID(tenantID int) ([]*entities.AlertRuleWithChannels, error)
}
