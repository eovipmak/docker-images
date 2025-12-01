package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/eovipmak/v-insight/backend/internal/domain/service"
	"github.com/eovipmak/v-insight/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

var tcpAddressRegex = regexp.MustCompile(`^[^:]+:\d+$`)

// MonitorHandler handles monitor-related HTTP requests
type MonitorHandler struct {
	monitorRepo      repository.MonitorRepository
	alertRuleRepo    repository.AlertRuleRepository
	alertChannelRepo repository.AlertChannelRepository
	monitorService   *service.MonitorService
}

// NewMonitorHandler creates a new monitor handler
func NewMonitorHandler(monitorRepo repository.MonitorRepository, alertRuleRepo repository.AlertRuleRepository, alertChannelRepo repository.AlertChannelRepository, monitorService *service.MonitorService) *MonitorHandler {
	return &MonitorHandler{
		monitorRepo:      monitorRepo,
		alertRuleRepo:    alertRuleRepo,
		alertChannelRepo: alertChannelRepo,
		monitorService:   monitorService,
	}
}

// CreateMonitorRequest represents the request body for creating a monitor
type CreateMonitorRequest struct {
	Name          string `json:"name" binding:"required"`
	URL           string `json:"url" binding:"required"`
	Type          string `json:"type" binding:"omitempty,oneof=http tcp ping icmp"`
	Keyword       *string `json:"keyword" binding:"omitempty"`
	CheckInterval int    `json:"check_interval" binding:"omitempty,min=60"`     // minimum 60 seconds
	Timeout       int    `json:"timeout" binding:"omitempty,min=5,max=120"`     // 5-120 seconds
	Enabled       *bool  `json:"enabled"`                                        // pointer to allow explicit false
	CheckSSL      *bool  `json:"check_ssl"`                                      // pointer to allow explicit false
	SSLAlertDays  int    `json:"ssl_alert_days" binding:"omitempty,min=1"`      // minimum 1 day
}

// UpdateMonitorRequest represents the request body for updating a monitor
type UpdateMonitorRequest struct {
	Name          string `json:"name" binding:"omitempty"`
	URL           string `json:"url" binding:"omitempty"`
	Type          string `json:"type" binding:"omitempty,oneof=http tcp ping icmp"`
	Keyword       *string `json:"keyword" binding:"omitempty"`
	CheckInterval int    `json:"check_interval" binding:"omitempty,min=60"`
	Timeout       int    `json:"timeout" binding:"omitempty,min=5,max=120"`
	Enabled       *bool  `json:"enabled"`
	CheckSSL      *bool  `json:"check_ssl"`
	SSLAlertDays  int    `json:"ssl_alert_days" binding:"omitempty,min=1"`
}

// Create godoc
// @Summary Create a new monitor
// @Description Create a new monitoring endpoint for a website or service
// @Tags Monitors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateMonitorRequest true "Monitor configuration"
// @Success 201 {object} entities.Monitor "Monitor created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /monitors [post]
func (h *MonitorHandler) Create(c *gin.Context) {
	var req CreateMonitorRequest
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

	// Validate URL based on monitor type
	monitorType := req.Type
	if monitorType == "" {
		monitorType = "http" // default
	}
	
	if monitorType == "http" {
		// Validate as URL
		if _, err := url.ParseRequestURI(req.URL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL format"})
			return
		}
	} else if monitorType == "tcp" {
		// Validate as host:port
		if !tcpAddressRegex.MatchString(req.URL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Host:Port format. Use format: host:port"})
			return
		}
	} else if monitorType == "ping" || monitorType == "icmp" {
		// Validate as hostname or IP (no protocol)
		if strings.Contains(req.URL, "://") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Hostname/IP format. Do not include protocol (http://, etc.)"})
			return
		}
	}

	// Sanitize user inputs to prevent XSS
	sanitizedName, valid := utils.SanitizeAndValidate(req.Name, 1, 255)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monitor name must be between 1 and 255 characters"})
		return
	}

	// Set defaults if not provided
	checkInterval := req.CheckInterval
	if checkInterval == 0 {
		checkInterval = 300 // 5 minutes default
	}

	timeout := req.Timeout
	if timeout == 0 {
		timeout = 30 // 30 seconds default
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	checkSSL := true
	if req.CheckSSL != nil {
		checkSSL = *req.CheckSSL
	}

	sslAlertDays := 30
	if req.SSLAlertDays > 0 {
		sslAlertDays = req.SSLAlertDays
	}

	keyword := ""
	if req.Keyword != nil {
		keyword = *req.Keyword
	}

	monitor := &entities.Monitor{
		TenantID:      tenantID,
		Name:          sanitizedName,
		URL:           req.URL,
		Type:          monitorType,
		Keyword:       keyword,
		CheckInterval: checkInterval,
		Timeout:       timeout,
		Enabled:       enabled,
		CheckSSL:      checkSSL,
		SSLAlertDays:  sslAlertDays,
	}

	if err := h.monitorRepo.Create(monitor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create monitor"})
		return
	}

	// Auto-create SSL expiry alert rule if SSL checking is enabled
	// (Only for HTTP)
	if monitorType == "http" && monitor.CheckSSL && monitor.SSLAlertDays > 0 {
		if err := h.createOrUpdateSSLAlertRule(tenantID, monitor); err != nil {
			// Log error but don't fail the monitor creation
			fmt.Printf("Warning: Failed to create SSL alert rule for monitor %s: %v\n", monitor.Name, err)
		}
	}

	c.JSON(http.StatusCreated, monitor)
}

