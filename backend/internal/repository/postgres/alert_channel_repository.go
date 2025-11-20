package postgres

import (
	"database/sql"
	"fmt"

	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
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
		INSERT INTO alert_channels (tenant_id, type, name, config, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		channel.TenantID,
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
		SELECT id, tenant_id, type, name, config, enabled, created_at, updated_at
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

// GetByTenantID retrieves all alert channels for a specific tenant
func (r *alertChannelRepository) GetByTenantID(tenantID int) ([]*entities.AlertChannel, error) {
	var channels []*entities.AlertChannel
	query := `
		SELECT id, tenant_id, type, name, config, enabled, created_at, updated_at
		FROM alert_channels
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	err := r.db.Select(&channels, query, tenantID)
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
