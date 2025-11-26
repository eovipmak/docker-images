import { browser } from '$app/environment';
import { writable, type Writable } from 'svelte/store';

export interface SSEEvent {
	type: string;
	data: Record<string, any>;
	timestamp: string;
}

export interface MonitorCheckEvent {
	monitor_id: string;
	monitor_name: string;
	success: boolean;
	status_code?: number;
	response_time_ms?: number;
	error_message?: string;
	ssl_valid?: boolean;
	ssl_expires_at?: string;
	checked_at: string;
}

export interface IncidentEvent {
	monitor_id: string;
	alert_rule_id: string;
	rule_name: string;
	trigger_value?: string;
	status: 'open' | 'resolved';
}

// Store for latest events
export const latestMonitorChecks: Writable<Map<string, MonitorCheckEvent>> = writable(new Map());
export const latestIncidents: Writable<IncidentEvent[]> = writable([]);

// Event source connection
let eventSource: EventSource | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 10;
const INITIAL_RECONNECT_DELAY = 1000; // 1 second
const MAX_RECONNECT_DELAY = 30000; // 30 seconds

// Cache for public API URL
let cachedPublicApiUrl: string | null = null;

/**
 * Fetch public API URL from server config
 */
async function getPublicApiUrl(): Promise<string> {
	if (cachedPublicApiUrl !== null) {
		return cachedPublicApiUrl;
	}
	
	try {
		const response = await fetch('/api/config');
		if (response.ok) {
			const config = await response.json();
			cachedPublicApiUrl = config.publicApiUrl || '';
			return cachedPublicApiUrl as string;
		}
	} catch (error) {
		console.warn('Failed to fetch config:', error);
	}
	
	// Fallback: use window variable or auto-detect
	const windowUrl = (window as unknown as { __PUBLIC_API_URL__?: string }).__PUBLIC_API_URL__;
	if (windowUrl) {
		cachedPublicApiUrl = windowUrl;
		return cachedPublicApiUrl as string;
	}
	
	// Auto-detect based on current origin
	const currentOrigin = window.location.origin;
	if (currentOrigin.includes(':3000')) {
		cachedPublicApiUrl = currentOrigin.replace(':3000', ':8080');
	} else {
		cachedPublicApiUrl = currentOrigin.replace(/:\d+$/, '') + ':8080';
	}
	
	return cachedPublicApiUrl as string;
}

/**
 * Connect to SSE stream
 */
export async function connectEventStream(): Promise<void> {
	if (!browser) {
		return;
	}

	// Don't reconnect if already connected
	if (eventSource && eventSource.readyState === EventSource.OPEN) {
		return;
	}

	// Get auth token
	const token = localStorage.getItem('auth_token');
	if (!token) {
		return;
	}

	// Get the backend URL for SSE connection
	const backendUrl = await getPublicApiUrl();
	
	// Create EventSource with auth token as query parameter
	// EventSource doesn't support custom headers, so we pass token in URL
	const url = `${backendUrl}/api/v1/stream/events?token=${encodeURIComponent(token)}`;
	
	try {
		// Note: We don't use withCredentials since token is passed via URL query param
		// This avoids CORS issues with credentials and wildcard origins
		eventSource = new EventSource(url);

		// Connection opened
		eventSource.addEventListener('open', () => {
			reconnectAttempts = 0; // Reset reconnect attempts on successful connection
		});

		// Connection established
		eventSource.addEventListener('connected', (e) => {
		});

		// Monitor check events
		eventSource.addEventListener('monitor_check', (e) => {
			try {
				const eventData = JSON.parse(e.data);
				const checkData = eventData.data as MonitorCheckEvent;

				// Update store
				latestMonitorChecks.update((checks) => {
					const newChecks = new Map(checks);
					newChecks.set(checkData.monitor_id, checkData);
					return newChecks;
				});
			} catch (error) {
				console.error('Error parsing monitor_check event:', error);
			}
		});

		// Incident created events
		eventSource.addEventListener('incident_created', (e) => {
			try {
				const eventData = JSON.parse(e.data);
				const incidentData = eventData.data as IncidentEvent;

				// Add to incidents store
				latestIncidents.update((incidents) => {
					return [incidentData, ...incidents];
				});
			} catch (error) {
				console.error('Error parsing incident_created event:', error);
			}
		});

		// Incident resolved events
		eventSource.addEventListener('incident_resolved', (e) => {
			try {
				const eventData = JSON.parse(e.data);
				const incidentData = eventData.data as IncidentEvent;

				// Update incidents store
				latestIncidents.update((incidents) => {
					return incidents.filter(
						(inc) =>
							inc.monitor_id !== incidentData.monitor_id ||
							inc.alert_rule_id !== incidentData.alert_rule_id
					);
				});
			} catch (error) {
				console.error('Error parsing incident_resolved event:', error);
			}
		});

		// Error handler
		eventSource.addEventListener('error', (e) => {
			console.error('Error:', e);

			if (eventSource?.readyState === EventSource.CLOSED) {
				scheduleReconnect();
			}
		});
	} catch (error) {
		console.error('Failed to create EventSource:', error);
		scheduleReconnect();
	}
}

/**
 * Disconnect from SSE stream
 */
export function disconnectEventStream(): void {
	if (reconnectTimer) {
		clearTimeout(reconnectTimer);
		reconnectTimer = null;
	}

	if (eventSource) {
		eventSource.close();
		eventSource = null;
	}

	reconnectAttempts = 0;
}

/**
 * Schedule reconnection with exponential backoff
 */
function scheduleReconnect(): void {
	if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
		console.error(
			`Max reconnection attempts (${MAX_RECONNECT_ATTEMPTS}) reached, giving up`
		);
		return;
	}

	// Calculate delay with exponential backoff
	const delay = Math.min(
		INITIAL_RECONNECT_DELAY * Math.pow(2, reconnectAttempts),
		MAX_RECONNECT_DELAY
	);

	reconnectAttempts++;

	if (reconnectTimer) {
		clearTimeout(reconnectTimer);
	}

	reconnectTimer = setTimeout(() => {
		connectEventStream();
	}, delay);
}

/**
 * Get connection status
 */
export function getConnectionStatus(): 'connecting' | 'open' | 'closed' {
	if (!eventSource) return 'closed';

	switch (eventSource.readyState) {
		case EventSource.CONNECTING:
			return 'connecting';
		case EventSource.OPEN:
			return 'open';
		case EventSource.CLOSED:
			return 'closed';
		default:
			return 'closed';
	}
}

/**
 * Clear all event data
 */
export function clearEventData(): void {
	latestMonitorChecks.set(new Map());
	latestIncidents.set([]);
}
