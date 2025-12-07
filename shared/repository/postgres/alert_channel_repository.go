package postgres

import (
	"database/sql"
	"fmt"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/jmoiron/sqlx"
)

// alertChannelRepository implements the AlertChannelRepository interface using PostgreSQL
type alertChannelRepository struct {
	db *sqlx.DB
}

// NewAlertChannelRepository creates a new PostgreSQL alert channel repository
func NewAlertChannelRepository(db *sqlx.DB) repository.AlertChannelRepository {
	return &alertChannelRepository{db: db}
}

// Create creates a new alert channel in the database
func (r *alertChannelRepository) Create(channel *entities.AlertChannel) error {
	query := `
		INSERT INTO alert_channels (user_id, type, name, config, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		channel.UserID,
		channel.Type,
		channel.Name,
		channel.Config,
		channel.Enabled,
	).Scan(&channel.ID, &channel.CreatedAt, &channel.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create alert channel: %w", err)
	}

	return nil
}

// GetByID retrieves an alert channel by its ID
func (r *alertChannelRepository) GetByID(id string) (*entities.AlertChannel, error) {
	channel := &entities.AlertChannel{}
	query := `
		SELECT id, user_id, type, name, config, enabled, created_at, updated_at
		FROM alert_channels
		WHERE id = $1
	`

	err := r.db.Get(channel, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alert channel not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get alert channel: %w", err)
	}

	return channel, nil
}

// GetByUserID retrieves all alert channels for a specific user
func (r *alertChannelRepository) GetByUserID(userID int) ([]*entities.AlertChannel, error) {
	var channels []*entities.AlertChannel
	query := `
		SELECT id, user_id, type, name, config, enabled, created_at, updated_at
		FROM alert_channels
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	err := r.db.Select(&channels, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert channels by tenant: %w", err)
	}

	return channels, nil
}

// Update updates an existing alert channel
func (r *alertChannelRepository) Update(channel *entities.AlertChannel) error {
	query := `
		UPDATE alert_channels
		SET type = $1, name = $2, config = $3, enabled = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		channel.Type,
		channel.Name,
		channel.Config,
		channel.Enabled,
		channel.ID,
	).Scan(&channel.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("alert channel not found: %w", err)
		}
		return fmt.Errorf("failed to update alert channel: %w", err)
	}

	return nil
}

// Delete deletes an alert channel by its ID
func (r *alertChannelRepository) Delete(id string) error {
	query := `DELETE FROM alert_channels WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert channel: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert channel not found")
	}

	return nil
}

// GetByAlertRuleID retrieves all alert channels associated with a specific alert rule
func (r *alertChannelRepository) GetByAlertRuleID(userID int, alertRuleID string) ([]*entities.AlertChannel, error) {
	var channels []*entities.AlertChannel
	query := `
		SELECT ac.id, ac.user_id, ac.type, ac.name, ac.config, ac.enabled, ac.created_at, ac.updated_at
		FROM alert_channels ac
		JOIN alert_rule_channels arc ON ac.id = arc.alert_channel_id
		WHERE ac.user_id = $1 AND arc.alert_rule_id = $2
		ORDER BY ac.created_at ASC
	`

	err := r.db.Select(&channels, query, userID, alertRuleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert channels by alert rule: %w", err)
	}

	return channels, nil
}

// GetAll retrieves all alert channels across all users (Admin only)
func (r *alertChannelRepository) GetAll() ([]*entities.AlertChannel, error) {
	channels := make([]*entities.AlertChannel, 0)
	query := `
		SELECT id, user_id, type, name, config, enabled, created_at, updated_at
		FROM alert_channels
		ORDER BY created_at DESC
	`

	err := r.db.Select(&channels, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all alert channels: %w", err)
	}

	return channels, nil
}
