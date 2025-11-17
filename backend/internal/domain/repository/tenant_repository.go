package repository

import "github.com/eovipmak/v-insight/backend/internal/domain/entities"

// TenantRepository defines the interface for tenant data operations
type TenantRepository interface {
	// Create creates a new tenant
	Create(tenant *entities.Tenant) error

	// GetByID retrieves a tenant by its ID
	GetByID(id int) (*entities.Tenant, error)

	// GetBySlug retrieves a tenant by its slug
	GetBySlug(slug string) (*entities.Tenant, error)

	// GetUserTenants retrieves all tenants associated with a user
	GetUserTenants(userID int) ([]*entities.Tenant, error)
}
