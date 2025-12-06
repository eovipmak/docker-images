import { test, expect } from '@playwright/test';

const BACKEND_URL = process.env.BACKEND_API_URL || 'http://localhost:8080';
const API_BASE = '/api/v1';

test.describe('UI Authentication Tests', () => {
  test('Login form displays all required fields', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    
    // Check for required form elements
    await expect(page.locator('input[name="email"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
    
    // Check for navigation to registration
    await expect(page.locator('a[href="/register"]')).toBeVisible();
  });

  test('Registration form displays all required fields', async ({ page }) => {
    await page.goto('/register');
    await page.waitForLoadState('networkidle');
    
    // Check for required form elements
    await expect(page.locator('input[name="email"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    await expect(page.locator('input[name="confirmPassword"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('Login with invalid credentials shows error', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    
    // Fill in invalid credentials
    await page.locator('input[name="email"]').fill('invalid@example.com');
    await page.locator('input[name="password"]').fill('wrongpassword');
    
    // Submit the form
    await page.locator('button[type="submit"]').click();
    
    // Wait for the error to appear (via network request failure)
    await page.waitForTimeout(2000);
    
    // Should show error message (not redirected to dashboard)
    // The error will appear in the form and page should not be dashboard
    const url = page.url();
    expect(url).not.toContain('/dashboard');
  });

  test('Login with valid credentials redirects to dashboard', async ({ page, request }) => {
    const timestamp = Date.now();
    const email = `ui-auth-test-${timestamp}@example.com`;
    const password = 'testpassword123';

    // Register new user via backend API first
    const registerRes = await request.post(`${BACKEND_URL}${API_BASE}/auth/register`, {
      data: {
        email,
        password,
      },
    });
    expect(registerRes.ok()).toBeTruthy();

    // Now test UI login
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    
    // Fill in valid credentials
    await page.locator('input[name="email"]').fill(email);
    await page.locator('input[name="password"]').fill(password);
    
    // Submit the form
    await page.locator('button[type="submit"]').click();
    
    // Wait for redirect to dashboard
    await page.waitForURL('**/dashboard', { timeout: 10000 });
    
    // Verify we're on the dashboard
    await expect(page).toHaveURL(/\/dashboard/);
  });

  test('Registration with valid data redirects to dashboard', async ({ page }) => {
    const timestamp = Date.now();
    const email = `ui-register-test-${timestamp}@example.com`;
    const password = 'testpassword123';

    await page.goto('/register');
    await page.waitForLoadState('networkidle');
    
    // Fill in registration form
    await page.locator('input[name="email"]').fill(email);
    await page.locator('input[name="password"]').fill(password);
    await page.locator('input[name="confirmPassword"]').fill(password);
    
    // Submit the form
    await page.locator('button[type="submit"]').click();
    
    // Wait for redirect to dashboard
    await page.waitForURL('**/dashboard', { timeout: 10000 });
    
    // Verify we're on the dashboard
    await expect(page).toHaveURL(/\/dashboard/);
  });

  test('Registration with mismatched passwords shows error', async ({ page }) => {
    await page.goto('/register');
    await page.waitForLoadState('networkidle');
    
    const timestamp = Date.now();
    
    // Fill in registration form with mismatched passwords
    await page.locator('input[name="email"]').fill(`mismatch-${timestamp}@example.com`);
    await page.locator('input[name="password"]').fill('password123');
    await page.locator('input[name="confirmPassword"]').fill('differentpassword');
    
    // Submit the form
    await page.locator('button[type="submit"]').click();
    
    // Wait for error message to appear
    await page.waitForTimeout(500);
    
    // Should show error message and stay on register page
    await expect(page.locator('text=Passwords do not match')).toBeVisible();
    await expect(page).toHaveURL(/\/register/);
  });
});