// List godoc
// @Summary List all monitors
// @Description Get all monitors for the current tenant
// @Tags Monitors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entities.Monitor "List of monitors"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /monitors [get]
func (h *MonitorHandler) List(c *gin.Context) {
	// Get tenant ID from context (set by middleware)
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	monitors, err := h.monitorRepo.GetByTenantID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve monitors"})
		return
	}

	// Return empty array instead of null if no monitors found
	if monitors == nil {
		monitors = []*entities.Monitor{}
	}

	c.JSON(http.StatusOK, monitors)
}

// GetByID godoc
// @Summary Get a monitor by ID
// @Description Get detailed information about a specific monitor
// @Tags Monitors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Monitor ID"
// @Success 200 {object} entities.Monitor "Monitor details"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Monitor not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /monitors/{id} [get]
func (h *MonitorHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monitor ID required"})
		return
	}

	// Get tenant ID from context for authorization check
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	monitor, err := h.monitorRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "monitor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve monitor"})
		return
	}

	// Verify that the monitor belongs to the current tenant
	if monitor.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, monitor)
}

// Update godoc
// @Summary Update a monitor
// @Description Update an existing monitor's configuration
// @Tags Monitors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Monitor ID"
// @Param request body UpdateMonitorRequest true "Updated monitor configuration"
// @Success 200 {object} entities.Monitor "Monitor updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Monitor not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /monitors/{id} [put]
func (h *MonitorHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monitor ID required"})
		return
	}

	var req UpdateMonitorRequest
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

	// Get existing monitor
	monitor, err := h.monitorRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "monitor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve monitor"})
		return
	}

	// Verify that the monitor belongs to the current tenant
	if monitor.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Determine monitor type for validation
	monitorType := monitor.Type
	if req.Type != "" {
		monitorType = req.Type
	}

	// Validate URL if provided
	if req.URL != "" {
		if monitorType == "http" {
			// Validate as URL
			if _, err := url.ParseRequestURI(req.URL); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL format"})
				return
			}
		} else if monitorType == "tcp" {
			// Validate as host:port
			if !tcpAddressRegex.MatchString(req.URL) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Host:Port format. Use format: host:port"})
				return
			}
		} else if monitorType == "ping" || monitorType == "icmp" {
			// Validate as hostname or IP (no protocol)
			if strings.Contains(req.URL, "://") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Hostname/IP format. Do not include protocol (http://, etc.)"})
				return
			}
		}
	}

	// Update fields if provided
	if req.Name != "" {
		// Sanitize name to prevent XSS
		sanitizedName, valid := utils.SanitizeAndValidate(req.Name, 1, 255)
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "monitor name must be between 1 and 255 characters"})
			return
		}
		monitor.Name = sanitizedName
	}
	if req.URL != "" {
		monitor.URL = req.URL
	}
	if req.Type != "" {
		monitor.Type = req.Type
	}

	// Update keyword if provided (including empty string if explicitly sent)
	if req.Keyword != nil {
		monitor.Keyword = *req.Keyword
	}

	if req.CheckInterval > 0 {
		monitor.CheckInterval = req.CheckInterval
	}
	if req.Timeout > 0 {
		monitor.Timeout = req.Timeout
	}
	if req.Enabled != nil {
		monitor.Enabled = *req.Enabled
	}
	if req.CheckSSL != nil {
		monitor.CheckSSL = *req.CheckSSL
	}
	if req.SSLAlertDays > 0 {
		monitor.SSLAlertDays = req.SSLAlertDays
	}

	if err := h.monitorRepo.Update(monitor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update monitor"})
		return
	}

	// Update SSL alert rule if SSL settings changed
	// (Only if type is HTTP)
	if monitorType == "http" && (req.CheckSSL != nil || req.SSLAlertDays > 0) {
		if monitor.CheckSSL && monitor.SSLAlertDays > 0 {
			if err := h.createOrUpdateSSLAlertRule(tenantID, monitor); err != nil {
				fmt.Printf("Warning: Failed to update SSL alert rule for monitor %s: %v\n", monitor.Name, err)
			}
		} else {
			// Disable SSL alert rule if SSL checking is disabled
			if err := h.disableSSLAlertRule(tenantID, monitor.ID); err != nil {
				fmt.Printf("Warning: Failed to disable SSL alert rule for monitor %s: %v\n", monitor.Name, err)
			}
		}
	} else if monitorType != "http" {
        // If type changed from HTTP to something else, disable SSL rule?
        // We probably should.
        // If monitor.Type was updated, we should check if we need to clean up SSL rules.
        // But simplifying: just leave it for now or disable if type changed.
        // If we switched away from HTTP, we should probably disable SSL monitoring rules.
        if req.Type != "" && req.Type != "http" {
             if err := h.disableSSLAlertRule(tenantID, monitor.ID); err != nil {
				fmt.Printf("Warning: Failed to disable SSL alert rule for monitor %s: %v\n", monitor.Name, err)
			}
        }
    }

	c.JSON(http.StatusOK, monitor)
}

