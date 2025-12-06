package service

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/eovipmak/v-insight/backend/internal"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// MetricsService provides business logic for metrics and analytics
type MetricsService struct {
	db *sqlx.DB
}

// NewMetricsService creates a new metrics service
func NewMetricsService(db *sqlx.DB) *MetricsService {
	return &MetricsService{
		db: db,
	}
}

// DataPoint represents a time-series data point
type DataPoint struct {
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Value     float64   `json:"value" db:"value"`
}

// UptimeMetrics represents uptime statistics
type UptimeMetrics struct {
	Percentage    float64 `json:"percentage"`
	TotalChecks   int     `json:"total_checks"`
	SuccessChecks int     `json:"success_checks"`
	FailedChecks  int     `json:"failed_checks"`
}

// StatusCodeDistribution represents status code counts
type StatusCodeDistribution struct {
	StatusCode int `json:"status_code" db:"status_code"`
	Count      int `json:"count" db:"count"`
}

// CalculateUptime calculates uptime percentage for a monitor over a period
func (s *MetricsService) CalculateUptime(monitorID string, period string) (*UptimeMetrics, error) {
	duration, err := parsePeriodToDuration(period)
	if err != nil {
		return nil, err
	}

	startTime := time.Now().UTC().Add(-duration)

	// Debugging: log start time and last check timestamp for this monitor
	if internal.Log != nil {
		var latestChecked sql.NullTime
		_ = s.db.Get(&latestChecked, `SELECT MAX(checked_at) FROM monitor_checks WHERE monitor_id = $1`, monitorID)
		var totalAll int
		_ = s.db.Get(&totalAll, `SELECT COUNT(*) FROM monitor_checks WHERE monitor_id = $1`, monitorID)
		var countInRange int
		_ = s.db.Get(&countInRange, `SELECT COUNT(*) FROM monitor_checks WHERE monitor_id = $1 AND checked_at >= $2`, monitorID, startTime)
		// Only log if available
		serverTZ := time.Now().Location().String()
		if latestChecked.Valid {
			internal.Log.Debug("CalculateUptime: time range", zap.String("monitor_id", monitorID), zap.Time("start_time", startTime), zap.Time("latest_check", latestChecked.Time), zap.String("server_tz", serverTZ), zap.String("latest_check_tz", latestChecked.Time.Location().String()), zap.Int("total_checks_all_time", totalAll), zap.Int("count_in_range", countInRange))
		} else {
			internal.Log.Debug("CalculateUptime: time range", zap.String("monitor_id", monitorID), zap.Time("start_time", startTime), zap.String("latest_check", "<none>"), zap.String("server_tz", serverTZ), zap.Int("total_checks_all_time", totalAll), zap.Int("count_in_range", countInRange))
		}
	}

	query := `
		SELECT 
			COUNT(*) as total_checks,
			COALESCE(SUM(CASE WHEN success = true THEN 1 ELSE 0 END), 0) as success_checks,
			COALESCE(SUM(CASE WHEN success = false THEN 1 ELSE 0 END), 0) as failed_checks
		FROM monitor_checks
		WHERE monitor_id = $1
		  AND checked_at >= $2
	`

	var totalChecks, successChecks, failedChecks int
	err = s.db.QueryRow(query, monitorID, startTime).Scan(&totalChecks, &successChecks, &failedChecks)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate uptime: %w", err)
	}

	var percentage float64
	if totalChecks > 0 {
		percentage = (float64(successChecks) / float64(totalChecks)) * 100
	}

	return &UptimeMetrics{
		Percentage:    percentage,
		TotalChecks:   totalChecks,
		SuccessChecks: successChecks,
		FailedChecks:  failedChecks,
	}, nil
}

// GetResponseTimeHistory retrieves response time data points over a period
func (s *MetricsService) GetResponseTimeHistory(monitorID string, period string) ([]DataPoint, error) {
	duration, err := parsePeriodToDuration(period)
	if err != nil {
		return nil, err
	}

	startTime := time.Now().UTC().Add(-duration)

	// Use time buckets for aggregation to reduce data points
	// For 24h, use 5-minute buckets; for 7d, use 1-hour buckets; for 30d, use 6-hour buckets
	var intervalSeconds int
	switch period {
	case "1h":
		intervalSeconds = 60 // 1 minute
	case "6h":
		intervalSeconds = 300 // 5 minutes
	case "12h":
		intervalSeconds = 300 // 5 minutes
	case "24h":
		intervalSeconds = 300 // 5 minutes
	case "7d", "1w":
		intervalSeconds = 3600 // 1 hour
	case "30d":
		intervalSeconds = 21600 // 6 hours
	default:
		intervalSeconds = 300
	}

	// Use PostgreSQL's date_trunc and EXTRACT(EPOCH) for time bucketing
	query := `
		SELECT 
			TO_TIMESTAMP(FLOOR(EXTRACT(EPOCH FROM checked_at) / $1) * $1) as timestamp,
			AVG(response_time_ms) as value
		FROM monitor_checks
		WHERE monitor_id = $2
		  AND checked_at >= $3
		  AND response_time_ms IS NOT NULL
		  AND success = true
		GROUP BY FLOOR(EXTRACT(EPOCH FROM checked_at) / $1)
		ORDER BY timestamp ASC
	`

	var dataPoints []DataPoint
	err = s.db.Select(&dataPoints, query, intervalSeconds, monitorID, startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get response time history: %w", err)
	}

	return dataPoints, nil
}

