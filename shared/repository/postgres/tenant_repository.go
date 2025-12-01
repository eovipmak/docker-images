package postgres

import (
	"database/sql"
	"fmt"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/jmoiron/sqlx"
)

// tenantRepository implements the TenantRepository interface using PostgreSQL
type tenantRepository struct {
	db *sqlx.DB
}

// NewTenantRepository creates a new PostgreSQL tenant repository
func NewTenantRepository(db *sqlx.DB) repository.TenantRepository {
	return &tenantRepository{db: db}
}

// Create creates a new tenant in the database
func (r *tenantRepository) Create(tenant *entities.Tenant) error {
	query := `
		INSERT INTO tenants (name, slug, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(query, tenant.Name, tenant.Slug, tenant.OwnerID).Scan(
		&tenant.ID,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

// GetByID retrieves a tenant by its ID
func (r *tenantRepository) GetByID(id int) (*entities.Tenant, error) {
	tenant := &entities.Tenant{}
	query := `
		SELECT id, name, slug, owner_id, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`

	err := r.db.Get(tenant, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// GetBySlug retrieves a tenant by its slug
func (r *tenantRepository) GetBySlug(slug string) (*entities.Tenant, error) {
	tenant := &entities.Tenant{}
	query := `
		SELECT id, name, slug, owner_id, created_at, updated_at
		FROM tenants
		WHERE slug = $1
	`

	err := r.db.Get(tenant, query, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// GetUserTenants retrieves all tenants associated with a user
func (r *tenantRepository) GetUserTenants(userID int) ([]*entities.Tenant, error) {
	var tenants []*entities.Tenant
	query := `
		SELECT t.id, t.name, t.slug, t.owner_id, t.created_at, t.updated_at
		FROM tenants t
		INNER JOIN tenant_users tu ON t.id = tu.tenant_id
		WHERE tu.user_id = $1
		ORDER BY t.created_at DESC
	`

	err := r.db.Select(&tenants, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}

	return tenants, nil
}
