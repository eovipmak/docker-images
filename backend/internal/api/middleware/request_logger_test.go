package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/eovipmak/v-insight/backend/internal"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestLogger(t *testing.T) {
	// Initialize logger for tests
	_ = internal.InitLogger("development")
	defer internal.SyncLogger()

	gin.SetMode(gin.TestMode)

	t.Run("logs successful request", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestID())
		router.Use(RequestLogger())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("logs request with user and role context", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestID())
		router.Use(RequestLogger())
		router.GET("/test", func(c *gin.Context) {
			c.Set("user_id", 456)
			c.Set("role", "admin")
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("logs request with error status", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestID())
		router.Use(RequestLogger())
		router.GET("/test", func(c *gin.Context) {
			c.String(404, "Not Found")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})

	t.Run("logs request with server error status", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestID())
		router.Use(RequestLogger())
		router.GET("/test", func(c *gin.Context) {
			c.String(500, "Internal Server Error")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
	})
}
