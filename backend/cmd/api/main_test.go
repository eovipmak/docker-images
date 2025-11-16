package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                string
		env                 string
		corsOrigins         string
		requestOrigin       string
		requestHost         string
		expectHeaderSet     bool
		expectedAllowOrigin string
		expectedCredentials string
	}{
		{
			name:                "Development mode with wildcard allows any origin",
			env:                 "development",
			corsOrigins:         "*",
			requestOrigin:       "http://localhost:3000",
			requestHost:         "localhost:8080",
			expectHeaderSet:     true,
			expectedAllowOrigin: "*",
			expectedCredentials: "",
		},
		{
			name:                "Development mode with wildcard allows different origin",
			env:                 "development",
			corsOrigins:         "*",
			requestOrigin:       "http://example.com:8080",
			requestHost:         "localhost:8080",
			expectHeaderSet:     true,
			expectedAllowOrigin: "*",
			expectedCredentials: "",
		},
		{
			name:                "Development mode with empty CORS (defaults to wildcard)",
			env:                 "development",
			corsOrigins:         "",
			requestOrigin:       "http://example.com:8080",
			requestHost:         "api.server.com",
			expectHeaderSet:     true,
			expectedAllowOrigin: "*",
			expectedCredentials: "",
		},
		{
			name:                "Development mode with specific origins - matching origin",
			env:                 "development",
			corsOrigins:         "http://localhost:3000,http://127.0.0.1:3000",
			requestOrigin:       "http://localhost:3000",
			requestHost:         "localhost:8080",
			expectHeaderSet:     true,
			expectedAllowOrigin: "http://localhost:3000",
			expectedCredentials: "true",
		},
		{
			name:                "Development mode with specific origins - non-matching origin",
			env:                 "development",
			corsOrigins:         "http://localhost:3000,http://127.0.0.1:3000",
			requestOrigin:       "http://evil.com",
			requestHost:         "localhost:8080",
			expectHeaderSet:     false,
			expectedAllowOrigin: "",
			expectedCredentials: "",
		},
		{
			name:                "Production mode with specific origins - matching",
			env:                 "production",
			corsOrigins:         "https://example.com",
			requestOrigin:       "https://example.com",
			requestHost:         "api.example.com",
			expectHeaderSet:     true,
			expectedAllowOrigin: "https://example.com",
			expectedCredentials: "true",
		},
		{
			name:                "Production mode with wildcard (falls back to default origins)",
			env:                 "production",
			corsOrigins:         "*",
			requestOrigin:       "http://localhost:3000",
			requestHost:         "localhost:8080",
			expectHeaderSet:     true,
			expectedAllowOrigin: "http://localhost:3000",
			expectedCredentials: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("ENV", tt.env)
			os.Setenv("CORS_ALLOWED_ORIGINS", tt.corsOrigins)

			// Create router with CORS middleware
			router := gin.New()
			router.Use(setupCORSMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"status": "ok"})
			})

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", tt.requestOrigin)
			req.Host = tt.requestHost
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Check if request was aborted (forbidden)
			if !tt.expectHeaderSet {
				if w.Code == http.StatusForbidden {
					// Expected forbidden response for non-matching origin
					return
				}
			}

			// Check Allow-Origin header
			allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
			if tt.expectHeaderSet {
				if allowOrigin != tt.expectedAllowOrigin {
					t.Errorf("Expected Access-Control-Allow-Origin: %s, got: %s", tt.expectedAllowOrigin, allowOrigin)
				}

				// Check Allow-Credentials header
				allowCredentials := w.Header().Get("Access-Control-Allow-Credentials")
				if allowCredentials != tt.expectedCredentials {
					t.Errorf("Expected Access-Control-Allow-Credentials: %s, got: %s", tt.expectedCredentials, allowCredentials)
				}
			} else {
				if allowOrigin != "" {
					t.Errorf("Expected no Access-Control-Allow-Origin header, got: %s", allowOrigin)
				}
			}

			// Cleanup
			os.Unsetenv("ENV")
			os.Unsetenv("CORS_ALLOWED_ORIGINS")
		})
	}
}

func TestCORSPreflightRequest(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Set environment for development with wildcard
	os.Setenv("ENV", "development")
	os.Setenv("CORS_ALLOWED_ORIGINS", "*")

	// Create router with CORS middleware
	router := gin.New()
	router.Use(setupCORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Create preflight OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	req.Host = "localhost:8080"
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d, got: %d", http.StatusNoContent, w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin: *, got: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	allowMethods := w.Header().Get("Access-Control-Allow-Methods")
	if allowMethods == "" {
		t.Error("Expected Access-Control-Allow-Methods header to be set")
	}

	// Cleanup
	os.Unsetenv("ENV")
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
}
