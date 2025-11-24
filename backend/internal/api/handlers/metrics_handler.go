package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
	"github.com/eovipmak/v-insight/backend/internal/domain/service"
	"github.com/gin-gonic/gin"
)

// MetricsHandler handles metrics-related HTTP requests
type MetricsHandler struct {
	metricsService *service.MetricsService
	monitorRepo    repository.MonitorRepository
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(metricsService *service.MetricsService, monitorRepo repository.MonitorRepository) *MetricsHandler {
	return &MetricsHandler{
		metricsService: metricsService,
		monitorRepo:    monitorRepo,
	}
}

// GetMonitorMetrics retrieves metrics for a specific monitor
// GET /api/v1/monitors/:id/metrics?period=24h|7d|30d
func (h *MetricsHandler) GetMonitorMetrics(c *gin.Context) {
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

	// Verify monitor ownership
	monitor, err := h.monitorRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "monitor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve monitor"})
		return
	}

	if monitor.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Get period parameter (default to 24h)
	period := c.DefaultQuery("period", "24h")

	// Validate period
	if period != "24h" && period != "7d" && period != "30d" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period (must be 24h, 7d, or 30d)"})
		return
	}

	// Calculate uptime
	uptime, err := h.metricsService.CalculateUptime(id, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate uptime"})
		return
	}

	// Get response time history
	responseTimeHistory, err := h.metricsService.GetResponseTimeHistory(id, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get response time history"})
		return
	}

	// Get status code distribution
	statusCodeDistribution, err := h.metricsService.GetStatusCodeDistribution(id, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get status code distribution"})
		return
	}

	// Get average response time
	avgResponseTime, err := h.metricsService.GetAverageResponseTime(id, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get average response time"})
		return
	}

	// Return empty arrays instead of null
	if responseTimeHistory == nil {
		responseTimeHistory = []service.DataPoint{}
	}
	if statusCodeDistribution == nil {
		statusCodeDistribution = []service.StatusCodeDistribution{}
	}

	c.JSON(http.StatusOK, gin.H{
		"period":                   period,
		"uptime":                   uptime,
		"response_time_history":    responseTimeHistory,
		"status_code_distribution": statusCodeDistribution,
		"average_response_time":    avgResponseTime,
	})
}
