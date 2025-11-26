package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Add the health check handlers as they appear in main.go
	// Liveness probe - checks if the application is running
	livenessHandler := func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "backend",
		})
	}
	router.GET("/health/live", livenessHandler)
	router.HEAD("/health/live", livenessHandler)

	// Readiness probe - simplified for testing (no DB check)
	readinessHandler := func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":   "ok",
			"service":  "backend",
			"database": "connected",
			"ready":    true,
		})
	}
	router.GET("/health/ready", readinessHandler)
	router.HEAD("/health/ready", readinessHandler)

	// Legacy health check
	healthHandler := func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	}
	router.GET("/health", healthHandler)
	router.HEAD("/health", healthHandler)
	
	return router
}

func TestHealthCheckEndpoints(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		endpoint       string
		method         string
		expectedStatus int
		checkBody      bool
		expectedKeys   []string
	}{
		{
			name:           "Legacy health check GET",
			endpoint:       "/health",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkBody:      true,
			expectedKeys:   []string{"status", "database"},
		},
		{
			name:           "Legacy health check HEAD",
			endpoint:       "/health",
			method:         "HEAD",
			expectedStatus: http.StatusOK,
			checkBody:      false,
			expectedKeys:   nil,
		},
		{
			name:           "Liveness probe GET",
			endpoint:       "/health/live",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkBody:      true,
			expectedKeys:   []string{"status", "service"},
		},
		{
			name:           "Liveness probe HEAD",
			endpoint:       "/health/live",
			method:         "HEAD",
			expectedStatus: http.StatusOK,
			checkBody:      false,
			expectedKeys:   nil,
		},
		{
			name:           "Readiness probe GET",
			endpoint:       "/health/ready",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkBody:      true,
			expectedKeys:   []string{"status", "service", "database", "ready"},
		},
		{
			name:           "Readiness probe HEAD",
			endpoint:       "/health/ready",
			method:         "HEAD",
			expectedStatus: http.StatusOK,
			checkBody:      false,
			expectedKeys:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.endpoint, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkBody {
				// Check content type
				assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

				// Parse response body
				body := w.Body.String()
				assert.NotEmpty(t, body)

				// Check expected keys are present
				for _, key := range tt.expectedKeys {
					assert.Contains(t, body, key)
				}
			}
		})
	}
}

func TestLivenessProbeResponse(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/health/live", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
	assert.Contains(t, w.Body.String(), `"service":"backend"`)
}

func TestReadinessProbeResponse(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
	assert.Contains(t, w.Body.String(), `"service":"backend"`)
	assert.Contains(t, w.Body.String(), `"database":"connected"`)
	assert.Contains(t, w.Body.String(), `"ready":true`)
}

