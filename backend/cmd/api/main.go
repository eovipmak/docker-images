package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eovipmak/v-insight/backend/internal/api/handlers"
	"github.com/eovipmak/v-insight/backend/internal/api/middleware"
	"github.com/eovipmak/v-insight/backend/internal/config"
	"github.com/eovipmak/v-insight/backend/internal/database"
	"github.com/eovipmak/v-insight/backend/internal/domain/service"
	"github.com/eovipmak/v-insight/backend/internal/repository/postgres"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	dbCfg := database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	db, err := database.New(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Set Gin mode based on environment
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.Default()

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db.DB)
	tenantRepo := postgres.NewTenantRepository(db.DB)
	tenantUserRepo := postgres.NewTenantUserRepository(db.DB)

	// Initialize services
	authService := service.NewAuthService(userRepo, tenantRepo, tenantUserRepo, cfg.JWT.Secret)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, userRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		// Create context with timeout for health check
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// Check database health
		if err := db.HealthContext(ctx); err != nil {
			c.JSON(503, gin.H{
				"status": "error",
				"error":  "database unhealthy",
			})
			return
		}

		c.JSON(200, gin.H{
			"status":   "ok",
			"database": "connected",
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

	// Auth routes (public)
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", authMiddleware.AuthRequired(), authHandler.Me)
	}

	// Start server
	log.Printf("Starting backend API on port %s", cfg.Server.Port)
	if err := router.Run(fmt.Sprintf(":%s", cfg.Server.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
