package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
	"github.com/eovipmak/v-insight/backend/internal/domain/service"
	"github.com/gin-gonic/gin"
)

// MonitorHandler handles monitor-related HTTP requests
type MonitorHandler struct {
	monitorRepo    repository.MonitorRepository
	monitorService *service.MonitorService
}

// NewMonitorHandler creates a new monitor handler
func NewMonitorHandler(monitorRepo repository.MonitorRepository, monitorService *service.MonitorService) *MonitorHandler {
	return &MonitorHandler{
		monitorRepo:    monitorRepo,
		monitorService: monitorService,
	}
}

// CreateMonitorRequest represents the request body for creating a monitor
type CreateMonitorRequest struct {
	Name          string `json:"name" binding:"required"`
	URL           string `json:"url" binding:"required,url"`
	CheckInterval int    `json:"check_interval" binding:"omitempty,min=60"`     // minimum 60 seconds
	Timeout       int    `json:"timeout" binding:"omitempty,min=5,max=120"`     // 5-120 seconds
	Enabled       *bool  `json:"enabled"`                                        // pointer to allow explicit false
	CheckSSL      *bool  `json:"check_ssl"`                                      // pointer to allow explicit false
	SSLAlertDays  int    `json:"ssl_alert_days" binding:"omitempty,min=1"`      // minimum 1 day
}

// UpdateMonitorRequest represents the request body for updating a monitor
type UpdateMonitorRequest struct {
	Name          string `json:"name" binding:"omitempty"`
	URL           string `json:"url" binding:"omitempty,url"`
	CheckInterval int    `json:"check_interval" binding:"omitempty,min=60"`
	Timeout       int    `json:"timeout" binding:"omitempty,min=5,max=120"`
	Enabled       *bool  `json:"enabled"`
	CheckSSL      *bool  `json:"check_ssl"`
	SSLAlertDays  int    `json:"ssl_alert_days" binding:"omitempty,min=1"`
}

// Create handles monitor creation
// POST /api/v1/monitors
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

	monitor := &entities.Monitor{
		TenantID:      tenantID,
		Name:          req.Name,
		URL:           req.URL,
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

	c.JSON(http.StatusCreated, monitor)
}

// List handles retrieving all monitors for the current tenant
// GET /api/v1/monitors
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

// GetByID handles retrieving a specific monitor
// GET /api/v1/monitors/:id
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

// Update handles updating a monitor
// PUT /api/v1/monitors/:id
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

	// Update fields if provided
	if req.Name != "" {
		monitor.Name = req.Name
	}
	if req.URL != "" {
		monitor.URL = req.URL
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

	c.JSON(http.StatusOK, monitor)
}

// Delete handles deleting a monitor
// DELETE /api/v1/monitors/:id
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

	if err := h.monitorRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete monitor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "monitor deleted successfully"})
}

// GetChecks handles retrieving check history for a monitor
// GET /api/v1/monitors/:id/checks
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
