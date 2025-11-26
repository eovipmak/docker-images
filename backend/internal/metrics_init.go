package internal

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// MonitorCheckTotal tracks total number of monitor checks
	MonitorCheckTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "monitor_check_total",
			Help: "Total number of monitor checks performed",
		},
		[]string{"status"}, // success, failure
	)

	// MonitorCheckFailed tracks failed monitor checks
	MonitorCheckFailed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "monitor_check_failed_total",
			Help: "Total number of failed monitor checks",
		},
	)

	// MonitorCount tracks current number of monitors
	MonitorCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "monitor_count",
			Help: "Current number of active monitors",
		},
	)

	// OpenIncidentsCount tracks number of open incidents
	OpenIncidentsCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "open_incidents_count",
			Help: "Current number of open incidents",
		},
	)

	// MonitorCheckDuration tracks duration of monitor checks
	MonitorCheckDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "monitor_check_duration_seconds",
			Help:    "Duration of monitor checks in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"monitor_id"},
	)

	// MonitorResponseTime tracks HTTP response times
	MonitorResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "monitor_response_time_ms",
			Help:    "HTTP response time in milliseconds",
			Buckets: []float64{50, 100, 200, 500, 1000, 2000, 5000, 10000},
		},
		[]string{"monitor_id"},
	)

	// HTTPRequestDuration tracks API request duration
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestTotal tracks total HTTP requests
	HTTPRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
)
