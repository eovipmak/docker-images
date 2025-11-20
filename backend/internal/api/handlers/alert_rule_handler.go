package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
	"github.com/gin-gonic/gin"
)

// AlertRuleHandler handles alert rule-related HTTP requests
type AlertRuleHandler struct {
	alertRuleRepo    repository.AlertRuleRepository
	alertChannelRepo repository.AlertChannelRepository
	monitorRepo      repository.MonitorRepository
}

// NewAlertRuleHandler creates a new alert rule handler
func NewAlertRuleHandler(alertRuleRepo repository.AlertRuleRepository, alertChannelRepo repository.AlertChannelRepository, monitorRepo repository.MonitorRepository) *AlertRuleHandler {
	return &AlertRuleHandler{
		alertRuleRepo:    alertRuleRepo,
		alertChannelRepo: alertChannelRepo,
		monitorRepo:      monitorRepo,
	}
}

// CreateAlertRuleRequest represents the request body for creating an alert rule
type CreateAlertRuleRequest struct {
	MonitorID      *string  `json:"monitor_id"`
	Name           string   `json:"name" binding:"required"`
	TriggerType    string   `json:"trigger_type" binding:"required,oneof=down ssl_expiry slow_response"`
	ThresholdValue int      `json:"threshold_value" binding:"required,min=0"`
	Enabled        *bool    `json:"enabled"`
	ChannelIDs     []string `json:"channel_ids"`
}

// UpdateAlertRuleRequest represents the request body for updating an alert rule
type UpdateAlertRuleRequest struct {
	MonitorID      *string  `json:"monitor_id"`
	Name           string   `json:"name" binding:"omitempty"`
	TriggerType    string   `json:"trigger_type" binding:"omitempty,oneof=down ssl_expiry slow_response"`
	ThresholdValue *int     `json:"threshold_value" binding:"omitempty,min=0"`
	Enabled        *bool    `json:"enabled"`
	ChannelIDs     []string `json:"channel_ids"`
}

// Create handles alert rule creation
// POST /api/v1/alert-rules
func (h *AlertRuleHandler) Create(c *gin.Context) {
	var req CreateAlertRuleRequest
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

	// Validate channel IDs belong to the tenant
	if len(req.ChannelIDs) > 0 {
		for _, channelID := range req.ChannelIDs {
			channel, err := h.alertChannelRepo.GetByID(channelID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "channel not found: " + channelID})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate channel"})
				return
			}
			if channel.TenantID != tenantID {
				c.JSON(http.StatusForbidden, gin.H{"error": "channel access denied: " + channelID})
				return
			}
		}
	}

	// Validate monitor ID belongs to the tenant if provided
	var monitorID sql.NullString
	if req.MonitorID != nil && *req.MonitorID != "" {
		monitor, err := h.monitorRepo.GetByID(*req.MonitorID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "monitor not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate monitor"})
			return
		}
		if monitor.TenantID != tenantID {
			c.JSON(http.StatusForbidden, gin.H{"error": "monitor access denied"})
			return
		}
		monitorID = sql.NullString{String: *req.MonitorID, Valid: true}
	}

	// Set defaults
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	rule := &entities.AlertRule{
		TenantID:       tenantID,
		MonitorID:      monitorID,
		Name:           req.Name,
		TriggerType:    req.TriggerType,
		ThresholdValue: req.ThresholdValue,
		Enabled:        enabled,
	}

	if err := h.alertRuleRepo.Create(rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create alert rule"})
		return
	}

	// Attach channels if provided
	if len(req.ChannelIDs) > 0 {
		if err := h.alertRuleRepo.AttachChannels(tenantID, rule.ID, req.ChannelIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to attach channels"})
			return
		}
	}

	// Return rule with channels
	ruleWithChannels, err := h.alertRuleRepo.GetWithChannels(tenantID, rule.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve created rule"})
		return
	}

	c.JSON(http.StatusCreated, ruleWithChannels)
}