// Delete godoc
// @Summary Delete a monitor
// @Description Delete a monitor and its associated alert rules
// @Tags Monitors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Monitor ID"
// @Success 200 {object} map[string]string "Monitor deleted successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Monitor not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /monitors/{id} [delete]
func (h *MonitorHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monitor ID required"})
		return
	}

	// Get tenant ID from context for authorization check
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	// Get existing monitor to verify ownership
	monitor, err := h.monitorRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "monitor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve monitor"})
		return
	}

	// Verify that the monitor belongs to the current tenant
	if monitor.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Delete associated SSL alert rule first
	if err := h.disableSSLAlertRule(tenantID, id); err != nil {
		fmt.Printf("Warning: Failed to delete SSL alert rule for monitor %s: %v\n", monitor.Name, err)
	}

	if err := h.monitorRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete monitor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "monitor deleted successfully"})
}

// GetChecks godoc
// @Summary Get monitor check history
// @Description Get the health check history for a specific monitor
// @Tags Monitors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Monitor ID"
// @Success 200 {array} entities.MonitorCheck "List of monitor checks"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Monitor not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /monitors/{id}/checks [get]
func (h *MonitorHandler) GetChecks(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monitor ID required"})
		return
	}

	// Get tenant ID from context for authorization check
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	// Get existing monitor to verify ownership
	monitor, err := h.monitorRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "monitor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve monitor"})
		return
	}

	// Verify that the monitor belongs to the current tenant
	if monitor.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Get optional limit parameter (default to 100 in repository)
	limit := 100
	if limitParam := c.Query("limit"); limitParam != "" {
		var err error
		if _, err = fmt.Sscanf(limitParam, "%d", &limit); err != nil || limit <= 0 {
			limit = 100
		}
	}

	checks, err := h.monitorRepo.GetChecksByMonitorID(id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve checks"})
		return
	}

	// Return empty array instead of null if no checks found
	if checks == nil {
		checks = []*entities.MonitorCheck{}
	}

	c.JSON(http.StatusOK, checks)
}

// GetSSLStatus handles retrieving SSL certificate status for a monitor
// GET /api/v1/monitors/:id/ssl-status
// GetSSLStatus godoc
// @Summary Get SSL certificate status
// @Description Get SSL certificate information and expiration status for a monitor
// @Tags Monitors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Monitor ID"
// @Success 200 {object} map[string]interface{} "SSL status and alert threshold"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Monitor not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /monitors/{id}/ssl-status [get]
func (h *MonitorHandler) GetSSLStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monitor ID required"})
		return
	}

	// Get tenant ID from context for authorization check
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	// Get existing monitor to verify ownership
	monitor, err := h.monitorRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "monitor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve monitor"})
		return
	}

	// Verify that the monitor belongs to the current tenant
	if monitor.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Check if SSL checking is enabled
	if !monitor.CheckSSL {
		c.JSON(http.StatusOK, gin.H{
			"message": "SSL checking is disabled for this monitor",
			"ssl_status": nil,
		})
		return
	}

	// Get SSL status
	sslStatus, err := h.monitorService.GetSSLStatus(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve SSL status"})
		return
	}

	// If no SSL status available yet
	if sslStatus == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "No SSL check data available yet",
			"ssl_status": nil,
		})
		return
	}

	// Check if expiring soon
	sslStatus.ExpiringSoon = sslStatus.Valid && sslStatus.DaysUntilExpiry <= monitor.SSLAlertDays && sslStatus.DaysUntilExpiry >= 0

	c.JSON(http.StatusOK, gin.H{
		"ssl_status": sslStatus,
		"alert_threshold": monitor.SSLAlertDays,
	})
}

