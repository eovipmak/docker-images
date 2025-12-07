package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/eovipmak/v-insight/backend/internal/utils"
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

// Create godoc
// @Summary Create a new alert rule
// @Description Create a new alert rule with optional alert channels
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateAlertRuleRequest true "Alert rule configuration"
// @Success 201 {object} entities.AlertRuleWithChannels "Alert rule created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /alert-rules [post]
func (h *AlertRuleHandler) Create(c *gin.Context) {
	var req CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user context not found"})
		return
	}
	userID := userIDValue.(int)

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
			if channel.UserID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "channel access denied: " + channelID})
				return
			}
		}
	}

	// Validate monitor ID belongs to the tenant if provided
	var monitorID *string
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
		if monitor.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "monitor access denied"})
			return
		}
		
		// Validate SSL expiry rules cannot be created for TCP monitors
		if req.TriggerType == "ssl_expiry" && monitor.Type == "tcp" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "SSL expiry rules cannot be created for TCP monitors"})
			return
		}
		
		monitorID = req.MonitorID
	}

	// Set defaults
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	// Sanitize name to prevent XSS
	sanitizedName, valid := utils.SanitizeAndValidate(req.Name, 1, 255)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule name must be between 1 and 255 characters"})
		return
	}

	rule := &entities.AlertRule{
		UserID:         userID,
		MonitorID:      monitorID,
		Name:           sanitizedName,
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
		if err := h.alertRuleRepo.AttachChannels(userID, rule.ID, req.ChannelIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to attach channels"})
			return
		}
	}

	// Return rule with channels
	ruleWithChannels, err := h.alertRuleRepo.GetWithChannels(userID, rule.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve created rule"})
		return
	}

	c.JSON(http.StatusCreated, ruleWithChannels)
}

// List godoc
// @Summary List all alert rules
// @Description Get all alert rules with their associated channels for the current tenant
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entities.AlertRuleWithChannels "List of alert rules"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /alert-rules [get]
func (h *AlertRuleHandler) List(c *gin.Context) {
	// Get user ID from context (set by middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user context not found"})
		return
	}
	userID := userIDValue.(int)

	rules, err := h.alertRuleRepo.GetAllWithChannelsByUserID(userID)
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

// GetByID godoc
// @Summary Get an alert rule by ID
// @Description Get detailed information about a specific alert rule including its channels
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Alert Rule ID"
// @Success 200 {object} entities.AlertRuleWithChannels "Alert rule details"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Alert rule not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /alert-rules/{id} [get]
func (h *AlertRuleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert rule ID required"})
		return
	}

	// Get user ID from context for authorization check
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user context not found"})
		return
	}
	userID := userIDValue.(int)

	// Get rule with channels (user-scoped)
	ruleWithChannels, err := h.alertRuleRepo.GetWithChannels(userID, id)
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

// Update godoc
// @Summary Update an alert rule
// @Description Update an existing alert rule's configuration and associated channels
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Alert Rule ID"
// @Param request body UpdateAlertRuleRequest true "Updated alert rule configuration"
// @Success 200 {object} entities.AlertRuleWithChannels "Alert rule updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Alert rule not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /alert-rules/{id} [put]
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

	// Get user ID from context for authorization check
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user context not found"})
		return
	}
	userID := userIDValue.(int)

	// Get existing rule (user-scoped)
	rule, err := h.alertRuleRepo.GetByID(userID, id)
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
			if channel.UserID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "channel access denied: " + channelID})
				return
			}
		}
	}

	// Update fields if provided
	if req.MonitorID != nil {
		if *req.MonitorID == "" {
			rule.MonitorID = nil
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
			if monitor.UserID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "monitor access denied"})
				return
			}
			
			// Validate SSL expiry rules cannot be created for TCP monitors
			triggerType := req.TriggerType
			if triggerType == "" {
				triggerType = rule.TriggerType
			}
			if triggerType == "ssl_expiry" && monitor.Type == "tcp" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "SSL expiry rules cannot be created for TCP monitors"})
				return
			}
			
			rule.MonitorID = req.MonitorID
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
		currentChannels, err := h.alertRuleRepo.GetChannelsByRuleID(userID, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get current channels"})
			return
		}

		// Detach all current channels
		if len(currentChannels) > 0 {
			if err := h.alertRuleRepo.DetachChannels(userID, id, currentChannels); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to detach channels"})
				return
			}
		}

		// Attach new channels
		if len(req.ChannelIDs) > 0 {
			if err := h.alertRuleRepo.AttachChannels(userID, id, req.ChannelIDs); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to attach channels"})
				return
			}
		}
	}

	// Return updated rule with channels
	ruleWithChannels, err := h.alertRuleRepo.GetWithChannels(userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve updated rule"})
		return
	}

	c.JSON(http.StatusOK, ruleWithChannels)
}

// Delete godoc
// @Summary Delete an alert rule
// @Description Delete an alert rule and its channel associations
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Alert Rule ID"
// @Success 200 {object} map[string]string "Alert rule deleted successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Alert rule not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /alert-rules/{id} [delete]
func (h *AlertRuleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert rule ID required"})
		return
	}

	// Get user ID from context for authorization check
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user context not found"})
		return
	}
	userID := userIDValue.(int)

	// Delete rule (user-scoped)
	if err := h.alertRuleRepo.Delete(userID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete alert rule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alert rule deleted successfully"})
}

// Test godoc
// @Summary Test an alert rule
// @Description Validate an alert rule's configuration
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Alert Rule ID"
// @Success 200 {object} map[string]interface{} "Alert rule validation result"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Alert rule not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /alert-rules/{id}/test [post]
func (h *AlertRuleHandler) Test(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert rule ID required"})
		return
	}

	// Get user ID from context for authorization check
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user context not found"})
		return
	}
	userID := userIDValue.(int)

	// Get existing rule
	rule, err := h.alertRuleRepo.GetWithChannels(userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve alert rule"})
		return
	}

	// Validate rule configuration
	issues := h.validateRuleConfiguration(rule)

	if len(issues) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"valid":  false,
			"issues": issues,
			"message": "alert rule has configuration issues",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "alert rule configuration is valid",
	})
}

// validateRuleConfiguration validates an alert rule's configuration
func (h *AlertRuleHandler) validateRuleConfiguration(rule *entities.AlertRuleWithChannels) []string {
	var issues []string

	// Check if rule has channels
	if len(rule.ChannelIDs) == 0 {
		issues = append(issues, "no notification channels configured")
	}

	// Check if monitor exists (if monitor-specific)
	if rule.MonitorID != nil {
		_, err := h.monitorRepo.GetByID(*rule.MonitorID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				issues = append(issues, "associated monitor not found")
			} else {
				issues = append(issues, "failed to validate monitor")
			}
		}
	}

	// Validate trigger-specific logic
	switch rule.TriggerType {
	case "ssl_expiry":
		if rule.MonitorID != nil {
			// Check if monitor has SSL checking enabled
			monitor, err := h.monitorRepo.GetByID(*rule.MonitorID)
			if err == nil && !monitor.CheckSSL {
				issues = append(issues, "monitor does not have SSL checking enabled")
			}
		}
	}

	return issues
}
