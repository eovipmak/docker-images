package repository

import (
	"time"

	"github.com/eovipmak/v-insight/shared/domain/entities"
)

// MonitorRepository defines the interface for monitor data operations
type MonitorRepository interface {
	// Create creates a new monitor
	Create(monitor *entities.Monitor) error

	// GetByID retrieves a monitor by its ID
	GetByID(id string) (*entities.Monitor, error)

	// GetByTenantID retrieves all monitors for a specific tenant
	GetByTenantID(tenantID int) ([]*entities.Monitor, error)

	// Update updates an existing monitor
	Update(monitor *entities.Monitor) error

	// Delete deletes a monitor by its ID
	Delete(id string) error

	// GetChecksByMonitorID retrieves all check history for a specific monitor
	GetChecksByMonitorID(monitorID string, limit int) ([]*entities.MonitorCheck, error)

	// GetStatsByMonitorID retrieves response time statistics for a monitor over the last 24 hours
	GetStatsByMonitorID(monitorID string) ([]*entities.MonitorStat, error)

	// GetMonitorsNeedingCheck retrieves enabled monitors that need to be checked
	GetMonitorsNeedingCheck(now time.Time) ([]*entities.Monitor, error)

	// SaveCheck saves a monitor check result to the database
	SaveCheck(check *entities.MonitorCheck) error

	// UpdateLastCheckedAt updates the last_checked_at timestamp for a monitor
	UpdateLastCheckedAt(monitorID string, checkedAt time.Time) error

	// GetLatestMonitorChecks retrieves the latest check for each monitor within the given duration
	GetLatestMonitorChecks(duration time.Duration) ([]*entities.MonitorCheck, error)
}
