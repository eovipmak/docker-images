package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/eovipmak/v-insight/backend/internal/config"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode based on environment
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.Default()

	// Configure CORS middleware
	router.Use(corsMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	api.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "V-Insight API v1",
			"version": "1.0.0",
		})
	})

	// Start server
	log.Printf("Starting backend API on port %s", cfg.Server.Port)
	if err := router.Run(fmt.Sprintf(":%s", cfg.Server.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// corsMiddleware configures CORS based on environment
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get allowed origins from environment
		corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
		if corsOrigins == "" {
			// Default to safe local development origins
			corsOrigins = "http://localhost:3000,http://127.0.0.1:3000"
		}

		// Parse comma-separated origins
		allowedOrigins := strings.Split(corsOrigins, ",")
		originMap := make(map[string]bool)
		for _, origin := range allowedOrigins {
			originMap[strings.TrimSpace(origin)] = true
		}

		// Check if request origin is allowed
		origin := c.Request.Header.Get("Origin")
		if originMap[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
