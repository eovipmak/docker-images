package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
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

// List handles listing incidents with filters
// GET /api/v1/incidents?status=&monitor_id=&from=&to=&limit=&offset=
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

// GetByID handles retrieving a single incident
// GET /api/v1/incidents/:id
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

// Resolve handles manually resolving an incident
// POST /api/v1/incidents/:id/resolve
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
