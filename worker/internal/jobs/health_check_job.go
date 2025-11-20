package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/eovipmak/v-insight/worker/internal/database"
	"github.com/eovipmak/v-insight/worker/internal/executor"
)

// Monitor represents a domain monitoring configuration
type Monitor struct {
	ID            string       `db:"id"`
	TenantID      int          `db:"tenant_id"`
	Name          string       `db:"name"`
	URL           string       `db:"url"`
	CheckInterval int          `db:"check_interval"`
	Timeout       int          `db:"timeout"`
	Enabled       bool         `db:"enabled"`
	CheckSSL      bool         `db:"check_ssl"`
	SSLAlertDays  int          `db:"ssl_alert_days"`
	LastCheckedAt sql.NullTime `db:"last_checked_at"`
	CreatedAt     time.Time    `db:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at"`
}

// MonitorCheck represents a single monitoring check result
type MonitorCheck struct {
	ID             string         `db:"id"`
	MonitorID      string         `db:"monitor_id"`
	CheckedAt      time.Time      `db:"checked_at"`
	StatusCode     sql.NullInt64  `db:"status_code"`
	ResponseTimeMs sql.NullInt64  `db:"response_time_ms"`
	SSLValid       sql.NullBool   `db:"ssl_valid"`
	SSLExpiresAt   sql.NullTime   `db:"ssl_expires_at"`
	ErrorMessage   sql.NullString `db:"error_message"`
	Success        bool           `db:"success"`
}

// HealthCheckJob performs HTTP health checks on monitors
type HealthCheckJob struct {
	db          *database.DB
	httpChecker *executor.HTTPChecker
	sslChecker  *executor.SSLChecker
}

// NewHealthCheckJob creates a new health check job
func NewHealthCheckJob(db *database.DB) *HealthCheckJob {
	return &HealthCheckJob{
		db:          db,
		httpChecker: executor.NewHTTPChecker(),
		sslChecker:  executor.NewSSLChecker(30 * time.Second),
	}
}

// Name returns the name of the job
func (j *HealthCheckJob) Name() string {
	return "HealthCheckJob"
}

// Run executes the health check job
func (j *HealthCheckJob) Run(ctx context.Context) error {
	startTime := time.Now()
	log.Println("[HealthCheckJob] Starting health check run")

	// Get monitors that need checking
	monitors, err := j.getMonitorsNeedingCheck(time.Now())
	if err != nil {
		log.Printf("[HealthCheckJob] Failed to get monitors: %v", err)
		return err
	}

	if len(monitors) == 0 {
		log.Println("[HealthCheckJob] No monitors need checking at this time")
		return nil
	}

	log.Printf("[HealthCheckJob] Found %d monitors to check", len(monitors))

	// Process monitors concurrently with worker pool
	j.checkMonitorsConcurrently(ctx, monitors)

	duration := time.Since(startTime)
	log.Printf("[HealthCheckJob] Health check completed in %v", duration)

	return nil
}

// getMonitorsNeedingCheck retrieves enabled monitors that need to be checked
func (j *HealthCheckJob) getMonitorsNeedingCheck(now time.Time) ([]*Monitor, error) {
	var monitors []*Monitor
	query := `
		SELECT id, tenant_id, name, url, check_interval, timeout, enabled, 
		       check_ssl, ssl_alert_days, last_checked_at, created_at, updated_at
		FROM monitors
		WHERE enabled = true
		  AND (
		      last_checked_at IS NULL
		      OR last_checked_at + (check_interval || ' seconds')::INTERVAL <= $1
		  )
		ORDER BY last_checked_at ASC NULLS FIRST
	`

	err := j.db.Select(&monitors, query, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitors needing check: %w", err)
	}

	return monitors, nil
}

// checkMonitorsConcurrently checks multiple monitors concurrently using a worker pool
func (j *HealthCheckJob) checkMonitorsConcurrently(ctx context.Context, monitors []*Monitor) {
	const maxConcurrent = 10 // Maximum 10 monitors checked concurrently
	
	// Create semaphore channel for limiting concurrency
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, monitor := range monitors {
		wg.Add(1)
		
		// Launch goroutine for each monitor
		go func(m *Monitor) {
			defer wg.Done()
			
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Create monitor-specific context with timeout
			checkCtx, cancel := context.WithTimeout(ctx, time.Duration(m.Timeout)*time.Second)
			defer cancel()

			// Perform the check
			j.checkMonitor(checkCtx, m)
		}(monitor)
	}

	// Wait for all checks to complete
	wg.Wait()
}

