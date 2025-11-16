package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/eovipmak/v-insight/backend/internal/config"
	"github.com/gin-contrib/cors"
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
	router.Use(setupCORSMiddleware())

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

// setupCORSMiddleware configures CORS using gin-contrib/cors based on environment
func setupCORSMiddleware() gin.HandlerFunc {
	// Get allowed origins from environment
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsOrigins == "" {
		// Default to safe local development origins
		corsOrigins = "http://localhost:3000,http://127.0.0.1:3000"
	}

	// Parse comma-separated origins
	allowedOrigins := strings.Split(corsOrigins, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	// Configure CORS
	config := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // Cache preflight requests for 12 hours
	}

	return cors.New(config)
}
