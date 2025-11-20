package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
	"github.com/jmoiron/sqlx"
)

// incidentRepository implements the IncidentRepository interface using PostgreSQL
type incidentRepository struct {
	db *sqlx.DB
}

// NewIncidentRepository creates a new PostgreSQL incident repository
func NewIncidentRepository(db *sqlx.DB) repository.IncidentRepository {
	return &incidentRepository{db: db}
}

// Create creates a new incident in the database
func (r *incidentRepository) Create(incident *entities.Incident) error {
	query := `
		INSERT INTO incidents (monitor_id, alert_rule_id, started_at, status, trigger_value, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		query,
		incident.MonitorID,
		incident.AlertRuleID,
		incident.StartedAt,
		incident.Status,
		incident.TriggerValue,
	).Scan(&incident.ID, &incident.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	return nil
}

// GetByID retrieves an incident by its ID
func (r *incidentRepository) GetByID(id string) (*entities.Incident, error) {
	incident := &entities.Incident{}
	query := `
		SELECT id, monitor_id, alert_rule_id, started_at, resolved_at, status, trigger_value, created_at
		FROM incidents
		WHERE id = $1
	`

	err := r.db.Get(incident, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("incident not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	return incident, nil
}

// GetOpenIncident retrieves an open incident for a specific monitor and alert rule
func (r *incidentRepository) GetOpenIncident(monitorID, alertRuleID string) (*entities.Incident, error) {
	incident := &entities.Incident{}
	query := `
		SELECT id, monitor_id, alert_rule_id, started_at, resolved_at, status, trigger_value, created_at
		FROM incidents
		WHERE monitor_id = $1 AND alert_rule_id = $2 AND status = 'open'
		ORDER BY started_at DESC
		LIMIT 1
	`

	err := r.db.Get(incident, query, monitorID, alertRuleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No open incident found, not an error
		}
		return nil, fmt.Errorf("failed to get open incident: %w", err)
	}

	return incident, nil
}

// GetByMonitorID retrieves all incidents for a specific monitor
func (r *incidentRepository) GetByMonitorID(monitorID string) ([]*entities.Incident, error) {
	var incidents []*entities.Incident
	query := `
		SELECT id, monitor_id, alert_rule_id, started_at, resolved_at, status, trigger_value, created_at
		FROM incidents
		WHERE monitor_id = $1
		ORDER BY started_at DESC
	`

	err := r.db.Select(&incidents, query, monitorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get incidents by monitor: %w", err)
	}

	return incidents, nil
}

// Update updates an existing incident
func (r *incidentRepository) Update(incident *entities.Incident) error {
	query := `
		UPDATE incidents
		SET resolved_at = $1, status = $2, trigger_value = $3
		WHERE id = $4
	`

	result, err := r.db.Exec(
		query,
		incident.ResolvedAt,
		incident.Status,
		incident.TriggerValue,
		incident.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("incident not found")
	}

	return nil
}

// Resolve marks an incident as resolved
func (r *incidentRepository) Resolve(id string) error {
	query := `
		UPDATE incidents
		SET resolved_at = $1, status = 'resolved'
		WHERE id = $2 AND status = 'open'
	`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to resolve incident: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("incident not found or already resolved")
	}

	return nil
}
