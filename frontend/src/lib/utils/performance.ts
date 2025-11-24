/**
 * Performance measurement utilities for V-Insight
 * Provides functions to measure and log web vitals metrics
 */

/**
 * Web Vitals metric interface
 */
export interface WebVitalMetric {
	name: string;
	value: number;
	rating: 'good' | 'needs-improvement' | 'poor';
	delta: number;
	id: string;
}

/**
 * Performance metrics interface
 */
export interface PerformanceMetrics {
	fcp: number | null;
	lcp: number | null;
	fid: number | null;
	cls: number | null;
	tti: number | null;
	ttfb: number | null;
}

/**
 * Measure First Contentful Paint (FCP)
 * FCP measures the time from page load to when the first content is painted
 * @returns Promise<number | null> - FCP value in milliseconds or null if not available
 */
export async function measureFCP(): Promise<number | null> {
	if (typeof window === 'undefined' || !window.performance) {
		return null;
	}

	return new Promise((resolve) => {
		// Check if FCP is already available
		const entries = performance.getEntriesByName('first-contentful-paint');
		if (entries.length > 0) {
			resolve(Math.round(entries[0].startTime));
			return;
		}

		// Use PerformanceObserver to wait for FCP
		if ('PerformanceObserver' in window) {
			const observer = new PerformanceObserver((entryList) => {
				const entries = entryList.getEntriesByName('first-contentful-paint');
				if (entries.length > 0) {
					observer.disconnect();
					resolve(Math.round(entries[0].startTime));
				}
			});

			try {
				observer.observe({ type: 'paint', buffered: true });
				// Timeout after 10 seconds
				setTimeout(() => {
					observer.disconnect();
					resolve(null);
				}, 10000);
			} catch {
				resolve(null);
			}
		} else {
			resolve(null);
		}
	});
}

/**
 * Measure Time to Interactive (TTI)
 * TTI measures when the page becomes fully interactive
 * Note: This is an approximation using domInteractive and loadEventEnd
 * @returns Promise<number | null> - TTI value in milliseconds or null if not available
 */
export async function measureTTI(): Promise<number | null> {
	if (typeof window === 'undefined' || !window.performance) {
		return null;
	}

	return new Promise((resolve) => {
		const checkTTI = () => {
			const timing = performance.timing || performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
			
			if (timing) {
				// Use Navigation Timing API v2 if available
				if ('domInteractive' in timing && 'fetchStart' in timing) {
					const navTiming = timing as PerformanceNavigationTiming;
					if (navTiming.domInteractive > 0) {
						// TTI approximation: time when DOM is interactive
						const tti = Math.round(navTiming.domInteractive - navTiming.fetchStart);
						resolve(tti);
						return;
					}
				}
				
				// Fallback to legacy timing API
				const legacyTiming = performance.timing;
				if (legacyTiming && legacyTiming.domInteractive > 0) {
					const tti = Math.round(legacyTiming.domInteractive - legacyTiming.navigationStart);
					resolve(tti);
					return;
				}
			}
			
			// Wait for page to finish loading
			if (document.readyState === 'complete') {
				// Use loadEventEnd as a fallback
				const entries = performance.getEntriesByType('navigation') as PerformanceNavigationTiming[];
				if (entries.length > 0 && entries[0].domInteractive > 0) {
					resolve(Math.round(entries[0].domInteractive));
				} else {
					resolve(null);
				}
			} else {
				// Retry after a short delay
				setTimeout(checkTTI, 100);
			}
		};

		if (document.readyState === 'complete') {
			checkTTI();
		} else {
			window.addEventListener('load', () => {
				// Wait a bit after load for interactive state
				setTimeout(checkTTI, 100);
			});
		}

		// Timeout after 30 seconds
		setTimeout(() => resolve(null), 30000);
	});
}

/**
 * Measure Largest Contentful Paint (LCP)
 * LCP measures when the largest content element is painted
 * @returns Promise<number | null> - LCP value in milliseconds or null if not available
 */
export async function measureLCP(): Promise<number | null> {
	if (typeof window === 'undefined' || !('PerformanceObserver' in window)) {
		return null;
	}

	return new Promise((resolve) => {
		let lcpValue: number | null = null;

		const observer = new PerformanceObserver((entryList) => {
			const entries = entryList.getEntries();
			if (entries.length > 0) {
				// LCP is the last entry (largest content painted so far)
				lcpValue = Math.round(entries[entries.length - 1].startTime);
			}
		});

		try {
			observer.observe({ type: 'largest-contentful-paint', buffered: true });

			// Resolve on page visibility change or after timeout
			const resolveAndDisconnect = () => {
				observer.disconnect();
				resolve(lcpValue);
			};

			document.addEventListener('visibilitychange', () => {
				if (document.visibilityState === 'hidden') {
					resolveAndDisconnect();
				}
			}, { once: true });

			// Timeout after 10 seconds
			setTimeout(resolveAndDisconnect, 10000);
		} catch {
			resolve(null);
		}
	});
}

