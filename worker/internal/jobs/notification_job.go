package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/eovipmak/v-insight/worker/internal"
	"github.com/eovipmak/v-insight/worker/internal/config"
	"github.com/eovipmak/v-insight/worker/internal/database"
	"go.uber.org/zap"
)

// IncidentNotificationData represents the full incident data for notifications
type IncidentNotificationData struct {
	IncidentID   string
	MonitorID    string
	MonitorName  string
	MonitorURL   string
	AlertRuleID  string
	AlertName    string
	TriggerType  string
	Status       string // "open" or "resolved"
	Message      string
	Timestamp    time.Time
}

// AlertChannelConfig represents an alert channel from database
type AlertChannelConfig struct {
	ID      string
	Type    string // "webhook", "discord", "email"
	Name    string
	Config  map[string]interface{}
	Enabled bool
}

// WebhookPayload represents the generic webhook payload structure
type WebhookPayload struct {
	IncidentID  string `json:"incident_id"`
	MonitorName string `json:"monitor_name"`
	MonitorURL  string `json:"monitor_url"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	Timestamp   string `json:"timestamp"`
}

// DiscordWebhookPayload represents the Discord-specific webhook format
type DiscordWebhookPayload struct {
	Embeds []DiscordEmbed `json:"embeds"`
}

// DiscordEmbed represents a Discord embed object
type DiscordEmbed struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Color       int                    `json:"color"`
	Fields      []DiscordEmbedField    `json:"fields"`
	Timestamp   string                 `json:"timestamp"`
	Footer      map[string]interface{} `json:"footer,omitempty"`
}

// DiscordEmbedField represents a field in a Discord embed
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

// NotificationJob processes and sends notifications for incidents
type NotificationJob struct {
	db         *database.DB
	httpClient *http.Client
	smtpConfig config.SMTPConfig
}

// NewNotificationJob creates a new notification job
func NewNotificationJob(db *database.DB, smtpConfig config.SMTPConfig) *NotificationJob {
	return &NotificationJob{
		db: db,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		smtpConfig: smtpConfig,
	}
}

// Name returns the name of the job
func (j *NotificationJob) Name() string {
	return "NotificationJob"
}

// Run executes the notification job
func (j *NotificationJob) Run(ctx context.Context) error {
	startTime := time.Now()
	
	// Record job execution metrics
	defer func() {
		duration := time.Since(startTime)
		internal.JobExecutionDuration.WithLabelValues("NotificationJob").Observe(duration.Seconds())
	}()

	if internal.Log != nil {
		internal.Log.Info("Starting notification processing run")
	}

	// Get incidents that need notification
	incidents, err := j.getUnnotifiedIncidents()
	if err != nil {
		if internal.Log != nil {
			internal.Log.Error("Failed to get unnotified incidents", zap.Error(err))
		}
		internal.JobExecutionTotal.WithLabelValues("NotificationJob", "failure").Inc()
		return err
	}

	if len(incidents) == 0 {
		if internal.Log != nil {
			internal.Log.Debug("No unnotified incidents found")
		}
		internal.JobExecutionTotal.WithLabelValues("NotificationJob", "success").Inc()
		return nil
	}

	if internal.Log != nil {
		internal.Log.Info("Found unnotified incidents", zap.Int("count", len(incidents)))
	}

	notificationsSent := 0
	notificationsFailed := 0

	// Process each incident
	for _, incident := range incidents {
		sent, err := j.processIncidentNotifications(incident)
		if err != nil {
			if internal.Log != nil {
				internal.Log.Error("Failed to process incident",
					zap.String("incident_id", incident.IncidentID),
					zap.Error(err),
				)
			}
			notificationsFailed++
			continue
		}
		if sent > 0 {
			// Mark incident as notified
			if err := j.markIncidentAsNotified(incident.IncidentID); err != nil {
				if internal.Log != nil {
					internal.Log.Error("Failed to mark incident as notified",
						zap.String("incident_id", incident.IncidentID),
						zap.Error(err),
					)
				}
				continue
			}
			notificationsSent += sent
		}
	}

	duration := time.Since(startTime)
	if internal.Log != nil {
		internal.Log.Info("Notification processing completed",
			zap.Duration("duration", duration),
			zap.Int("notifications_sent", notificationsSent),
			zap.Int("notifications_failed", notificationsFailed),
		)
	}

	internal.JobExecutionTotal.WithLabelValues("NotificationJob", "success").Inc()
	return nil
}

// getUnnotifiedIncidents retrieves incidents that haven't been notified yet
func (j *NotificationJob) getUnnotifiedIncidents() ([]*IncidentNotificationData, error) {
	query := `
		SELECT 
			i.id as incident_id,
			i.monitor_id,
			m.name as monitor_name,
			m.url as monitor_url,
			i.alert_rule_id,
			ar.name as alert_name,
			ar.trigger_type,
			i.status,
			i.trigger_value as message,
			i.started_at as timestamp
		FROM incidents i
		JOIN monitors m ON i.monitor_id = m.id
		JOIN alert_rules ar ON i.alert_rule_id = ar.id
		WHERE i.notified_at IS NULL
		ORDER BY i.created_at ASC
		LIMIT 100
	`

	rows, err := j.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unnotified incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*IncidentNotificationData
	for rows.Next() {
		var incident IncidentNotificationData
		err := rows.Scan(
			&incident.IncidentID,
			&incident.MonitorID,
			&incident.MonitorName,
			&incident.MonitorURL,
			&incident.AlertRuleID,
			&incident.AlertName,
			&incident.TriggerType,
			&incident.Status,
			&incident.Message,
			&incident.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}
		incidents = append(incidents, &incident)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating incidents: %w", err)
	}

	return incidents, nil
}

// processIncidentNotifications processes notifications for a single incident
func (j *NotificationJob) processIncidentNotifications(incident *IncidentNotificationData) (int, error) {
	// Get alert channels associated with this incident's alert rule
	channels, err := j.getAlertChannels(incident.AlertRuleID)
	if err != nil {
		return 0, fmt.Errorf("failed to get alert channels: %w", err)
	}

	if len(channels) == 0 {
		if internal.Log != nil {
			internal.Log.Debug("No alert channels configured for incident",
				zap.String("incident_id", incident.IncidentID),
			)
		}
		return 0, nil
	}

	sentCount := 0
	for _, channel := range channels {
		if !channel.Enabled {
			continue
		}

		var err error
		switch channel.Type {
		case "webhook":
			err = j.sendWebhookNotification(incident, channel)
		case "discord":
			err = j.sendDiscordNotification(incident, channel)
		case "email":
			err = j.sendEmailNotification(incident, channel)
		default:
			if internal.Log != nil {
				internal.Log.Warn("Unsupported channel type",
					zap.String("channel_type", channel.Type),
				)
			}
			continue
		}

		if err != nil {
			internal.NotificationSent.WithLabelValues(channel.Type, "failure").Inc()
			if internal.Log != nil {
				internal.Log.Error("Failed to send notification",
					zap.String("channel_type", channel.Type),
					zap.String("channel_name", channel.Name),
					zap.Error(err),
				)
			}
			continue
		}

		internal.NotificationSent.WithLabelValues(channel.Type, "success").Inc()
		if internal.Log != nil {
			internal.Log.Info("Notification sent",
				zap.String("channel_type", channel.Type),
				zap.String("incident_id", incident.IncidentID),
				zap.String("channel_name", channel.Name),
			)
		}
		sentCount++
	}

	return sentCount, nil
}

// getAlertChannels retrieves all alert channels for a given alert rule
func (j *NotificationJob) getAlertChannels(alertRuleID string) ([]*AlertChannelConfig, error) {
	query := `
		SELECT 
			ac.id,
			ac.type,
			ac.name,
			ac.config,
			ac.enabled
		FROM alert_channels ac
		JOIN alert_rule_channels arc ON ac.id = arc.alert_channel_id
		WHERE arc.alert_rule_id = $1 AND ac.enabled = true
		ORDER BY ac.created_at ASC
	`

	rows, err := j.db.Query(query, alertRuleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert channels: %w", err)
	}
	defer rows.Close()

	var channels []*AlertChannelConfig
	for rows.Next() {
		var channel AlertChannelConfig
		var configJSON []byte

		err := rows.Scan(
			&channel.ID,
			&channel.Type,
			&channel.Name,
			&configJSON,
			&channel.Enabled,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert channel: %w", err)
		}

		// Parse JSONB config
		if err := json.Unmarshal(configJSON, &channel.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal channel config: %w", err)
		}

		channels = append(channels, &channel)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert channels: %w", err)
	}

	return channels, nil
}

// sendWebhookNotification sends a generic webhook notification
func (j *NotificationJob) sendWebhookNotification(incident *IncidentNotificationData, channel *AlertChannelConfig) error {
	// Get webhook URL from config
	webhookURL, ok := channel.Config["url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhook URL not configured")
	}

	// Create payload
	payload := WebhookPayload{
		IncidentID:  incident.IncidentID,
		MonitorName: incident.MonitorName,
		MonitorURL:  incident.MonitorURL,
		Status:      incident.Status,
		Message:     incident.Message,
		Timestamp:   incident.Timestamp.Format(time.RFC3339),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Send HTTP POST request
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}

	return nil
}

// sendDiscordNotification sends a Discord-specific webhook notification
func (j *NotificationJob) sendDiscordNotification(incident *IncidentNotificationData, channel *AlertChannelConfig) error {
	// Get webhook URL from config
	webhookURL, ok := channel.Config["url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("Discord webhook URL not configured")
	}

	// Determine embed color and title based on status and trigger type
	var color int
	var title string
	if incident.Status == "open" {
		color = 0xFF0000 // Red for open incidents
		switch incident.TriggerType {
		case "down":
			title = "🔴 Monitor Down Alert"
		case "slow_response":
			title = "🐌 Slow Response Alert"
		case "ssl_expiry":
			title = "🔒 SSL Certificate Expiry Alert"
		default:
			title = "🚨 New Incident Detected"
		}
	} else {
		color = 0x00FF00 // Green for resolved incidents
		switch incident.TriggerType {
		case "down":
			title = "✅ Monitor Restored"
		case "slow_response":
			title = "✅ Response Time Normalized"
		case "ssl_expiry":
			title = "✅ SSL Certificate Renewed"
		default:
			title = "✅ Incident Resolved"
		}
	}

	// Create Discord embed
	embed := DiscordEmbed{
		Title:       title,
		Description: fmt.Sprintf("**%s**", incident.MonitorName),
		Color:       color,
		Timestamp:   incident.Timestamp.Format(time.RFC3339),
		Fields: []DiscordEmbedField{
			{
				Name:   "Monitor",
				Value:  incident.MonitorName,
				Inline: true,
			},
			{
				Name:   "URL",
				Value:  incident.MonitorURL,
				Inline: true,
			},
			{
				Name:   "Status",
				Value:  incident.Status,
				Inline: true,
			},
			{
				Name:   "Message",
				Value:  incident.Message,
				Inline: false,
			},
			{
				Name:   "Alert Rule",
				Value:  incident.AlertName,
				Inline: false,
			},
		},
		Footer: map[string]interface{}{
			"text": fmt.Sprintf("Incident ID: %s", incident.IncidentID),
		},
	}

	// Create Discord payload
	payload := DiscordWebhookPayload{
		Embeds: []DiscordEmbed{embed},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord payload: %w", err)
	}

	// Send HTTP POST request
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Discord request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Discord request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Discord webhook returned non-success status: %d", resp.StatusCode)
	}

	return nil
}

// sendEmailNotification sends an email notification
func (j *NotificationJob) sendEmailNotification(incident *IncidentNotificationData, channel *AlertChannelConfig) error {
	to, ok := channel.Config["to"].(string)
	if !ok || to == "" {
		return fmt.Errorf("email address not configured")
	}

	// Parse and validate email address
	addr, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("invalid email address: %w", err)
	}
	// Use parsed address to be safe
	to = addr.Address

	// Extra safety check for header injection
	if strings.ContainsAny(to, "\r\n") {
		return fmt.Errorf("invalid email address: contains control characters")
	}

	if j.smtpConfig.Host == "" {
		return fmt.Errorf("SMTP host not configured")
	}
	// Extract SMTP config from channel
	smtpHost, ok := channel.Config["smtp_host"].(string)
	if !ok || smtpHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}
	smtpPortFloat, ok := channel.Config["smtp_port"].(float64)
	if !ok || smtpPortFloat <= 0 {
		return fmt.Errorf("SMTP port not configured or invalid")
	}
	smtpPort := int(smtpPortFloat)
	smtpUser, ok := channel.Config["smtp_user"].(string)
	if !ok {
		smtpUser = ""
	}
	smtpPassword, ok := channel.Config["smtp_password"].(string)
	if !ok {
		smtpPassword = ""
	}
	smtpFrom, ok := channel.Config["smtp_from"].(string)
	if !ok || smtpFrom == "" {
		return fmt.Errorf("SMTP from email not configured")
	}

	// Determine title and color (text based representation)
	var title string
	if incident.Status == "open" {
		switch incident.TriggerType {
		case "down":
			title = "🔴 Monitor Down: " + incident.MonitorName
		case "slow_response":
			title = "🐌 Slow Response: " + incident.MonitorName
		case "ssl_expiry":
			title = "🔒 SSL Expiry: " + incident.MonitorName
		default:
			title = "🚨 Incident: " + incident.MonitorName
		}
	} else {
		title = "✅ Resolved: " + incident.MonitorName
	}

	// Simple text body (using \n)
	// Simple text body
	body := fmt.Sprintf(`Subject: %s
From: %s
To: %s

%s

Monitor: %s
URL: %s
Status: %s
Message: %s
Time: %s

--
V-Insight Monitoring
`, title, j.smtpConfig.From, to, title, incident.MonitorName, incident.MonitorURL, incident.Status, incident.Message, incident.Timestamp.Format(time.RFC3339))
`, title, smtpFrom, to, title, incident.MonitorName, incident.MonitorURL, incident.Status, incident.Message, incident.Timestamp.Format(time.RFC3339))

	// Replace \n with \r\n for SMTP compliance
	body = strings.ReplaceAll(body, "\n", "\r\n")

	auth := smtp.PlainAuth("", j.smtpConfig.User, j.smtpConfig.Password, j.smtpConfig.Host)
	smtpAddr := fmt.Sprintf("%s:%d", j.smtpConfig.Host, j.smtpConfig.Port)

	// Note: smtp.SendMail requires valid auth. If no auth is needed, auth should be nil.
	// We assume auth is needed if User is set.
	if j.smtpConfig.User == "" {
		auth = nil
	}

	err = smtp.SendMail(smtpAddr, auth, j.smtpConfig.From, []string{to}, []byte(body))
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)
	smtpAddr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)

	// Note: smtp.SendMail requires valid auth. If no auth is needed, auth should be nil.
	// We assume auth is needed if User is set.
	if smtpUser == "" {
		auth = nil
	}

	err = smtp.SendMail(smtpAddr, auth, smtpFrom, []string{to}, []byte(body))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// markIncidentAsNotified marks an incident as notified
func (j *NotificationJob) markIncidentAsNotified(incidentID string) error {
	query := `
		UPDATE incidents
		SET notified_at = $1
		WHERE id = $2 AND notified_at IS NULL
	`

	result, err := j.db.Exec(query, time.Now(), incidentID)
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
