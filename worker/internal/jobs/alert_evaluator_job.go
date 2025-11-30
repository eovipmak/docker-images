package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/eovipmak/v-insight/worker/internal"
	"github.com/eovipmak/v-insight/worker/internal/database"
	"go.uber.org/zap"
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
	
	// Record job execution metrics
	defer func() {
		duration := time.Since(startTime)
		internal.JobExecutionDuration.WithLabelValues("AlertEvaluatorJob").Observe(duration.Seconds())
	}()

	if internal.Log != nil {
		internal.Log.Info("Starting alert evaluation run")
	}

	// Get all enabled alert rules
	rules, err := j.getAllEnabledRules()
	if err != nil {
		if internal.Log != nil {
			internal.Log.Error("Failed to get alert rules", zap.Error(err))
		}
		internal.JobExecutionTotal.WithLabelValues("AlertEvaluatorJob", "failure").Inc()
		internal.AlertEvaluationTotal.WithLabelValues("failure").Inc()
		return err
	}

	if len(rules) == 0 {
		if internal.Log != nil {
			internal.Log.Debug("No enabled alert rules found")
		}
		internal.JobExecutionTotal.WithLabelValues("AlertEvaluatorJob", "success").Inc()
		return nil
	}

	if internal.Log != nil {
		internal.Log.Info("Found enabled alert rules", zap.Int("count", len(rules)))
	}

	// Get latest monitor checks (last 5 minutes)
	checks, err := j.getLatestMonitorChecks(5 * time.Minute)
	if err != nil {
		if internal.Log != nil {
			internal.Log.Error("Failed to get monitor checks", zap.Error(err))
		}
		internal.JobExecutionTotal.WithLabelValues("AlertEvaluatorJob", "failure").Inc()
		internal.AlertEvaluationTotal.WithLabelValues("failure").Inc()
		return err
	}

	if len(checks) == 0 {
		if internal.Log != nil {
			internal.Log.Debug("No recent monitor checks found")
		}
		internal.JobExecutionTotal.WithLabelValues("AlertEvaluatorJob", "success").Inc()
		return nil
	}

	if internal.Log != nil {
		internal.Log.Info("Found recent monitor checks", zap.Int("count", len(checks)))
	}

	// Evaluate each check against rules
	incidentsCreated := 0
	incidentsResolved := 0

	for _, check := range checks {
		created, resolved, err := j.evaluateCheckAgainstRules(check, rules)
		if err != nil {
			if internal.Log != nil {
				internal.Log.Error("Failed to evaluate check",
					zap.String("check_id", check.ID),
					zap.Error(err),
				)
			}
			continue
		}
		incidentsCreated += created
		incidentsResolved += resolved
	}

	duration := time.Since(startTime)
	if internal.Log != nil {
		internal.Log.Info("Alert evaluation completed",
			zap.Duration("duration", duration),
			zap.Int("incidents_created", incidentsCreated),
			zap.Int("incidents_resolved", incidentsResolved),
		)
	}

	// Update metrics
	internal.AlertEvaluationTotal.WithLabelValues("success").Inc()
	internal.JobExecutionTotal.WithLabelValues("AlertEvaluatorJob", "success").Inc()

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
		SELECT DISTINCT ON (mc.monitor_id) 
			mc.id, mc.monitor_id, m.tenant_id, m.type as monitor_type, mc.checked_at, mc.status_code, mc.response_time_ms, 
			mc.ssl_valid, mc.ssl_expires_at, mc.error_message, mc.success
		FROM monitor_checks mc
		JOIN monitors m ON mc.monitor_id = m.id
		WHERE mc.checked_at >= $1
		ORDER BY mc.monitor_id, mc.checked_at DESC
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
		// Skip rules that don't belong to the same tenant as the monitor
		if rule.TenantID != check.TenantID {
			continue
		}

		// Skip if rule is monitor-specific and doesn't match this check's monitor
		if rule.MonitorID.Valid && rule.MonitorID.String != check.MonitorID {
			continue
		}

		// Skip SSL expiry rules for TCP monitors when rule applies to all monitors
		if rule.TriggerType == "ssl_expiry" && !rule.MonitorID.Valid && check.MonitorType == "tcp" {
			continue
		}

		// Evaluate the rule
		triggered, triggerValue := j.evaluateRule(check, rule)

		if triggered {
			// Alert triggered - create incident if not already open
			created, err := j.createIncidentIfNeeded(check.MonitorID, rule.ID, triggerValue)
			if err != nil {
				if internal.Log != nil {
					internal.Log.Error("Failed to create incident",
						zap.String("monitor_id", check.MonitorID),
						zap.String("rule_id", rule.ID),
						zap.Error(err),
					)
				}
				continue
			}
			if created {
				incidentsCreated++
				internal.IncidentCreated.Inc()
				if internal.Log != nil {
					internal.Log.Warn("Incident created",
						zap.String("monitor_id", check.MonitorID),
						zap.String("rule_name", rule.Name),
						zap.String("trigger_value", triggerValue),
					)
				}
				
				// Get monitor info for tenant ID (already have it)
				j.broadcastIncidentCreatedEvent(check.MonitorID, rule.ID, rule.TenantID, rule.Name, triggerValue)
			}
		} else {
			// Alert not triggered - resolve incident if open
			resolved, err := j.resolveIncidentIfOpen(check.MonitorID, rule.ID)
			if err != nil {
				if internal.Log != nil {
					internal.Log.Error("Failed to resolve incident",
						zap.String("monitor_id", check.MonitorID),
						zap.String("rule_id", rule.ID),
						zap.Error(err),
					)
				}
				continue
			}
			if resolved {
				incidentsResolved++
				internal.IncidentResolved.Inc()
				if internal.Log != nil {
					internal.Log.Info("Incident resolved",
						zap.String("monitor_id", check.MonitorID),
						zap.String("rule_name", rule.Name),
					)
				}
				
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
