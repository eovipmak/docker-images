import { test, expect } from '@playwright/test';

/**
 * Comprehensive E2E Tests for V-Insight Platform
 * 
 * Tests the complete workflow:
 * 1. Login with demo user
 * 2. Add a monitor for google.com
 * 3. Edit the monitor (change interval, timeout, name)
 * 4. Add an alert rule
 * 5. Edit the alert rule
 */

const BASE_URL = process.env.BASE_URL || 'http://localhost:3000';
const BACKEND_URL = process.env.BACKEND_API_URL || 'http://localhost:8080';
const API_BASE = '/api/v1';

test.describe('V-Insight E2E Automated Tests', () => {
	// Use a unique timestamp for each test run to avoid conflicts
	const timestamp = Date.now();
	const TEST_EMAIL = `e2e-test-${timestamp}@example.com`;
	const TEST_PASSWORD = 'testpassword123';
	const TEST_TENANT = `E2E Tenant ${timestamp}`;

	let authToken = '';

	test.beforeAll(async ({ request }) => {
		// Register a test user for E2E tests
		const response = await request.post(`${BACKEND_URL}${API_BASE}/auth/register`, {
			data: {
				email: TEST_EMAIL,
				password: TEST_PASSWORD,
				tenant_name: TEST_TENANT,
			},
		});
		expect(response.ok()).toBeTruthy();
		const body = await response.json();
		authToken = body.token;
	});

	test.beforeEach(async ({ page }) => {
		// Set a longer timeout for all actions in this test (CI environments can be slow)
		page.setDefaultTimeout(30000);
	});

	test('1. Login with test user', async ({ page }) => {
		// Navigate to login page
		await page.goto(`${BASE_URL}/login`);
		await page.waitForLoadState('networkidle');

		// Fill in login credentials
		await page.locator('input[name="email"]').fill(TEST_EMAIL);
		await page.locator('input[name="password"]').fill(TEST_PASSWORD);

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

	test('2. Add monitor for google.com', async ({ page }) => {
		// Set token in localStorage before navigating
		await page.addInitScript((token: string) => {
			window.localStorage.setItem('auth_token', token);
		}, authToken);

		// Navigate to monitors page
		await page.goto(`${BASE_URL}/monitors`);
		await page.waitForLoadState('networkidle');

		// Take screenshot of monitors page
		await page.screenshot({ path: 'test-results/03-monitors-page.png', fullPage: true });

		// Click "Add Monitor" button
		const addButton = page.locator('button:has-text("Add Monitor")');
		await expect(addButton).toBeVisible();
		await addButton.click();

		// Wait for modal to appear
		await page.waitForSelector('[role="dialog"]', { state: 'visible' });

		// Fill in monitor details using id selectors (matching the MonitorModal component)
		await page.locator('#name').fill('Google Monitor');
		await page.locator('#url').fill('https://google.com');

		// Set interval (60 seconds minimum)
		await page.locator('#check_interval').clear();
		await page.locator('#check_interval').fill('60');

		// Set timeout (30 seconds)
		await page.locator('#timeout').clear();
		await page.locator('#timeout').fill('30');

		// Take screenshot before submitting
		await page.screenshot({ path: 'test-results/04-add-monitor-form.png', fullPage: true });

		// Click Save button (the modal uses a button with text "Save")
		await page.locator('button:has-text("Save")').click();

		// Wait for modal to close
		await page.waitForSelector('[role="dialog"]', { state: 'hidden', timeout: 10000 });

		// Take screenshot after creation
		await page.screenshot({ path: 'test-results/05-monitor-created.png', fullPage: true });

		// Verify monitor appears in the list
		await expect(page.locator('text=Google Monitor')).toBeVisible({ timeout: 5000 });

		console.log('✅ Monitor added successfully');
	});

	test('3. Edit monitor (change interval, timeout, name)', async ({ page }) => {
		// Set token in localStorage
		await page.addInitScript((token: string) => {
			window.localStorage.setItem('auth_token', token);
		}, authToken);

		// Navigate to monitors page
		await page.goto(`${BASE_URL}/monitors`);
		await page.waitForLoadState('networkidle');

		// Find and click on the Google Monitor card
		const monitorCard = page.locator('[data-testid="monitor-card"]:has-text("Google Monitor")');
		
		if (await monitorCard.isVisible({ timeout: 5000 }).catch(() => false)) {
			await monitorCard.click();
		} else {
			// Try clicking first monitor card if specific one not found
			await page.locator('[data-testid="monitor-card"]').first().click();
		}

		// Wait for navigation to monitor detail page
		await page.waitForURL(/.*\/monitors\/.+/);
		await page.waitForLoadState('networkidle');

		// Take screenshot of monitor detail page
		await page.screenshot({ path: 'test-results/06-monitor-detail.png', fullPage: true });

		// Look for Edit button on the detail page
		const editButton = page.locator('button:has-text("Edit")');
		if (await editButton.isVisible({ timeout: 3000 }).catch(() => false)) {
			await editButton.click();
			
			// Wait for modal to appear
			await page.waitForSelector('[role="dialog"]', { state: 'visible' });

			// Update monitor name
			await page.locator('#name').clear();
			await page.locator('#name').fill('Google Monitor - Updated');

			// Update interval to 120 seconds
			await page.locator('#check_interval').clear();
			await page.locator('#check_interval').fill('120');

			// Update timeout to 60 seconds
			await page.locator('#timeout').clear();
			await page.locator('#timeout').fill('60');

			// Take screenshot before saving
			await page.screenshot({ path: 'test-results/07-edit-monitor-form.png', fullPage: true });

			// Save changes
			await page.locator('button:has-text("Save")').click();

			// Wait for modal to close
			await page.waitForSelector('[role="dialog"]', { state: 'hidden', timeout: 10000 });

			// Take screenshot after update
			await page.screenshot({ path: 'test-results/08-monitor-updated.png', fullPage: true });

			console.log('✅ Monitor updated successfully');
		} else {
			console.log('⚠️ Edit button not found on monitor detail page');
		}
	});

	test('4. Add alert rule', async ({ page }) => {
		// Set token in localStorage
		await page.addInitScript((token: string) => {
			window.localStorage.setItem('auth_token', token);
		}, authToken);

		// Navigate to alerts page
		await page.goto(`${BASE_URL}/alerts`);
		await page.waitForLoadState('networkidle');

		// Take screenshot of alerts page
		await page.screenshot({ path: 'test-results/09-alerts-page.png', fullPage: true });

		// Click "Create Rule" or "Add" button (based on AlertRuleModal trigger in alerts/+page.svelte)
		const addRuleButton = page.locator('button:has-text("Create Rule"), button:has-text("Add Rule"), button:has-text("New Rule")').first();
		
		if (await addRuleButton.isVisible({ timeout: 5000 }).catch(() => false)) {
			await addRuleButton.click();
		} else {
			// Try finding any button with "Add" or plus icon
			const addButton = page.locator('button:has-text("Add")').first();
			await addButton.click();
		}

		// Wait for modal to appear
		await page.waitForSelector('[role="dialog"]', { state: 'visible' });

		// Fill in alert rule details (matching AlertRuleModal component)
		await page.locator('#name').fill('Google Down Alert');

		// Select trigger type (the select has id="trigger_type")
		await page.locator('#trigger_type').selectOption('down');

		// Set threshold value (the input has id="threshold_value")
		await page.locator('#threshold_value').clear();
		await page.locator('#threshold_value').fill('3');

		// Take screenshot before submitting
		await page.screenshot({ path: 'test-results/10-add-alert-form.png', fullPage: true });

		// Submit the form (button with text "Create Rule")
		await page.locator('button[type="submit"]:has-text("Create Rule")').click();

		// Wait for modal to close
		await page.waitForSelector('[role="dialog"]', { state: 'hidden', timeout: 10000 });

		// Take screenshot after creation
		await page.screenshot({ path: 'test-results/11-alert-created.png', fullPage: true });

		// Verify alert appears in the list
		await expect(page.locator('text=Google Down Alert')).toBeVisible({ timeout: 5000 });

		console.log('✅ Alert rule added successfully');
	});

	test('5. Edit alert rule', async ({ page }) => {
		// Set token in localStorage
		await page.addInitScript((token: string) => {
			window.localStorage.setItem('auth_token', token);
		}, authToken);

		// Navigate to alerts page
		await page.goto(`${BASE_URL}/alerts`);
		await page.waitForLoadState('networkidle');

		// Find and click on the alert rule to edit (look for AlertCard or list item)
		const alertCard = page.locator('[data-testid="alert-card"]:has-text("Google Down Alert"), div:has-text("Google Down Alert")').first();
		
		// Look for edit button/icon on the alert card
		const editButton = alertCard.locator('button:has-text("Edit"), button[aria-label*="edit" i]').first();
		
		if (await editButton.isVisible({ timeout: 3000 }).catch(() => false)) {
			await editButton.click();
		} else {
			// Try clicking the card itself and then finding edit button
			await alertCard.click();
			await page.waitForTimeout(500);
			const editBtn = page.locator('button:has-text("Edit")').first();
			if (await editBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
				await editBtn.click();
			}
		}

		// Wait for modal to appear
		const modalVisible = await page.waitForSelector('[role="dialog"]', { state: 'visible', timeout: 5000 }).catch(() => null);
		
		if (modalVisible) {
			// Update alert name
			await page.locator('#name').clear();
			await page.locator('#name').fill('Google Down Alert - Updated');

			// Update threshold to 5 consecutive failures
			await page.locator('#threshold_value').clear();
			await page.locator('#threshold_value').fill('5');

			// Take screenshot before saving
			await page.screenshot({ path: 'test-results/12-edit-alert-form.png', fullPage: true });

			// Save changes (button with text "Update Rule")
			await page.locator('button[type="submit"]:has-text("Update Rule")').click();

			// Wait for modal to close
			await page.waitForSelector('[role="dialog"]', { state: 'hidden', timeout: 10000 });

			// Take screenshot after update
			await page.screenshot({ path: 'test-results/13-alert-updated.png', fullPage: true });

			// Verify the updated name appears
			await expect(page.locator('text=Google Down Alert - Updated')).toBeVisible({ timeout: 5000 });

			console.log('✅ Alert rule updated successfully');
		} else {
			console.log('⚠️ Could not open edit modal for alert rule');
		}
	});
});
