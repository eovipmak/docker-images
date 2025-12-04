package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eovipmak/v-insight/backend/internal"
	"github.com/eovipmak/v-insight/backend/internal/api/handlers"
	"github.com/eovipmak/v-insight/backend/internal/api/middleware"
	"github.com/eovipmak/v-insight/backend/internal/config"
	"github.com/eovipmak/v-insight/backend/internal/database"
	"github.com/eovipmak/v-insight/backend/internal/domain/service"
	"github.com/eovipmak/v-insight/shared/repository/postgres"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "github.com/eovipmak/v-insight/backend/docs"
)

// @title V-Insight API
// @version 1.0
// @description Multi-tenant monitoring SaaS platform API for website health checks, SSL monitoring, and intelligent alerting.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/eovipmak/v-insight
// @contact.email support@v-insight.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize structured logger
	if err := internal.InitLogger(cfg.Server.Env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer internal.SyncLogger()

	internal.Log.Info("Starting V-Insight Backend API",
		zap.String("env", cfg.Server.Env),
		zap.String("port", cfg.Server.Port),
	)

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
		internal.Log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	internal.Log.Info("Database connection established")

	// Set Gin mode based on environment
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.New() // Use gin.New() instead of Default() to have full control

	// Apply global middleware
	// Request logging (structured JSON logs)
	router.Use(middleware.RequestLogger())
	
	// Recovery middleware
	router.Use(gin.Recovery())
	
	// Security headers (all routes)
	router.Use(middleware.SecurityHeaders(middleware.SecurityConfig{
		HSTSMaxAge:            cfg.Security.HSTSMaxAge,
		HSTSIncludeSubdomains: cfg.Security.HSTSIncludeSubdomains,
	}))

	// Request ID (all routes)
	router.Use(middleware.RequestID())

	// Request size limit (all routes) - 10MB default
	router.Use(middleware.RequestSizeLimit(10 * 1024 * 1024))

	// Performance middleware
	router.Use(middleware.GzipCompression())
	router.Use(middleware.CacheHeaders())

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db.DB)
	tenantRepo := postgres.NewTenantRepository(db.DB)
	tenantUserRepo := postgres.NewTenantUserRepository(db.DB)
	monitorRepo := postgres.NewMonitorRepository(db.DB)
	alertRuleRepo := postgres.NewAlertRuleRepository(db.DB)
	alertChannelRepo := postgres.NewAlertChannelRepository(db.DB)
	incidentRepo := postgres.NewIncidentRepository(db.DB)
	statusPageRepo := postgres.NewStatusPageRepository(db.DB)
	maintenanceWindowRepo := postgres.NewMaintenanceWindowRepository(db.DB)

	// Initialize services
	authService := service.NewAuthService(userRepo, tenantRepo, tenantUserRepo, cfg.JWT.Secret)
	monitorService := service.NewMonitorService(db)
	metricsService := service.NewMetricsService(db.DB)
	statusPageService := service.NewStatusPageService(db, statusPageRepo, monitorRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, userRepo)
	monitorHandler := handlers.NewMonitorHandler(monitorRepo, alertRuleRepo, alertChannelRepo, monitorService)
	metricsHandler := handlers.NewMetricsHandler(metricsService, monitorRepo)
	alertRuleHandler := handlers.NewAlertRuleHandler(alertRuleRepo, alertChannelRepo, monitorRepo)
	alertChannelHandler := handlers.NewAlertChannelHandler(alertChannelRepo)
	dashboardHandler := handlers.NewDashboardHandler(monitorRepo, incidentRepo, metricsService)
	incidentHandler := handlers.NewIncidentHandler(incidentRepo, monitorRepo, alertRuleRepo, alertChannelRepo)
	streamHandler := handlers.NewStreamHandler()
	statusPageHandler := handlers.NewStatusPageHandler(statusPageService)
	publicStatusPageHandler := handlers.NewPublicStatusPageHandler(statusPageService)
	maintenanceWindowHandler := handlers.NewMaintenanceWindowHandler(maintenanceWindowRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)
	tenantMiddleware := middleware.NewTenantMiddleware(tenantUserRepo)

	// Health check endpoint (legacy)
	healthHandler := func(c *gin.Context) {
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
	}
	router.GET("/health", healthHandler)
	router.HEAD("/health", healthHandler)

	// Liveness probe - checks if the application is running
	livenessHandler := func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "backend",
		})
	}
	router.GET("/health/live", livenessHandler)
	router.HEAD("/health/live", livenessHandler)

	// Readiness probe - checks if the application is ready to accept traffic
	readinessHandler := func(c *gin.Context) {
		// Create context with timeout for health check
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// Check database health
		if err := db.HealthContext(ctx); err != nil {
			c.JSON(503, gin.H{
				"status":   "error",
				"service":  "backend",
				"database": "unhealthy",
				"ready":    false,
			})
			return
		}

		c.JSON(200, gin.H{
			"status":   "ok",
			"service":  "backend",
			"database": "connected",
			"ready":    true,
		})
	}
	router.GET("/health/ready", readinessHandler)
	router.HEAD("/health/ready", readinessHandler)

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation (only in development)
	if cfg.Server.Env != "production" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// API routes
	api := router.Group("/api/v1")
	api.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "V-Insight API v1",
			"version": "1.0.0",
		})
	})

	// Auth routes (public) with rate limiting
	auth := api.Group("/auth")
	auth.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
		PerIP:   cfg.RateLimit.PerIP,
		PerUser: cfg.RateLimit.PerUser,
	}))
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", authMiddleware.AuthRequired(), authHandler.Me)
	}

	// Internal routes (no auth - should be internal network only)
	internal := router.Group("/internal")
	{
		internal.POST("/broadcast", streamHandler.HandleBroadcast)
	}

	// Public status page routes (no auth required)
	public := router.Group("/api/public")
	{
		public.GET("/status/:slug", publicStatusPageHandler.GetPublicStatusPage)
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
		protected.GET("/monitors/:id/stats", monitorHandler.GetStats)
		protected.GET("/monitors/:id/metrics", metricsHandler.GetMonitorMetrics)

		// Alert rule endpoints
		protected.POST("/alert-rules", alertRuleHandler.Create)
		protected.GET("/alert-rules", alertRuleHandler.List)
		protected.GET("/alert-rules/:id", alertRuleHandler.GetByID)
		protected.PUT("/alert-rules/:id", alertRuleHandler.Update)
		protected.DELETE("/alert-rules/:id", alertRuleHandler.Delete)
		protected.POST("/alert-rules/:id/test", alertRuleHandler.Test)

		// Alert channel endpoints
		protected.POST("/alert-channels", alertChannelHandler.Create)
		protected.GET("/alert-channels", alertChannelHandler.List)
		protected.GET("/alert-channels/:id", alertChannelHandler.GetByID)
		protected.PUT("/alert-channels/:id", alertChannelHandler.Update)
		protected.DELETE("/alert-channels/:id", alertChannelHandler.Delete)
		protected.POST("/alert-channels/:id/test", alertChannelHandler.Test)

		// Dashboard endpoints
		protected.GET("/dashboard/stats", dashboardHandler.GetStats)
		protected.GET("/dashboard", dashboardHandler.GetDashboard)

		// Incident endpoints
		protected.GET("/incidents", incidentHandler.List)
		protected.GET("/incidents/:id", incidentHandler.GetByID)
		protected.POST("/incidents/:id/resolve", incidentHandler.Resolve)

		// Status page endpoints
		protected.POST("/status-pages", statusPageHandler.CreateStatusPage)
		protected.GET("/status-pages", statusPageHandler.GetStatusPages)
		protected.GET("/status-pages/:id", statusPageHandler.GetStatusPage)
		protected.PUT("/status-pages/:id", statusPageHandler.UpdateStatusPage)
		protected.DELETE("/status-pages/:id", statusPageHandler.DeleteStatusPage)
		protected.GET("/status-pages/:id/monitors", statusPageHandler.GetStatusPageMonitors)
		protected.POST("/status-pages/:id/monitors/:monitor_id", statusPageHandler.AddMonitorToStatusPage)
		protected.DELETE("/status-pages/:id/monitors/:monitor_id", statusPageHandler.RemoveMonitorFromStatusPage)

		// Maintenance window endpoints
		protected.POST("/maintenance-windows", maintenanceWindowHandler.Create)
		protected.GET("/maintenance-windows", maintenanceWindowHandler.List)
		protected.GET("/maintenance-windows/:id", maintenanceWindowHandler.GetByID)
		protected.PUT("/maintenance-windows/:id", maintenanceWindowHandler.Update)
		protected.DELETE("/maintenance-windows/:id", maintenanceWindowHandler.Delete)

		// SSE Stream endpoint
		protected.GET("/stream/events", streamHandler.HandleSSE)

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
