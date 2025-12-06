package repository

import "github.com/eovipmak/v-insight/shared/domain/entities"

// AlertRuleRepository defines the interface for alert rule data operations
type AlertRuleRepository interface {
	// Create creates a new alert rule
	Create(rule *entities.AlertRule) error

	// GetByID retrieves an alert rule by its ID (user-scoped)
	GetByID(userID int, id string) (*entities.AlertRule, error)

	// GetByUserID retrieves all alert rules for a specific user
	GetByUserID(userID int) ([]*entities.AlertRule, error)

	// Update updates an existing alert rule
	Update(rule *entities.AlertRule) error

	// Delete deletes an alert rule by its ID (user-scoped)
	Delete(userID int, id string) error

	// AttachChannels attaches channels to an alert rule (user-scoped)
	AttachChannels(userID int, ruleID string, channelIDs []string) error

	// DetachChannels detaches channels from an alert rule (user-scoped)
	DetachChannels(userID int, ruleID string, channelIDs []string) error

	// GetChannelsByRuleID retrieves all channel IDs associated with an alert rule (user-scoped)
	GetChannelsByRuleID(userID int, ruleID string) ([]string, error)

	// GetWithChannels retrieves an alert rule with its associated channel IDs (user-scoped)
	GetWithChannels(userID int, id string) (*entities.AlertRuleWithChannels, error)

	// GetAllWithChannelsByUserID retrieves all alert rules with channels for a user
	GetAllWithChannelsByUserID(userID int) ([]*entities.AlertRuleWithChannels, error)

	// GetAllEnabled retrieves all enabled alert rules across all users
	GetAllEnabled() ([]*entities.AlertRule, error)

	// GetAll retrieves all alert rules across all users (Admin only)
	GetAll() ([]*entities.AlertRule, error)
}
