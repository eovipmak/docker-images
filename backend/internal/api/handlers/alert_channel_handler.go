package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
	"github.com/eovipmak/v-insight/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// AlertChannelHandler handles alert channel-related HTTP requests
type AlertChannelHandler struct {
	alertChannelRepo repository.AlertChannelRepository
}

// NewAlertChannelHandler creates a new alert channel handler
func NewAlertChannelHandler(alertChannelRepo repository.AlertChannelRepository) *AlertChannelHandler {
	return &AlertChannelHandler{
		alertChannelRepo: alertChannelRepo,
	}
}

// CreateAlertChannelRequest represents the request body for creating an alert channel
type CreateAlertChannelRequest struct {
	Type    string                 `json:"type" binding:"required,oneof=webhook discord email"`
	Name    string                 `json:"name" binding:"required"`
	Config  map[string]interface{} `json:"config" binding:"required"`
	Enabled *bool                  `json:"enabled"`
}

// UpdateAlertChannelRequest represents the request body for updating an alert channel
type UpdateAlertChannelRequest struct {
	Type    string                 `json:"type" binding:"omitempty,oneof=webhook discord email"`
	Name    string                 `json:"name" binding:"omitempty"`
	Config  map[string]interface{} `json:"config"`
	Enabled *bool                  `json:"enabled"`
}

// Create handles alert channel creation
// POST /api/v1/alert-channels
func (h *AlertChannelHandler) Create(c *gin.Context) {
	var req CreateAlertChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant ID from context (set by middleware)
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	// Sanitize name to prevent XSS
	sanitizedName, valid := utils.SanitizeAndValidate(req.Name, 1, 255)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel name must be between 1 and 255 characters"})
		return
	}

	// Set defaults
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	channel := &entities.AlertChannel{
		TenantID: tenantID,
		Type:     req.Type,
		Name:     sanitizedName,
		Config:   req.Config,
		Enabled:  enabled,
	}

	if err := h.alertChannelRepo.Create(channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create alert channel"})
		return
	}

	c.JSON(http.StatusCreated, channel)
}

// List handles retrieving all alert channels for the current tenant
// GET /api/v1/alert-channels
func (h *AlertChannelHandler) List(c *gin.Context) {
	// Get tenant ID from context (set by middleware)
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	channels, err := h.alertChannelRepo.GetByTenantID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve alert channels"})
		return
	}

	// Return empty array instead of null if no channels found
	if channels == nil {
		channels = []*entities.AlertChannel{}
	}

	c.JSON(http.StatusOK, channels)
}

// GetByID handles retrieving a specific alert channel
// GET /api/v1/alert-channels/:id
func (h *AlertChannelHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert channel ID required"})
		return
	}

	// Get tenant ID from context for authorization check
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	channel, err := h.alertChannelRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert channel not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve alert channel"})
		return
	}

	// Verify that the alert channel belongs to the current tenant
	if channel.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, channel)
}

// Update handles updating an alert channel
// PUT /api/v1/alert-channels/:id
func (h *AlertChannelHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert channel ID required"})
		return
	}

	var req UpdateAlertChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant ID from context for authorization check
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	// Get existing channel
	channel, err := h.alertChannelRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert channel not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve alert channel"})
		return
	}

	// Verify that the alert channel belongs to the current tenant
	if channel.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Update fields if provided
	if req.Type != "" {
		channel.Type = req.Type
	}
	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.Config != nil {
		channel.Config = req.Config
	}
	if req.Enabled != nil {
		channel.Enabled = *req.Enabled
	}

	if err := h.alertChannelRepo.Update(channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update alert channel"})
		return
	}

	c.JSON(http.StatusOK, channel)
}

// Delete handles deleting an alert channel
// DELETE /api/v1/alert-channels/:id
func (h *AlertChannelHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert channel ID required"})
		return
	}

	// Get tenant ID from context for authorization check
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	// Get existing channel to verify ownership
	channel, err := h.alertChannelRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert channel not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve alert channel"})
		return
	}

	// Verify that the alert channel belongs to the current tenant
	if channel.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.alertChannelRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete alert channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alert channel deleted successfully"})
}

// Test handles testing an alert channel by sending a test notification
// POST /api/v1/alert-channels/:id/test
func (h *AlertChannelHandler) Test(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert channel ID required"})
		return
	}

	// Get tenant ID from context for authorization check
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	// Get existing channel
	channel, err := h.alertChannelRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert channel not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve alert channel"})
		return
	}

	// Verify that the alert channel belongs to the current tenant
	if channel.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Send test notification
	if err := h.sendTestNotification(channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send test notification: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "test notification sent successfully"})
}

// sendTestNotification sends a test notification to the given channel
func (h *AlertChannelHandler) sendTestNotification(channel *entities.AlertChannel) error {
	// Create test data similar to notification job
	testData := map[string]interface{}{
		"incident_id":  "test-incident-123",
		"monitor_name": "Test Monitor",
		"monitor_url":  "https://example.com",
		"status":       "test",
		"message":      "This is a test notification from V-Insight",
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	switch channel.Type {
	case "webhook":
		return h.sendTestWebhook(channel, testData)
	case "discord":
		return h.sendTestDiscord(channel, testData)
	case "email":
		return h.sendTestEmail(channel, testData)
	default:
		return fmt.Errorf("unsupported channel type: %s", channel.Type)
	}
}

// sendTestWebhook sends a test webhook notification
func (h *AlertChannelHandler) sendTestWebhook(channel *entities.AlertChannel, data map[string]interface{}) error {
	webhookURL, ok := channel.Config["url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhook URL not configured")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}

	return nil
}

// sendTestDiscord sends a test Discord notification
func (h *AlertChannelHandler) sendTestDiscord(channel *entities.AlertChannel, data map[string]interface{}) error {
	webhookURL, ok := channel.Config["url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("Discord webhook URL not configured")
	}

	embed := map[string]interface{}{
		"title":       "🧪 Test Notification",
		"description": "This is a test notification from V-Insight",
		"color":       0x00FF00, // Green
		"fields": []map[string]interface{}{
			{
				"name":   "Monitor",
				"value":  data["monitor_name"].(string),
				"inline": true,
			},
			{
				"name":   "URL",
				"value":  data["monitor_url"].(string),
				"inline": true,
			},
			{
				"name":   "Status",
				"value":  "Test",
				"inline": true,
			},
			{
				"name":   "Message",
				"value":  data["message"].(string),
				"inline": false,
			},
		},
		"timestamp": data["timestamp"],
		"footer": map[string]interface{}{
			"text": "V-Insight Test Notification",
		},
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{embed},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Discord request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Discord request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Discord webhook returned non-success status: %d", resp.StatusCode)
	}

	return nil
}

// sendTestEmail sends a test email notification (placeholder)
func (h *AlertChannelHandler) sendTestEmail(channel *entities.AlertChannel, data map[string]interface{}) error {
	// Email implementation is ready for future development
	return fmt.Errorf("email notifications not yet implemented")
}
