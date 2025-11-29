import { test, expect } from '@playwright/test';

const BACKEND_URL = process.env.BACKEND_API_URL || 'http://localhost:8080';
const API_BASE = '/api/v1';

test('Dashboard monitors & alerts preview grid renders', async ({ request, page }) => {
  const timestamp = Date.now();
  const email = `dashboard-test-${timestamp}@example.com`;
  const password = 'testpassword123';
  const tenantName = `Dashboard Tenant ${timestamp}`;

  // Register new user via backend API
  const registerRes = await request.post(`${BACKEND_URL}${API_BASE}/auth/register`, {
    data: {
      email,
      password,
      tenant_name: tenantName,
    },
  });

  expect(registerRes.ok()).toBeTruthy();
  const body = await registerRes.json();
  const token = body.token;

  // Create monitors via API for this user
  const monitorsToCreate = [
    { name: 'Example 1', url: 'https://google.com', check_interval: 60, timeout: 10, enabled: true },
    { name: 'Example 2', url: 'https://youtube.com', check_interval: 60, timeout: 10, enabled: true },
    { name: 'Example 3', url: 'https://facebook.com', check_interval: 60, timeout: 10, enabled: true }
  ];

  for (const m of monitorsToCreate) {
    const res = await request.post(`${BACKEND_URL}${API_BASE}/monitors`, {
      data: m,
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
    expect(res.ok()).toBeTruthy();
  }

  // Create a couple of alert rules
  const alertRulesToCreate = [
    { name: 'Down rule', trigger_type: 'down', threshold_value: 3, enabled: true },
    { name: 'Slow', trigger_type: 'slow_response', threshold_value: 2000, enabled: true }
  ];

  for (const r of alertRulesToCreate) {
    const res = await request.post(`${BACKEND_URL}${API_BASE}/alert-rules`, {
      data: r,
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
    expect(res.ok()).toBeTruthy();
  }

  // Ensure token is present in localStorage
  await page.addInitScript((t: string) => {
    window.localStorage.setItem('auth_token', t);
  }, token);

  // Navigate to dashboard
  await page.goto('/dashboard');
  await page.waitForLoadState('domcontentloaded');

  // Wait for the dashboard stats to render
  const totalMonitorsStat = page.locator('text=Total Monitors');
  await expect(totalMonitorsStat).toBeVisible();

  // Check that stats are present
  await expect(page.locator('p:has-text("Total Monitors")')).toBeVisible();
  await expect(page.locator('p:has-text("Operational")')).toBeVisible();
  await expect(page.locator('p:has-text("Downtime")')).toBeVisible();
  await expect(page.locator('p:has-text("Open Incidents")')).toBeVisible();

});
