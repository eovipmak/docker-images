package handlers

import (
	"net/http"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/eovipmak/v-insight/backend/internal/domain/service"
	"github.com/gin-gonic/gin"
)

// DashboardHandler handles dashboard-related HTTP requests
type DashboardHandler struct {
	monitorRepo    repository.MonitorRepository
	incidentRepo   repository.IncidentRepository
	metricsService *service.MetricsService
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(monitorRepo repository.MonitorRepository, incidentRepo repository.IncidentRepository, metricsService *service.MetricsService) *DashboardHandler {
	return &DashboardHandler{
		monitorRepo:    monitorRepo,
		incidentRepo:   incidentRepo,
		metricsService: metricsService,
	}
}

// DashboardStats represents the dashboard statistics response
type DashboardStats struct {
	TotalMonitors       int     `json:"total_monitors"`
	UpCount             int     `json:"up_count"`
	DownCount           int     `json:"down_count"`
	OpenIncidents       int     `json:"open_incidents"`
	AverageResponseTime float64 `json:"average_response_time"`
	OverallUptime       float64 `json:"overall_uptime"`
}

// MonitorCheckWithMonitor represents a monitor check with associated monitor details
type MonitorCheckWithMonitor struct {
	Check   *entities.MonitorCheck `json:"check"`
	Monitor *entities.Monitor      `json:"monitor"`
}

// IncidentWithDetails represents an incident with associated monitor and alert rule details
type IncidentWithDetails struct {
	Incident *entities.Incident `json:"incident"`
	Monitor  *entities.Monitor  `json:"monitor"`
}

// DashboardData represents the full dashboard response
type DashboardData struct {
	Stats         DashboardStats            `json:"stats"`
	RecentChecks  []MonitorCheckWithMonitor `json:"recent_checks"`
	OpenIncidents []IncidentWithDetails     `json:"open_incidents"`
}

// GetStats godoc
// @Summary Get dashboard statistics
// @Description Get summary statistics for the dashboard including monitor counts, incidents, and uptime
// @Tags Dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} DashboardStats "Dashboard statistics"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /dashboard/stats [get]
func (h *DashboardHandler) GetStats(c *gin.Context) {
	// Get user ID from context (set by middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user context not found"})
		return
	}
	userID := userIDValue.(int)

	// Get all monitors for user
	monitors, err := h.monitorRepo.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get monitors"})
		return
	}

	totalMonitors := len(monitors)
	upCount := 0
	downCount := 0

	// Calculate up/down counts based on latest check
	for _, monitor := range monitors {
		if !monitor.Enabled {
			continue
		}

		checks, err := h.monitorRepo.GetChecksByMonitorID(monitor.ID, 1)
		if err != nil || len(checks) == 0 {
			// No checks yet, don't count as up or down
			continue
		}

		if checks[0].Success {
			upCount++
		} else {
			downCount++
		}
	}

	// Get open incidents count (we need to implement this in repository)
	// For now, we'll get all incidents for all monitors and count open ones
	openIncidentsCount := 0
	for _, monitor := range monitors {
		incidents, err := h.incidentRepo.GetByMonitorID(monitor.ID)
		if err != nil {
			continue
		}
		for _, incident := range incidents {
			if incident.Status == "open" {
				openIncidentsCount++
			}
		}
	}

	stats := DashboardStats{
		TotalMonitors: totalMonitors,
		UpCount:       upCount,
		DownCount:     downCount,
		OpenIncidents: openIncidentsCount,
	}

	// Get global average response time (24h period)
	avgResponseTime, err := h.metricsService.GetGlobalAverageResponseTime(userID, "24h")
	if err == nil {
		stats.AverageResponseTime = avgResponseTime
	}

	// Get overall uptime (24h period)
	uptime, err := h.metricsService.GetGlobalUptime(userID, "24h")
	if err == nil && uptime != nil {
		stats.OverallUptime = uptime.Percentage
	}

	c.JSON(http.StatusOK, stats)
}

// GetDashboard godoc
// @Summary Get complete dashboard data
// @Description Get comprehensive dashboard data including stats, recent checks, and open incidents
// @Tags Dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} DashboardData "Complete dashboard data"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /dashboard [get]
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	// Get user ID from context (set by middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user context not found"})
		return
	}
	userID := userIDValue.(int)

	// Get all monitors for user
	monitors, err := h.monitorRepo.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get monitors"})
		return
	}

	totalMonitors := len(monitors)
	upCount := 0
	downCount := 0

	// Map monitors by ID for quick lookup
	monitorMap := make(map[string]*entities.Monitor)
	for _, monitor := range monitors {
		monitorMap[monitor.ID] = monitor
	}

	// Calculate up/down counts and get recent checks
	recentChecks := []MonitorCheckWithMonitor{}
	checksAdded := 0
	maxRecentChecks := 10

	for _, monitor := range monitors {
		if !monitor.Enabled {
			continue
		}

		checks, err := h.monitorRepo.GetChecksByMonitorID(monitor.ID, 1)
		if err != nil || len(checks) == 0 {
			continue
		}

		// Count up/down
		if checks[0].Success {
			upCount++
		} else {
			downCount++
		}

		// Add to recent checks if we haven't reached the limit
		if checksAdded < maxRecentChecks {
			recentChecks = append(recentChecks, MonitorCheckWithMonitor{
				Check:   checks[0],
				Monitor: monitor,
			})
			checksAdded++
		}
	}

	// Get open incidents with monitor details
	openIncidentsList := []IncidentWithDetails{}
	openIncidentsCount := 0

	for _, monitor := range monitors {
		incidents, err := h.incidentRepo.GetByMonitorID(monitor.ID)
		if err != nil {
			continue
		}
		for _, incident := range incidents {
			if incident.Status == "open" {
				openIncidentsCount++
				openIncidentsList = append(openIncidentsList, IncidentWithDetails{
					Incident: incident,
					Monitor:  monitor,
				})
			}
		}
	}

	stats := DashboardStats{
		TotalMonitors: totalMonitors,
		UpCount:       upCount,
		DownCount:     downCount,
		OpenIncidents: openIncidentsCount,
	}

	// Get global average response time (24h period)
	avgResponseTime, err := h.metricsService.GetGlobalAverageResponseTime(userID, "24h")
	if err == nil {
		stats.AverageResponseTime = avgResponseTime
	}

	// Get overall uptime (24h period)
	uptime, err := h.metricsService.GetGlobalUptime(userID, "24h")
	if err == nil && uptime != nil {
		stats.OverallUptime = uptime.Percentage
	}

	dashboardData := DashboardData{
		Stats:         stats,
		RecentChecks:  recentChecks,
		OpenIncidents: openIncidentsList,
	}

	c.JSON(http.StatusOK, dashboardData)
}
