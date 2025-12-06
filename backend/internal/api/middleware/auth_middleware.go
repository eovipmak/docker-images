package middleware

import (
	"net/http"
	"strings"

	"github.com/eovipmak/v-insight/backend/internal/domain/service"
	"github.com/eovipmak/v-insight/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	authService *service.AuthService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// AuthRequired is a middleware that validates JWT tokens
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header or query parameter (for SSE)
		var token string
		authHeader := c.GetHeader("Authorization")
		
		if authHeader != "" {
			// Check if it's a Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
				c.Abort()
				return
			}
			token = parts[1]
		} else {
			// Fallback to query parameter for SSE (EventSource can't set headers)
			token = c.Query("token")
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization token required"})
				c.Abort()
				return
			}
		}

		// Validate token
		userID, role, err := m.authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Set user ID and role in Gin context for handlers to use
		c.Set("user_id", userID)
		c.Set("role", role)

		// Also set in request context using context utilities
		ctx := c.Request.Context()
		ctx = utils.SetUserID(ctx, userID)
		ctx = utils.SetRole(ctx, role)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// AdminRequired is a middleware that ensures the user has admin role
func (m *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "role not found in context"})
			c.Abort()
			return
		}

		if role.(string) != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
