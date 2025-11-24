package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetricsService_Creation(t *testing.T) {
	// Test that metrics service can be created
	service := NewMetricsService(nil)
	assert.NotNil(t, service, "NewMetricsService should not return nil")
}

func TestParsePeriodToDuration(t *testing.T) {
	tests := []struct {
		name     string
		period   string
		expected time.Duration
		hasError bool
	}{
		{
			name:     "24h period",
			period:   "24h",
			expected: 24 * time.Hour,
			hasError: false,
		},
		{
			name:     "7d period",
			period:   "7d",
			expected: 7 * 24 * time.Hour,
			hasError: false,
		},
		{
			name:     "30d period",
			period:   "30d",
			expected: 30 * 24 * time.Hour,
			hasError: false,
		},
		{
			name:     "invalid period",
			period:   "invalid",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := parsePeriodToDuration(tt.period)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, duration)
			}
		})
	}
}

func TestUptimeMetrics_Structure(t *testing.T) {
	metrics := &UptimeMetrics{
		Percentage:    99.5,
		TotalChecks:   200,
		SuccessChecks: 199,
		FailedChecks:  1,
	}

	assert.Equal(t, 99.5, metrics.Percentage)
	assert.Equal(t, 200, metrics.TotalChecks)
	assert.Equal(t, 199, metrics.SuccessChecks)
	assert.Equal(t, 1, metrics.FailedChecks)
}

func TestDataPoint_Structure(t *testing.T) {
	now := time.Now()
	dataPoint := &DataPoint{
		Timestamp: now,
		Value:     123.45,
	}

	assert.Equal(t, now, dataPoint.Timestamp)
	assert.Equal(t, 123.45, dataPoint.Value)
}

func TestStatusCodeDistribution_Structure(t *testing.T) {
	dist := &StatusCodeDistribution{
		StatusCode: 200,
		Count:      100,
	}

	assert.Equal(t, 200, dist.StatusCode)
	assert.Equal(t, 100, dist.Count)
}
