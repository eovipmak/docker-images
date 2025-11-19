import { test, expect } from '@playwright/test';

test('Debug registration', async ({ page }) => {
  const timestamp = Date.now();
  const email = `test-${timestamp}@example.com`;
  const password = 'testpassword123';
  const tenantName = `Test Tenant ${timestamp}`;

  // Enable detailed logging
  page.on('console', msg => console.log('[CONSOLE]', msg.text()));
  page.on('pageerror', err => console.log('[PAGE ERROR]', err.message));
  page.on('request', req => console.log('[REQUEST]', req.method(), req.url()));
  page.on('response', async res => {
    console.log('[RESPONSE]', res.status(), res.url());
    if (res.url().includes('/api/v1/auth/register')) {
      const text = await res.text();
      console.log('[REGISTER RESPONSE BODY]', text);
    }
  });

  console.log('=== Starting test ===');
  console.log('Email:', email);
  console.log('Password:', password);
  console.log('Tenant:', tenantName);

  console.log('\n=== Navigating to /register ===');
  await page.goto('/register');
  
  console.log('\n=== Filling form ===');
  await page.fill('input[name="email"]', email);
  await page.fill('input[name="password"]', password);
  await page.fill('input[name="confirm_password"]', password);
  await page.fill('input[name="tenant_name"]', tenantName);

  console.log('\n=== Taking screenshot before submit ===');
  await page.screenshot({ path: `before-submit-${timestamp}.png` });

  console.log('\n=== Clicking submit button ===');
  await page.click('button[type="submit"]');

  console.log('\n=== Waiting 3 seconds for any async operations ===');
  await page.waitForTimeout(3000);

  console.log('\n=== Taking screenshot after submit ===');
  await page.screenshot({ path: `after-submit-${timestamp}.png` });

  console.log('\n=== Current URL:', page.url(), '===');

  console.log('\n=== Waiting for navigation to /dashboard (10s timeout) ===');
  await page.waitForURL('/dashboard', { timeout: 10000 });

  console.log('\n=== SUCCESS ===');
});
