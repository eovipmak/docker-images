package handlers

import (
	"net/http"

	"github.com/eovipmak/v-insight/backend/internal/domain/service"
	"github.com/gin-gonic/gin"
)

// MonitorWithStatus represents a monitor with its current status
type MonitorWithStatus struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"` // "up", "down", "unknown"
}

// PublicStatusPageHandler handles public status page requests (no auth required)
type PublicStatusPageHandler struct {
	statusPageService *service.StatusPageService
}
func NewPublicStatusPageHandler(statusPageService *service.StatusPageService) *PublicStatusPageHandler {
	return &PublicStatusPageHandler{
		statusPageService: statusPageService,
	}
}

// GetPublicStatusPage retrieves a public status page by slug
func (h *PublicStatusPageHandler) GetPublicStatusPage(c *gin.Context) {
	slug := c.Param("slug")

	statusPage, monitors, err := h.statusPageService.GetPublicStatusPage(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Status page not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status_page": statusPage,
		"monitors":    monitors,
	})
}