// GetStats godoc
// @Summary Get monitor response time statistics
// @Description Get response time statistics for a monitor over the last 24 hours
// @Tags Monitors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Monitor ID"
// @Success 200 {array} entities.MonitorStat "Array of response time statistics"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Monitor not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /monitors/{id}/stats [get]
func (h *MonitorHandler) GetStats(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monitor ID required"})
		return
	}

	// Get tenant ID from context for authorization check
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	// Get existing monitor to verify ownership
	monitor, err := h.monitorRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "monitor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve monitor"})
		return
	}

	// Verify that the monitor belongs to the current tenant
	if monitor.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	stats, err := h.monitorRepo.GetStatsByMonitorID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve stats"})
		return
	}

	// Return empty array instead of null if no stats found
	if stats == nil {
		stats = []*entities.MonitorStat{}
	}

	c.JSON(http.StatusOK, stats)
}

// createOrUpdateSSLAlertRule creates or updates an SSL expiry alert rule for a monitor
func (h *MonitorHandler) createOrUpdateSSLAlertRule(tenantID int, monitor *entities.Monitor) error {
	ruleName := fmt.Sprintf("SSL Expiry Alert - %s (%d days)", monitor.Name, monitor.SSLAlertDays)

	// Check if alert rule already exists for this monitor
	existingRules, err := h.alertRuleRepo.GetAllWithChannelsByTenantID(tenantID)
	if err != nil {
		return fmt.Errorf("failed to check existing alert rules: %w", err)
	}

	var existingRule *entities.AlertRuleWithChannels
	for _, rule := range existingRules {
		// Check if this is an auto-generated SSL rule for this monitor
		if rule.MonitorID.Valid && rule.MonitorID.String == monitor.ID && 
		   rule.TriggerType == "ssl_expiry" && 
		   strings.HasPrefix(rule.Name, "SSL Expiry Alert - ") {
			existingRule = rule
			break
		}
	}

	if existingRule != nil {
		// Update existing rule
		existingRule.Name = ruleName
		existingRule.ThresholdValue = monitor.SSLAlertDays
		existingRule.Enabled = monitor.Enabled

		alertRule := &entities.AlertRule{
			ID:             existingRule.ID,
			TenantID:       existingRule.TenantID,
			MonitorID:      existingRule.MonitorID,
			Name:           existingRule.Name,
			TriggerType:    existingRule.TriggerType,
			ThresholdValue: existingRule.ThresholdValue,
			Enabled:        existingRule.Enabled,
		}

		return h.alertRuleRepo.Update(alertRule)
	} else {
		// Create new rule
		alertRule := &entities.AlertRule{
			TenantID:       tenantID,
			MonitorID:      sql.NullString{String: monitor.ID, Valid: true},
			Name:           ruleName,
			TriggerType:    "ssl_expiry",
			ThresholdValue: monitor.SSLAlertDays,
			Enabled:        monitor.Enabled,
		}

		if err := h.alertRuleRepo.Create(alertRule); err != nil {
			return fmt.Errorf("failed to create SSL alert rule: %w", err)
		}

		return nil
	}
}

// disableSSLAlertRule disables SSL alert rules for a monitor
func (h *MonitorHandler) disableSSLAlertRule(tenantID int, monitorID string) error {
	// Find and disable all auto-generated SSL expiry rules for this monitor
	rules, err := h.alertRuleRepo.GetAllWithChannelsByTenantID(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get alert rules: %w", err)
	}

	for _, rule := range rules {
		if rule.MonitorID.Valid && rule.MonitorID.String == monitorID && 
		   rule.TriggerType == "ssl_expiry" && 
		   strings.HasPrefix(rule.Name, "SSL Expiry Alert - ") {
			// Disable the auto-generated rule
			rule.Enabled = false
			alertRule := &entities.AlertRule{
				ID:             rule.ID,
				TenantID:       rule.TenantID,
				MonitorID:      rule.MonitorID,
				Name:           rule.Name,
				TriggerType:    rule.TriggerType,
				ThresholdValue: rule.ThresholdValue,
				Enabled:        rule.Enabled,
			}

			if err := h.alertRuleRepo.Update(alertRule); err != nil {
				return fmt.Errorf("failed to disable SSL alert rule %s: %w", rule.ID, err)
			}
		}
	}

	return nil
}
