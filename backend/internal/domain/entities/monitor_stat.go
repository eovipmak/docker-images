package entities

import (
	"time"
)

// MonitorStat represents aggregated response time statistics for a monitor
type MonitorStat struct {
	Timestamp      time.Time `db:"timestamp" json:"timestamp"`
	ResponseTimeMs int64     `db:"response_time_ms" json:"response_time_ms"`
}