package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/jmoiron/sqlx"
)

type maintenanceWindowRepository struct {
	db *sqlx.DB
}

// NewMaintenanceWindowRepository creates a new PostgreSQL maintenance window repository
func NewMaintenanceWindowRepository(db *sqlx.DB) repository.MaintenanceWindowRepository {
	return &maintenanceWindowRepository{db: db}
}

func (r *maintenanceWindowRepository) Create(window *entities.MaintenanceWindow) error {
	query := `
		INSERT INTO maintenance_windows (user_id, name, start_time, end_time, repeat_interval, monitor_ids, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		window.UserID,
		window.Name,
		window.StartTime,
		window.EndTime,
		window.RepeatInterval,
		window.MonitorIDs,
		window.Tags,
	).Scan(&window.ID, &window.CreatedAt, &window.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create maintenance window: %w", err)
	}

	return nil
}

func (r *maintenanceWindowRepository) GetByID(id string) (*entities.MaintenanceWindow, error) {
	window := &entities.MaintenanceWindow{}
	query := `
		SELECT id, user_id, name, start_time, end_time, repeat_interval, monitor_ids, tags, created_at, updated_at
		FROM maintenance_windows
		WHERE id = $1
	`

	err := r.db.Get(window, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("maintenance window not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get maintenance window: %w", err)
	}

	return window, nil
}

func (r *maintenanceWindowRepository) GetByUserID(userID int) ([]*entities.MaintenanceWindow, error) {
	var windows []*entities.MaintenanceWindow
	query := `
		SELECT id, user_id, name, start_time, end_time, repeat_interval, monitor_ids, tags, created_at, updated_at
		FROM maintenance_windows
		WHERE user_id = $1
		ORDER BY start_time DESC
	`

	err := r.db.Select(&windows, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance windows: %w", err)
	}

	return windows, nil
}

func (r *maintenanceWindowRepository) GetActiveWindows(now time.Time) ([]*entities.MaintenanceWindow, error) {
	var windows []*entities.MaintenanceWindow
	// This query handles one-time windows.
	// For recurring windows, logic is more complex and might be better handled in application code or sophisticated query.
	// For now, let's just fetch windows where now is between start and end.

	query := `
		SELECT id, user_id, name, start_time, end_time, repeat_interval, monitor_ids, tags, created_at, updated_at
		FROM maintenance_windows
		WHERE start_time <= $1 AND ((end_time >= $1) OR (repeat_interval > 0))
	`

	err := r.db.Select(&windows, query, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get active maintenance windows: %w", err)
	}

	return windows, nil
}

func (r *maintenanceWindowRepository) Update(window *entities.MaintenanceWindow) error {
	query := `
		UPDATE maintenance_windows
		SET name = $1, start_time = $2, end_time = $3, repeat_interval = $4, monitor_ids = $5, tags = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		window.Name,
		window.StartTime,
		window.EndTime,
		window.RepeatInterval,
		window.MonitorIDs,
		window.Tags,
		window.ID,
	).Scan(&window.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("maintenance window not found: %w", err)
		}
		return fmt.Errorf("failed to update maintenance window: %w", err)
	}

	return nil
}

func (r *maintenanceWindowRepository) Delete(id string) error {
	query := `DELETE FROM maintenance_windows WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete maintenance window: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("maintenance window not found")
	}

	return nil
}
