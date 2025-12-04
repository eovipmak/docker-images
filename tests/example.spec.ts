import { test, expect } from '@playwright/test';

/**
 * V-Insight Basic E2E Tests
 * 
 * Tests basic functionality of the V-Insight platform:
 * - Landing page loads correctly
 * - Login page is accessible
 * - Backend health endpoint responds
 */

const BASE_URL = process.env.BASE_URL || 'http://localhost:3000';
const BACKEND_URL = process.env.BACKEND_API_URL || 'http://localhost:8080';

test.describe('V-Insight Basic Tests', () => {
  test('landing page loads and shows V-Insight branding', async ({ page }) => {
    await page.goto(BASE_URL);
    
    // Wait for the page to load
    await page.waitForLoadState('networkidle');
    
    // Check the page title contains V-Insight
    await expect(page).toHaveTitle(/V-Insight/i);
  });

  test('login page is accessible', async ({ page }) => {
    await page.goto(`${BASE_URL}/login`);
    
    // Wait for the page to load
    await page.waitForLoadState('networkidle');
    
    // Check for login form elements
    await expect(page.locator('input[name="email"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('register page is accessible', async ({ page }) => {
    await page.goto(`${BASE_URL}/register`);
    
    // Wait for the page to load
    await page.waitForLoadState('networkidle');
    
    // Check for registration form elements (email, password, and tenant name)
    await expect(page.locator('input[name="email"], input[type="email"]').first()).toBeVisible();
    await expect(page.locator('input[name="password"], input[type="password"]').first()).toBeVisible();
  });

  test('backend health endpoint responds', async ({ request }) => {
    const response = await request.get(`${BACKEND_URL}/health`);
    
    expect(response.ok()).toBeTruthy();
    expect(response.status()).toBe(200);
    
    const body = await response.json();
    expect(body).toHaveProperty('status');
    expect(body.status).toBe('ok');
  });

  test('backend liveness probe responds', async ({ request }) => {
    const response = await request.get(`${BACKEND_URL}/health/live`);
    
    expect(response.ok()).toBeTruthy();
    expect(response.status()).toBe(200);
    
    const body = await response.json();
    expect(body.status).toBe('ok');
    expect(body.service).toBe('backend');
  });

  test('API v1 endpoint responds', async ({ request }) => {
    const response = await request.get(`${BACKEND_URL}/api/v1/`);
    
    expect(response.ok()).toBeTruthy();
    expect(response.status()).toBe(200);
    
    const body = await response.json();
    expect(body).toHaveProperty('message');
    expect(body.message).toBe('V-Insight API v1');
  });
});
