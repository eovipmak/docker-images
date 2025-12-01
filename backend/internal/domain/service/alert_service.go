package service

import (
	"fmt"
	"time"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
)

// AlertService handles alert evaluation and incident management
type AlertService struct {
	incidentRepo repository.IncidentRepository
	alertRepo    repository.AlertRuleRepository
}

// NewAlertService creates a new alert service
func NewAlertService(incidentRepo repository.IncidentRepository, alertRepo repository.AlertRuleRepository) *AlertService {
	return &AlertService{
		incidentRepo: incidentRepo,
		alertRepo:    alertRepo,
	}
}

// TriggeredAlert represents an alert that has been triggered by a check
type TriggeredAlert struct {
	AlertRule    *entities.AlertRule
	TriggerValue string
}

// EvaluateCheck evaluates a monitor check against alert rules and returns triggered alerts
func (s *AlertService) EvaluateCheck(check *entities.MonitorCheck, rules []*entities.AlertRule) ([]*TriggeredAlert, error) {
	var triggered []*TriggeredAlert

	for _, rule := range rules {
		// Skip disabled rules
		if !rule.Enabled {
			continue
		}

		// Skip if rule is monitor-specific and doesn't match this check's monitor
		if rule.MonitorID.Valid && rule.MonitorID.String != check.MonitorID {
			continue
		}

		// Evaluate based on trigger type
		switch rule.TriggerType {
		case "down":
			if s.evaluateDownTrigger(check) {
				triggered = append(triggered, &TriggeredAlert{
					AlertRule:    rule,
					TriggerValue: s.getDownTriggerValue(check),
				})
			}

		case "slow_response":
			if s.evaluateSlowResponseTrigger(check, rule.ThresholdValue) {
				triggered = append(triggered, &TriggeredAlert{
					AlertRule:    rule,
					TriggerValue: s.getSlowResponseTriggerValue(check),
				})
			}

		case "ssl_expiry":
			if s.evaluateSSLExpiryTrigger(check, rule.ThresholdValue) {
				triggered = append(triggered, &TriggeredAlert{
					AlertRule:    rule,
					TriggerValue: s.getSSLExpiryTriggerValue(check),
				})
			}
		}
	}

	return triggered, nil
}

// evaluateDownTrigger checks if the monitor is down
func (s *AlertService) evaluateDownTrigger(check *entities.MonitorCheck) bool {
	return !check.Success
}

// evaluateSlowResponseTrigger checks if response time exceeds threshold
func (s *AlertService) evaluateSlowResponseTrigger(check *entities.MonitorCheck, thresholdMs int) bool {
	if !check.Success || !check.ResponseTimeMs.Valid {
		return false
	}
	return check.ResponseTimeMs.Int64 > int64(thresholdMs)
}

// evaluateSSLExpiryTrigger checks if SSL certificate expires within threshold days
func (s *AlertService) evaluateSSLExpiryTrigger(check *entities.MonitorCheck, thresholdDays int) bool {
	if !check.SSLExpiresAt.Valid {
		return false
	}

	daysUntilExpiry := int(time.Until(check.SSLExpiresAt.Time).Hours() / 24)
	return daysUntilExpiry <= thresholdDays && daysUntilExpiry >= 0
}

// getDownTriggerValue returns a description of the down trigger
func (s *AlertService) getDownTriggerValue(check *entities.MonitorCheck) string {
	if check.ErrorMessage.Valid {
		return fmt.Sprintf("Monitor is down: %s", check.ErrorMessage.String)
	}
	return "Monitor is down"
}

// getSlowResponseTriggerValue returns a description of the slow response trigger
func (s *AlertService) getSlowResponseTriggerValue(check *entities.MonitorCheck) string {
	if check.ResponseTimeMs.Valid {
		return fmt.Sprintf("Response time: %dms", check.ResponseTimeMs.Int64)
	}
	return "Slow response detected"
}

// getSSLExpiryTriggerValue returns a description of the SSL expiry trigger
func (s *AlertService) getSSLExpiryTriggerValue(check *entities.MonitorCheck) string {
	if check.SSLExpiresAt.Valid {
		daysUntilExpiry := int(time.Until(check.SSLExpiresAt.Time).Hours() / 24)
		return fmt.Sprintf("SSL certificate expires in %d days (on %s)", daysUntilExpiry, check.SSLExpiresAt.Time.Format("2006-01-02"))
	}
	return "SSL certificate expiring soon"
}

// CreateIncident creates a new incident for a triggered alert
func (s *AlertService) CreateIncident(monitorID, alertRuleID, triggerValue string) error {
	// Check if there's already an open incident for this monitor+rule combination
	existingIncident, err := s.incidentRepo.GetOpenIncident(monitorID, alertRuleID)
	if err != nil {
		return fmt.Errorf("failed to check for existing incident: %w", err)
	}

	// Don't create duplicate incidents
	if existingIncident != nil {
		return nil
	}

	// Create new incident
	incident := &entities.Incident{
		MonitorID:    monitorID,
		AlertRuleID:  alertRuleID,
		StartedAt:    time.Now(),
		Status:       "open",
		TriggerValue: triggerValue,
	}

	if err := s.incidentRepo.Create(incident); err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	return nil
}

// ResolveIncident resolves an incident by its ID
func (s *AlertService) ResolveIncident(incidentID string) error {
	if err := s.incidentRepo.Resolve(incidentID); err != nil {
		return fmt.Errorf("failed to resolve incident: %w", err)
	}
	return nil
}

// ResolveMonitorIncidents resolves all open incidents for a monitor when it recovers
func (s *AlertService) ResolveMonitorIncidents(monitorID string, alertRuleID string) error {
	// Get the open incident for this monitor+rule
	incident, err := s.incidentRepo.GetOpenIncident(monitorID, alertRuleID)
	if err != nil {
		return fmt.Errorf("failed to get open incident: %w", err)
	}

	// No open incident to resolve
	if incident == nil {
		return nil
	}

	// Resolve the incident
	if err := s.incidentRepo.Resolve(incident.ID); err != nil {
		return fmt.Errorf("failed to resolve incident: %w", err)
	}

	return nil
}

// GetAllEnabledRules retrieves all enabled alert rules across all tenants
func (s *AlertService) GetAllEnabledRules() ([]*entities.AlertRule, error) {
	rules, err := s.alertRepo.GetAllEnabled()
	if err != nil {
		return nil, fmt.Errorf("failed to get all enabled rules: %w", err)
	}
	return rules, nil
}
