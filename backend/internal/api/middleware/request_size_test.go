package middleware

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestSizeLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("allows requests within size limit", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestSizeLimit(1024)) // 1KB limit
		router.POST("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		// Send a small request (100 bytes)
		body := bytes.Repeat([]byte("a"), 100)
		req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("blocks requests exceeding size limit", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestSizeLimit(1024)) // 1KB limit
		router.POST("/test", func(c *gin.Context) {
			// Try to read the body
			_, err := c.GetRawData()
			if err != nil {
				c.String(413, "Request too large")
				return
			}
			c.String(200, "OK")
		})

		// Send a large request (2KB)
		body := bytes.Repeat([]byte("a"), 2048)
		req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should get 413 or error due to size limit
		assert.NotEqual(t, 200, w.Code)
	})
}
