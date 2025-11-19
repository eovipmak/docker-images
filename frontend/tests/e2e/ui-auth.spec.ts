import { test, expect } from '@playwright/test';

test.describe('UI Authentication Tests', () => {
  test('Registration Test', async ({ page }) => {
    const timestamp = Date.now();
    const email = `test-${timestamp}@example.com`;
    const password = 'testpassword123';
    const tenantName = `Test Tenant ${timestamp}`;

    // Navigate to register page
    await page.goto('/register');
    
    // Wait for Svelte to hydrate
    await page.waitForLoadState('networkidle');

    // Fill register form and trigger input events
    await page.locator('input[name="email"]').fill(email);
    await page.locator('input[name="password"]').fill(password);
    await page.locator('input[name="confirm_password"]').fill(password);
    await page.locator('input[name="tenant_name"]').fill(tenantName);

    // Wait a moment for Svelte reactivity
    await page.waitForTimeout(500);

    // Take screenshot before submit
    await page.screenshot({ path: `register-form-${timestamp}.png` });

    // Submit form by clicking button
    await page.locator('button[type="submit"]').click();

    // Wait for navigation or success
    await page.waitForURL('/dashboard');

    // Screenshot after register
    await page.screenshot({ path: `after-register-${timestamp}.png` });
  });
});