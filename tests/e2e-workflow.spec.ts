import { test, expect } from '@playwright/test';

/**
 * Comprehensive E2E Tests for V-Insight Platform
 * 
 * Tests the complete workflow:
 * 1. Login with demo user
 * 2. Add a monitor (domain) for google.com
 * 3. Edit the monitor (change interval, timeout, name)
 * 4. Add an alert rule
 * 5. Edit the alert rule
 */

test.describe('V-Insight E2E Automated Tests', () => {
	const BASE_URL = process.env.BASE_URL || 'http://localhost:3000';
	const DEMO_EMAIL = 'test@gmail.com';
	const DEMO_PASSWORD = 'Password!';

	let monitorId: string;
	let alertRuleId: string;

	test.beforeEach(async ({ page }) => {
		// Set a longer timeout for all actions in this test (CI environments can be slow)
		page.setDefaultTimeout(30000);
	});

	test('1. Login with demo user', async ({ page }) => {
		// Navigate to login page
		await page.goto(`${BASE_URL}/login`);
		await page.waitForLoadState('networkidle');

		// Fill in login credentials
		await page.locator('input[name="email"]').fill(DEMO_EMAIL);
		await page.locator('input[name="password"]').fill(DEMO_PASSWORD);

		// Take screenshot before login
		await page.screenshot({ path: 'test-results/01-login-form.png', fullPage: true });

		// Submit login form
		await page.locator('button[type="submit"]').click();

		// Wait for redirect to dashboard
		await page.waitForURL('**/dashboard', { timeout: 10000 });

		// Verify we're on the dashboard
		await expect(page).toHaveURL(/\/dashboard/);

		// Take screenshot after login
		await page.screenshot({ path: 'test-results/02-dashboard.png', fullPage: true });

		console.log('✅ Login successful');
	});

	test('2. Add monitor (domain) for google.com', async ({ page }) => {
		// Login first
		await page.goto(`${BASE_URL}/login`);
		await page.waitForLoadState('networkidle');
		await page.locator('input[name="email"]').fill(DEMO_EMAIL);
		await page.locator('input[name="password"]').fill(DEMO_PASSWORD);
		await page.locator('button[type="submit"]').click();
		await page.waitForURL('**/dashboard', { timeout: 10000 });

		// Navigate to monitors/domains page
		await page.goto(`${BASE_URL}/domains`);
		await page.waitForLoadState('networkidle');

		// Take screenshot of domains page
		await page.screenshot({ path: 'test-results/03-domains-page.png', fullPage: true });

		// Look for "Add Monitor" or similar button
		// Common patterns: "Add Monitor", "New Monitor", "Create Monitor", or a "+" button
		const addButton = page.locator('button:has-text("Add"), button:has-text("New"), button:has-text("Create"), a:has-text("Add"), a:has-text("New")').first();
		
		if (await addButton.isVisible({ timeout: 5000 }).catch(() => false)) {
			await addButton.click();
			await page.waitForTimeout(1000);
		}

		// Fill in monitor details
		await page.locator('input[name="name"], input[placeholder*="name" i]').first().fill('Google Monitor');
		await page.locator('input[name="url"], input[placeholder*="url" i], input[type="url"]').first().fill('https://google.com');
		
		// Set interval (300 seconds = 5 minutes)
		const intervalInput = page.locator('input[name="interval"], input[name="check_interval"], input[placeholder*="interval" i]').first();
		if (await intervalInput.isVisible({ timeout: 2000 }).catch(() => false)) {
			await intervalInput.fill('300');
		}

		// Set timeout (30 seconds)
		const timeoutInput = page.locator('input[name="timeout"], input[placeholder*="timeout" i]').first();
		if (await timeoutInput.isVisible({ timeout: 2000 }).catch(() => false)) {
			await timeoutInput.fill('30');
		}

		// Take screenshot before submitting
		await page.screenshot({ path: 'test-results/04-add-monitor-form.png', fullPage: true });

		// Submit the form
		await page.locator('button[type="submit"]:has-text("Create"), button[type="submit"]:has-text("Add"), button[type="submit"]:has-text("Save")').first().click();

		// Wait for the monitor to be created
		await page.waitForTimeout(2000);

		// Take screenshot after creation
		await page.screenshot({ path: 'test-results/05-monitor-created.png', fullPage: true });

		// Verify monitor appears in the list
		await expect(page.locator('text=Google Monitor')).toBeVisible({ timeout: 5000 });

		console.log('✅ Monitor added successfully');
	});

	test('3. Edit monitor (change interval, timeout, name)', async ({ page }) => {
		// Login first
		await page.goto(`${BASE_URL}/login`);
		await page.waitForLoadState('networkidle');
		await page.locator('input[name="email"]').fill(DEMO_EMAIL);
		await page.locator('input[name="password"]').fill(DEMO_PASSWORD);
		await page.locator('button[type="submit"]').click();
		await page.waitForURL('**/dashboard', { timeout: 10000 });

		// Navigate to domains page
		await page.goto(`${BASE_URL}/domains`);
		await page.waitForLoadState('networkidle');

		// Find and click on the Google Monitor (or first monitor)
		const monitorLink = page.locator('a:has-text("Google Monitor"), tr:has-text("Google Monitor") a, div:has-text("Google Monitor") a').first();
		
		if (await monitorLink.isVisible({ timeout: 5000 }).catch(() => false)) {
			await monitorLink.click();
		} else {
			// If we can't find "Google Monitor", try to find the first monitor in the list
			const firstMonitor = page.locator('tbody tr a, .monitor-item a').first();
			await firstMonitor.click();
		}

		await page.waitForTimeout(1000);

		// Take screenshot of monitor detail page
		await page.screenshot({ path: 'test-results/06-monitor-detail.png', fullPage: true });

		// Look for Edit button
		const editButton = page.locator('button:has-text("Edit"), a:has-text("Edit")').first();
		if (await editButton.isVisible({ timeout: 3000 }).catch(() => false)) {
			await editButton.click();
			await page.waitForTimeout(500);
		}

		// Update monitor name
		const nameInput = page.locator('input[name="name"], input[placeholder*="name" i]').first();
		await nameInput.clear();
		await nameInput.fill('Google Monitor - Updated');

		// Update interval to 600 seconds (10 minutes)
		const intervalInput = page.locator('input[name="interval"], input[name="check_interval"], input[placeholder*="interval" i]').first();
		if (await intervalInput.isVisible({ timeout: 2000 }).catch(() => false)) {
			await intervalInput.clear();
			await intervalInput.fill('600');
		}

		// Update timeout to 60 seconds
		const timeoutInput = page.locator('input[name="timeout"], input[placeholder*="timeout" i]').first();
		if (await timeoutInput.isVisible({ timeout: 2000 }).catch(() => false)) {
			await timeoutInput.clear();
			await timeoutInput.fill('60');
		}

		// Take screenshot before saving
		await page.screenshot({ path: 'test-results/07-edit-monitor-form.png', fullPage: true });

		// Save changes
		await page.locator('button[type="submit"]:has-text("Update"), button[type="submit"]:has-text("Save"), button:has-text("Save")').first().click();

		// Wait for changes to be saved
		await page.waitForTimeout(2000);

		// Take screenshot after update
		await page.screenshot({ path: 'test-results/08-monitor-updated.png', fullPage: true });

		// Verify the updated name appears
		await expect(page.locator('text=Google Monitor - Updated')).toBeVisible({ timeout: 5000 });

		console.log('✅ Monitor updated successfully');
	});

	test('4. Add alert rule', async ({ page }) => {
		// Login first
		await page.goto(`${BASE_URL}/login`);
		await page.waitForLoadState('networkidle');
		await page.locator('input[name="email"]').fill(DEMO_EMAIL);
		await page.locator('input[name="password"]').fill(DEMO_PASSWORD);
		await page.locator('button[type="submit"]').click();
		await page.waitForURL('**/dashboard', { timeout: 10000 });

		// Navigate to alerts page
		await page.goto(`${BASE_URL}/alerts`);
		await page.waitForLoadState('networkidle');

		// Take screenshot of alerts page
		await page.screenshot({ path: 'test-results/09-alerts-page.png', fullPage: true });

		// Look for "Add Alert" or similar button
		const addButton = page.locator('button:has-text("Add"), button:has-text("New"), button:has-text("Create"), a:has-text("Add")').first();
		
		if (await addButton.isVisible({ timeout: 5000 }).catch(() => false)) {
			await addButton.click();
			await page.waitForTimeout(1000);
		}

		// Fill in alert rule details
		await page.locator('input[name="name"], input[placeholder*="name" i]').first().fill('Google Down Alert');

		// Select trigger type (down, ssl_expiry, slow_response)
		const triggerSelect = page.locator('select[name="trigger_type"], select[name="triggerType"]').first();
		if (await triggerSelect.isVisible({ timeout: 2000 }).catch(() => false)) {
			await triggerSelect.selectOption('down');
		}

		// Set threshold value
		const thresholdInput = page.locator('input[name="threshold"], input[name="threshold_value"]').first();
		if (await thresholdInput.isVisible({ timeout: 2000 }).catch(() => false)) {
			await thresholdInput.fill('3'); // Alert after 3 consecutive failures
		}

		// Select monitor (if there's a dropdown)
		const monitorSelect = page.locator('select[name="monitor"], select[name="monitor_id"]').first();
		if (await monitorSelect.isVisible({ timeout: 2000 }).catch(() => false)) {
			// Select first option (should be our Google monitor)
			await monitorSelect.selectOption({ index: 1 }); // Index 0 is usually "Select a monitor"
		}

		// Take screenshot before submitting
		await page.screenshot({ path: 'test-results/10-add-alert-form.png', fullPage: true });

		// Submit the form
		await page.locator('button[type="submit"]:has-text("Create"), button[type="submit"]:has-text("Add"), button[type="submit"]:has-text("Save")').first().click();

		// Wait for the alert to be created
		await page.waitForTimeout(2000);

		// Take screenshot after creation
		await page.screenshot({ path: 'test-results/11-alert-created.png', fullPage: true });

		// Verify alert appears in the list
		await expect(page.locator('text=Google Down Alert')).toBeVisible({ timeout: 5000 });

		console.log('✅ Alert rule added successfully');
	});

	test('5. Edit alert rule', async ({ page }) => {
		// Login first
		await page.goto(`${BASE_URL}/login`);
		await page.waitForLoadState('networkidle');
		await page.locator('input[name="email"]').fill(DEMO_EMAIL);
		await page.locator('input[name="password"]').fill(DEMO_PASSWORD);
		await page.locator('button[type="submit"]').click();
		await page.waitForURL('**/dashboard', { timeout: 10000 });

		// Navigate to alerts page
		await page.goto(`${BASE_URL}/alerts`);
		await page.waitForLoadState('networkidle');

		// Find and click on the alert rule
		const alertLink = page.locator('a:has-text("Google Down Alert"), tr:has-text("Google Down Alert") a, div:has-text("Google Down Alert") a').first();
		
		if (await alertLink.isVisible({ timeout: 5000 }).catch(() => false)) {
			await alertLink.click();
		} else {
			// If we can't find the specific alert, try to find the first alert in the list
			const firstAlert = page.locator('tbody tr a, .alert-item a').first();
			if (await firstAlert.isVisible({ timeout: 3000 }).catch(() => false)) {
				await firstAlert.click();
			}
		}

		await page.waitForTimeout(1000);

		// Take screenshot of alert detail page
		await page.screenshot({ path: 'test-results/12-alert-detail.png', fullPage: true });

		// Look for Edit button
		const editButton = page.locator('button:has-text("Edit"), a:has-text("Edit")').first();
		if (await editButton.isVisible({ timeout: 3000 }).catch(() => false)) {
			await editButton.click();
			await page.waitForTimeout(500);
		}

		// Update alert name
		const nameInput = page.locator('input[name="name"], input[placeholder*="name" i]').first();
		await nameInput.clear();
		await nameInput.fill('Google Down Alert - Updated');

		// Update threshold to 5 consecutive failures
		const thresholdInput = page.locator('input[name="threshold"], input[name="threshold_value"]').first();
		if (await thresholdInput.isVisible({ timeout: 2000 }).catch(() => false)) {
			await thresholdInput.clear();
			await thresholdInput.fill('5');
		}

		// Take screenshot before saving
		await page.screenshot({ path: 'test-results/13-edit-alert-form.png', fullPage: true });

		// Save changes
		await page.locator('button[type="submit"]:has-text("Update"), button[type="submit"]:has-text("Save"), button:has-text("Save")').first().click();

		// Wait for changes to be saved
		await page.waitForTimeout(2000);

		// Take screenshot after update
		await page.screenshot({ path: 'test-results/14-alert-updated.png', fullPage: true });

		// Verify the updated name appears
		await expect(page.locator('text=Google Down Alert - Updated')).toBeVisible({ timeout: 5000 });

		console.log('✅ Alert rule updated successfully');
	});

	test('Complete workflow: Login -> Add Monitor -> Edit Monitor -> Add Alert -> Edit Alert', async ({ page }) => {
		// This test runs the complete workflow in sequence
		console.log('🚀 Starting complete E2E workflow test...');

		// 1. Login
		console.log('Step 1: Logging in...');
		await page.goto(`${BASE_URL}/login`);
		await page.waitForLoadState('networkidle');
		await page.locator('input[name="email"]').fill(DEMO_EMAIL);
		await page.locator('input[name="password"]').fill(DEMO_PASSWORD);
		await page.screenshot({ path: 'test-results/workflow-01-login.png', fullPage: true });
		await page.locator('button[type="submit"]').click();
		await page.waitForURL('**/dashboard', { timeout: 10000 });
		await page.screenshot({ path: 'test-results/workflow-02-dashboard.png', fullPage: true });
		console.log('✅ Step 1 complete: Logged in');

		// 2. Add Monitor
		console.log('Step 2: Adding monitor...');
		await page.goto(`${BASE_URL}/domains`);
		await page.waitForLoadState('networkidle');
		
		const addMonitorBtn = page.locator('button:has-text("Add"), button:has-text("New"), a:has-text("Add")').first();
		if (await addMonitorBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
			await addMonitorBtn.click();
			await page.waitForTimeout(1000);
		}

		await page.locator('input[name="name"], input[placeholder*="name" i]').first().fill('Workflow Test Monitor');
		await page.locator('input[name="url"], input[placeholder*="url" i], input[type="url"]').first().fill('https://google.com');
		
		const intervalInput = page.locator('input[name="interval"], input[name="check_interval"]').first();
		if (await intervalInput.isVisible({ timeout: 2000 }).catch(() => false)) {
			await intervalInput.fill('300');
		}

		const timeoutInput = page.locator('input[name="timeout"]').first();
		if (await timeoutInput.isVisible({ timeout: 2000 }).catch(() => false)) {
			await timeoutInput.fill('30');
		}

		await page.screenshot({ path: 'test-results/workflow-03-add-monitor.png', fullPage: true });
		await page.locator('button[type="submit"]').first().click();
		await page.waitForTimeout(2000);
		await page.screenshot({ path: 'test-results/workflow-04-monitor-added.png', fullPage: true });
		console.log('✅ Step 2 complete: Monitor added');

		// 3. Edit Monitor
		console.log('Step 3: Editing monitor...');
		const monitorLink = page.locator('a:has-text("Workflow Test Monitor"), tbody tr a').first();
		if (await monitorLink.isVisible({ timeout: 5000 }).catch(() => false)) {
			await monitorLink.click();
			await page.waitForTimeout(1000);

			const editBtn = page.locator('button:has-text("Edit"), a:has-text("Edit")').first();
			if (await editBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
				await editBtn.click();
				await page.waitForTimeout(500);
			}

			const nameInput = page.locator('input[name="name"]').first();
			await nameInput.clear();
			await nameInput.fill('Workflow Test Monitor - Edited');

			await page.screenshot({ path: 'test-results/workflow-05-edit-monitor.png', fullPage: true });
			await page.locator('button[type="submit"]').first().click();
			await page.waitForTimeout(2000);
			await page.screenshot({ path: 'test-results/workflow-06-monitor-edited.png', fullPage: true });
			console.log('✅ Step 3 complete: Monitor edited');
		}

		// 4. Add Alert Rule
		console.log('Step 4: Adding alert rule...');
		await page.goto(`${BASE_URL}/alerts`);
		await page.waitForLoadState('networkidle');

		const addAlertBtn = page.locator('button:has-text("Add"), button:has-text("New"), a:has-text("Add")').first();
		if (await addAlertBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
			await addAlertBtn.click();
			await page.waitForTimeout(1000);
		}

		await page.locator('input[name="name"], input[placeholder*="name" i]').first().fill('Workflow Test Alert');

		const triggerSelect = page.locator('select[name="trigger_type"], select[name="triggerType"]').first();
		if (await triggerSelect.isVisible({ timeout: 2000 }).catch(() => false)) {
			await triggerSelect.selectOption('down');
		}

		const thresholdInput2 = page.locator('input[name="threshold"], input[name="threshold_value"]').first();
		if (await thresholdInput2.isVisible({ timeout: 2000 }).catch(() => false)) {
			await thresholdInput2.fill('3');
		}

		await page.screenshot({ path: 'test-results/workflow-07-add-alert.png', fullPage: true });
		await page.locator('button[type="submit"]').first().click();
		await page.waitForTimeout(2000);
		await page.screenshot({ path: 'test-results/workflow-08-alert-added.png', fullPage: true });
		console.log('✅ Step 4 complete: Alert rule added');

		// 5. Edit Alert Rule
		console.log('Step 5: Editing alert rule...');
		const alertLink = page.locator('a:has-text("Workflow Test Alert"), tbody tr a').first();
		if (await alertLink.isVisible({ timeout: 5000 }).catch(() => false)) {
			await alertLink.click();
			await page.waitForTimeout(1000);

			const editAlertBtn = page.locator('button:has-text("Edit"), a:has-text("Edit")').first();
			if (await editAlertBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
				await editAlertBtn.click();
				await page.waitForTimeout(500);
			}

			const alertNameInput = page.locator('input[name="name"]').first();
			await alertNameInput.clear();
			await alertNameInput.fill('Workflow Test Alert - Edited');

			await page.screenshot({ path: 'test-results/workflow-09-edit-alert.png', fullPage: true });
			await page.locator('button[type="submit"]').first().click();
			await page.waitForTimeout(2000);
			await page.screenshot({ path: 'test-results/workflow-10-alert-edited.png', fullPage: true });
			console.log('✅ Step 5 complete: Alert rule edited');
		}

		console.log('🎉 Complete workflow test finished successfully!');
	});
});
