package service

import (
	"database/sql"
	"testing"
	"time"

	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
)

// Mock incident repository for testing
type mockIncidentRepository struct {
	incidents map[string]*entities.Incident
}

func newMockIncidentRepository() *mockIncidentRepository {
	return &mockIncidentRepository{
		incidents: make(map[string]*entities.Incident),
	}
}

func (m *mockIncidentRepository) Create(incident *entities.Incident) error {
	incident.ID = "test-incident-id"
	incident.CreatedAt = time.Now()
	m.incidents[incident.ID] = incident
	return nil
}

func (m *mockIncidentRepository) GetByID(id string) (*entities.Incident, error) {
	if incident, ok := m.incidents[id]; ok {
		return incident, nil
	}
	return nil, sql.ErrNoRows
}

func (m *mockIncidentRepository) GetOpenIncident(monitorID, alertRuleID string) (*entities.Incident, error) {
	for _, incident := range m.incidents {
		if incident.MonitorID == monitorID && incident.AlertRuleID == alertRuleID && incident.Status == "open" {
			return incident, nil
		}
	}
	return nil, nil
}

func (m *mockIncidentRepository) GetByMonitorID(monitorID string) ([]*entities.Incident, error) {
	var result []*entities.Incident
	for _, incident := range m.incidents {
		if incident.MonitorID == monitorID {
			result = append(result, incident)
		}
	}
	return result, nil
}

func (m *mockIncidentRepository) Update(incident *entities.Incident) error {
	if _, ok := m.incidents[incident.ID]; ok {
		m.incidents[incident.ID] = incident
		return nil
	}
	return sql.ErrNoRows
}

func (m *mockIncidentRepository) Resolve(id string) error {
	if incident, ok := m.incidents[id]; ok {
		incident.Status = "resolved"
		incident.ResolvedAt = sql.NullTime{Time: time.Now(), Valid: true}
		return nil
	}
	return sql.ErrNoRows
}

// Mock alert rule repository for testing
type mockAlertRuleRepository struct{}

func (m *mockAlertRuleRepository) Create(rule *entities.AlertRule) error                       { return nil }
func (m *mockAlertRuleRepository) GetByID(tenantID int, id string) (*entities.AlertRule, error) { return nil, nil }
func (m *mockAlertRuleRepository) GetByTenantID(tenantID int) ([]*entities.AlertRule, error)    { return nil, nil }
func (m *mockAlertRuleRepository) Update(rule *entities.AlertRule) error                        { return nil }
func (m *mockAlertRuleRepository) Delete(tenantID int, id string) error                         { return nil }
func (m *mockAlertRuleRepository) AttachChannels(tenantID int, ruleID string, channelIDs []string) error { return nil }
func (m *mockAlertRuleRepository) DetachChannels(tenantID int, ruleID string, channelIDs []string) error { return nil }
func (m *mockAlertRuleRepository) GetChannelsByRuleID(tenantID int, ruleID string) ([]string, error) { return nil, nil }
func (m *mockAlertRuleRepository) GetWithChannels(tenantID int, id string) (*entities.AlertRuleWithChannels, error) { return nil, nil }
func (m *mockAlertRuleRepository) GetAllWithChannelsByTenantID(tenantID int) ([]*entities.AlertRuleWithChannels, error) { return nil, nil }
func (m *mockAlertRuleRepository) GetAllEnabled() ([]*entities.AlertRule, error) { return nil, nil }

func TestAlertService_EvaluateCheck_Down(t *testing.T) {
	incidentRepo := newMockIncidentRepository()
	alertRepo := &mockAlertRuleRepository{}
	service := NewAlertService(incidentRepo, alertRepo)

	check := &entities.MonitorCheck{
		MonitorID: "test-monitor",
		Success:   false,
		ErrorMessage: sql.NullString{
			String: "Connection refused",
			Valid:  true,
		},
	}

	rules := []*entities.AlertRule{
		{
			ID:          "rule-1",
			TriggerType: "down",
			Enabled:     true,
		},
	}

	triggered, err := service.EvaluateCheck(check, rules)
	if err != nil {
		t.Fatalf("EvaluateCheck failed: %v", err)
	}

	if len(triggered) != 1 {
		t.Fatalf("Expected 1 triggered alert, got %d", len(triggered))
	}

	if triggered[0].AlertRule.ID != "rule-1" {
		t.Errorf("Expected rule-1 to be triggered")
	}

	if triggered[0].TriggerValue == "" {
		t.Error("Expected trigger value to be set")
	}
}

func TestAlertService_EvaluateCheck_SlowResponse(t *testing.T) {
	incidentRepo := newMockIncidentRepository()
	alertRepo := &mockAlertRuleRepository{}
	service := NewAlertService(incidentRepo, alertRepo)

	check := &entities.MonitorCheck{
		MonitorID: "test-monitor",
		Success:   true,
		ResponseTimeMs: sql.NullInt64{
			Int64: 5000,
			Valid: true,
		},
	}

	rules := []*entities.AlertRule{
		{
			ID:             "rule-1",
			TriggerType:    "slow_response",
			ThresholdValue: 3000,
			Enabled:        true,
		},
	}

	triggered, err := service.EvaluateCheck(check, rules)
	if err != nil {
		t.Fatalf("EvaluateCheck failed: %v", err)
	}

	if len(triggered) != 1 {
		t.Fatalf("Expected 1 triggered alert, got %d", len(triggered))
	}

	if triggered[0].AlertRule.ID != "rule-1" {
		t.Errorf("Expected rule-1 to be triggered")
	}
}

