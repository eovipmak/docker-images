export interface Monitor {
    id: string;
    user_id: number;
    name: string;
    url: string;
    type: 'http' | 'tcp' | 'ping' | 'icmp';
    keyword?: string;
    check_interval: number;
    timeout: number;
    enabled: boolean;
    check_ssl: boolean;
    ssl_alert_days: number;
    tags?: string[];
    expected_status_codes?: number[];
    last_checked_at?: string;
    created_at: string;
    updated_at: string;

    // UI specific
    status?: 'up' | 'down' | 'unknown';
    last_check?: MonitorCheck;
}

export interface MonitorCheck {
    id?: string;
    monitor_id: string;
    monitor_name?: string;
    checked_at: string;
    status_code?: number;
    response_time_ms?: number;
    ssl_valid?: boolean;
    ssl_expires_at?: string;
    error_message?: string;
    success: boolean;
}

export interface MaintenanceWindow {
    id: string;
    user_id: number;
    name: string;
    start_time: string;
    end_time: string;
    repeat_interval: number; // in seconds, 0 = one-time
    monitor_ids: string[];
    tags: string[];
    created_at: string;
    updated_at: string;
}