// GetStatusCodeDistribution retrieves status code distribution over a period
func (s *MetricsService) GetStatusCodeDistribution(monitorID string, period string) ([]StatusCodeDistribution, error) {
	duration, err := parsePeriodToDuration(period)
	if err != nil {
		return nil, err
	}

	startTime := time.Now().UTC().Add(-duration)

	query := `
		SELECT 
			status_code,
			COUNT(*) as count
		FROM monitor_checks
		WHERE monitor_id = $1
		  AND checked_at >= $2
		  AND status_code IS NOT NULL
		GROUP BY status_code
		ORDER BY status_code ASC
	`

	var distribution []StatusCodeDistribution
	err = s.db.Select(&distribution, query, monitorID, startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get status code distribution: %w", err)
	}

	return distribution, nil
}

// GetAverageResponseTime calculates average response time for a monitor over a period
func (s *MetricsService) GetAverageResponseTime(monitorID string, period string) (float64, error) {
	duration, err := parsePeriodToDuration(period)
	if err != nil {
		return 0, err
	}

	startTime := time.Now().UTC().Add(-duration)

	query := `
		SELECT COALESCE(AVG(response_time_ms), 0) as avg_response_time
		FROM monitor_checks
		WHERE monitor_id = $1
		  AND checked_at >= $2
		  AND response_time_ms IS NOT NULL
		  AND success = true
	`

	var avgResponseTime float64
	err = s.db.QueryRow(query, monitorID, startTime).Scan(&avgResponseTime)
	if err != nil {
		return 0, fmt.Errorf("failed to get average response time: %w", err)
	}

	return avgResponseTime, nil
}

// GetGlobalAverageResponseTime calculates average response time across all monitors for a user
func (s *MetricsService) GetGlobalAverageResponseTime(userID int, period string) (float64, error) {
	duration, err := parsePeriodToDuration(period)
	if err != nil {
		return 0, err
	}

	startTime := time.Now().UTC().Add(-duration)

	query := `
		SELECT COALESCE(AVG(mc.response_time_ms), 0) as avg_response_time
		FROM monitor_checks mc
		INNER JOIN monitors m ON mc.monitor_id = m.id
		WHERE m.user_id = $1
		  AND mc.checked_at >= $2
		  AND mc.response_time_ms IS NOT NULL
		  AND mc.success = true
	`

	var avgResponseTime float64
	err = s.db.QueryRow(query, userID, startTime).Scan(&avgResponseTime)
	if err != nil {
		return 0, fmt.Errorf("failed to get global average response time: %w", err)
	}

	return avgResponseTime, nil
}

// GetGlobalUptime calculates overall uptime across all monitors for a user
func (s *MetricsService) GetGlobalUptime(userID int, period string) (*UptimeMetrics, error) {
	duration, err := parsePeriodToDuration(period)
	if err != nil {
		return nil, err
	}

	startTime := time.Now().Add(-duration)

	query := `
		SELECT 
			COUNT(*) as total_checks,
			COALESCE(SUM(CASE WHEN mc.success = true THEN 1 ELSE 0 END), 0) as success_checks,
			COALESCE(SUM(CASE WHEN mc.success = false THEN 1 ELSE 0 END), 0) as failed_checks
		FROM monitor_checks mc
		INNER JOIN monitors m ON mc.monitor_id = m.id
		WHERE m.user_id = $1
		  AND mc.checked_at >= $2
	`

	var totalChecks, successChecks, failedChecks int
	err = s.db.QueryRow(query, userID, startTime).Scan(&totalChecks, &successChecks, &failedChecks)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate global uptime: %w", err)
	}

	var percentage float64
	if totalChecks > 0 {
		percentage = (float64(successChecks) / float64(totalChecks)) * 100
	}

	return &UptimeMetrics{
		Percentage:    percentage,
		TotalChecks:   totalChecks,
		SuccessChecks: successChecks,
		FailedChecks:  failedChecks,
	}, nil
}

// parsePeriodToDuration converts period string to time.Duration
func parsePeriodToDuration(period string) (time.Duration, error) {
	switch period {
	case "1h":
		return time.Hour, nil
	case "6h":
		return 6 * time.Hour, nil
	case "12h":
		return 12 * time.Hour, nil
	case "24h":
		return 24 * time.Hour, nil
	case "7d", "1w":
		return 7 * 24 * time.Hour, nil
	case "30d":
		return 30 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid period: %s (must be 1h, 6h, 12h, 24h, 1w/7d, or 30d)", period)
	}
}
