package executor

import (
	"context"
	"testing"
	"time"
)

func TestICMPChecker_Check_Success(t *testing.T) {
	checker := NewICMPChecker()
	ctx := context.Background()

	// Ping localhost
	result := checker.Check(ctx, "127.0.0.1", 2*time.Second)

	if !result.Success {
		t.Errorf("Expected success, got failure: %v", result.Error)
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	if result.ResponseTime <= 0 {
		t.Errorf("Expected positive response time, got %v", result.ResponseTime)
	}
}

func TestICMPChecker_Check_Failure(t *testing.T) {
	checker := NewICMPChecker()
	ctx := context.Background()

	// Ping non-existent host (should fail or timeout)
	// Using a reserved IP that is unlikely to respond or route
	result := checker.Check(ctx, "203.0.113.1", 1*time.Second)

	if result.Success {
		t.Errorf("Expected failure for unreachable host, got success")
	}
	if result.Error == nil {
		t.Errorf("Expected error for unreachable host, got nil")
	}
}
