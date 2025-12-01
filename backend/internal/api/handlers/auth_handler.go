package handlers

import (
	"net/http"

	"github.com/eovipmak/v-insight/backend/internal"
	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/eovipmak/v-insight/backend/internal/domain/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *service.AuthService
	userRepo    repository.UserRepository
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *service.AuthService, userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userRepo:    userRepo,
	}
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	TenantName string `json:"tenant_name" binding:"required"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string `json:"token"`
}

// Register godoc
// @Summary Register a new user and tenant
// @Description Register a new user account and create an associated tenant organization
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} AuthResponse "Successfully registered, returns JWT token"
// @Failure 400 {object} map[string]string "Invalid request or user already exists"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if internal.Log != nil {
		internal.Log.Info("User registration attempt",
			zap.String("email", req.Email),
			zap.String("tenant_name", req.TenantName),
		)
	}

	token, err := h.authService.Register(req.Email, req.Password, req.TenantName)
	if err != nil {
		if internal.Log != nil {
			internal.Log.Warn("User registration failed",
				zap.String("email", req.Email),
				zap.Error(err),
			)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if internal.Log != nil {
		internal.Log.Info("User registration successful",
			zap.String("email", req.Email),
		)
	}

	c.JSON(http.StatusCreated, AuthResponse{Token: token})
}

// Login godoc
// @Summary Login to get JWT token
// @Description Authenticate with email and password to receive a JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse "Successfully logged in, returns JWT token"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if internal.Log != nil {
		internal.Log.Info("User login attempt",
			zap.String("email", req.Email),
		)
	}

	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if internal.Log != nil {
			internal.Log.Warn("User login failed",
				zap.String("email", req.Email),
				zap.Error(err),
			)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if internal.Log != nil {
		internal.Log.Info("User login successful",
			zap.String("email", req.Email),
		)
	}

	c.JSON(http.StatusOK, AuthResponse{Token: token})
}

// Me godoc
// @Summary Get current user info
// @Description Get the authenticated user's information
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User information including id, email, and tenant_id"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get user from database
	user, err := h.userRepo.GetByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Also get tenant_id from context if available
	tenantID, _ := c.Get("tenant_id")

	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID,
		"email":     user.Email,
		"tenant_id": tenantID,
	})
}
