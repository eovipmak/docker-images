package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Set test environment variables
	os.Setenv("POSTGRES_HOST", "testhost")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	os.Setenv("POSTGRES_DB", "testdb")
	os.Setenv("PORT", "9090")
	os.Setenv("ENV", "testing")

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Database Host", cfg.Database.Host, "testhost"},
		{"Database Port", cfg.Database.Port, "5433"},
		{"Database User", cfg.Database.User, "testuser"},
		{"Database Password", cfg.Database.Password, "testpass"},
		{"Database Name", cfg.Database.DBName, "testdb"},
		{"Server Port", cfg.Server.Port, "9090"},
		{"Server Env", cfg.Server.Env, "testing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: expected %s, got %s", tt.name, tt.expected, tt.got)
			}
		})
	}

	// Test DSN generation
	expectedDSN := "host=testhost port=5433 user=testuser password=testpass dbname=testdb sslmode=disable"
	if cfg.GetDSN() != expectedDSN {
		t.Errorf("DSN: expected %s, got %s", expectedDSN, cfg.GetDSN())
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("POSTGRES_HOST")
	os.Unsetenv("POSTGRES_PORT")
	os.Unsetenv("POSTGRES_USER")
	os.Unsetenv("POSTGRES_PASSWORD")
	os.Unsetenv("POSTGRES_DB")
	os.Unsetenv("PORT")
	os.Unsetenv("ENV")

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify default values
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Database Host Default", cfg.Database.Host, "localhost"},
		{"Database Port Default", cfg.Database.Port, "5432"},
		{"Database User Default", cfg.Database.User, "postgres"},
		{"Database Password Default", cfg.Database.Password, "postgres"},
		{"Database Name Default", cfg.Database.DBName, "v_insight"},
		{"Server Port Default", cfg.Server.Port, "8080"},
		{"Server Env Default", cfg.Server.Env, "development"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: expected %s, got %s", tt.name, tt.expected, tt.got)
			}
		})
	}
}
