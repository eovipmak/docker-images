package internal

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// JobExecutionTotal tracks total number of job executions
	JobExecutionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_job_execution_total",
			Help: "Total number of worker job executions",
		},
		[]string{"job_name", "status"}, // success, failure
	)

	// JobExecutionDuration tracks duration of job executions
	JobExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "worker_job_execution_duration_seconds",
			Help:    "Duration of worker job executions in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"job_name"},
	)

	// MonitorCheckTotal tracks total number of monitor checks
	MonitorCheckTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_monitor_check_total",
			Help: "Total number of monitor checks performed by worker",
		},
		[]string{"status"}, // success, failure
	)

	// AlertEvaluationTotal tracks total number of alert evaluations
	AlertEvaluationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_alert_evaluation_total",
			Help: "Total number of alert evaluations",
		},
		[]string{"status"}, // success, failure
	)

	// IncidentCreated tracks incidents created
	IncidentCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_incident_created_total",
			Help: "Total number of incidents created",
		},
	)

	// IncidentResolved tracks incidents resolved
	IncidentResolved = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_incident_resolved_total",
			Help: "Total number of incidents resolved",
		},
	)

	// NotificationSent tracks notifications sent
	NotificationSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_notification_sent_total",
			Help: "Total number of notifications sent",
		},
		[]string{"channel_type", "status"}, // webhook, discord, email; success, failure
	)
)
