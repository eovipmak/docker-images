package service

import (
	"testing"
	"time"

	"github.com/eovipmak/v-insight/backend/internal/database"
)

func TestMonitorService_Creation(t *testing.T) {
	// This is a simple test to verify the service can be created
	// More comprehensive tests would require a test database
	
	// Create a nil database for basic structure validation
	var db *database.DB = nil
	
	// Should not panic
	service := NewMonitorService(db)
	
	if service == nil {
		t.Error("NewMonitorService returned nil")
	}
}

func TestSSLStatus_Structure(t *testing.T) {
	// Test SSL status structure
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	
	status := &SSLStatus{
		Valid:           true,
		ExpiresAt:       &expiresAt,
		DaysUntilExpiry: 30,
		ExpiringSoon:    false,
	}
	
	if !status.Valid {
		t.Error("Expected Valid to be true")
	}
	
	if status.DaysUntilExpiry != 30 {
		t.Errorf("Expected DaysUntilExpiry to be 30, got %d", status.DaysUntilExpiry)
	}
	
	if status.ExpiringSoon {
		t.Error("Expected ExpiringSoon to be false")
	}
}
