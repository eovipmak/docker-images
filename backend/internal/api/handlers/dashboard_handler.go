package handlers

import (
	"database/sql"
	"net/http"

	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
	"github.com/gin-gonic/gin"
)

// DashboardHandler handles dashboard-related HTTP requests
type DashboardHandler struct {
	monitorRepo  repository.MonitorRepository
	incidentRepo repository.IncidentRepository
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(monitorRepo repository.MonitorRepository, incidentRepo repository.IncidentRepository) *DashboardHandler {
	return &DashboardHandler{
		monitorRepo:  monitorRepo,
		incidentRepo: incidentRepo,
	}
}

// DashboardStats represents the dashboard statistics response
type DashboardStats struct {
	TotalMonitors int `json:"total_monitors"`
	UpCount       int `json:"up_count"`
	DownCount     int `json:"down_count"`
	OpenIncidents int `json:"open_incidents"`
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

// GetStats returns dashboard statistics
// GET /api/v1/dashboard/stats
func (h *DashboardHandler) GetStats(c *gin.Context) {
	// Get tenant ID from context (set by middleware)
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	// Get all monitors for tenant
	monitors, err := h.monitorRepo.GetByTenantID(tenantID)
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

	c.JSON(http.StatusOK, stats)
}

// GetDashboard returns complete dashboard data including stats, recent checks, and open incidents
// GET /api/v1/dashboard
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	// Get tenant ID from context (set by middleware)
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	// Get all monitors for tenant
	monitors, err := h.monitorRepo.GetByTenantID(tenantID)
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

	dashboardData := DashboardData{
		Stats:         stats,
		RecentChecks:  recentChecks,
		OpenIncidents: openIncidentsList,
	}

	c.JSON(http.StatusOK, dashboardData)
}
