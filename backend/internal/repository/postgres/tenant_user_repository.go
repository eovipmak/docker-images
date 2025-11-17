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
