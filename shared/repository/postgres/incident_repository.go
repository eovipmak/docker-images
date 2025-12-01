package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
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
		SELECT i.id, i.monitor_id, i.alert_rule_id, i.started_at, i.resolved_at, i.status, i.trigger_value, i.created_at, i.notified_at,
		       ar.name as alert_rule_name
		FROM incidents i
		LEFT JOIN alert_rules ar ON i.alert_rule_id = ar.id
		WHERE i.monitor_id = $1
		ORDER BY i.started_at DESC
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

// List retrieves incidents with filtering options
func (r *incidentRepository) List(filters repository.IncidentFilters) ([]*entities.Incident, error) {
	query := `
		SELECT i.id, i.monitor_id, i.alert_rule_id, i.started_at, i.resolved_at, i.status, i.trigger_value, i.created_at, i.notified_at,
		       m.name as monitor_name, m.url as monitor_url,
		       ar.name as alert_rule_name
		FROM incidents i
		INNER JOIN monitors m ON i.monitor_id = m.id
		LEFT JOIN alert_rules ar ON i.alert_rule_id = ar.id
		WHERE m.tenant_id = $1
	`
	
	args := []interface{}{filters.TenantID}
	argCount := 1

	// Add status filter
	if filters.Status != "" {
		argCount++
		query += fmt.Sprintf(" AND i.status = $%d", argCount)
		args = append(args, filters.Status)
	}

	// Add monitor_id filter
	if filters.MonitorID != "" {
		argCount++
		query += fmt.Sprintf(" AND i.monitor_id = $%d", argCount)
		args = append(args, filters.MonitorID)
	}

	// Add date range filters
	if filters.From != nil {
		argCount++
		query += fmt.Sprintf(" AND i.started_at >= $%d", argCount)
		args = append(args, filters.From)
	}

	if filters.To != nil {
		argCount++
		query += fmt.Sprintf(" AND i.started_at <= $%d", argCount)
		args = append(args, filters.To)
	}

	// Order by most recent first
	query += " ORDER BY i.started_at DESC"

	// Add pagination
	if filters.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
	}

	if filters.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filters.Offset)
	}

	var incidents []*entities.Incident
	err := r.db.Select(&incidents, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}

	return incidents, nil
}

// GetUnnotifiedIncidents retrieves incidents that haven't been notified yet
func (r *incidentRepository) GetUnnotifiedIncidents() ([]*entities.Incident, error) {
	var incidents []*entities.Incident
	query := `
		SELECT
			i.id,
			i.monitor_id,
			m.tenant_id,
			m.name as monitor_name,
			m.url as monitor_url,
			i.alert_rule_id,
			ar.name as alert_rule_name,
			ar.trigger_type,
			i.status,
			i.trigger_value,
			i.started_at,
			i.created_at
		FROM incidents i
		JOIN monitors m ON i.monitor_id = m.id
		JOIN alert_rules ar ON i.alert_rule_id = ar.id
		WHERE i.notified_at IS NULL
		ORDER BY i.created_at ASC
		LIMIT 100
	`

	err := r.db.Select(&incidents, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get unnotified incidents: %w", err)
	}

	return incidents, nil
}

// MarkAsNotified marks an incident as notified
func (r *incidentRepository) MarkAsNotified(id string) error {
	query := "UPDATE incidents SET notified_at = NOW() WHERE id = $1 AND notified_at IS NULL"

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to mark incident as notified: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("incident not found or already notified")
	}

	return nil
}
