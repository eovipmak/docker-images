package jobs

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// TestMonitor_StructFields tests that Monitor struct has all required fields
func TestMonitor_StructFields(t *testing.T) {
	monitor := Monitor{
		ID:            "test-id",
		TenantID:      1,
		Name:          "Test Monitor",
		URL:           "https://example.com",
		CheckInterval: 300,
		Timeout:       30,
		Enabled:       true,
		LastCheckedAt: sql.NullTime{Time: time.Now(), Valid: true},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if monitor.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", monitor.ID)
	}
	if monitor.TenantID != 1 {
		t.Errorf("Expected TenantID 1, got %d", monitor.TenantID)
	}
	if monitor.Name != "Test Monitor" {
		t.Errorf("Expected Name 'Test Monitor', got '%s'", monitor.Name)
	}
	if monitor.URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got '%s'", monitor.URL)
	}
}

// TestMonitorCheck_StructFields tests that MonitorCheck struct has all required fields
func TestMonitorCheck_StructFields(t *testing.T) {
	check := MonitorCheck{
		ID:             "check-id",
		MonitorID:      "monitor-id",
		CheckedAt:      time.Now(),
		StatusCode:     sql.NullInt64{Int64: 200, Valid: true},
		ResponseTimeMs: sql.NullInt64{Int64: 150, Valid: true},
		SSLValid:       sql.NullBool{Bool: true, Valid: true},
		SSLExpiresAt:   sql.NullTime{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
		ErrorMessage:   sql.NullString{String: "", Valid: false},
		Success:        true,
	}

	if check.ID != "check-id" {
		t.Errorf("Expected ID 'check-id', got '%s'", check.ID)
	}
	if check.MonitorID != "monitor-id" {
		t.Errorf("Expected MonitorID 'monitor-id', got '%s'", check.MonitorID)
	}
	if !check.Success {
		t.Error("Expected Success true, got false")
	}
	if check.StatusCode.Int64 != 200 {
		t.Errorf("Expected StatusCode 200, got %d", check.StatusCode.Int64)
	}
	if check.ResponseTimeMs.Int64 != 150 {
		t.Errorf("Expected ResponseTimeMs 150, got %d", check.ResponseTimeMs.Int64)
	}
}

// TestHealthCheckJob_NewHealthCheckJob tests that NewHealthCheckJob creates a valid job
func TestHealthCheckJob_NewHealthCheckJob(t *testing.T) {
	job := NewHealthCheckJob(nil)
	
	if job == nil {
		t.Fatal("Expected job to be created, got nil")
	}
	
	if job.httpChecker == nil {
		t.Error("Expected httpChecker to be initialized, got nil")
	}
	
	if job.Name() != "HealthCheckJob" {
		t.Errorf("Expected job name 'HealthCheckJob', got '%s'", job.Name())
	}
}

// TestHealthCheckJob_Run_ContextCancellation tests graceful handling of context cancellation
func TestHealthCheckJob_Run_ContextCancellation(t *testing.T) {
	job := NewHealthCheckJob(nil)
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	// Should handle cancelled context gracefully
	err := job.Run(ctx)
	// With nil DB, it should return an error when trying to query
	if err == nil {
		t.Error("Expected error with cancelled context and nil DB, got nil")
	}
}

// TestHealthCheckJob_CheckMonitorsConcurrently_EmptyList tests concurrent checking with empty list
func TestHealthCheckJob_CheckMonitorsConcurrently_EmptyList(t *testing.T) {
	job := NewHealthCheckJob(nil)
	ctx := context.Background()
	
	// Should handle empty list gracefully
	monitors := []*Monitor{}
	
	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("checkMonitorsConcurrently panicked with empty list: %v", r)
		}
	}()
	
	job.checkMonitorsConcurrently(ctx, monitors)
}
