package jobs

import (
	"testing"
	"time"
)

// TestIncidentNotificationData_StructFields tests that IncidentNotificationData struct has all required fields
func TestIncidentNotificationData_StructFields(t *testing.T) {
	incident := IncidentNotificationData{
		IncidentID:   "test-incident-id",
		MonitorID:    "test-monitor-id",
		MonitorName:  "Test Monitor",
		MonitorURL:   "https://example.com",
		AlertRuleID:  "test-alert-rule-id",
		AlertName:    "Test Alert",
		Status:       "open",
		Message:      "Monitor is down",
		Timestamp:    time.Now(),
	}

	if incident.IncidentID != "test-incident-id" {
		t.Errorf("Expected IncidentID 'test-incident-id', got '%s'", incident.IncidentID)
	}
	if incident.MonitorName != "Test Monitor" {
		t.Errorf("Expected MonitorName 'Test Monitor', got '%s'", incident.MonitorName)
	}
	if incident.Status != "open" {
		t.Errorf("Expected Status 'open', got '%s'", incident.Status)
	}
}

// TestAlertChannelConfig_StructFields tests that AlertChannelConfig struct has all required fields
func TestAlertChannelConfig_StructFields(t *testing.T) {
	channel := AlertChannelConfig{
		ID:   "test-channel-id",
		Type: "webhook",
		Name: "Test Webhook",
		Config: map[string]interface{}{
			"url": "https://example.com/webhook",
		},
		Enabled: true,
	}

	if channel.ID != "test-channel-id" {
		t.Errorf("Expected ID 'test-channel-id', got '%s'", channel.ID)
	}
	if channel.Type != "webhook" {
		t.Errorf("Expected Type 'webhook', got '%s'", channel.Type)
	}
	if channel.Name != "Test Webhook" {
		t.Errorf("Expected Name 'Test Webhook', got '%s'", channel.Name)
	}
	if !channel.Enabled {
		t.Error("Expected Enabled to be true")
	}
	url, ok := channel.Config["url"].(string)
	if !ok || url != "https://example.com/webhook" {
		t.Errorf("Expected Config URL 'https://example.com/webhook', got '%v'", url)
	}

	// Test email channel config
	emailChannel := AlertChannelConfig{
		ID:   "test-email-channel-id",
		Type: "email",
		Name: "Test Email",
		Config: map[string]interface{}{
			"to":            "test@example.com",
			"smtp_host":     "smtp.gmail.com",
			"smtp_port":     587.0,
			"smtp_user":     "user@gmail.com",
			"smtp_password": "password",
			"smtp_from":     "noreply@example.com",
		},
		Enabled: true,
	}

	if emailChannel.Type != "email" {
		t.Errorf("Expected Type 'email', got '%s'", emailChannel.Type)
	}
	to, ok := emailChannel.Config["to"].(string)
	if !ok || to != "test@example.com" {
		t.Errorf("Expected Config to 'test@example.com', got '%v'", to)
	}
	smtpHost, ok := emailChannel.Config["smtp_host"].(string)
	if !ok || smtpHost != "smtp.gmail.com" {
		t.Errorf("Expected Config smtp_host 'smtp.gmail.com', got '%v'", smtpHost)
	}
}

// TestWebhookPayload_StructFields tests that WebhookPayload struct has all required fields
func TestWebhookPayload_StructFields(t *testing.T) {
	payload := WebhookPayload{
		IncidentID:  "test-incident",
		MonitorName: "Test Monitor",
		MonitorURL:  "https://example.com",
		Status:      "open",
		Message:     "Test message",
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	if payload.IncidentID != "test-incident" {
		t.Errorf("Expected IncidentID 'test-incident', got '%s'", payload.IncidentID)
	}
	if payload.MonitorName != "Test Monitor" {
		t.Errorf("Expected MonitorName 'Test Monitor', got '%s'", payload.MonitorName)
	}
	if payload.Status != "open" {
		t.Errorf("Expected Status 'open', got '%s'", payload.Status)
	}
}

// TestDiscordWebhookPayload_StructFields tests that Discord payload structures are correct
func TestDiscordWebhookPayload_StructFields(t *testing.T) {
	embed := DiscordEmbed{
		Title:       "Test Title",
		Description: "Test Description",
		Color:       0xFF0000,
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields: []DiscordEmbedField{
			{
				Name:   "Field 1",
				Value:  "Value 1",
				Inline: true,
			},
		},
		Footer: map[string]interface{}{
			"text": "Test Footer",
		},
	}

	if embed.Title != "Test Title" {
		t.Errorf("Expected Title 'Test Title', got '%s'", embed.Title)
	}
	if embed.Color != 0xFF0000 {
		t.Errorf("Expected Color 0xFF0000 (red), got %d", embed.Color)
	}
	if len(embed.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(embed.Fields))
	}
	if embed.Fields[0].Name != "Field 1" {
		t.Errorf("Expected Field Name 'Field 1', got '%s'", embed.Fields[0].Name)
	}

	payload := DiscordWebhookPayload{
		Embeds: []DiscordEmbed{embed},
	}

	if len(payload.Embeds) != 1 {
		t.Errorf("Expected 1 embed, got %d", len(payload.Embeds))
	}
}

// TestNotificationJob_Name tests that the job returns correct name
func TestNotificationJob_Name(t *testing.T) {
	job := &NotificationJob{}
	expectedName := "NotificationJob"

	if job.Name() != expectedName {
		t.Errorf("Expected job name '%s', got '%s'", expectedName, job.Name())
	}
}
