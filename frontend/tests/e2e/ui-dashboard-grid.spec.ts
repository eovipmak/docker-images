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
    { name: 'Example 1', url: 'https://example-1.com', check_interval: 60, timeout: 10, enabled: true },
    { name: 'Example 2', url: 'https://example-2.com', check_interval: 60, timeout: 10, enabled: true },
    { name: 'Example 3', url: 'https://example-3.com', check_interval: 60, timeout: 10, enabled: true }
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
  await page.waitForLoadState('networkidle');

  // Wait for the monitors preview to render
  const monitorsPreview = page.locator('[data-testid="dashboard-monitors-preview"]');
  await expect(monitorsPreview).toBeVisible();

  // Check that the monitor cards are present
  const monitorCards = page.locator('[data-testid="monitor-card"]');
  await expect(monitorCards).toHaveCount(monitorsToCreate.length);

  // Check that alert cards are present
  const alertCards = page.locator('[data-testid="alert-card"]');
  await expect(alertCards).toHaveCount(alertRulesToCreate.length);

  // Take visual snapshot of monitors preview
  await expect(monitorsPreview).toHaveScreenshot('dashboard-monitors-preview.png');

  // Clicking a monitor preview should navigate to monitor details
  const firstMonitor = monitorCards.first();
  await firstMonitor.click();
  await page.waitForURL(/.*\/monitors\/.+/);
  await expect(page).toHaveURL(/.*\/monitors\/.+/);

});
