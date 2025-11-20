package repository

import "github.com/eovipmak/v-insight/backend/internal/domain/entities"

// IncidentRepository defines the interface for incident data operations
type IncidentRepository interface {
	// Create creates a new incident
	Create(incident *entities.Incident) error

	// GetByID retrieves an incident by its ID
	GetByID(id string) (*entities.Incident, error)

	// GetOpenIncident retrieves an open incident for a specific monitor and alert rule
	GetOpenIncident(monitorID, alertRuleID string) (*entities.Incident, error)

	// GetByMonitorID retrieves all incidents for a specific monitor
	GetByMonitorID(monitorID string) ([]*entities.Incident, error)

	// Update updates an existing incident
	Update(incident *entities.Incident) error

	// Resolve marks an incident as resolved
	Resolve(id string) error
}
