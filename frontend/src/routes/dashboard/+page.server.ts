import type { PageServerLoad } from './$types';
import { error } from '@sveltejs/kit';

interface MonitorCheck {
	id: string;
	monitor_id: string;
	checked_at: string;
	status_code?: number;
	response_time_ms?: number;
	ssl_valid?: boolean;
	ssl_expires_at?: string;
	error_message?: string;
	success: boolean;
}

interface Monitor {
	id: string;
	tenant_id: number;
	name: string;
	url: string;
	check_interval: number;
	timeout: number;
	enabled: boolean;
	check_ssl: boolean;
	ssl_alert_days: number;
	last_checked_at?: string;
	created_at: string;
	updated_at: string;
}

interface Incident {
	id: string;
	monitor_id: string;
	alert_rule_id: string;
	started_at: string;
	resolved_at?: string;
	status: string;
	trigger_value?: string;
	notified_at?: string;
	created_at: string;
}

interface DashboardStats {
	total_monitors: number;
	up_count: number;
	down_count: number;
	open_incidents: number;
}

interface MonitorCheckWithMonitor {
	check: MonitorCheck;
	monitor: Monitor;
}

interface IncidentWithDetails {
	incident: Incident;
	monitor: Monitor;
}

interface DashboardData {
	stats: DashboardStats;
	recent_checks: MonitorCheckWithMonitor[];
	open_incidents: IncidentWithDetails[];
}

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	const authToken = cookies.get('auth_token');
	
	if (!authToken) {
		throw error(401, 'Not authenticated');
	}

	try {
		const response = await fetch('/api/v1/dashboard', {
			headers: {
				'Authorization': `Bearer ${authToken}`
			}
		});

		if (!response.ok) {
			if (response.status === 401) {
				throw error(401, 'Not authenticated');
			}
			throw error(response.status, 'Failed to load dashboard data');
		}

		const data: DashboardData = await response.json();

		return {
			stats: data.stats,
			recentChecks: data.recent_checks,
			openIncidents: data.open_incidents
		};
	} catch (err) {
		console.error('Error loading dashboard:', err);
		
		// Return empty data if there's an error
		return {
			stats: {
				total_monitors: 0,
				up_count: 0,
				down_count: 0,
				open_incidents: 0
			},
			recentChecks: [],
			openIncidents: []
		};
	}
};
