package middleware

import (
	"time"

	"github.com/eovipmak/v-insight/backend/internal"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger logs all HTTP requests with structured logging
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get request ID from context
		requestID, _ := c.Get("request_id")

		// Get tenant ID and user ID if available
		tenantID, _ := c.Get("tenant_id")
		userID, _ := c.Get("user_id")

		// Build log fields
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
		}

		// Add request ID if present
		if reqID, ok := requestID.(string); ok {
			fields = append(fields, zap.String("request_id", reqID))
		}

		// Add tenant ID if present
		if tID, ok := tenantID.(int); ok {
			fields = append(fields, zap.Int("tenant_id", tID))
		}

		// Add user ID if present
		if uID, ok := userID.(int); ok {
			fields = append(fields, zap.Int("user_id", uID))
		}

		// Record metrics
		status := c.Writer.Status()
		internal.HTTPRequestTotal.WithLabelValues(
			c.Request.Method,
			c.Request.URL.Path,
			string(rune(status/100))+"xx",
		).Inc()

		internal.HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			c.Request.URL.Path,
			string(rune(status/100))+"xx",
		).Observe(duration.Seconds())

		// Log at appropriate level based on status code
		if status >= 500 {
			if internal.Log != nil {
				internal.Log.Error("HTTP request", fields...)
			}
		} else if status >= 400 {
			if internal.Log != nil {
				internal.Log.Warn("HTTP request", fields...)
			}
		} else {
			if internal.Log != nil {
				internal.Log.Info("HTTP request", fields...)
			}
		}
	}
}
