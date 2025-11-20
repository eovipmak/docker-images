package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
	"github.com/jmoiron/sqlx"
)

// monitorRepository implements the MonitorRepository interface using PostgreSQL
type monitorRepository struct {
	db *sqlx.DB
}

// NewMonitorRepository creates a new PostgreSQL monitor repository
func NewMonitorRepository(db *sqlx.DB) repository.MonitorRepository {
	return &monitorRepository{db: db}
}

// Create creates a new monitor in the database
func (r *monitorRepository) Create(monitor *entities.Monitor) error {
	query := `
		INSERT INTO monitors (tenant_id, name, url, check_interval, timeout, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		monitor.TenantID,
		monitor.Name,
		monitor.URL,
		monitor.CheckInterval,
		monitor.Timeout,
		monitor.Enabled,
	).Scan(&monitor.ID, &monitor.CreatedAt, &monitor.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create monitor: %w", err)
	}

	return nil
}

// GetByID retrieves a monitor by its ID
func (r *monitorRepository) GetByID(id string) (*entities.Monitor, error) {
	monitor := &entities.Monitor{}
	query := `
		SELECT id, tenant_id, name, url, check_interval, timeout, enabled, created_at, updated_at
		FROM monitors
		WHERE id = $1
	`

	err := r.db.Get(monitor, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("monitor not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get monitor: %w", err)
	}

	return monitor, nil
}

// GetByTenantID retrieves all monitors for a specific tenant
func (r *monitorRepository) GetByTenantID(tenantID int) ([]*entities.Monitor, error) {
	var monitors []*entities.Monitor
	query := `
		SELECT id, tenant_id, name, url, check_interval, timeout, enabled, created_at, updated_at
		FROM monitors
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	err := r.db.Select(&monitors, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitors by tenant: %w", err)
	}

	return monitors, nil
}

// Update updates an existing monitor
func (r *monitorRepository) Update(monitor *entities.Monitor) error {
	query := `
		UPDATE monitors
		SET name = $1, url = $2, check_interval = $3, timeout = $4, enabled = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		monitor.Name,
		monitor.URL,
		monitor.CheckInterval,
		monitor.Timeout,
		monitor.Enabled,
		monitor.ID,
	).Scan(&monitor.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("monitor not found: %w", err)
		}
		return fmt.Errorf("failed to update monitor: %w", err)
	}

	return nil
}

// Delete deletes a monitor by its ID
func (r *monitorRepository) Delete(id string) error {
	query := `DELETE FROM monitors WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete monitor: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("monitor not found")
	}

	return nil
}

// GetChecksByMonitorID retrieves all check history for a specific monitor
func (r *monitorRepository) GetChecksByMonitorID(monitorID string, limit int) ([]*entities.MonitorCheck, error) {
	var checks []*entities.MonitorCheck
	
	// Default limit if not provided or invalid
	if limit <= 0 {
		limit = 100
	}
	
	query := `
		SELECT id, monitor_id, checked_at, status_code, response_time_ms, 
		       ssl_valid, ssl_expires_at, error_message, success
		FROM monitor_checks
		WHERE monitor_id = $1
		ORDER BY checked_at DESC
		LIMIT $2
	`

	err := r.db.Select(&checks, query, monitorID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitor checks: %w", err)
	}

	return checks, nil
}

// GetMonitorsNeedingCheck retrieves enabled monitors that need to be checked
// A monitor needs checking if:
// - It is enabled
// - last_checked_at is NULL (never checked), OR
// - last_checked_at + check_interval has passed
func (r *monitorRepository) GetMonitorsNeedingCheck(now time.Time) ([]*entities.Monitor, error) {
	var monitors []*entities.Monitor
	query := `
		SELECT id, tenant_id, name, url, check_interval, timeout, enabled, 
		       last_checked_at, created_at, updated_at
		FROM monitors
		WHERE enabled = true
		  AND (
		      last_checked_at IS NULL
		      OR last_checked_at + (check_interval || ' seconds')::INTERVAL <= $1
		  )
		ORDER BY last_checked_at ASC NULLS FIRST
	`

	err := r.db.Select(&monitors, query, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitors needing check: %w", err)
	}

	return monitors, nil
}

// SaveCheck saves a monitor check result to the database
func (r *monitorRepository) SaveCheck(check *entities.MonitorCheck) error {
	query := `
		INSERT INTO monitor_checks (
			monitor_id, checked_at, status_code, response_time_ms, 
			ssl_valid, ssl_expires_at, error_message, success
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		check.MonitorID,
		check.CheckedAt,
		check.StatusCode,
		check.ResponseTimeMs,
		check.SSLValid,
		check.SSLExpiresAt,
		check.ErrorMessage,
		check.Success,
	).Scan(&check.ID)

	if err != nil {
		return fmt.Errorf("failed to save monitor check: %w", err)
	}

	return nil
}

// UpdateLastCheckedAt updates the last_checked_at timestamp for a monitor
func (r *monitorRepository) UpdateLastCheckedAt(monitorID string, checkedAt time.Time) error {
	query := `
		UPDATE monitors
		SET last_checked_at = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(query, checkedAt, monitorID)
	if err != nil {
		return fmt.Errorf("failed to update last_checked_at: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("monitor not found")
	}

	return nil
}
