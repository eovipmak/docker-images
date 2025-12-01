package postgres

import (
	"testing"
	"time"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func setupIncidentTestDB(t *testing.T) *sqlx.DB {
	// This is a placeholder for database setup
	// In a real test, you would either:
	// 1. Use a test database
	// 2. Use an in-memory database like SQLite
	// 3. Use a mocking library
	t.Skip("Database integration tests require a running PostgreSQL instance")
	return nil
}

func TestIncidentRepository_Create(t *testing.T) {
	db := setupIncidentTestDB(t)
	defer db.Close()

	repo := NewIncidentRepository(db)

	incident := &entities.Incident{
		MonitorID:    "test-monitor-id",
		AlertRuleID:  "test-rule-id",
		StartedAt:    time.Now(),
		Status:       "open",
		TriggerValue: "Test trigger",
	}

	err := repo.Create(incident)
	if err != nil {
		t.Fatalf("Failed to create incident: %v", err)
	}

	if incident.ID == "" {
		t.Error("Expected incident ID to be set after creation")
	}
}

func TestIncidentRepository_GetOpenIncident(t *testing.T) {
	db := setupIncidentTestDB(t)
	defer db.Close()

	repo := NewIncidentRepository(db)

	// Create an incident first
	incident := &entities.Incident{
		MonitorID:    "test-monitor-id",
		AlertRuleID:  "test-rule-id",
		StartedAt:    time.Now(),
		Status:       "open",
		TriggerValue: "Test trigger",
	}

	err := repo.Create(incident)
	if err != nil {
		t.Fatalf("Failed to create incident: %v", err)
	}

	// Retrieve the open incident
	retrieved, err := repo.GetOpenIncident("test-monitor-id", "test-rule-id")
	if err != nil {
		t.Fatalf("Failed to get open incident: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected to find open incident, got nil")
	}

	if retrieved.ID != incident.ID {
		t.Errorf("Expected incident ID %s, got %s", incident.ID, retrieved.ID)
	}

	if retrieved.Status != "open" {
		t.Errorf("Expected status 'open', got %s", retrieved.Status)
	}
}

func TestIncidentRepository_GetOpenIncident_NotFound(t *testing.T) {
	db := setupIncidentTestDB(t)
	defer db.Close()

	repo := NewIncidentRepository(db)

	// Try to retrieve a non-existent incident
	retrieved, err := repo.GetOpenIncident("non-existent-monitor", "non-existent-rule")
	if err != nil {
		t.Fatalf("Expected no error for non-existent incident, got: %v", err)
	}

	if retrieved != nil {
		t.Error("Expected nil for non-existent incident")
	}
}

func TestIncidentRepository_Resolve(t *testing.T) {
	db := setupIncidentTestDB(t)
	defer db.Close()

	repo := NewIncidentRepository(db)

	// Create an incident first
	incident := &entities.Incident{
		MonitorID:    "test-monitor-id",
		AlertRuleID:  "test-rule-id",
		StartedAt:    time.Now(),
		Status:       "open",
		TriggerValue: "Test trigger",
	}

	err := repo.Create(incident)
	if err != nil {
		t.Fatalf("Failed to create incident: %v", err)
	}

	// Resolve the incident
	err = repo.Resolve(incident.ID)
	if err != nil {
		t.Fatalf("Failed to resolve incident: %v", err)
	}

	// Verify it was resolved
	retrieved, err := repo.GetByID(incident.ID)
	if err != nil {
		t.Fatalf("Failed to get resolved incident: %v", err)
	}

	if retrieved.Status != "resolved" {
		t.Errorf("Expected status 'resolved', got %s", retrieved.Status)
	}

	if !retrieved.ResolvedAt.Valid {
		t.Error("Expected resolved_at to be set")
	}

	// Try to get it as an open incident - should return nil
	openIncident, err := repo.GetOpenIncident("test-monitor-id", "test-rule-id")
	if err != nil {
		t.Fatalf("Error checking for open incident: %v", err)
	}

	if openIncident != nil {
		t.Error("Expected no open incident after resolving")
	}
}

func TestIncidentRepository_GetByMonitorID(t *testing.T) {
	db := setupIncidentTestDB(t)
	defer db.Close()

	repo := NewIncidentRepository(db)

	monitorID := "test-monitor-id"

	// Create multiple incidents for the same monitor
	for i := 0; i < 3; i++ {
		incident := &entities.Incident{
			MonitorID:    monitorID,
			AlertRuleID:  "test-rule-id",
			StartedAt:    time.Now(),
			Status:       "open",
			TriggerValue: "Test trigger",
		}

		err := repo.Create(incident)
		if err != nil {
			t.Fatalf("Failed to create incident %d: %v", i, err)
		}
	}

	// Retrieve all incidents for the monitor
	incidents, err := repo.GetByMonitorID(monitorID)
	if err != nil {
		t.Fatalf("Failed to get incidents by monitor ID: %v", err)
	}

	if len(incidents) != 3 {
		t.Errorf("Expected 3 incidents, got %d", len(incidents))
	}
}
