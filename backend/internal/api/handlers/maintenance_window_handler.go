package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/eovipmak/v-insight/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type MaintenanceWindowHandler struct {
	repo repository.MaintenanceWindowRepository
}

func NewMaintenanceWindowHandler(repo repository.MaintenanceWindowRepository) *MaintenanceWindowHandler {
	return &MaintenanceWindowHandler{repo: repo}
}

type CreateMaintenanceWindowRequest struct {
	Name           string    `json:"name" binding:"required"`
	StartTime      time.Time `json:"start_time" binding:"required"`
	EndTime        time.Time `json:"end_time" binding:"required"`
	RepeatInterval int       `json:"repeat_interval"`
	MonitorIDs     []string  `json:"monitor_ids"`
	Tags           []string  `json:"tags"`
}

type UpdateMaintenanceWindowRequest struct {
	Name           string     `json:"name"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	RepeatInterval *int       `json:"repeat_interval"`
	MonitorIDs     []string   `json:"monitor_ids"`
	Tags           []string   `json:"tags"`
}

func (h *MaintenanceWindowHandler) Create(c *gin.Context) {
	var req CreateMaintenanceWindowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	sanitizedName, valid := utils.SanitizeAndValidate(req.Name, 1, 255)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name must be between 1 and 255 characters"})
		return
	}

	if req.EndTime.Before(req.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end time must be after start time"})
		return
	}

	window := &entities.MaintenanceWindow{
		TenantID:       tenantID,
		Name:           sanitizedName,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		RepeatInterval: req.RepeatInterval,
		MonitorIDs:     req.MonitorIDs,
		Tags:           req.Tags,
	}

	if err := h.repo.Create(window); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create maintenance window"})
		return
	}

	c.JSON(http.StatusCreated, window)
}

func (h *MaintenanceWindowHandler) List(c *gin.Context) {
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	windows, err := h.repo.GetByTenantID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve maintenance windows"})
		return
	}

	if windows == nil {
		windows = []*entities.MaintenanceWindow{}
	}

	c.JSON(http.StatusOK, windows)
}

func (h *MaintenanceWindowHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}

	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	window, err := h.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "maintenance window not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve maintenance window"})
		return
	}

	if window.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, window)
}

func (h *MaintenanceWindowHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}

	var req UpdateMaintenanceWindowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	window, err := h.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "maintenance window not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve maintenance window"})
		return
	}

	if window.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if req.Name != "" {
		sanitizedName, valid := utils.SanitizeAndValidate(req.Name, 1, 255)
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name must be between 1 and 255 characters"})
			return
		}
		window.Name = sanitizedName
	}

	if req.StartTime != nil {
		window.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		window.EndTime = *req.EndTime
	}

	if window.EndTime.Before(window.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end time must be after start time"})
		return
	}

	if req.RepeatInterval != nil {
		window.RepeatInterval = *req.RepeatInterval
	}
	if req.MonitorIDs != nil {
		window.MonitorIDs = req.MonitorIDs
	}
	if req.Tags != nil {
		window.Tags = req.Tags
	}

	if err := h.repo.Update(window); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update maintenance window"})
		return
	}

	c.JSON(http.StatusOK, window)
}

func (h *MaintenanceWindowHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}

	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	window, err := h.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "maintenance window not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve maintenance window"})
		return
	}

	if window.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete maintenance window"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "maintenance window deleted successfully"})
}
