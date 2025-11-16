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
	// Get environment mode
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	// Get allowed origins from environment
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")

	// Configure CORS based on environment and origins setting
	config := cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		MaxAge:       12 * time.Hour, // Cache preflight requests for 12 hours
	}

	// In development mode, check if wildcard is requested
	if env == "development" && (corsOrigins == "*" || corsOrigins == "") {
		// Allow all origins in development for ease of development
		config.AllowAllOrigins = true
		log.Println("CORS: Allowing all origins (development mode)")
	} else {
		// In production, never allow wildcard for security
		// Parse comma-separated origins for specific origins
		if corsOrigins == "" || corsOrigins == "*" {
			// Default to safe local development origins even if wildcard is specified in production
			corsOrigins = "http://localhost:3000,http://127.0.0.1:3000"
			if env == "production" {
				log.Println("CORS: Warning - wildcard not allowed in production, using default origins")
			}
		}

		allowedOrigins := strings.Split(corsOrigins, ",")
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
		}

		config.AllowOrigins = allowedOrigins
		config.AllowCredentials = true
		log.Printf("CORS: Allowing specific origins: %v", allowedOrigins)
	}

	return cors.New(config)
}
