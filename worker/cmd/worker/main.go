package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eovipmak/v-insight/worker/internal/config"
	"github.com/eovipmak/v-insight/worker/internal/database"
	"github.com/eovipmak/v-insight/worker/internal/executor"
	"github.com/eovipmak/v-insight/worker/internal/jobs"
	"github.com/eovipmak/v-insight/worker/internal/scheduler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
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

	// Initialize executor for concurrent job processing
	exec := executor.New(executor.DefaultConfig())
	exec.Start()
	defer exec.Stop()

	// Initialize cron scheduler
	sched := scheduler.New()

	// Register jobs
	healthCheckJob := jobs.NewHealthCheckJob(db)
	sslCheckJob := jobs.NewSSLCheckJob(db)

	// Schedule health check job to run every 1 minute
	if err := sched.AddJob("*/1 * * * *", healthCheckJob); err != nil {
		log.Fatalf("Failed to schedule health check job: %v", err)
	}

	// Schedule SSL check job to run every 5 minutes
	if err := sched.AddJob("*/5 * * * *", sslCheckJob); err != nil {
		log.Fatalf("Failed to schedule SSL check job: %v", err)
	}

	// Start the scheduler
	sched.Start()
	defer sched.Stop()

	log.Printf("Registered jobs: %v", sched.GetJobs())

	// Create HTTP server for health checks and metrics
	app := fiber.New(fiber.Config{
		AppName: "V-Insight Worker",
	})

	app.Use(logger.New())

	// Health check endpoint
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

	// Start HTTP server in a goroutine
	go func() {
		port := cfg.Worker.Port
		log.Printf("Starting worker HTTP server on port %s", port)
		if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down worker service...")

	// Graceful shutdown
	log.Println("Stopping scheduler...")
	sched.Stop()

	log.Println("Stopping executor...")
	exec.Stop()

	log.Println("Stopping HTTP server...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}

	log.Println("Closing database connection...")
	db.Close()

	log.Println("Worker service stopped successfully")
}
