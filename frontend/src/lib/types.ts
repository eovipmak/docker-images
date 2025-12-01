export interface Monitor {
    id: string;
    tenant_id: number;
    name: string;
    url: string;
    type: 'http' | 'tcp' | 'ping' | 'icmp';
    keyword?: string;
    check_interval: number;
    timeout: number;
    enabled: boolean;
    check_ssl: boolean;
    ssl_alert_days: number;
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
