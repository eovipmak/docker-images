package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eovipmak/v-insight/backend/internal/api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("security headers are applied", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.SecurityHeaders(middleware.SecurityConfig{
			HSTSMaxAge:            31536000,
			HSTSIncludeSubdomains: true,
		}))
		router.Use(middleware.RequestID())

		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	})

	t.Run("rate limiter blocks excessive requests", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
			PerIP:   2, // Very low limit for testing
			PerUser: 100,
		}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// First two requests should succeed
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.100:12345"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		}

		// Third request should be rate limited
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 429, w.Code)
		assert.Contains(t, w.Body.String(), "rate limit exceeded")
	})

	t.Run("request size limit works", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.RequestSizeLimit(1024)) // 1KB limit

		router.POST("/test", func(c *gin.Context) {
			_, err := c.GetRawData()
			if err != nil {
				c.JSON(413, gin.H{"error": "request too large"})
				return
			}
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Small request should succeed
		smallBody := bytes.Repeat([]byte("a"), 100)
		req1 := httptest.NewRequest("POST", "/test", bytes.NewReader(smallBody))
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, 200, w1.Code)

		// Large request should fail
		largeBody := bytes.Repeat([]byte("a"), 2048)
		req2 := httptest.NewRequest("POST", "/test", bytes.NewReader(largeBody))
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		assert.NotEqual(t, 200, w2.Code)
	})
}
