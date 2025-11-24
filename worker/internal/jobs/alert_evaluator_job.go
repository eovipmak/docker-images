package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/eovipmak/v-insight/worker/internal/database"
)

// AlertRule represents an alert rule (worker-side struct)
type AlertRule struct {
	ID             string         `db:"id"`
	TenantID       int            `db:"tenant_id"`
	MonitorID      sql.NullString `db:"monitor_id"`
	Name           string         `db:"name"`
	TriggerType    string         `db:"trigger_type"`
	ThresholdValue int            `db:"threshold_value"`
	Enabled        bool           `db:"enabled"`
}

// Incident represents an incident (worker-side struct)
type Incident struct {
	ID           string       `db:"id"`
	MonitorID    string       `db:"monitor_id"`
	AlertRuleID  string       `db:"alert_rule_id"`
	StartedAt    time.Time    `db:"started_at"`
	ResolvedAt   sql.NullTime `db:"resolved_at"`
	Status       string       `db:"status"`
	TriggerValue string       `db:"trigger_value"`
	NotifiedAt   sql.NullTime `db:"notified_at"`
	CreatedAt    time.Time    `db:"created_at"`
}

// AlertEvaluatorJob evaluates alert rules against monitor checks
type AlertEvaluatorJob struct {
	db *database.DB
}

// NewAlertEvaluatorJob creates a new alert evaluator job
func NewAlertEvaluatorJob(db *database.DB) *AlertEvaluatorJob {
	return &AlertEvaluatorJob{
		db: db,
	}
}

// Name returns the name of the job
func (j *AlertEvaluatorJob) Name() string {
	return "AlertEvaluatorJob"
}

// Run executes the alert evaluation job
func (j *AlertEvaluatorJob) Run(ctx context.Context) error {
	startTime := time.Now()
	log.Println("[AlertEvaluatorJob] Starting alert evaluation run")

	// Get all enabled alert rules
	rules, err := j.getAllEnabledRules()
	if err != nil {
		log.Printf("[AlertEvaluatorJob] Failed to get alert rules: %v", err)
		return err
	}

	if len(rules) == 0 {
		log.Println("[AlertEvaluatorJob] No enabled alert rules found")
		return nil
	}

	log.Printf("[AlertEvaluatorJob] Found %d enabled alert rules", len(rules))

	// Get latest monitor checks (last 5 minutes)
	checks, err := j.getLatestMonitorChecks(5 * time.Minute)
	if err != nil {
		log.Printf("[AlertEvaluatorJob] Failed to get monitor checks: %v", err)
		return err
	}

	if len(checks) == 0 {
		log.Println("[AlertEvaluatorJob] No recent monitor checks found")
		return nil
	}

	log.Printf("[AlertEvaluatorJob] Found %d recent monitor checks", len(checks))

	// Evaluate each check against rules
	incidentsCreated := 0
	incidentsResolved := 0

	for _, check := range checks {
		created, resolved, err := j.evaluateCheckAgainstRules(check, rules)
		if err != nil {
			log.Printf("[AlertEvaluatorJob] Failed to evaluate check %s: %v", check.ID, err)
			continue
		}
		incidentsCreated += created
		incidentsResolved += resolved
	}

	duration := time.Since(startTime)
	log.Printf("[AlertEvaluatorJob] Evaluation completed in %v - Created: %d incidents, Resolved: %d incidents",
		duration, incidentsCreated, incidentsResolved)

	return nil
}

// getAllEnabledRules retrieves all enabled alert rules
func (j *AlertEvaluatorJob) getAllEnabledRules() ([]*AlertRule, error) {
	var rules []*AlertRule
	query := `
		SELECT id, tenant_id, monitor_id, name, trigger_type, threshold_value, enabled
		FROM alert_rules
		WHERE enabled = true
		ORDER BY created_at DESC
	`

	err := j.db.Select(&rules, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled alert rules: %w", err)
	}

	return rules, nil
}

// getLatestMonitorChecks retrieves monitor checks from the last N duration
func (j *AlertEvaluatorJob) getLatestMonitorChecks(duration time.Duration) ([]*MonitorCheck, error) {
	var checks []*MonitorCheck
	query := `
		SELECT DISTINCT ON (monitor_id) 
			id, monitor_id, checked_at, status_code, response_time_ms, 
			ssl_valid, ssl_expires_at, error_message, success
		FROM monitor_checks
		WHERE checked_at >= $1
		ORDER BY monitor_id, checked_at DESC
	`

	cutoffTime := time.Now().Add(-duration)
	err := j.db.Select(&checks, query, cutoffTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest monitor checks: %w", err)
	}

	return checks, nil
}

