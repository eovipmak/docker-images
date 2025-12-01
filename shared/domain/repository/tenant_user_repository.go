package repository

import "github.com/eovipmak/v-insight/shared/domain/entities"

// TenantUserRepository defines the interface for tenant-user relationship operations
type TenantUserRepository interface {
	// AddUserToTenant adds a user to a tenant with a specific role
	AddUserToTenant(tenantUser *entities.TenantUser) error

	// HasAccess checks if a user has access to a tenant
	HasAccess(userID, tenantID int) (bool, error)
}
