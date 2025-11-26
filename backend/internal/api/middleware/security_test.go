package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("adds all security headers", func(t *testing.T) {
		router := gin.New()
		router.Use(SecurityHeaders(SecurityConfig{
			HSTSMaxAge:            31536000,
			HSTSIncludeSubdomains: true,
		}))
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
	})

	t.Run("adds HSTS without includeSubdomains", func(t *testing.T) {
		router := gin.New()
		router.Use(SecurityHeaders(SecurityConfig{
			HSTSMaxAge:            31536000,
			HSTSIncludeSubdomains: false,
		}))
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "max-age=31536000", w.Header().Get("Strict-Transport-Security"))
	})

	t.Run("skips HSTS when max age is 0", func(t *testing.T) {
		router := gin.New()
		router.Use(SecurityHeaders(SecurityConfig{
			HSTSMaxAge:            0,
			HSTSIncludeSubdomains: false,
		}))
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Empty(t, w.Header().Get("Strict-Transport-Security"))
		// Other headers should still be present
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	})
}