/**
 * Measure Time to First Byte (TTFB)
 * TTFB measures the time until the first byte of the response is received
 * @returns number | null - TTFB value in milliseconds or null if not available
 */
export function measureTTFB(): number | null {
	if (typeof window === 'undefined' || !window.performance) {
		return null;
	}

	const entries = performance.getEntriesByType('navigation') as PerformanceNavigationTiming[];
	if (entries.length > 0) {
		return Math.round(entries[0].responseStart - entries[0].requestStart);
	}

	// Fallback to legacy timing API
	const timing = performance.timing;
	if (timing && timing.responseStart > 0) {
		return Math.round(timing.responseStart - timing.requestStart);
	}

	return null;
}

/**
 * Log web vitals metrics to console (development only)
 * @param metrics - Object containing performance metrics
 */
export function logWebVitals(metrics: Partial<PerformanceMetrics>): void {
	if (typeof window === 'undefined') {
		return;
	}

	const formatValue = (value: number | null | undefined, unit = 'ms'): string => {
		if (value === null || value === undefined) return 'N/A';
		return `${value}${unit}`;
	};

	const getRating = (name: string, value: number): 'good' | 'needs-improvement' | 'poor' => {
		const thresholds: Record<string, [number, number]> = {
			fcp: [1800, 3000],
			lcp: [2500, 4000],
			fid: [100, 300],
			cls: [0.1, 0.25],
			tti: [3800, 7300],
			ttfb: [800, 1800]
		};

		const [good, poor] = thresholds[name] || [0, 0];
		if (value <= good) return 'good';
		if (value <= poor) return 'needs-improvement';
		return 'poor';
	};

	const getRatingEmoji = (rating: 'good' | 'needs-improvement' | 'poor'): string => {
		switch (rating) {
			case 'good': return '🟢';
			case 'needs-improvement': return '🟡';
			case 'poor': return '🔴';
		}
	};

	console.group('📊 Web Vitals');
	
	if (metrics.fcp !== undefined && metrics.fcp !== null) {
		const rating = getRating('fcp', metrics.fcp);
		console.log(`${getRatingEmoji(rating)} FCP (First Contentful Paint): ${formatValue(metrics.fcp)}`);
	}
	
	if (metrics.lcp !== undefined && metrics.lcp !== null) {
		const rating = getRating('lcp', metrics.lcp);
		console.log(`${getRatingEmoji(rating)} LCP (Largest Contentful Paint): ${formatValue(metrics.lcp)}`);
	}
	
	if (metrics.tti !== undefined && metrics.tti !== null) {
		const rating = getRating('tti', metrics.tti);
		console.log(`${getRatingEmoji(rating)} TTI (Time to Interactive): ${formatValue(metrics.tti)}`);
	}
	
	if (metrics.ttfb !== undefined && metrics.ttfb !== null) {
		const rating = getRating('ttfb', metrics.ttfb);
		console.log(`${getRatingEmoji(rating)} TTFB (Time to First Byte): ${formatValue(metrics.ttfb)}`);
	}
	
	if (metrics.cls !== undefined && metrics.cls !== null) {
		const rating = getRating('cls', metrics.cls);
		console.log(`${getRatingEmoji(rating)} CLS (Cumulative Layout Shift): ${metrics.cls.toFixed(3)}`);
	}
	
	console.groupEnd();
}

/**
 * Collect all available web vitals metrics
 * @returns Promise<PerformanceMetrics> - Object containing all metrics
 */
export async function collectWebVitals(): Promise<PerformanceMetrics> {
	const [fcp, lcp, tti] = await Promise.all([
		measureFCP(),
		measureLCP(),
		measureTTI()
	]);

	const ttfb = measureTTFB();

	return {
		fcp,
		lcp,
		fid: null, // FID requires user interaction
		cls: null, // CLS requires observation over time
		tti,
		ttfb
	};
}

/**
 * Initialize performance monitoring
 * Call this function once on app startup to begin collecting metrics
 * @param options - Configuration options
 */
export function initPerformanceMonitoring(options: { logToConsole?: boolean } = {}): void {
	if (typeof window === 'undefined') {
		return;
	}

	const { logToConsole = false } = options;

	// Collect metrics after page load
	if (document.readyState === 'complete') {
		setTimeout(async () => {
			const metrics = await collectWebVitals();
			if (logToConsole) {
				logWebVitals(metrics);
			}
		}, 1000);
	} else {
		window.addEventListener('load', async () => {
			// Wait for page to stabilize
			setTimeout(async () => {
				const metrics = await collectWebVitals();
				if (logToConsole) {
					logWebVitals(metrics);
				}
			}, 1000);
		});
	}
}
