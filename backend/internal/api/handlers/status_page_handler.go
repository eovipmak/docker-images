package handlers

import (
	"net/http"

	"github.com/eovipmak/v-insight/backend/internal/domain/service"
	"github.com/gin-gonic/gin"
)

// StatusPageHandler handles status page-related HTTP requests
type StatusPageHandler struct {
	statusPageService *service.StatusPageService
}

// NewStatusPageHandler creates a new status page handler
func NewStatusPageHandler(statusPageService *service.StatusPageService) *StatusPageHandler {
	return &StatusPageHandler{
		statusPageService: statusPageService,
	}
}

// CreateStatusPageRequest represents the request body for creating a status page
type CreateStatusPageRequest struct {
	Slug          string `json:"slug" binding:"required"`
	Name          string `json:"name" binding:"required"`
	PublicEnabled *bool  `json:"public_enabled"` // pointer to allow explicit false
}

// UpdateStatusPageRequest represents the request body for updating a status page
type UpdateStatusPageRequest struct {
	Slug          string `json:"slug" binding:"required"`
	Name          string `json:"name" binding:"required"`
	PublicEnabled *bool  `json:"public_enabled"` // pointer to allow explicit false
}

// CreateStatusPage creates a new status page
func (h *StatusPageHandler) CreateStatusPage(c *gin.Context) {
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	tenantID := tenantIDValue.(int)

	var req CreateStatusPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	publicEnabled := false
	if req.PublicEnabled != nil {
		publicEnabled = *req.PublicEnabled
	}

	statusPage, err := h.statusPageService.CreateStatusPage(tenantID, req.Slug, req.Name, publicEnabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, statusPage)
}

// GetStatusPages retrieves all status pages for the tenant
func (h *StatusPageHandler) GetStatusPages(c *gin.Context) {
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	tenantID := tenantIDValue.(int)

	statusPages, err := h.statusPageService.GetStatusPagesByTenant(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status_pages": statusPages})
}

// GetStatusPage retrieves a specific status page
func (h *StatusPageHandler) GetStatusPage(c *gin.Context) {
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	tenantID := tenantIDValue.(int)

	id := c.Param("id")

	statusPage, err := h.statusPageService.GetStatusPageByID(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Status page not found"})
		return
	}

	monitors, err := h.statusPageService.GetStatusPageMonitors(id, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status_page": statusPage,
		"monitors":    monitors,
	})
}

// UpdateStatusPage updates an existing status page
func (h *StatusPageHandler) UpdateStatusPage(c *gin.Context) {
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	tenantID := tenantIDValue.(int)

	id := c.Param("id")

	var req UpdateStatusPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	publicEnabled := false
	if req.PublicEnabled != nil {
		publicEnabled = *req.PublicEnabled
	}

	statusPage, err := h.statusPageService.UpdateStatusPage(id, tenantID, req.Slug, req.Name, publicEnabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, statusPage)
}

// DeleteStatusPage deletes a status page
func (h *StatusPageHandler) DeleteStatusPage(c *gin.Context) {
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	tenantID := tenantIDValue.(int)

	id := c.Param("id")

	err := h.statusPageService.DeleteStatusPage(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Status page not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status page deleted"})
}

// GetStatusPageMonitors retrieves monitors for a status page
func (h *StatusPageHandler) GetStatusPageMonitors(c *gin.Context) {
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	tenantID := tenantIDValue.(int)

	statusPageID := c.Param("id")

	monitors, err := h.statusPageService.GetStatusPageMonitors(statusPageID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"monitors": monitors})
}

// AddMonitorToStatusPage adds a monitor to a status page
func (h *StatusPageHandler) AddMonitorToStatusPage(c *gin.Context) {
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	tenantID := tenantIDValue.(int)

	statusPageID := c.Param("id")
	monitorID := c.Param("monitor_id")

	err := h.statusPageService.AddMonitorToStatusPage(statusPageID, monitorID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Monitor added to status page"})
}

// RemoveMonitorFromStatusPage removes a monitor from a status page
func (h *StatusPageHandler) RemoveMonitorFromStatusPage(c *gin.Context) {
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	tenantID := tenantIDValue.(int)

	statusPageID := c.Param("id")
	monitorID := c.Param("monitor_id")

	err := h.statusPageService.RemoveMonitorFromStatusPage(statusPageID, monitorID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Monitor removed from status page"})
}