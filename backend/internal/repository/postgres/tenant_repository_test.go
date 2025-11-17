package postgres

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewTenantRepository(sqlxDB)

	tenant := &entities.Tenant{
		Name:    "Test Tenant",
		Slug:    "test-tenant",
		OwnerID: 1,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(1, now, now)

	mock.ExpectQuery(`INSERT INTO tenants`).
		WithArgs(tenant.Name, tenant.Slug, tenant.OwnerID).
		WillReturnRows(rows)

	err = repo.Create(tenant)
	assert.NoError(t, err)
	assert.Equal(t, 1, tenant.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTenantRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewTenantRepository(sqlxDB)

	expectedTenant := &entities.Tenant{
		ID:      1,
		Name:    "Test Tenant",
		Slug:    "test-tenant",
		OwnerID: 1,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "slug", "owner_id", "created_at", "updated_at"}).
		AddRow(expectedTenant.ID, expectedTenant.Name, expectedTenant.Slug, expectedTenant.OwnerID, now, now)

	mock.ExpectQuery(`SELECT id, name, slug, owner_id, created_at, updated_at FROM tenants WHERE id`).
		WithArgs(1).
		WillReturnRows(rows)

	tenant, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, expectedTenant.Name, tenant.Name)
	assert.Equal(t, expectedTenant.Slug, tenant.Slug)
	assert.Equal(t, expectedTenant.ID, tenant.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTenantRepository_GetBySlug(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewTenantRepository(sqlxDB)

	expectedTenant := &entities.Tenant{
		ID:      1,
		Name:    "Test Tenant",
		Slug:    "test-tenant",
		OwnerID: 1,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "slug", "owner_id", "created_at", "updated_at"}).
		AddRow(expectedTenant.ID, expectedTenant.Name, expectedTenant.Slug, expectedTenant.OwnerID, now, now)

	mock.ExpectQuery(`SELECT id, name, slug, owner_id, created_at, updated_at FROM tenants WHERE slug`).
		WithArgs("test-tenant").
		WillReturnRows(rows)

	tenant, err := repo.GetBySlug("test-tenant")
	assert.NoError(t, err)
	assert.Equal(t, expectedTenant.Name, tenant.Name)
	assert.Equal(t, expectedTenant.Slug, tenant.Slug)
	assert.Equal(t, expectedTenant.ID, tenant.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTenantRepository_GetUserTenants(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewTenantRepository(sqlxDB)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "slug", "owner_id", "created_at", "updated_at"}).
		AddRow(1, "Tenant 1", "tenant-1", 1, now, now).
		AddRow(2, "Tenant 2", "tenant-2", 1, now, now)

	mock.ExpectQuery(`SELECT t.id, t.name, t.slug, t.owner_id, t.created_at, t.updated_at FROM tenants t`).
		WithArgs(1).
		WillReturnRows(rows)

	tenants, err := repo.GetUserTenants(1)
	assert.NoError(t, err)
	assert.Len(t, tenants, 2)
	assert.Equal(t, "Tenant 1", tenants[0].Name)
	assert.Equal(t, "Tenant 2", tenants[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}
