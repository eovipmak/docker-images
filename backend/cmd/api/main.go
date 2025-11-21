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
	monitorRepo := postgres.NewMonitorRepository(db.DB)
	alertRuleRepo := postgres.NewAlertRuleRepository(db.DB)
	alertChannelRepo := postgres.NewAlertChannelRepository(db.DB)
	incidentRepo := postgres.NewIncidentRepository(db.DB)

	// Initialize services
	authService := service.NewAuthService(userRepo, tenantRepo, tenantUserRepo, cfg.JWT.Secret)
	monitorService := service.NewMonitorService(db)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, userRepo)
	monitorHandler := handlers.NewMonitorHandler(monitorRepo, monitorService)
	alertRuleHandler := handlers.NewAlertRuleHandler(alertRuleRepo, alertChannelRepo, monitorRepo)
	alertChannelHandler := handlers.NewAlertChannelHandler(alertChannelRepo)
	dashboardHandler := handlers.NewDashboardHandler(monitorRepo, incidentRepo)
	incidentHandler := handlers.NewIncidentHandler(incidentRepo, monitorRepo, alertRuleRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)
	tenantMiddleware := middleware.NewTenantMiddleware(tenantUserRepo)

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

	// Protected routes requiring authentication and tenant context
	protected := api.Group("/")
	protected.Use(authMiddleware.AuthRequired(), tenantMiddleware.TenantRequired())
	{
		// Monitor endpoints
		protected.POST("/monitors", monitorHandler.Create)
		protected.GET("/monitors", monitorHandler.List)
		protected.GET("/monitors/:id", monitorHandler.GetByID)
		protected.PUT("/monitors/:id", monitorHandler.Update)
		protected.DELETE("/monitors/:id", monitorHandler.Delete)
		protected.GET("/monitors/:id/checks", monitorHandler.GetChecks)
		protected.GET("/monitors/:id/ssl-status", monitorHandler.GetSSLStatus)

		// Alert rule endpoints
		protected.POST("/alert-rules", alertRuleHandler.Create)
		protected.GET("/alert-rules", alertRuleHandler.List)
		protected.GET("/alert-rules/:id", alertRuleHandler.GetByID)
		protected.PUT("/alert-rules/:id", alertRuleHandler.Update)
		protected.DELETE("/alert-rules/:id", alertRuleHandler.Delete)

		// Alert channel endpoints
		protected.POST("/alert-channels", alertChannelHandler.Create)
		protected.GET("/alert-channels", alertChannelHandler.List)
		protected.GET("/alert-channels/:id", alertChannelHandler.GetByID)
		protected.PUT("/alert-channels/:id", alertChannelHandler.Update)
		protected.DELETE("/alert-channels/:id", alertChannelHandler.Delete)

		// Dashboard endpoints
		protected.GET("/dashboard/stats", dashboardHandler.GetStats)
		protected.GET("/dashboard", dashboardHandler.GetDashboard)

		// Incident endpoints
		protected.GET("/incidents", incidentHandler.List)
		protected.GET("/incidents/:id", incidentHandler.GetByID)
		protected.POST("/incidents/:id/resolve", incidentHandler.Resolve)

		// Example protected endpoints - placeholder for future tenant-specific routes
		protected.GET("/tenant/info", func(c *gin.Context) {
			// Get tenant ID from context
			tenantIDValue, _ := c.Get("tenant_id")
			c.JSON(200, gin.H{
				"message":   "Tenant context established",
				"tenant_id": tenantIDValue,
			})
		})
	}

	// Start server
	log.Printf("Starting backend API on port %s", cfg.Server.Port)
	if err := router.Run(fmt.Sprintf(":%s", cfg.Server.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
