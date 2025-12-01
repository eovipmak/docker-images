package repository

import (
	"github.com/eovipmak/v-insight/shared/domain/entities"
)

// StatusPageRepository defines the interface for status page data operations
type StatusPageRepository interface {
	// Create creates a new status page
	Create(statusPage *entities.StatusPage) error

	// GetByID retrieves a status page by its ID
	GetByID(id string) (*entities.StatusPage, error)

	// GetBySlug retrieves a status page by its slug
	GetBySlug(slug string) (*entities.StatusPage, error)

	// GetByTenantID retrieves all status pages for a specific tenant
	GetByTenantID(tenantID int) ([]*entities.StatusPage, error)

	// Update updates an existing status page
	Update(statusPage *entities.StatusPage) error

	// Delete deletes a status page by its ID
	Delete(id string) error

	// AddMonitor adds a monitor to a status page
	AddMonitor(statusPageID, monitorID string) error

	// RemoveMonitor removes a monitor from a status page
	RemoveMonitor(statusPageID, monitorID string) error

	// GetMonitors retrieves all monitors associated with a status page
	GetMonitors(statusPageID string) ([]*entities.Monitor, error)
}