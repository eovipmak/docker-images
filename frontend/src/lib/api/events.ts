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
let reconnectTimer: number | null = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 10;
const INITIAL_RECONNECT_DELAY = 1000; // 1 second
const MAX_RECONNECT_DELAY = 30000; // 30 seconds

/**
 * Connect to SSE stream
 */
export function connectEventStream(): void {
	if (!browser) {
		console.log('[SSE] Not in browser, skipping connection');
		return;
	}

	// Don't reconnect if already connected
	if (eventSource && eventSource.readyState === EventSource.OPEN) {
		console.log('[SSE] Already connected');
		return;
	}

	// Get auth token
	const token = localStorage.getItem('auth_token');
	if (!token) {
		console.log('[SSE] No auth token, skipping connection');
		return;
	}

	console.log('[SSE] Connecting to event stream...');

	// Create EventSource with auth header
	// Note: EventSource doesn't support custom headers directly,
	// so we pass the token as a query parameter for server-side auth
	const url = `/api/v1/stream/events`;

	try {
		eventSource = new EventSource(url, {
			withCredentials: true
		});

		// Connection opened
		eventSource.addEventListener('open', () => {
			console.log('[SSE] Connected to event stream');
			reconnectAttempts = 0; // Reset reconnect attempts on successful connection
		});

		// Connection established
		eventSource.addEventListener('connected', (e) => {
			console.log('[SSE] Connection established:', e.data);
		});

		// Monitor check events
		eventSource.addEventListener('monitor_check', (e) => {
			try {
				const eventData = JSON.parse(e.data);
				const checkData = eventData.data as MonitorCheckEvent;

				console.log('[SSE] Monitor check event:', checkData);

				// Update store
				latestMonitorChecks.update((checks) => {
					const newChecks = new Map(checks);
					newChecks.set(checkData.monitor_id, checkData);
					return newChecks;
				});
			} catch (error) {
				console.error('[SSE] Error parsing monitor_check event:', error);
			}
		});

		// Incident created events
		eventSource.addEventListener('incident_created', (e) => {
			try {
				const eventData = JSON.parse(e.data);
				const incidentData = eventData.data as IncidentEvent;

				console.log('[SSE] Incident created event:', incidentData);

				// Add to incidents store
				latestIncidents.update((incidents) => {
					return [incidentData, ...incidents];
				});
			} catch (error) {
				console.error('[SSE] Error parsing incident_created event:', error);
			}
		});

		// Incident resolved events
		eventSource.addEventListener('incident_resolved', (e) => {
			try {
				const eventData = JSON.parse(e.data);
				const incidentData = eventData.data as IncidentEvent;

				console.log('[SSE] Incident resolved event:', incidentData);

				// Update incidents store
				latestIncidents.update((incidents) => {
					return incidents.filter(
						(inc) =>
							inc.monitor_id !== incidentData.monitor_id ||
							inc.alert_rule_id !== incidentData.alert_rule_id
					);
				});
			} catch (error) {
				console.error('[SSE] Error parsing incident_resolved event:', error);
			}
		});

		// Error handler
		eventSource.addEventListener('error', (e) => {
			console.error('[SSE] Error:', e);

			if (eventSource?.readyState === EventSource.CLOSED) {
				console.log('[SSE] Connection closed, attempting to reconnect...');
				scheduleReconnect();
			}
		});
	} catch (error) {
		console.error('[SSE] Failed to create EventSource:', error);
		scheduleReconnect();
	}
}

/**
 * Disconnect from SSE stream
 */
export function disconnectEventStream(): void {
	console.log('[SSE] Disconnecting from event stream');

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
			`[SSE] Max reconnection attempts (${MAX_RECONNECT_ATTEMPTS}) reached, giving up`
		);
		return;
	}

	// Calculate delay with exponential backoff
	const delay = Math.min(
		INITIAL_RECONNECT_DELAY * Math.pow(2, reconnectAttempts),
		MAX_RECONNECT_DELAY
	);

	reconnectAttempts++;

	console.log(`[SSE] Scheduling reconnect attempt ${reconnectAttempts} in ${delay}ms`);

	if (reconnectTimer) {
		clearTimeout(reconnectTimer);
	}

	reconnectTimer = setTimeout(() => {
		console.log(`[SSE] Reconnect attempt ${reconnectAttempts}`);
		connectEventStream();
	}, delay) as unknown as number;
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
