package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/eovipmak/v-insight/worker/internal"
	"go.uber.org/zap"
)

// AlertEvaluatorJob evaluates alert rules against monitor checks
type AlertEvaluatorJob struct {
	alertRuleRepo repository.AlertRuleRepository
	incidentRepo  repository.IncidentRepository
	monitorRepo   repository.MonitorRepository
}

// NewAlertEvaluatorJob creates a new alert evaluator job
func NewAlertEvaluatorJob(
	alertRuleRepo repository.AlertRuleRepository,
	incidentRepo repository.IncidentRepository,
	monitorRepo repository.MonitorRepository,
) *AlertEvaluatorJob {
	return &AlertEvaluatorJob{
		alertRuleRepo: alertRuleRepo,
		incidentRepo:  incidentRepo,
		monitorRepo:   monitorRepo,
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
	rules, err := j.alertRuleRepo.GetAllEnabled()
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
	checks, err := j.monitorRepo.GetLatestMonitorChecks(5 * time.Minute)
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

// evaluateCheckAgainstRules evaluates a single check against all rules
func (j *AlertEvaluatorJob) evaluateCheckAgainstRules(check *entities.MonitorCheck, rules []*entities.AlertRule) (int, int, error) {
	incidentsCreated := 0
	incidentsResolved := 0

	for _, rule := range rules {
		// Skip rules that don't belong to the same user as the monitor
		if rule.UserID != check.UserID {
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
			created, err := j.createIncidentIfNeeded(check.MonitorID, rule.ID, rule.UserID, triggerValue)
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
				
				// Get monitor info for user ID (already have it)
				j.broadcastIncidentCreatedEvent(check.MonitorID, rule.ID, rule.UserID, rule.Name, triggerValue)
			}
		} else {
			// Alert not triggered - resolve incident if open
			resolved, err := j.resolveIncidentIfOpen(check.MonitorID, rule.ID, rule.UserID, rule.Name)
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
				j.broadcastIncidentResolvedEvent(check.MonitorID, rule.ID, rule.UserID, rule.Name)
			}
		}
	}

	return incidentsCreated, incidentsResolved, nil
}

// evaluateRule evaluates a single rule against a check
func (j *AlertEvaluatorJob) evaluateRule(check *entities.MonitorCheck, rule *entities.AlertRule) (bool, string) {
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
func (j *AlertEvaluatorJob) createIncidentIfNeeded(monitorID, ruleID string, userID int, triggerValue string) (bool, error) {
	// Check if there's already an open incident
	incident, err := j.incidentRepo.GetOpenIncident(monitorID, ruleID)
	if err != nil {
		return false, fmt.Errorf("failed to check for existing incident: %w", err)
	}

	// Don't create duplicate incidents
	if incident != nil {
		return false, nil
	}

	// Create new incident
	newIncident := &entities.Incident{
		MonitorID:    monitorID,
		AlertRuleID:  ruleID,
		UserID:       userID,
		StartedAt:    time.Now(),
		Status:       "open",
		TriggerValue: triggerValue,
	}

	err = j.incidentRepo.Create(newIncident)
	if err != nil {
		return false, fmt.Errorf("failed to create incident: %w", err)
	}

	return true, nil
}

// resolveIncidentIfOpen resolves an open incident if it exists
func (j *AlertEvaluatorJob) resolveIncidentIfOpen(monitorID, ruleID string, userID int, ruleName string) (bool, error) {
	incident, err := j.incidentRepo.GetOpenIncident(monitorID, ruleID)
	if err != nil {
		return false, fmt.Errorf("failed to check for open incident: %w", err)
	}

	if incident == nil {
		return false, nil
	}

	err = j.incidentRepo.Resolve(incident.ID)
	if err != nil {
		return false, fmt.Errorf("failed to resolve incident: %w", err)
	}

	return true, nil
}

// broadcastIncidentCreatedEvent broadcasts an incident created event
func (j *AlertEvaluatorJob) broadcastIncidentCreatedEvent(monitorID, ruleID string, userID int, ruleName, triggerValue string) {
	data := map[string]interface{}{
		"monitor_id":    monitorID,
		"alert_rule_id": ruleID,
		"rule_name":     ruleName,
		"trigger_value": triggerValue,
		"status":        "open",
	}

	broadcastEvent("incident_created", data, userID)
}

// broadcastIncidentResolvedEvent broadcasts an incident resolved event
func (j *AlertEvaluatorJob) broadcastIncidentResolvedEvent(monitorID, ruleID string, userID int, ruleName string) {
	data := map[string]interface{}{
		"monitor_id":    monitorID,
		"alert_rule_id": ruleID,
		"rule_name":     ruleName,
		"status":        "resolved",
	}

	broadcastEvent("incident_resolved", data, userID)
}
