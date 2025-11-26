package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestSizeLimit limits the size of request bodies
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set max bytes for request body
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}