// evaluateCheckAgainstRules evaluates a single check against all rules
func (j *AlertEvaluatorJob) evaluateCheckAgainstRules(check *MonitorCheck, rules []*AlertRule) (int, int, error) {
	incidentsCreated := 0
	incidentsResolved := 0

	for _, rule := range rules {
		// Skip if rule is monitor-specific and doesn't match this check's monitor
		if rule.MonitorID.Valid && rule.MonitorID.String != check.MonitorID {
			continue
		}

		// Evaluate the rule
		triggered, triggerValue := j.evaluateRule(check, rule)

		if triggered {
			// Alert triggered - create incident if not already open
			created, err := j.createIncidentIfNeeded(check.MonitorID, rule.ID, triggerValue)
			if err != nil {
				log.Printf("[AlertEvaluatorJob] Failed to create incident for monitor %s, rule %s: %v",
					check.MonitorID, rule.ID, err)
				continue
			}
			if created {
				incidentsCreated++
				log.Printf("[AlertEvaluatorJob] ⚠ Incident created: Monitor %s triggered rule '%s' - %s",
					check.MonitorID, rule.Name, triggerValue)
				
				// Get monitor info for tenant ID
				j.broadcastIncidentCreatedEvent(check.MonitorID, rule.ID, rule.TenantID, rule.Name, triggerValue)
			}
		} else {
			// Alert not triggered - resolve incident if open
			resolved, err := j.resolveIncidentIfOpen(check.MonitorID, rule.ID)
			if err != nil {
				log.Printf("[AlertEvaluatorJob] Failed to resolve incident for monitor %s, rule %s: %v",
					check.MonitorID, rule.ID, err)
				continue
			}
			if resolved {
				incidentsResolved++
				log.Printf("[AlertEvaluatorJob] ✓ Incident resolved: Monitor %s recovered from rule '%s'",
					check.MonitorID, rule.Name)
				
				// Broadcast incident resolved event
				j.broadcastIncidentResolvedEvent(check.MonitorID, rule.ID, rule.TenantID, rule.Name)
			}
		}
	}

	return incidentsCreated, incidentsResolved, nil
}

// evaluateRule evaluates a single rule against a check
func (j *AlertEvaluatorJob) evaluateRule(check *MonitorCheck, rule *AlertRule) (bool, string) {
	switch rule.TriggerType {
	case "down":
		if !check.Success {
			if check.ErrorMessage.Valid {
				return true, fmt.Sprintf("Monitor is down: %s", check.ErrorMessage.String)
			}
			return true, "Monitor is down"
		}

	case "slow_response":
		if check.Success && check.ResponseTimeMs.Valid {
			if check.ResponseTimeMs.Int64 > int64(rule.ThresholdValue) {
				return true, fmt.Sprintf("Response time: %dms (threshold: %dms)",
					check.ResponseTimeMs.Int64, rule.ThresholdValue)
			}
		}

	case "ssl_expiry":
		if check.SSLExpiresAt.Valid {
			daysUntilExpiry := int(time.Until(check.SSLExpiresAt.Time).Hours() / 24)
			if daysUntilExpiry <= rule.ThresholdValue && daysUntilExpiry >= 0 {
				return true, fmt.Sprintf("SSL certificate expires in %d days (on %s)",
					daysUntilExpiry, check.SSLExpiresAt.Time.Format("2006-01-02"))
			}
		}
	}

	return false, ""
}

// createIncidentIfNeeded creates an incident if one doesn't already exist
func (j *AlertEvaluatorJob) createIncidentIfNeeded(monitorID, ruleID, triggerValue string) (bool, error) {
	// Check if there's already an open incident
	var count int
	checkQuery := `
		SELECT COUNT(*)
		FROM incidents
		WHERE monitor_id = $1 AND alert_rule_id = $2 AND status = 'open'
	`

	err := j.db.QueryRow(checkQuery, monitorID, ruleID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check for existing incident: %w", err)
	}

	// Don't create duplicate incidents
	if count > 0 {
		return false, nil
	}

	// Create new incident
	insertQuery := `
		INSERT INTO incidents (monitor_id, alert_rule_id, started_at, status, trigger_value, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`

	_, err = j.db.Exec(insertQuery, monitorID, ruleID, time.Now(), "open", triggerValue)
	if err != nil {
		return false, fmt.Errorf("failed to create incident: %w", err)
	}

	return true, nil
}

// resolveIncidentIfOpen resolves an open incident if it exists
func (j *AlertEvaluatorJob) resolveIncidentIfOpen(monitorID, ruleID string) (bool, error) {
	query := `
		UPDATE incidents
		SET resolved_at = $1, status = 'resolved'
		WHERE monitor_id = $2 AND alert_rule_id = $3 AND status = 'open'
	`

	result, err := j.db.Exec(query, time.Now(), monitorID, ruleID)
	if err != nil {
		return false, fmt.Errorf("failed to resolve incident: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

// broadcastIncidentCreatedEvent broadcasts an incident created event
func (j *AlertEvaluatorJob) broadcastIncidentCreatedEvent(monitorID, ruleID string, tenantID int, ruleName, triggerValue string) {
	data := map[string]interface{}{
		"monitor_id":    monitorID,
		"alert_rule_id": ruleID,
		"rule_name":     ruleName,
		"trigger_value": triggerValue,
		"status":        "open",
	}

	broadcastEvent("incident_created", data, tenantID)
}

// broadcastIncidentResolvedEvent broadcasts an incident resolved event
func (j *AlertEvaluatorJob) broadcastIncidentResolvedEvent(monitorID, ruleID string, tenantID int, ruleName string) {
	data := map[string]interface{}{
		"monitor_id":    monitorID,
		"alert_rule_id": ruleID,
		"rule_name":     ruleName,
		"status":        "resolved",
	}

	broadcastEvent("incident_resolved", data, tenantID)
}
