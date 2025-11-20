package service

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/eovipmak/v-insight/backend/internal/database"
	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
)

// MonitorService provides business logic for monitor operations
type MonitorService struct {
	db *database.DB
}

// NewMonitorService creates a new monitor service
func NewMonitorService(db *database.DB) *MonitorService {
	return &MonitorService{
		db: db,
	}
}

// SSLStatus represents the SSL certificate status for a monitor
type SSLStatus struct {
	Valid         bool       `json:"valid"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	DaysUntilExpiry int      `json:"days_until_expiry"`
	ExpiringSoon  bool       `json:"expiring_soon"`
}

// GetSSLStatus retrieves the SSL certificate status for a monitor
// Returns the latest SSL check information from monitor_checks
func (s *MonitorService) GetSSLStatus(monitorID string) (*SSLStatus, error) {
	// Get the latest check with SSL information
	query := `
		SELECT ssl_valid, ssl_expires_at
		FROM monitor_checks
		WHERE monitor_id = $1
		  AND ssl_valid IS NOT NULL
		ORDER BY checked_at DESC
		LIMIT 1
	`

	var sslValid sql.NullBool
	var sslExpiresAt sql.NullTime

	err := s.db.QueryRow(query, monitorID).Scan(&sslValid, &sslExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// No SSL check data available yet
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get SSL status: %w", err)
	}

	// If no valid SSL data, return nil
	if !sslValid.Valid {
		return nil, nil
	}

	status := &SSLStatus{
		Valid: sslValid.Bool,
	}

	// Calculate days until expiry if we have expiry date
	if sslExpiresAt.Valid {
		status.ExpiresAt = &sslExpiresAt.Time
		daysUntil := int(time.Until(sslExpiresAt.Time).Hours() / 24)
		status.DaysUntilExpiry = daysUntil
	}

	return status, nil
}

// IsSSLExpiringSoon checks if a monitor's SSL certificate is expiring within the specified days
func (s *MonitorService) IsSSLExpiringSoon(monitorID string, days int) (bool, error) {
	// Get monitor to check if SSL checking is enabled
	var checkSSL bool
	err := s.db.QueryRow("SELECT check_ssl FROM monitors WHERE id = $1", monitorID).Scan(&checkSSL)
	if err != nil {
		return false, fmt.Errorf("failed to get monitor: %w", err)
	}

	// If SSL checking is disabled, return false
	if !checkSSL {
		return false, nil
	}

	// Get SSL status
	status, err := s.GetSSLStatus(monitorID)
	if err != nil {
		return false, err
	}

	// If no SSL status available, return false
	if status == nil {
		return false, nil
	}

	// Check if certificate is expiring soon
	return status.Valid && status.DaysUntilExpiry <= days && status.DaysUntilExpiry >= 0, nil
}

// GetMonitorWithSSLStatus retrieves a monitor along with its SSL status
func (s *MonitorService) GetMonitorWithSSLStatus(monitorID string) (*entities.Monitor, *SSLStatus, error) {
	// Get monitor
	var monitor entities.Monitor
	query := `
		SELECT id, tenant_id, name, url, check_interval, timeout, enabled,
		       check_ssl, ssl_alert_days, last_checked_at, created_at, updated_at
		FROM monitors
		WHERE id = $1
	`

	err := s.db.QueryRow(query, monitorID).Scan(
		&monitor.ID,
		&monitor.TenantID,
		&monitor.Name,
		&monitor.URL,
		&monitor.CheckInterval,
		&monitor.Timeout,
		&monitor.Enabled,
		&monitor.CheckSSL,
		&monitor.SSLAlertDays,
		&monitor.LastCheckedAt,
		&monitor.CreatedAt,
		&monitor.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get monitor: %w", err)
	}

	// Get SSL status if SSL checking is enabled
	var sslStatus *SSLStatus
	if monitor.CheckSSL {
		sslStatus, err = s.GetSSLStatus(monitorID)
		if err != nil {
			return &monitor, nil, fmt.Errorf("failed to get SSL status: %w", err)
		}
	}

	return &monitor, sslStatus, nil
}
