package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/eovipmak/v-insight/backend/internal"
	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
	"github.com/eovipmak/v-insight/backend/internal/domain/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

// GetMonitorMetrics godoc
// @Summary Get monitor metrics
// @Description Retrieve detailed metrics for a specific monitor including uptime, response times, and status codes
// @Tags Metrics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Monitor ID"
// @Param period query string false "Time period for metrics (24h, 7d, or 30d)" default(24h)
// @Success 200 {object} map[string]interface{} "Monitor metrics including uptime, response time history, and status code distribution"
// @Failure 400 {object} map[string]string "Invalid request or period parameter"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Monitor not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /monitors/{id}/metrics [get]
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
	validPeriods := map[string]bool{"1h": true, "6h": true, "12h": true, "24h": true, "1w": true, "7d": true, "30d": true}
	if !validPeriods[period] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period (must be one of 1h, 6h, 12h, 24h, 1w/7d, or 30d)"})
		return
	}

	// Calculate uptime
	uptime, err := h.metricsService.CalculateUptime(id, period)
	if err != nil {
		internal.Log.Error("failed to calculate uptime", zap.Error(err), zap.String("monitor_id", id), zap.String("period", period))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate uptime"})
		return
	}

	// Get response time history
	responseTimeHistory, err := h.metricsService.GetResponseTimeHistory(id, period)
	if err != nil {
		internal.Log.Error("failed to get response time history", zap.Error(err), zap.String("monitor_id", id), zap.String("period", period))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get response time history"})
		return
	}

	// Get status code distribution
	statusCodeDistribution, err := h.metricsService.GetStatusCodeDistribution(id, period)
	if err != nil {
		internal.Log.Error("failed to get status code distribution", zap.Error(err), zap.String("monitor_id", id), zap.String("period", period))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get status code distribution"})
		return
	}

	// Get average response time
	avgResponseTime, err := h.metricsService.GetAverageResponseTime(id, period)
	if err != nil {
		internal.Log.Error("failed to get average response time", zap.Error(err), zap.String("monitor_id", id), zap.String("period", period))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get average response time"})
		return
	}

	// Ensure uptime is never null (return 0-values if no checks)
	if uptime == nil {
		uptime = &service.UptimeMetrics{Percentage: 0, TotalChecks: 0, SuccessChecks: 0, FailedChecks: 0}
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
