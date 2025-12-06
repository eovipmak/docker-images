package handlers

import (
	"net/http"
	"strconv"

	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/gin-gonic/gin"
)

// AdminHandler handles admin-related HTTP requests
type AdminHandler struct {
	userRepo      repository.UserRepository
	monitorRepo   repository.MonitorRepository
	alertRuleRepo repository.AlertRuleRepository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(userRepo repository.UserRepository, monitorRepo repository.MonitorRepository, alertRuleRepo repository.AlertRuleRepository) *AdminHandler {
	return &AdminHandler{
		userRepo:      userRepo,
		monitorRepo:   monitorRepo,
		alertRuleRepo: alertRuleRepo,
	}
}

// ListUsers godoc
// @Summary List all users
// @Description Get a list of all registered users (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entities.User "List of users"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.userRepo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user by their ID (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]string "User deleted successfully"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	err = h.userRepo.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// ListMonitors godoc
// @Summary List all monitors
// @Description Get a list of all monitors across all users (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entities.Monitor "List of monitors"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/monitors [get]
func (h *AdminHandler) ListMonitors(c *gin.Context) {
	monitors, err := h.monitorRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get all monitors"})
		return
	}
	c.JSON(http.StatusOK, monitors)
}

// ListAlertRules godoc
// @Summary List all alert rules
// @Description Get a list of all alert rules across all users (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entities.AlertRule "List of alert rules"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/alert-rules [get]
func (h *AdminHandler) ListAlertRules(c *gin.Context) {
	rules, err := h.alertRuleRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get all alert rules"})
		return
	}
	c.JSON(http.StatusOK, rules)
}
