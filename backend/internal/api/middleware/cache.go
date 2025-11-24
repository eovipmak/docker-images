package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CacheConfig holds configuration for cache middleware
type CacheConfig struct {
	// DefaultMaxAge is the default max-age for Cache-Control header (in seconds)
	DefaultMaxAge int
	// HealthMaxAge is the max-age for health check endpoints (in seconds)
	HealthMaxAge int
	// APIMaxAge is the max-age for API data endpoints (in seconds)
	APIMaxAge int
}

// DefaultCacheConfig returns the default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		DefaultMaxAge: 0,  // No caching by default
		HealthMaxAge:  60, // 1 minute for health endpoints
		APIMaxAge:     30, // 30 seconds for API data
	}
}

// CacheHeaders returns a middleware that adds appropriate cache control headers
func CacheHeaders() gin.HandlerFunc {
	return CacheHeadersWithConfig(DefaultCacheConfig())
}

// CacheHeadersWithConfig returns a cache middleware with custom configuration
func CacheHeadersWithConfig(config CacheConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply caching to GET requests
		if c.Request.Method != http.MethodGet {
			c.Header("Cache-Control", "no-store")
			c.Next()
			return
		}

		path := c.Request.URL.Path

		// Skip caching for SSE streams
		if strings.Contains(path, "/stream/") {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Next()
			return
		}

		// Determine cache duration based on path
		var maxAge int
		var cacheType string

		switch {
		case path == "/health":
			maxAge = config.HealthMaxAge
			cacheType = "public"
		case strings.HasPrefix(path, "/api/v1/monitors"),
			strings.HasPrefix(path, "/api/v1/dashboard"),
			strings.HasPrefix(path, "/api/v1/incidents"),
			strings.HasPrefix(path, "/api/v1/alert-rules"),
			strings.HasPrefix(path, "/api/v1/alert-channels"):
			maxAge = config.APIMaxAge
			cacheType = "private" // API responses may contain user-specific data
		default:
			maxAge = config.DefaultMaxAge
			cacheType = "private"
		}

		// Process the request first
		c.Next()

		// Only add cache headers for successful responses
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			if maxAge > 0 {
				c.Header("Cache-Control", fmt.Sprintf("%s, max-age=%d", cacheType, maxAge))
			} else {
				c.Header("Cache-Control", "no-cache")
			}
		}
	}
}

// ETagMiddleware adds ETag support for conditional requests
func ETagMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Skip for SSE streams
		if strings.Contains(c.Request.URL.Path, "/stream/") {
			c.Next()
			return
		}

		// Process the request with a response recorder
		c.Next()

		// We can't easily generate ETags without buffering the response
		// For simplicity, we'll generate an ETag based on the response metadata
		// In production, you might want to use a more sophisticated approach

		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// Generate a weak ETag based on content length and last state
			etag := generateETag(c.Request.URL.Path, c.Writer.Size())

			// Check if client sent If-None-Match
			ifNoneMatch := c.GetHeader("If-None-Match")
			if ifNoneMatch != "" && ifNoneMatch == etag {
				c.Status(http.StatusNotModified)
				return
			}

			c.Header("ETag", etag)
		}
	}
}

// generateETag generates a weak ETag based on path and content size
func generateETag(path string, size int) string {
	data := fmt.Sprintf("%s:%d", path, size)
	hash := md5.Sum([]byte(data))
	return `W/"` + hex.EncodeToString(hash[:8]) + `"`
}
