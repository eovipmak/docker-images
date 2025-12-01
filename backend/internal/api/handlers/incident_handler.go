package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/gin-gonic/gin"
)

// IncidentHandler handles incident-related HTTP requests
type IncidentHandler struct {
	incidentRepo     repository.IncidentRepository
	monitorRepo      repository.MonitorRepository
	alertRuleRepo    repository.AlertRuleRepository
	alertChannelRepo repository.AlertChannelRepository
}

// NewIncidentHandler creates a new incident handler
func NewIncidentHandler(incidentRepo repository.IncidentRepository, monitorRepo repository.MonitorRepository, alertRuleRepo repository.AlertRuleRepository, alertChannelRepo repository.AlertChannelRepository) *IncidentHandler {
	return &IncidentHandler{
		incidentRepo:     incidentRepo,
		monitorRepo:      monitorRepo,
		alertRuleRepo:    alertRuleRepo,
		alertChannelRepo: alertChannelRepo,
	}
}

// IncidentResponse represents an incident with related details
type IncidentResponse struct {
	entities.Incident
	MonitorName    string `json:"monitor_name"`
	MonitorURL     string `json:"monitor_url"`
	AlertRuleName  string `json:"alert_rule_name"`
	Channels       []ChannelInfo `json:"channels,omitempty"`
	Duration       *int64 `json:"duration,omitempty"` // Duration in seconds
}

// ChannelInfo represents channel information for notifications
type ChannelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// List godoc
// @Summary List incidents with filters
// @Description Get a list of incidents filtered by status, monitor, and date range
// @Tags Incidents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (open or resolved)"
// @Param monitor_id query string false "Filter by monitor ID"
// @Param from query string false "Filter by start date (RFC3339 format)"
// @Param to query string false "Filter by end date (RFC3339 format)"
// @Param limit query int false "Maximum number of results (default 50, max 100)"
// @Param offset query int false "Offset for pagination (default 0)"
// @Success 200 {array} IncidentResponse "List of incidents with details"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /incidents [get]
func (h *IncidentHandler) List(c *gin.Context) {
	// Get tenant ID from context (set by middleware)
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	// Parse query parameters
	filters := repository.IncidentFilters{
		TenantID: tenantID,
		Status:   c.Query("status"),
		MonitorID: c.Query("monitor_id"),
		Limit:    50, // Default limit
		Offset:   0,
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 || limit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
			return
		}
		filters.Limit = limit
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset parameter"})
			return
		}
		filters.Offset = offset
	}

	// Parse from date
	if fromStr := c.Query("from"); fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date format, use RFC3339"})
			return
		}
		filters.From = &from
	}

	// Parse to date
	if toStr := c.Query("to"); toStr != "" {
		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date format, use RFC3339"})
			return
		}
		filters.To = &to
	}

	// Validate status if provided
	if filters.Status != "" && filters.Status != "open" && filters.Status != "resolved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be 'open' or 'resolved'"})
		return
	}

	// Get incidents
	incidents, err := h.incidentRepo.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list incidents"})
		return
	}

	// Build response with monitor and alert rule details
	response := make([]IncidentResponse, 0, len(incidents))
	for _, incident := range incidents {
		incidentResp := IncidentResponse{
			Incident: *incident,
		}

		// Get monitor details
		monitor, err := h.monitorRepo.GetByID(incident.MonitorID)
		if err == nil {
			incidentResp.MonitorName = monitor.Name
			incidentResp.MonitorURL = monitor.URL

			// Get alert rule details (requires tenant ID)
			alertRule, err := h.alertRuleRepo.GetByID(monitor.TenantID, incident.AlertRuleID)
			if err == nil {
				incidentResp.AlertRuleName = alertRule.Name
			}
		}

		// Calculate duration
		if incident.ResolvedAt.Valid {
			duration := int64(incident.ResolvedAt.Time.Sub(incident.StartedAt).Seconds())
			incidentResp.Duration = &duration
		} else {
			// For ongoing incidents, calculate duration from start to now
			duration := int64(time.Since(incident.StartedAt).Seconds())
			incidentResp.Duration = &duration
		}

		response = append(response, incidentResp)
	}

	c.JSON(http.StatusOK, response)
}

// GetByID godoc
// @Summary Get an incident by ID
// @Description Get detailed information about a specific incident including monitor and alert rule details
// @Tags Incidents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Incident ID"
// @Success 200 {object} IncidentResponse "Incident details"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Incident not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /incidents/{id} [get]
func (h *IncidentHandler) GetByID(c *gin.Context) {
	// Get tenant ID from context (set by middleware)
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	id := c.Param("id")

	// Get incident
	incident, err := h.incidentRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "incident not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get incident"})
		return
	}

	// Verify tenant ownership through monitor
	monitor, err := h.monitorRepo.GetByID(incident.MonitorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify access"})
		return
	}

	if monitor.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Build response with details
	response := IncidentResponse{
		Incident:    *incident,
		MonitorName: monitor.Name,
		MonitorURL:  monitor.URL,
	}

	// Get alert rule details
	alertRule, err := h.alertRuleRepo.GetByID(monitor.TenantID, incident.AlertRuleID)
	if err == nil {
		response.AlertRuleName = alertRule.Name
		
		// Get associated channels
		channels, err := h.alertChannelRepo.GetByAlertRuleID(monitor.TenantID, incident.AlertRuleID)
		if err == nil {
			response.Channels = make([]ChannelInfo, len(channels))
			for i, ch := range channels {
				response.Channels[i] = ChannelInfo{
					ID:   ch.ID,
					Name: ch.Name,
					Type: ch.Type,
				}
			}
		}
	}

	// Calculate duration
	if incident.ResolvedAt.Valid {
		duration := int64(incident.ResolvedAt.Time.Sub(incident.StartedAt).Seconds())
		response.Duration = &duration
	} else {
		duration := int64(time.Since(incident.StartedAt).Seconds())
		response.Duration = &duration
	}

	c.JSON(http.StatusOK, response)
}

// Resolve godoc
// @Summary Manually resolve an incident
// @Description Mark an incident as resolved manually
// @Tags Incidents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Incident ID"
// @Success 200 {object} map[string]string "Incident resolved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Incident not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /incidents/{id}/resolve [post]
func (h *IncidentHandler) Resolve(c *gin.Context) {
	// Get tenant ID from context (set by middleware)
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := tenantIDValue.(int)

	id := c.Param("id")

	// Get incident to verify ownership
	incident, err := h.incidentRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "incident not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get incident"})
		return
	}

	// Verify tenant ownership through monitor
	monitor, err := h.monitorRepo.GetByID(incident.MonitorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify access"})
		return
	}

	if monitor.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Resolve the incident
	if err := h.incidentRepo.Resolve(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "incident resolved successfully"})
}
