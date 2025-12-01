package repository

import (
	"time"

	"github.com/eovipmak/v-insight/shared/domain/entities"
)

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

	// List retrieves incidents with filtering options
	List(filters IncidentFilters) ([]*entities.Incident, error)

	// Update updates an existing incident
	Update(incident *entities.Incident) error

	// Resolve marks an incident as resolved
	Resolve(id string) error

	// GetUnnotifiedIncidents retrieves incidents that haven't been notified yet
	GetUnnotifiedIncidents() ([]*entities.Incident, error)

	// MarkAsNotified marks an incident as notified
	MarkAsNotified(id string) error
}

// IncidentFilters defines filtering options for incident queries
type IncidentFilters struct {
	TenantID  int
	Status    string // 'open', 'resolved', or empty for all
	MonitorID string
	From      *time.Time
	To        *time.Time
	Limit     int
	Offset    int
}
