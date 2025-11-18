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
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		userID, tenantID, err := m.authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Set user ID and tenant ID in Gin context for handlers to use
		c.Set("user_id", userID)
		c.Set("tenant_id", tenantID)

		// Also set in request context using context utilities
		ctx := c.Request.Context()
		ctx = utils.SetUserID(ctx, userID)
		ctx = utils.SetTenantID(ctx, tenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
