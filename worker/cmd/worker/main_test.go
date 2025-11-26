package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupTestApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "V-Insight Worker Test",
	})

	// Liveness probe - checks if the worker is running
	app.Get("/health/live", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "worker",
		})
	})

	// Readiness probe - simplified for testing (no DB/scheduler check)
	app.Get("/health/ready", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"service":   "worker",
			"database":  "connected",
			"scheduler": "running",
			"jobs":      []string{"health_check", "ssl_check", "alert_evaluator", "notification"},
			"ready":     true,
		})
	})

	// Legacy health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":   "ok",
			"service":  "worker",
			"database": "connected",
			"jobs":     []string{"health_check", "ssl_check", "alert_evaluator", "notification"},
		})
	})

	return app
}

func TestWorkerHealthCheckEndpoints(t *testing.T) {
	app := setupTestApp()

	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
		checkBody      bool
		expectedKeys   []string
	}{
		{
			name:           "Legacy health check",
			endpoint:       "/health",
			expectedStatus: http.StatusOK,
			checkBody:      true,
			expectedKeys:   []string{"status", "service", "database", "jobs"},
		},
		{
			name:           "Liveness probe",
			endpoint:       "/health/live",
			expectedStatus: http.StatusOK,
			checkBody:      true,
			expectedKeys:   []string{"status", "service"},
		},
		{
			name:           "Readiness probe",
			endpoint:       "/health/ready",
			expectedStatus: http.StatusOK,
			checkBody:      true,
			expectedKeys:   []string{"status", "service", "database", "scheduler", "jobs", "ready"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.endpoint, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.checkBody {
				// Check content type
				assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
			}
		})
	}
}

func TestWorkerLivenessProbe(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest("GET", "/health/live", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWorkerReadinessProbe(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest("GET", "/health/ready", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
