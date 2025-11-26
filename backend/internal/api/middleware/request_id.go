package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

// RequestID generates a unique request ID for each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists in header
		requestID := c.GetHeader(RequestIDHeader)
		
		// Generate new UUID if not provided
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// Set request ID in context for logging
		c.Set("request_id", requestID)
		
		// Add request ID to response header
		c.Header(RequestIDHeader, requestID)
		
		c.Next()
	}
}
