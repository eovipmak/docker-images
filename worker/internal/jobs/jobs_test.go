package jobs

import (
	"context"
	"testing"

	"github.com/eovipmak/v-insight/worker/internal/database"
)

func TestHealthCheckJob_Name(t *testing.T) {
	job := NewHealthCheckJob(nil)
	if job.Name() != "HealthCheckJob" {
		t.Fatalf("Expected job name 'HealthCheckJob', got '%s'", job.Name())
	}
}

func TestHealthCheckJob_Run_NilDB(t *testing.T) {
	job := NewHealthCheckJob(nil)
	ctx := context.Background()

	// Should handle nil DB gracefully by returning an error
	err := job.Run(ctx)
	if err == nil {
		t.Fatal("Expected error with nil database, got nil")
	}
}

func TestHealthCheckJob_Run_WithMockDB(t *testing.T) {
	// This test requires a proper database connection
	// Skip if DB_HOST is not set (CI environment without database)
	// In a real scenario, you would use a test database or mock
	t.Skip("Skipping integration test - requires database connection")
	
	// Example of how to test with a real database:
	// cfg := database.Config{
	// 	Host:     "localhost",
	// 	Port:     "5432",
	// 	User:     "test",
	// 	Password: "test",
	// 	DBName:   "test_db",
	// 	SSLMode:  "disable",
	// }
	// db, err := database.New(cfg)
	// if err != nil {
	// 	t.Fatalf("Failed to connect to test database: %v", err)
	// }
	// defer db.Close()
	//
	// job := NewHealthCheckJob(db)
	// ctx := context.Background()
	//
	// err = job.Run(ctx)
	// if err != nil {
	// 	t.Fatalf("Expected no error, got: %v", err)
	// }
}

func TestSSLCheckJob_Name(t *testing.T) {
	job := NewSSLCheckJob(nil)
	if job.Name() != "SSLCheckJob" {
		t.Fatalf("Expected job name 'SSLCheckJob', got '%s'", job.Name())
	}
}

func TestSSLCheckJob_Run(t *testing.T) {
	job := NewSSLCheckJob(nil)
	ctx := context.Background()

	err := job.Run(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}
