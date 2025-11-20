package repository

import "github.com/eovipmak/v-insight/backend/internal/domain/entities"

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
}
