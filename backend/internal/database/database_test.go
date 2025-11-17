package database

import (
	"testing"
	"time"
)

func TestNewDatabase(t *testing.T) {
	// Test with invalid configuration (should fail to connect)
	cfg := Config{
		Host:            "invalid-host",
		Port:            "5432",
		User:            "postgres",
		Password:        "postgres",
		DBName:          "test",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("Expected error when connecting to invalid host, got nil")
	}
}

func TestConfigDefaults(t *testing.T) {
	// This test validates that we can create a Config struct
	cfg := Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "test",
		SSLMode:  "disable",
	}

	if cfg.Host != "localhost" {
		t.Errorf("Expected Host to be 'localhost', got '%s'", cfg.Host)
	}

	if cfg.Port != "5432" {
		t.Errorf("Expected Port to be '5432', got '%s'", cfg.Port)
	}
}
