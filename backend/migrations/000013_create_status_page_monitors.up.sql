-- Create status_page_monitors linking table
CREATE TABLE IF NOT EXISTS status_page_monitors (
    status_page_id UUID NOT NULL,
    monitor_id UUID NOT NULL,
    PRIMARY KEY (status_page_id, monitor_id),
    CONSTRAINT fk_status_page_monitors_status_page FOREIGN KEY (status_page_id) REFERENCES status_pages(id) ON DELETE CASCADE,
    CONSTRAINT fk_status_page_monitors_monitor FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
);

-- Create index on status_page_id for faster queries
CREATE INDEX IF NOT EXISTS idx_status_page_monitors_status_page_id ON status_page_monitors(status_page_id);

-- Create index on monitor_id for faster queries
CREATE INDEX IF NOT EXISTS idx_status_page_monitors_monitor_id ON status_page_monitors(monitor_id);