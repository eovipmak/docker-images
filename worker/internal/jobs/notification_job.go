package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/eovipmak/v-insight/worker/internal/database"
)

// IncidentNotificationData represents the full incident data for notifications
type IncidentNotificationData struct {
	IncidentID   string
	MonitorID    string
	MonitorName  string
	MonitorURL   string
	AlertRuleID  string
	AlertName    string
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
}

// NewNotificationJob creates a new notification job
func NewNotificationJob(db *database.DB) *NotificationJob {
	return &NotificationJob{
		db: db,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name returns the name of the job
func (j *NotificationJob) Name() string {
	return "NotificationJob"
}

// Run executes the notification job
func (j *NotificationJob) Run(ctx context.Context) error {
	startTime := time.Now()
	log.Println("[NotificationJob] Starting notification processing run")

	// Get incidents that need notification
	incidents, err := j.getUnnotifiedIncidents()
	if err != nil {
		log.Printf("[NotificationJob] Failed to get unnotified incidents: %v", err)
		return err
	}

	if len(incidents) == 0 {
		log.Println("[NotificationJob] No unnotified incidents found")
		return nil
	}

	log.Printf("[NotificationJob] Found %d unnotified incidents", len(incidents))

	notificationsSent := 0
	notificationsFailed := 0

	// Process each incident
	for _, incident := range incidents {
		sent, err := j.processIncidentNotifications(incident)
		if err != nil {
			log.Printf("[NotificationJob] Failed to process incident %s: %v", incident.IncidentID, err)
			notificationsFailed++
			continue
		}
		if sent > 0 {
			// Mark incident as notified
			if err := j.markIncidentAsNotified(incident.IncidentID); err != nil {
				log.Printf("[NotificationJob] Failed to mark incident %s as notified: %v", incident.IncidentID, err)
				continue
			}
			notificationsSent += sent
		}
	}

	duration := time.Since(startTime)
	log.Printf("[NotificationJob] Notification processing completed in %v - Sent: %d, Failed: %d",
		duration, notificationsSent, notificationsFailed)

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
		log.Printf("[NotificationJob] No alert channels configured for incident %s", incident.IncidentID)
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
		default:
			log.Printf("[NotificationJob] Unsupported channel type: %s", channel.Type)
			continue
		}

		if err != nil {
			log.Printf("[NotificationJob] Failed to send notification to %s channel '%s': %v",
				channel.Type, channel.Name, err)
			continue
		}

		log.Printf("[NotificationJob] ✓ Sent %s notification for incident %s via channel '%s'",
			channel.Type, incident.IncidentID, channel.Name)
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

	// Determine embed color based on status
	var color int
	var title string
	if incident.Status == "open" {
		color = 0xFF0000 // Red for open incidents
		title = "🚨 New Incident Detected"
	} else {
		color = 0x00FF00 // Green for resolved incidents
		title = "✅ Incident Resolved"
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
