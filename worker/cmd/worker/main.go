package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eovipmak/v-insight/worker/internal"
	"github.com/eovipmak/v-insight/worker/internal/config"
	"github.com/eovipmak/v-insight/shared/repository/postgres"
	"github.com/eovipmak/v-insight/worker/internal/database"
	"github.com/eovipmak/v-insight/worker/internal/executor"
	"github.com/eovipmak/v-insight/worker/internal/jobs"
	"github.com/eovipmak/v-insight/worker/internal/scheduler"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize structured logger
	env := "development"
	if cfg.Worker.Env != "" {
		env = cfg.Worker.Env
	}
	if err := internal.InitLogger(env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer internal.SyncLogger()

	internal.Log.Info("Starting V-Insight Worker Service",
		zap.String("env", env),
		zap.String("port", cfg.Worker.Port),
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

	// Initialize executor for concurrent job processing
	exec := executor.New(executor.DefaultConfig())
	exec.Start()
	defer exec.Stop()

	// Initialize cron scheduler
	sched := scheduler.New()

	// Initialize repositories
	monitorRepo := postgres.NewMonitorRepository(db.DB)
	alertRuleRepo := postgres.NewAlertRuleRepository(db.DB)
	incidentRepo := postgres.NewIncidentRepository(db.DB)
	alertChannelRepo := postgres.NewAlertChannelRepository(db.DB)

	// Register jobs
	healthCheckJob := jobs.NewHealthCheckJob(monitorRepo)
	sslCheckJob := jobs.NewSSLCheckJob(db)
	alertEvaluatorJob := jobs.NewAlertEvaluatorJob(alertRuleRepo, incidentRepo, monitorRepo)
	notificationJob := jobs.NewNotificationJob(incidentRepo, alertChannelRepo, cfg.SMTP)

	// Schedule health check job to run every 30 seconds
	if err := sched.AddJob("*/30 * * * * *", healthCheckJob); err != nil {
		log.Fatalf("Failed to schedule health check job: %v", err)
	}

	// Schedule SSL check job to run every 5 minutes
	if err := sched.AddJob("*/5 * * * *", sslCheckJob); err != nil {
		log.Fatalf("Failed to schedule SSL check job: %v", err)
	}

	// Schedule alert evaluator job to run every minute
	if err := sched.AddJob("* * * * *", alertEvaluatorJob); err != nil {
		log.Fatalf("Failed to schedule alert evaluator job: %v", err)
	}

	// Schedule notification job to run every 30 seconds
	if err := sched.AddJob("*/30 * * * * *", notificationJob); err != nil {
		log.Fatalf("Failed to schedule notification job: %v", err)
	}

	// Start the scheduler
	sched.Start()
	defer sched.Stop()

	internal.Log.Info("Scheduler started", zap.Strings("jobs", sched.GetJobs()))

	// Create HTTP server for health checks and metrics
	app := fiber.New(fiber.Config{
		AppName: "V-Insight Worker",
	})

	app.Use(logger.New())

	// Health check endpoint (legacy)
	app.Get("/health", func(c *fiber.Ctx) error {
		// Check database health
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		if err := db.HealthContext(ctx); err != nil {
			return c.Status(503).JSON(fiber.Map{
				"status":   "error",
				"service":  "worker",
				"database": "unhealthy",
				"error":    err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status":   "ok",
			"service":  "worker",
			"database": "connected",
			"jobs":     sched.GetJobs(),
		})
	})

	// Liveness probe - checks if the worker is running
	app.Get("/health/live", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "worker",
		})
	})

	// Readiness probe - checks if the worker is ready to process jobs
	app.Get("/health/ready", func(c *fiber.Ctx) error {
		// Check database health
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		if err := db.HealthContext(ctx); err != nil {
			return c.Status(503).JSON(fiber.Map{
				"status":   "error",
				"service":  "worker",
				"database": "unhealthy",
				"ready":    false,
			})
		}

		// Check if scheduler is running
		jobs := sched.GetJobs()
		if len(jobs) == 0 {
			return c.Status(503).JSON(fiber.Map{
				"status":    "error",
				"service":   "worker",
				"database":  "connected",
				"scheduler": "no jobs registered",
				"ready":     false,
			})
		}

		return c.JSON(fiber.Map{
			"status":    "ok",
			"service":   "worker",
			"database":  "connected",
			"scheduler": "running",
			"jobs":      jobs,
			"ready":     true,
		})
	})

	// Prometheus metrics endpoint
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// Start HTTP server in a goroutine
	go func() {
		port := cfg.Worker.Port
		internal.Log.Info("Starting worker HTTP server", zap.String("port", port))
		if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
			internal.Log.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	internal.Log.Info("Shutting down worker service...")

	// Graceful shutdown
	internal.Log.Info("Stopping scheduler...")
	sched.Stop()

	internal.Log.Info("Stopping executor...")
	exec.Stop()

	internal.Log.Info("Stopping HTTP server...")
	if err := app.Shutdown(); err != nil {
		internal.Log.Error("Error shutting down HTTP server", zap.Error(err))
	}

	internal.Log.Info("Closing database connection...")
	db.Close()

	internal.Log.Info("Worker service stopped successfully")
}
