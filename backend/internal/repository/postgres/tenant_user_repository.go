package postgres

import (
	"fmt"

	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
	"github.com/jmoiron/sqlx"
)

// tenantUserRepository implements the TenantUserRepository interface using PostgreSQL
type tenantUserRepository struct {
	db *sqlx.DB
}

// NewTenantUserRepository creates a new PostgreSQL tenant-user repository
func NewTenantUserRepository(db *sqlx.DB) repository.TenantUserRepository {
	return &tenantUserRepository{db: db}
}

// AddUserToTenant adds a user to a tenant with a specific role
func (r *tenantUserRepository) AddUserToTenant(tenantUser *entities.TenantUser) error {
	query := `
		INSERT INTO tenant_users (tenant_id, user_id, role, created_at)
		VALUES ($1, $2, $3, NOW())
	`

	_, err := r.db.Exec(query, tenantUser.TenantID, tenantUser.UserID, tenantUser.Role)
	if err != nil {
		return fmt.Errorf("failed to add user to tenant: %w", err)
	}

	return nil
}

// HasAccess checks if a user has access to a tenant
func (r *tenantUserRepository) HasAccess(userID, tenantID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM tenant_users
			WHERE user_id = $1 AND tenant_id = $2
		)
	`

	var hasAccess bool
	err := r.db.Get(&hasAccess, query, userID, tenantID)
	if err != nil {
		return false, fmt.Errorf("failed to check user tenant access: %w", err)
	}

	return hasAccess, nil
}
