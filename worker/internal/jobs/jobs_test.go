package jobs

import (
	"context"
	"testing"
)

func TestHealthCheckJob_Name(t *testing.T) {
	job := NewHealthCheckJob(nil)
	if job.Name() != "HealthCheckJob" {
		t.Fatalf("Expected job name 'HealthCheckJob', got '%s'", job.Name())
	}
}

func TestHealthCheckJob_Run(t *testing.T) {
	job := NewHealthCheckJob(nil)
	ctx := context.Background()

	err := job.Run(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
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