// List handles retrieving all alert rules for the current tenant
// GET /api/v1/alert-rules
func (h *AlertRuleHandler) List(c *gin.Context) {
	// Get tenant ID from context (set by middleware)
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	rules, err := h.alertRuleRepo.GetAllWithChannelsByTenantID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve alert rules"})
		return
	}

	// Return empty array instead of null if no rules found
	if rules == nil {
		rules = []*entities.AlertRuleWithChannels{}
	}

	c.JSON(http.StatusOK, rules)
}

// GetByID handles retrieving a specific alert rule
// GET /api/v1/alert-rules/:id
func (h *AlertRuleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert rule ID required"})
		return
	}

	// Get tenant ID from context for authorization check
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	// Get rule with channels (tenant-scoped)
	ruleWithChannels, err := h.alertRuleRepo.GetWithChannels(tenantID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve alert rule with channels"})
		return
	}

	c.JSON(http.StatusOK, ruleWithChannels)
}

// Update handles updating an alert rule
// PUT /api/v1/alert-rules/:id
func (h *AlertRuleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert rule ID required"})
		return
	}

	var req UpdateAlertRuleRequest
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

	// Get existing rule (tenant-scoped)
	rule, err := h.alertRuleRepo.GetByID(tenantID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve alert rule"})
		return
	}

	// Validate channel IDs belong to the tenant if provided
	if len(req.ChannelIDs) > 0 {
		for _, channelID := range req.ChannelIDs {
			channel, err := h.alertChannelRepo.GetByID(channelID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "channel not found: " + channelID})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate channel"})
				return
			}
			if channel.TenantID != tenantID {
				c.JSON(http.StatusForbidden, gin.H{"error": "channel access denied: " + channelID})
				return
			}
		}
	}

	// Update fields if provided
	if req.MonitorID != nil {
		if *req.MonitorID == "" {
			rule.MonitorID = sql.NullString{Valid: false}
		} else {
			// Validate monitor ID belongs to the tenant
			monitor, err := h.monitorRepo.GetByID(*req.MonitorID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "monitor not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate monitor"})
				return
			}
			if monitor.TenantID != tenantID {
				c.JSON(http.StatusForbidden, gin.H{"error": "monitor access denied"})
				return
			}
			rule.MonitorID = sql.NullString{String: *req.MonitorID, Valid: true}
		}
	}
	if req.Name != "" {
		rule.Name = req.Name
	}
	if req.TriggerType != "" {
		rule.TriggerType = req.TriggerType
	}
	if req.ThresholdValue != nil {
		rule.ThresholdValue = *req.ThresholdValue
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}

	if err := h.alertRuleRepo.Update(rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update alert rule"})
		return
	}

	// Update channels if provided (replace all)
	if req.ChannelIDs != nil {
		// Get current channels
		currentChannels, err := h.alertRuleRepo.GetChannelsByRuleID(tenantID, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get current channels"})
			return
		}

		// Detach all current channels
		if len(currentChannels) > 0 {
			if err := h.alertRuleRepo.DetachChannels(tenantID, id, currentChannels); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to detach channels"})
				return
			}
		}

		// Attach new channels
		if len(req.ChannelIDs) > 0 {
			if err := h.alertRuleRepo.AttachChannels(tenantID, id, req.ChannelIDs); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to attach channels"})
				return
			}
		}
	}

	// Return updated rule with channels
	ruleWithChannels, err := h.alertRuleRepo.GetWithChannels(tenantID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve updated rule"})
		return
	}

	c.JSON(http.StatusOK, ruleWithChannels)
}

// Delete handles deleting an alert rule
// DELETE /api/v1/alert-rules/:id
func (h *AlertRuleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert rule ID required"})
		return
	}

	// Get tenant ID from context for authorization check
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	// Delete rule (tenant-scoped)
	if err := h.alertRuleRepo.Delete(tenantID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete alert rule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alert rule deleted successfully"})
}
