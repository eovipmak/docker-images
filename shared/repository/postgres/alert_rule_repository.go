package postgres

import (
	"database/sql"
	"fmt"


	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// alertRuleRepository implements the AlertRuleRepository interface using PostgreSQL
type alertRuleRepository struct {
	db *sqlx.DB
}

// NewAlertRuleRepository creates a new PostgreSQL alert rule repository
func NewAlertRuleRepository(db *sqlx.DB) repository.AlertRuleRepository {
	return &alertRuleRepository{db: db}
}

// Create creates a new alert rule in the database
func (r *alertRuleRepository) Create(rule *entities.AlertRule) error {
	query := `
		INSERT INTO alert_rules (user_id, monitor_id, name, trigger_type, threshold_value, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		rule.UserID,
		rule.MonitorID,
		rule.Name,
		rule.TriggerType,
		rule.ThresholdValue,
		rule.Enabled,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create alert rule: %w", err)
	}

	return nil
}

// GetByID retrieves an alert rule by its ID (user-scoped)
func (r *alertRuleRepository) GetByID(userID int, id string) (*entities.AlertRule, error) {
	rule := &entities.AlertRule{}
	query := `
		SELECT id, user_id, monitor_id, name, trigger_type, threshold_value, enabled, created_at, updated_at
		FROM alert_rules
		WHERE user_id = $1 AND id = $2
	`

	err := r.db.Get(rule, query, userID, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alert rule not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}

	return rule, nil
}

// GetByUserID retrieves all alert rules for a specific user
func (r *alertRuleRepository) GetByUserID(userID int) ([]*entities.AlertRule, error) {
	var rules []*entities.AlertRule
	query := `
		SELECT id, user_id, monitor_id, name, trigger_type, threshold_value, enabled, created_at, updated_at
		FROM alert_rules
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	err := r.db.Select(&rules, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rules by tenant: %w", err)
	}

	return rules, nil
}

// GetAll retrieves all alert rules across all users (Admin only)
func (r *alertRuleRepository) GetAll() ([]*entities.AlertRule, error) {
	var rules []*entities.AlertRule
	query := `
		SELECT id, user_id, monitor_id, name, trigger_type, threshold_value, enabled, created_at, updated_at
		FROM alert_rules
		ORDER BY created_at DESC
	`

	err := r.db.Select(&rules, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all alert rules: %w", err)
	}

	return rules, nil
}

// GetByIDAdmin retrieves an alert rule by its ID ignoring user scope (Admin only)
func (r *alertRuleRepository) GetByIDAdmin(id string) (*entities.AlertRule, error) {
	rule := &entities.AlertRule{}
	query := `
		SELECT id, user_id, monitor_id, name, trigger_type, threshold_value, enabled, created_at, updated_at
		FROM alert_rules
		WHERE id = $1
	`

	err := r.db.Get(rule, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alert rule not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}

	return rule, nil
}

// Update updates an existing alert rule
func (r *alertRuleRepository) Update(rule *entities.AlertRule) error {
	query := `
		UPDATE alert_rules
		SET monitor_id = $1, name = $2, trigger_type = $3, threshold_value = $4, enabled = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		rule.MonitorID,
		rule.Name,
		rule.TriggerType,
		rule.ThresholdValue,
		rule.Enabled,
		rule.ID,
	).Scan(&rule.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("alert rule not found: %w", err)
		}
		return fmt.Errorf("failed to update alert rule: %w", err)
	}

	return nil
}

// Delete deletes an alert rule by its ID (user-scoped)
func (r *alertRuleRepository) Delete(userID int, id string) error {
	query := `DELETE FROM alert_rules WHERE user_id = $1 AND id = $2`

	result, err := r.db.Exec(query, userID, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert rule not found")
	}

	return nil
}

// AttachChannels attaches channels to an alert rule (user-scoped)
func (r *alertRuleRepository) AttachChannels(userID int, ruleID string, channelIDs []string) error {
	if len(channelIDs) == 0 {
		return nil
	}

	// Verify the rule belongs to the user before attaching channels
	query := `
		INSERT INTO alert_rule_channels (alert_rule_id, alert_channel_id)
		SELECT $1, $2
		WHERE EXISTS (SELECT 1 FROM alert_rules WHERE id = $1 AND user_id = $3)
		ON CONFLICT (alert_rule_id, alert_channel_id) DO NOTHING
	`

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, channelID := range channelIDs {
		if _, err := tx.Exec(query, ruleID, channelID, userID); err != nil {
			return fmt.Errorf("failed to attach channel %s: %w", channelID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DetachChannels detaches channels from an alert rule (user-scoped)
func (r *alertRuleRepository) DetachChannels(userID int, ruleID string, channelIDs []string) error {
	if len(channelIDs) == 0 {
		return nil
	}

	query := `
		DELETE FROM alert_rule_channels
		WHERE alert_rule_id = $1 AND alert_channel_id = ANY($2)
		AND EXISTS (SELECT 1 FROM alert_rules WHERE id = $1 AND user_id = $3)
	`

	_, err := r.db.Exec(query, ruleID, pq.Array(channelIDs), userID)
	if err != nil {
		return fmt.Errorf("failed to detach channels: %w", err)
	}

	return nil
}

// GetChannelsByRuleID retrieves all channel IDs associated with an alert rule (user-scoped)
func (r *alertRuleRepository) GetChannelsByRuleID(userID int, ruleID string) ([]string, error) {
	var channelIDs []string
	query := `
		SELECT arc.alert_channel_id
		FROM alert_rule_channels arc
		INNER JOIN alert_rules ar ON arc.alert_rule_id = ar.id
		WHERE arc.alert_rule_id = $1 AND ar.user_id = $2
		ORDER BY arc.created_at
	`

	err := r.db.Select(&channelIDs, query, ruleID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels by rule ID: %w", err)
	}

	return channelIDs, nil
}

// GetWithChannels retrieves an alert rule with its associated channel IDs (user-scoped)
func (r *alertRuleRepository) GetWithChannels(userID int, id string) (*entities.AlertRuleWithChannels, error) {
	rule, err := r.GetByID(userID, id)
	if err != nil {
		return nil, err
	}

	channelIDs, err := r.GetChannelsByRuleID(userID, id)
	if err != nil {
		return nil, err
	}

	return &entities.AlertRuleWithChannels{
		AlertRule:  *rule,
		ChannelIDs: channelIDs,
	}, nil
}

// GetAllWithChannelsByUserID retrieves all alert rules with channels for a user
func (r *alertRuleRepository) GetAllWithChannelsByUserID(userID int) ([]*entities.AlertRuleWithChannels, error) {
	rules, err := r.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	rulesWithChannels := make([]*entities.AlertRuleWithChannels, len(rules))
	for i, rule := range rules {
		channelIDs, err := r.GetChannelsByRuleID(userID, rule.ID)
		if err != nil {
			return nil, err
		}

		rulesWithChannels[i] = &entities.AlertRuleWithChannels{
			AlertRule:  *rule,
			ChannelIDs: channelIDs,
		}
	}

	return rulesWithChannels, nil
}

// GetAllEnabled retrieves all enabled alert rules across all users
func (r *alertRuleRepository) GetAllEnabled() ([]*entities.AlertRule, error) {
	var rules []*entities.AlertRule
	query := `
		SELECT id, user_id, monitor_id, name, trigger_type, threshold_value, enabled, created_at, updated_at
		FROM alert_rules
		WHERE enabled = true
		ORDER BY created_at DESC
	`

	err := r.db.Select(&rules, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all enabled alert rules: %w", err)
	}

	return rules, nil
}
