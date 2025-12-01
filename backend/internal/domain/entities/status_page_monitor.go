package entities

// StatusPageMonitor represents the many-to-many relationship between status pages and monitors
type StatusPageMonitor struct {
	StatusPageID string `db:"status_page_id" json:"status_page_id"`
	MonitorID    string `db:"monitor_id" json:"monitor_id"`
}