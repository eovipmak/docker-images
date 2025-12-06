package postgres

import (
	"database/sql"
	"fmt"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/jmoiron/sqlx"
)

// statusPageRepository implements the StatusPageRepository interface using PostgreSQL
type statusPageRepository struct {
	db *sqlx.DB
}

// NewStatusPageRepository creates a new PostgreSQL status page repository
func NewStatusPageRepository(db *sqlx.DB) repository.StatusPageRepository {
	return &statusPageRepository{db: db}
}

// Create creates a new status page in the database
func (r *statusPageRepository) Create(statusPage *entities.StatusPage) error {
	query := `
		INSERT INTO status_pages (user_id, slug, name, public_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		statusPage.UserID,
		statusPage.Slug,
		statusPage.Name,
		statusPage.PublicEnabled,
	).Scan(&statusPage.ID, &statusPage.CreatedAt, &statusPage.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create status page: %w", err)
	}

	return nil
}

// GetByID retrieves a status page by its ID
func (r *statusPageRepository) GetByID(id string) (*entities.StatusPage, error) {
	statusPage := &entities.StatusPage{}
	query := `
		SELECT id, user_id, slug, name, public_enabled, created_at, updated_at
		FROM status_pages
		WHERE id = $1
	`

	err := r.db.Get(statusPage, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("status page not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get status page: %w", err)
	}

	return statusPage, nil
}

// GetBySlug retrieves a status page by its slug
func (r *statusPageRepository) GetBySlug(slug string) (*entities.StatusPage, error) {
	statusPage := &entities.StatusPage{}
	query := `
		SELECT id, user_id, slug, name, public_enabled, created_at, updated_at
		FROM status_pages
		WHERE slug = $1
	`

	err := r.db.Get(statusPage, query, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("status page not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get status page: %w", err)
	}

	return statusPage, nil
}

// GetByUserID retrieves all status pages for a specific user
func (r *statusPageRepository) GetByUserID(userID int) ([]*entities.StatusPage, error) {
	var statusPages []*entities.StatusPage
	query := `
		SELECT id, user_id, slug, name, public_enabled, created_at, updated_at
		FROM status_pages
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	err := r.db.Select(&statusPages, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status pages by tenant: %w", err)
	}

	return statusPages, nil
}

// Update updates an existing status page
func (r *statusPageRepository) Update(statusPage *entities.StatusPage) error {
	query := `
		UPDATE status_pages
		SET slug = $1, name = $2, public_enabled = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		statusPage.Slug,
		statusPage.Name,
		statusPage.PublicEnabled,
		statusPage.ID,
	).Scan(&statusPage.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update status page: %w", err)
	}

	return nil
}

// Delete deletes a status page by its ID
func (r *statusPageRepository) Delete(id string) error {
	query := `DELETE FROM status_pages WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete status page: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("status page not found")
	}

	return nil
}

// AddMonitor adds a monitor to a status page
func (r *statusPageRepository) AddMonitor(statusPageID, monitorID string) error {
	query := `
		INSERT INTO status_page_monitors (status_page_id, monitor_id)
		VALUES ($1, $2)
		ON CONFLICT (status_page_id, monitor_id) DO NOTHING
	`

	_, err := r.db.Exec(query, statusPageID, monitorID)
	if err != nil {
		return fmt.Errorf("failed to add monitor to status page: %w", err)
	}

	return nil
}

// RemoveMonitor removes a monitor from a status page
func (r *statusPageRepository) RemoveMonitor(statusPageID, monitorID string) error {
	query := `DELETE FROM status_page_monitors WHERE status_page_id = $1 AND monitor_id = $2`

	_, err := r.db.Exec(query, statusPageID, monitorID)
	if err != nil {
		return fmt.Errorf("failed to remove monitor from status page: %w", err)
	}

	return nil
}

// GetMonitors retrieves all monitors associated with a status page
func (r *statusPageRepository) GetMonitors(statusPageID string) ([]*entities.Monitor, error) {
	var monitors []*entities.Monitor
	query := `
		SELECT m.id, m.user_id, m.name, m.url, m.type, m.check_interval, m.timeout, m.enabled,
		       m.check_ssl, m.ssl_alert_days, m.last_checked_at, m.created_at, m.updated_at
		FROM monitors m
		INNER JOIN status_page_monitors spm ON m.id = spm.monitor_id
		WHERE spm.status_page_id = $1
		ORDER BY m.created_at DESC
	`

	err := r.db.Select(&monitors, query, statusPageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitors for status page: %w", err)
	}

	return monitors, nil
}