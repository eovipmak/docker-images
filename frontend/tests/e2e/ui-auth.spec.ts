import { test, expect } from '@playwright/test';

test.describe('UI Authentication Tests', () => {
  test('Register and Login with Screenshots', async ({ page }) => {
    const timestamp = Date.now();
    const email = `test-${timestamp}@example.com`;
    const password = 'testpassword123';
    const tenantName = `Test Tenant ${timestamp}`;

    // Navigate to register page
    await page.goto('/register');

    // Fill register form
    await page.fill('input[name="email"]', email);
    await page.fill('input[name="password"]', password);
    await page.fill('input[name="tenant_name"]', tenantName);

    // Take screenshot before submit
    await page.screenshot({ path: `register-form-${timestamp}.png` });

    // Submit
    await page.click('button[type="submit"]');

    // Wait for navigation or success
    await page.waitForURL('/dashboard');

    // Screenshot after register
    await page.screenshot({ path: `after-register-${timestamp}.png` });

    // Logout if needed, or navigate to login
    await page.goto('/login');

    // Fill login form
    await page.fill('input[name="email"]', email);
    await page.fill('input[name="password"]', password);

    // Screenshot login form
    await page.screenshot({ path: `login-form-${timestamp}.png` });

    // Submit
    await page.click('button[type="submit"]');

    // Wait for dashboard
    await page.waitForURL('/dashboard');

    // Screenshot after login
    await page.screenshot({ path: `after-login-${timestamp}.png` });
  });
});