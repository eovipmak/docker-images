package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/eovipmak/v-insight/worker/internal"
	"github.com/eovipmak/v-insight/worker/internal/database"
	"github.com/eovipmak/v-insight/worker/internal/executor"
	"go.uber.org/zap"
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
	if j.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	startTime := time.Now()
	
	// Record job execution metrics
	defer func() {
		duration := time.Since(startTime)
		internal.JobExecutionDuration.WithLabelValues("HealthCheckJob").Observe(duration.Seconds())
	}()

	if internal.Log != nil {
		internal.Log.Info("Starting health check run")
	}

	// Get monitors that need checking
	monitors, err := j.getMonitorsNeedingCheck(time.Now())
	if err != nil {
		if internal.Log != nil {
			internal.Log.Error("Failed to get monitors", zap.Error(err))
		}
		internal.JobExecutionTotal.WithLabelValues("HealthCheckJob", "failure").Inc()
		return err
	}

	if len(monitors) == 0 {
		if internal.Log != nil {
			internal.Log.Debug("No monitors need checking at this time")
		}
		internal.JobExecutionTotal.WithLabelValues("HealthCheckJob", "success").Inc()
		return nil
	}

	if internal.Log != nil {
		internal.Log.Info("Found monitors to check", zap.Int("count", len(monitors)))
	}

	// Process monitors concurrently with worker pool
	j.checkMonitorsConcurrently(ctx, monitors)

	duration := time.Since(startTime)
	if internal.Log != nil {
		internal.Log.Info("Health check completed", zap.Duration("duration", duration))
	}

	internal.JobExecutionTotal.WithLabelValues("HealthCheckJob", "success").Inc()
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
	if internal.Log != nil {
		internal.Log.Debug("Checking monitor",
			zap.String("monitor_name", monitor.Name),
			zap.String("url", monitor.URL),
		)
	}
	
	checkedAt := time.Now().UTC()

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
			if internal.Log != nil {
				internal.Log.Warn("SSL check warning",
					zap.String("monitor_name", monitor.Name),
					zap.Error(sslResult.Error),
				)
			}
		} else if sslResult.DaysUntil < monitor.SSLAlertDays {
			if internal.Log != nil {
				internal.Log.Warn("SSL certificate expiring soon",
					zap.String("monitor_name", monitor.Name),
					zap.Int("days_until_expiry", sslResult.DaysUntil),
				)
			}
		} else {
			if internal.Log != nil {
				internal.Log.Debug("SSL certificate valid",
					zap.String("monitor_name", monitor.Name),
					zap.Int("days_until_expiry", sslResult.DaysUntil),
				)
			}
		}
	}

	// Save check result to database
	if err := j.saveCheck(check); err != nil {
		if internal.Log != nil {
			internal.Log.Error("Failed to save check",
				zap.String("monitor_name", monitor.Name),
				zap.Error(err),
			)
		}
		return
	}

	// Update last_checked_at timestamp
	if err := j.updateLastCheckedAt(monitor.ID, checkedAt); err != nil {
		if internal.Log != nil {
			internal.Log.Error("Failed to update last_checked_at",
				zap.String("monitor_name", monitor.Name),
				zap.Error(err),
			)
		}
		return
	}

	// Record metrics
	if result.Success {
		internal.MonitorCheckTotal.WithLabelValues("success").Inc()
		if internal.Log != nil {
			internal.Log.Info("Monitor check successful",
				zap.String("monitor_name", monitor.Name),
				zap.Int("status_code", result.StatusCode),
				zap.Int64("response_time_ms", result.ResponseTime.Milliseconds()),
			)
		}
	} else {
		internal.MonitorCheckTotal.WithLabelValues("failure").Inc()
		if internal.Log != nil {
			internal.Log.Warn("Monitor check failed",
				zap.String("monitor_name", monitor.Name),
				zap.Error(result.Error),
			)
		}
	}

	// Broadcast monitor_check event
	j.broadcastMonitorCheckEvent(monitor, check)
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

// broadcastMonitorCheckEvent broadcasts a monitor check event to the backend SSE handler
func (j *HealthCheckJob) broadcastMonitorCheckEvent(monitor *Monitor, check *MonitorCheck) {
	// Prepare event data
	data := map[string]interface{}{
		"monitor_id":   monitor.ID,
		"monitor_name": monitor.Name,
		"success":      check.Success,
		"checked_at":   check.CheckedAt.Format(time.RFC3339),
	}

	if check.StatusCode.Valid {
		data["status_code"] = check.StatusCode.Int64
	}

	if check.ResponseTimeMs.Valid {
		data["response_time_ms"] = check.ResponseTimeMs.Int64
	}

	if check.ErrorMessage.Valid {
		data["error_message"] = check.ErrorMessage.String
	}

	if check.SSLValid.Valid {
		data["ssl_valid"] = check.SSLValid.Bool
	}

	if check.SSLExpiresAt.Valid {
		data["ssl_expires_at"] = check.SSLExpiresAt.Time.Format(time.RFC3339)
	}

	// Send broadcast request to backend
	broadcastEvent("monitor_check", data, monitor.TenantID)
}