func TestAlertService_EvaluateCheck_SSLExpiry(t *testing.T) {
	incidentRepo := newMockIncidentRepository()
	alertRepo := &mockAlertRuleRepository{}
	service := NewAlertService(incidentRepo, alertRepo)

	// SSL expires in 5 days
	expiryDate := time.Now().Add(5 * 24 * time.Hour)

	check := &entities.MonitorCheck{
		MonitorID: "test-monitor",
		Success:   true,
		SSLExpiresAt: sql.NullTime{
			Time:  expiryDate,
			Valid: true,
		},
	}

	rules := []*entities.AlertRule{
		{
			ID:             "rule-1",
			TriggerType:    "ssl_expiry",
			ThresholdValue: 7, // Alert if expires within 7 days
			Enabled:        true,
		},
	}

	triggered, err := service.EvaluateCheck(check, rules)
	if err != nil {
		t.Fatalf("EvaluateCheck failed: %v", err)
	}

	if len(triggered) != 1 {
		t.Fatalf("Expected 1 triggered alert, got %d", len(triggered))
	}

	if triggered[0].AlertRule.ID != "rule-1" {
		t.Errorf("Expected rule-1 to be triggered")
	}
}

func TestAlertService_EvaluateCheck_NoTrigger(t *testing.T) {
	incidentRepo := newMockIncidentRepository()
	alertRepo := &mockAlertRuleRepository{}
	service := NewAlertService(incidentRepo, alertRepo)

	check := &entities.MonitorCheck{
		MonitorID: "test-monitor",
		Success:   true,
		ResponseTimeMs: sql.NullInt64{
			Int64: 100,
			Valid: true,
		},
	}

	rules := []*entities.AlertRule{
		{
			ID:             "rule-1",
			TriggerType:    "slow_response",
			ThresholdValue: 3000,
			Enabled:        true,
		},
	}

	triggered, err := service.EvaluateCheck(check, rules)
	if err != nil {
		t.Fatalf("EvaluateCheck failed: %v", err)
	}

	if len(triggered) != 0 {
		t.Errorf("Expected no triggered alerts, got %d", len(triggered))
	}
}

func TestAlertService_CreateIncident(t *testing.T) {
	incidentRepo := newMockIncidentRepository()
	alertRepo := &mockAlertRuleRepository{}
	service := NewAlertService(incidentRepo, alertRepo)

	err := service.CreateIncident("monitor-1", "rule-1", "Test trigger")
	if err != nil {
		t.Fatalf("CreateIncident failed: %v", err)
	}

	// Verify incident was created
	incident, err := incidentRepo.GetOpenIncident("monitor-1", "rule-1")
	if err != nil {
		t.Fatalf("Failed to get created incident: %v", err)
	}

	if incident == nil {
		t.Fatal("Expected incident to be created")
	}

	if incident.Status != "open" {
		t.Errorf("Expected status 'open', got %s", incident.Status)
	}
}

func TestAlertService_CreateIncident_NoDuplicate(t *testing.T) {
	incidentRepo := newMockIncidentRepository()
	alertRepo := &mockAlertRuleRepository{}
	service := NewAlertService(incidentRepo, alertRepo)

	// Create first incident
	err := service.CreateIncident("monitor-1", "rule-1", "Test trigger")
	if err != nil {
		t.Fatalf("First CreateIncident failed: %v", err)
	}

	// Try to create duplicate - should not error
	err = service.CreateIncident("monitor-1", "rule-1", "Test trigger 2")
	if err != nil {
		t.Fatalf("Second CreateIncident failed: %v", err)
	}

	// Verify only one incident exists
	incidents, err := incidentRepo.GetByMonitorID("monitor-1")
	if err != nil {
		t.Fatalf("Failed to get incidents: %v", err)
	}

	if len(incidents) != 1 {
		t.Errorf("Expected 1 incident, got %d", len(incidents))
	}
}

func TestAlertService_ResolveMonitorIncidents(t *testing.T) {
	incidentRepo := newMockIncidentRepository()
	alertRepo := &mockAlertRuleRepository{}
	service := NewAlertService(incidentRepo, alertRepo)

	// Create an incident
	err := service.CreateIncident("monitor-1", "rule-1", "Test trigger")
	if err != nil {
		t.Fatalf("CreateIncident failed: %v", err)
	}

	// Resolve the incident
	err = service.ResolveMonitorIncidents("monitor-1", "rule-1")
	if err != nil {
		t.Fatalf("ResolveMonitorIncidents failed: %v", err)
	}

	// Verify no open incident exists
	openIncident, err := incidentRepo.GetOpenIncident("monitor-1", "rule-1")
	if err != nil {
		t.Fatalf("Failed to check for open incident: %v", err)
	}

	if openIncident != nil {
		t.Error("Expected no open incident after resolving")
	}
}