// checkMonitor performs a health check on a single monitor
func (j *HealthCheckJob) checkMonitor(ctx context.Context, monitor *Monitor) {
	log.Printf("[HealthCheckJob] Checking monitor: %s (%s)", monitor.Name, monitor.URL)
	
	checkedAt := time.Now()

	// Perform HTTP check
	result := j.httpChecker.CheckURL(ctx, monitor.URL, time.Duration(monitor.Timeout)*time.Second)

	// Create check record
	check := &MonitorCheck{
		MonitorID: monitor.ID,
		CheckedAt: checkedAt,
		Success:   result.Success,
	}

	// Set status code if available
	if result.StatusCode > 0 {
		check.StatusCode = sql.NullInt64{Int64: int64(result.StatusCode), Valid: true}
	}

	// Set response time in milliseconds
	if result.ResponseTime > 0 {
		check.ResponseTimeMs = sql.NullInt64{
			Int64: int64(result.ResponseTime.Milliseconds()),
			Valid: true,
		}
	}

	// Set error message if check failed
	if result.Error != nil {
		check.ErrorMessage = sql.NullString{
			String: result.Error.Error(),
			Valid:  true,
		}
	}

	// Check SSL certificate for HTTPS URLs if enabled
	if monitor.CheckSSL && (len(monitor.URL) >= 5 && monitor.URL[:5] == "https") {
		sslResult := j.sslChecker.CheckSSL(monitor.URL)
		
		// Set SSL validity
		check.SSLValid = sql.NullBool{
			Bool:  sslResult.Valid,
			Valid: true,
		}

		// Set SSL expiry date
		if !sslResult.ExpiresAt.IsZero() {
			check.SSLExpiresAt = sql.NullTime{
				Time:  sslResult.ExpiresAt,
				Valid: true,
			}
		}

		// Log SSL check result
		if sslResult.Error != nil {
			log.Printf("[HealthCheckJob] SSL check warning for %s: %v", monitor.Name, sslResult.Error)
		} else if sslResult.DaysUntil < monitor.SSLAlertDays {
			log.Printf("[HealthCheckJob] ⚠ SSL certificate for %s expires in %d days", monitor.Name, sslResult.DaysUntil)
		} else {
			log.Printf("[HealthCheckJob] ✓ SSL certificate for %s is valid (expires in %d days)", monitor.Name, sslResult.DaysUntil)
		}
	}

	// Save check result to database
	if err := j.saveCheck(check); err != nil {
		log.Printf("[HealthCheckJob] Failed to save check for monitor %s: %v", monitor.Name, err)
		return
	}

	// Update last_checked_at timestamp
	if err := j.updateLastCheckedAt(monitor.ID, checkedAt); err != nil {
		log.Printf("[HealthCheckJob] Failed to update last_checked_at for monitor %s: %v", monitor.Name, err)
		return
	}

	// Log result
	if result.Success {
		log.Printf("[HealthCheckJob] ✓ Monitor %s is UP - Status: %d, Response: %dms",
			monitor.Name, result.StatusCode, result.ResponseTime.Milliseconds())
	} else {
		log.Printf("[HealthCheckJob] ✗ Monitor %s is DOWN - Error: %v",
			monitor.Name, result.Error)
	}
}

// saveCheck saves a monitor check result to the database
func (j *HealthCheckJob) saveCheck(check *MonitorCheck) error {
	query := `
		INSERT INTO monitor_checks (
			monitor_id, checked_at, status_code, response_time_ms, 
			ssl_valid, ssl_expires_at, error_message, success
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := j.db.QueryRow(
		query,
		check.MonitorID,
		check.CheckedAt,
		check.StatusCode,
		check.ResponseTimeMs,
		check.SSLValid,
		check.SSLExpiresAt,
		check.ErrorMessage,
		check.Success,
	).Scan(&check.ID)

	if err != nil {
		return fmt.Errorf("failed to save monitor check: %w", err)
	}

	return nil
}

// updateLastCheckedAt updates the last_checked_at timestamp for a monitor
func (j *HealthCheckJob) updateLastCheckedAt(monitorID string, checkedAt time.Time) error {
	query := `
		UPDATE monitors
		SET last_checked_at = $1
		WHERE id = $2
	`

	result, err := j.db.Exec(query, checkedAt, monitorID)
	if err != nil {
		return fmt.Errorf("failed to update last_checked_at: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("monitor not found")
	}

	return nil
}
