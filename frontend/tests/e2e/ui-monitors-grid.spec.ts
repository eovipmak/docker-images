import { test, expect } from '@playwright/test';

const BACKEND_URL = process.env.BACKEND_API_URL || 'http://localhost:8080';
const API_BASE = '/api/v1';

test('Monitors grid view renders with cards', async ({ request, page }) => {
  const timestamp = Date.now();
  const email = `monitors-test-${timestamp}@example.com`;
  const password = 'testpassword123';
  // Register new user via backend API
  const registerRes = await request.post(`${BACKEND_URL}${API_BASE}/auth/register`, {
    data: {
      email,
      password,
    },
  });

  expect(registerRes.ok()).toBeTruthy();
  const body = await registerRes.json();
  const token = body.token;

  // Create monitors via API for this user
  const monitorsToCreate = [
    { name: 'Grid 1', url: 'https://grid-1.example.com', check_interval: 60, timeout: 10, enabled: true },
    { name: 'Grid 2', url: 'https://grid-2.example.com', check_interval: 60, timeout: 10, enabled: true },
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

  // Ensure token is present in localStorage
  await page.addInitScript((t: string) => {
    window.localStorage.setItem('auth_token', t);
  }, token);

  // Navigate to monitors page
  await page.goto('/monitors');
  await page.waitForLoadState('networkidle');

  // Wait for the monitors grid to render and take a screenshot
  const monitorsGrid = page.locator('[data-testid="monitor-card"]').first();
  await expect(monitorsGrid).toBeVisible();

  const allMonitorCards = page.locator('[data-testid="monitor-card"]');
  await expect(allMonitorCards).toHaveCount(monitorsToCreate.length);

  await expect(page.locator('div.grid')).toHaveScreenshot('monitors-grid.png');

  // Clicking the first monitor should navigate to its details page
  const firstMonitorCard = page.locator('[data-testid="monitor-card"]').first();
  await firstMonitorCard.click();
  // Wait for SPA navigation to route to /monitors/:id
  await page.waitForURL(/.*\/monitors\/.+/);
  await expect(page).toHaveURL(/.*\/monitors\/.+/);

});
