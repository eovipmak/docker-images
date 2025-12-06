import { test, expect } from '@playwright/test';

const BACKEND_URL = process.env.BACKEND_API_URL || 'http://localhost:8080';
const API_BASE = '/api/v1';

test.describe('Landing redirect behavior', () => {
  test('Logged-in users are redirected from landing to dashboard', async ({ request, page }) => {
    const timestamp = Date.now();
    const email = `redirect-test-${timestamp}@example.com`;
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

    // Ensure token is present and set in localStorage before page load
    await page.addInitScript((t: string) => {
      window.localStorage.setItem('auth_token', t);
    }, token);

    // Navigate to landing page
    await page.goto('/');

    // Wait for redirect to dashboard
    await page.waitForURL('/dashboard');
    await expect(page).toHaveURL('/dashboard');
  });
});
