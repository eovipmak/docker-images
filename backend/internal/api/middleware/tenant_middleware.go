package middleware

import (
	"net/http"
	"strconv"

	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
	"github.com/eovipmak/v-insight/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// TenantMiddleware handles tenant isolation and context management
type TenantMiddleware struct {
	tenantUserRepo repository.TenantUserRepository
}

// NewTenantMiddleware creates a new tenant middleware
func NewTenantMiddleware(tenantUserRepo repository.TenantUserRepository) *TenantMiddleware {
	return &TenantMiddleware{
		tenantUserRepo: tenantUserRepo,
	}
}

// TenantRequired is a middleware that ensures tenant context is set and validates access
// This middleware must be used after AuthRequired middleware
func (m *TenantMiddleware) TenantRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by AuthRequired middleware)
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}
		userID, ok := userIDValue.(int)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID in context"})
			c.Abort()
			return
		}

		// Get tenant ID from context (set by AuthRequired middleware) or X-Tenant-ID header
		var tenantID int
		
		// Check X-Tenant-ID header first (allows tenant switching)
		if tenantIDHeader := c.GetHeader("X-Tenant-ID"); tenantIDHeader != "" {
			var err error
			tenantID, err = strconv.Atoi(tenantIDHeader)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid X-Tenant-ID header"})
				c.Abort()
				return
			}
		} else {
			// Fall back to tenant ID from JWT token
			tenantIDValue, exists := c.Get("tenant_id")
			if !exists {
				c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID not found in context or header"})
				c.Abort()
				return
			}
			var ok bool
			tenantID, ok = tenantIDValue.(int)
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid tenant ID in context"})
				c.Abort()
				return
			}
		}

		// Validate that user has access to the requested tenant
		hasAccess, err := m.tenantUserRepo.HasAccess(userID, tenantID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate tenant access"})
			c.Abort()
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "user does not have access to this tenant"})
			c.Abort()
			return
		}

		// Set user ID and tenant ID in request context using context utilities
		ctx := c.Request.Context()
		ctx = utils.SetUserID(ctx, userID)
		ctx = utils.SetTenantID(ctx, tenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
