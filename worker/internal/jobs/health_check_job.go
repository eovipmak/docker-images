package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/eovipmak/v-insight/worker/internal"
	"github.com/eovipmak/v-insight/worker/internal/executor"
	"go.uber.org/zap"
)

// HealthCheckJob performs HTTP health checks on monitors
type HealthCheckJob struct {
	monitorRepo repository.MonitorRepository
	httpChecker *executor.HTTPChecker
	tcpChecker  *executor.TCPChecker
	sslChecker  *executor.SSLChecker
	icmpChecker *executor.ICMPChecker
}

// NewHealthCheckJob creates a new health check job
func NewHealthCheckJob(monitorRepo repository.MonitorRepository) *HealthCheckJob {
	return &HealthCheckJob{
		monitorRepo: monitorRepo,
		httpChecker: executor.NewHTTPChecker(),
		tcpChecker:  executor.NewTCPChecker(),
		sslChecker:  executor.NewSSLChecker(30 * time.Second),
		icmpChecker: executor.NewICMPChecker(),
	}
}

// Name returns the name of the job
func (j *HealthCheckJob) Name() string {
	return "HealthCheckJob"
}

// Run executes the health check job
func (j *HealthCheckJob) Run(ctx context.Context) error {
	if j.monitorRepo == nil {
		return fmt.Errorf("monitor repository is nil")
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
	monitors, err := j.monitorRepo.GetMonitorsNeedingCheck(time.Now())
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

// checkMonitorsConcurrently checks multiple monitors concurrently using a worker pool
func (j *HealthCheckJob) checkMonitorsConcurrently(ctx context.Context, monitors []*entities.Monitor) {
	const maxConcurrent = 10 // Maximum 10 monitors checked concurrently
	
	// Create semaphore channel for limiting concurrency
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, monitor := range monitors {
		wg.Add(1)
		
		// Launch goroutine for each monitor
		go func(m *entities.Monitor) {
			defer wg.Done()
			
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Create monitor-specific context with timeout
			checkCtx, cancel := context.WithTimeout(ctx, time.Duration(m.Timeout)*time.Second)
			defer cancel()

			// Perform the check
			if internal.Log != nil {
				internal.Log.Info("Checking monitor",
					zap.String("type", m.Type),
					zap.String("url", m.URL),
					zap.String("id", m.ID),
				)
			}
			j.checkMonitor(checkCtx, m)
		}(monitor)
	}

	// Wait for all checks to complete
	wg.Wait()
}

// checkMonitor performs a health check on a single monitor
func (j *HealthCheckJob) checkMonitor(ctx context.Context, monitor *entities.Monitor) {
	// Clean up monitor type to ensure reliable comparisons
	monitor.Type = strings.ToLower(strings.TrimSpace(monitor.Type))

	if internal.Log != nil {
		internal.Log.Debug("Checking monitor",
			zap.String("monitor_name", monitor.Name),
			zap.String("type", monitor.Type),
			zap.String("url", monitor.URL),
		)
	}
	
	checkedAt := time.Now().UTC()

	var success bool
	var responseTime time.Duration
	var statusCode int
	var checkError error

	// Perform check based on monitor type
	if internal.Log != nil {
		internal.Log.Info("Processing monitor check",
			zap.String("monitor_id", monitor.ID),
			zap.String("type", monitor.Type),
			zap.String("url", monitor.URL),
		)
	}

	if monitor.Type == "tcp" {
		// Parse TCP URL (expected format: host:port or tcp://host:port)
		host, port, err := j.parseTCPAddress(monitor.URL)
		if err != nil {
			checkError = fmt.Errorf("invalid TCP address: %w", err)
			success = false
			responseTime = 0
		} else {
			tcpResult := j.tcpChecker.Check(host, port, time.Duration(monitor.Timeout)*time.Second)
			success = tcpResult.Success
			responseTime = tcpResult.ResponseTime
			if tcpResult.Error != nil {
				checkError = tcpResult.Error
			}
		}
	} else if monitor.Type == "ping" || monitor.Type == "icmp" {
		// Prepare host for ping (strip protocol)
		host := monitor.URL
		if strings.HasPrefix(host, "http://") {
			host = host[7:]
		} else if strings.HasPrefix(host, "https://") {
			host = host[8:]
		}
		// Strip path/query/fragment if present, simplistic approach
		// Ideally we rely on validation, but let's be safe against basic mistakes
		// e.g. google.com/foo -> google.com
		if idx := strings.Index(host, "/"); idx != -1 {
			host = host[:idx]
		}

		icmpResult := j.icmpChecker.Check(ctx, host, time.Duration(monitor.Timeout)*time.Second)
		success = icmpResult.Success
		responseTime = icmpResult.ResponseTime
		if icmpResult.Error != nil {
			checkError = icmpResult.Error
		}
	} else {
		if internal.Log != nil {
			internal.Log.Info("Falling back to HTTP check", zap.String("type", monitor.Type))
		}
		// HTTP/HTTPS check with expected status codes support
		httpResult := j.httpChecker.CheckURLWithExpectedCodes(ctx, monitor.URL, time.Duration(monitor.Timeout)*time.Second, monitor.Keyword, monitor.ExpectedStatusCodes)
		success = httpResult.Success
		responseTime = httpResult.ResponseTime
		statusCode = httpResult.StatusCode
		if httpResult.Error != nil {
			checkError = httpResult.Error
		}
	}

	// Create check record
	check := &entities.MonitorCheck{
		MonitorID: monitor.ID,
		CheckedAt: checkedAt,
		Success:   success,
	}

	// Set status code if available (only for HTTP checks)
	if statusCode > 0 {
		check.StatusCode = sql.NullInt64{Int64: int64(statusCode), Valid: true}
	}

	// Set response time in milliseconds
	if responseTime > 0 {
		check.ResponseTimeMs = sql.NullInt64{
			Int64: int64(responseTime.Milliseconds()),
			Valid: true,
		}
	}

	// Set error message if check failed
	if checkError != nil {
		check.ErrorMessage = sql.NullString{
			String: checkError.Error(),
			Valid:  true,
		}
	}

	// Check SSL certificate for HTTPS URLs if enabled (only for HTTP monitors)
	if monitor.Type != "tcp" && monitor.Type != "ping" && monitor.Type != "icmp" && monitor.CheckSSL && (len(monitor.URL) >= 5 && monitor.URL[:5] == "https") {
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
	if err := j.monitorRepo.SaveCheck(check); err != nil {
		if internal.Log != nil {
			internal.Log.Error("Failed to save check",
				zap.String("monitor_name", monitor.Name),
				zap.Error(err),
			)
		}
		return
	}

	// Update last_checked_at timestamp
	if err := j.monitorRepo.UpdateLastCheckedAt(monitor.ID, checkedAt); err != nil {
		if internal.Log != nil {
			internal.Log.Error("Failed to update last_checked_at",
				zap.String("monitor_name", monitor.Name),
				zap.Error(err),
			)
		}
		return
	}

	// Record metrics
	if success {
		internal.MonitorCheckTotal.WithLabelValues("success").Inc()
		if internal.Log != nil {
			if monitor.Type == "tcp" {
				internal.Log.Info("TCP monitor check successful",
					zap.String("monitor_name", monitor.Name),
					zap.String("address", monitor.URL),
					zap.Int64("response_time_ms", responseTime.Milliseconds()),
				)
			} else if monitor.Type == "ping" || monitor.Type == "icmp" {
				internal.Log.Info("Ping monitor check successful",
					zap.String("monitor_name", monitor.Name),
					zap.String("address", monitor.URL),
					zap.Int64("response_time_ms", responseTime.Milliseconds()),
				)
			} else {
				internal.Log.Info("HTTP monitor check successful",
					zap.String("monitor_name", monitor.Name),
					zap.Int("status_code", statusCode),
					zap.Int64("response_time_ms", responseTime.Milliseconds()),
				)
			}
		}
	} else {
		internal.MonitorCheckTotal.WithLabelValues("failure").Inc()
		if internal.Log != nil {
			if monitor.Type == "tcp" {
				internal.Log.Warn("TCP monitor check failed",
					zap.String("monitor_name", monitor.Name),
					zap.String("address", monitor.URL),
					zap.Error(checkError),
				)
			} else if monitor.Type == "ping" || monitor.Type == "icmp" {
				internal.Log.Warn("Ping monitor check failed",
					zap.String("monitor_name", monitor.Name),
					zap.String("address", monitor.URL),
					zap.Error(checkError),
				)
			} else {
				internal.Log.Warn("HTTP monitor check failed",
					zap.String("monitor_name", monitor.Name),
					zap.Error(checkError),
				)
			}
		}
	}

	// Broadcast monitor_check event
	j.broadcastMonitorCheckEvent(monitor, check)
}

// broadcastMonitorCheckEvent broadcasts a monitor check event to the backend SSE handler
func (j *HealthCheckJob) broadcastMonitorCheckEvent(monitor *entities.Monitor, check *entities.MonitorCheck) {
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
	broadcastEvent("monitor_check", data, monitor.UserID)
}

// parseTCPAddress parses a TCP address string and returns host and port
// Supports formats: "host:port" or "tcp://host:port"
func (j *HealthCheckJob) parseTCPAddress(address string) (string, int, error) {
	// Remove tcp:// prefix if present
	if len(address) > 6 && address[:6] == "tcp://" {
		address = address[6:]
	}

	// Parse host:port
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, fmt.Errorf("invalid address format: %w", err)
	}

	// Parse port number
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port number: %w", err)
	}

	if port < 1 || port > 65535 {
		return "", 0, fmt.Errorf("port number out of range: %d", port)
	}

	return host, port, nil
}
