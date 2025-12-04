package repository

import (
	"time"

	"github.com/eovipmak/v-insight/shared/domain/entities"
)

// MaintenanceWindowRepository defines the interface for maintenance window persistence
type MaintenanceWindowRepository interface {
	Create(window *entities.MaintenanceWindow) error
	GetByID(id string) (*entities.MaintenanceWindow, error)
	GetByTenantID(tenantID int) ([]*entities.MaintenanceWindow, error)
	GetActiveWindows(now time.Time) ([]*entities.MaintenanceWindow, error)
	Update(window *entities.MaintenanceWindow) error
	Delete(id string) error
}
