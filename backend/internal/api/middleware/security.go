package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

// SecurityConfig holds security header configuration
type SecurityConfig struct {
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
}

// SecurityHeaders adds security headers to all responses
func SecurityHeaders(cfg SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// HTTP Strict Transport Security
		if cfg.HSTSMaxAge > 0 {
			hstsValue := fmt.Sprintf("max-age=%d", cfg.HSTSMaxAge)
			if cfg.HSTSIncludeSubdomains {
				hstsValue += "; includeSubDomains"
			}
			c.Header("Strict-Transport-Security", hstsValue)
		}
		
		c.Next()
	}
}
