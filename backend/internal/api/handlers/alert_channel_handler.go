package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
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

	// Set defaults
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	channel := &entities.AlertChannel{
		TenantID: tenantID,
		Type:     req.Type,
		Name:     req.Name,
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
