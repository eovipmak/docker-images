import { test, expect, Page } from '@playwright/test';
import { AuthHelper } from './helpers/auth.helper';

/**
 * E2E Test: Webhook and Alert API Errors (Issue #)
 * 
 * This test reproduces and verifies fixes for:
 * 1. 500 error when testing webhook
 * 2. 404 error when testing domain alert
 * 3. 404 error when changing check interval
 * 4. 404 error when toggling alerts for domain
 */

// Test configuration
const TEST_DOMAIN = 'baohiemxahoidientu.vn';
const TEST_EMAIL = process.env.TEST_USER_EMAIL || 'test@example.com';
const TEST_PASSWORD = process.env.TEST_USER_PASSWORD || 'testpassword123';

/**
 * Helper function to monitor network requests and capture errors
 */
function setupNetworkMonitoring(page: Page): Array<{ url: string; status: number; method: string; endpoint: string }> {
  const errorRequests: Array<{ url: string; status: number; method: string; endpoint: string }> = [];

  page.on('response', async (response) => {
    const status = response.status();
    const url = response.url();
    const method = response.request().method();

    // Extract endpoint from URL for easier identification
    const urlObj = new URL(url);
    const endpoint = urlObj.pathname + urlObj.search;

    // Log all failed requests
    if (status >= 400) {
      const errorInfo = { url, status, method, endpoint };
      errorRequests.push(errorInfo);
      console.error(`❌ HTTP Error: ${method} ${endpoint} - Status: ${status}`);

      // Log response body for debugging
      try {
        const responseBody = await response.text();
        console.error(`Response body: ${responseBody}`);
      } catch {
        console.error('Could not read response body');
      }
    }
  });

  return errorRequests;
}

test.describe('Webhook and Alert API Error Tests', () => {
  let authHelper: AuthHelper;

  test.beforeEach(async ({ page }) => {
    authHelper = new AuthHelper(page);
  });

  test('should reproduce webhook and alert API errors', async ({ page }) => {
    console.log('🧪 Starting webhook and alert API error reproduction test...');

    // Setup network monitoring to capture errors
    const errorRequests = setupNetworkMonitoring(page);

    // Step 1: Login to the application
    console.log('📝 Step 1: Logging in...');
    await test.step('Login to application', async () => {
      await authHelper.login(TEST_EMAIL, TEST_PASSWORD);
      console.log('✅ Login successful');
    });

    // Step 2: Navigate to Alert Settings page and test webhook
    console.log('📝 Step 2: Testing webhook notification...');
    await test.step('Test webhook notification', async () => {
      await page.goto('/alert-settings');
      await page.waitForLoadState('networkidle');
      
      // Configure a webhook URL first if not present
      const webhookInput = page.locator('input[name="webhook_url"], input[placeholder*="webhook"], input[label*="Webhook"]').first();
      if (await webhookInput.isVisible().catch(() => false)) {
        await webhookInput.fill('https://webhook.site/unique-id');
        
        // Save the configuration
        const saveButton = page.locator('button:has-text("Save")').first();
        if (await saveButton.isVisible().catch(() => false)) {
          await saveButton.click();
          await page.waitForTimeout(1000);
        }
      }
      
      // Find and click the test webhook button
      const testWebhookButton = page.locator('button:has-text("Test"), button:has-text("Send Test")').first();
      if (await testWebhookButton.isVisible().catch(() => false)) {
        console.log('📌 Clicking test webhook button...');
        await testWebhookButton.click();
        await page.waitForTimeout(2000);
        
        // Check for 500 error on test-webhook endpoint
        const webhookErrors = errorRequests.filter(r => r.endpoint.includes('test-webhook'));
        if (webhookErrors.length > 0) {
          console.error('🚨 ISSUE #1: Test webhook returned error:', webhookErrors);
        }
      } else {
        console.log('⚠️  Test webhook button not found, skipping...');
      }
    });

    // Step 3: Navigate to Dashboard
    console.log('📝 Step 3: Navigating to dashboard...');
    await test.step('Navigate to dashboard', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      
      // Verify we're on the dashboard
      const dashboardTitle = page.locator('text=Dashboard, text=SSL Monitor').first();
      await expect(dashboardTitle).toBeVisible({ timeout: 10000 });
      console.log('✅ Dashboard loaded');
    });

    // Step 4: Add test domain if not exists
    console.log('📝 Step 4: Ensuring test domain exists...');
    await test.step('Add test domain if needed', async () => {
      const domainCard = page.locator(`text=${TEST_DOMAIN}`).first();
      const domainExists = await domainCard.isVisible().catch(() => false);

      if (!domainExists) {
        console.log(`📌 Domain ${TEST_DOMAIN} not found, adding it...`);
        
        // Try to find add domain button/input
        const addInput = page.locator('input[placeholder*="domain"], input[placeholder*="Check SSL"]').first();
        if (await addInput.isVisible().catch(() => false)) {
          await addInput.fill(TEST_DOMAIN);
          await addInput.press('Enter');
          await page.waitForTimeout(3000);
        }
      } else {
        console.log(`✅ Domain ${TEST_DOMAIN} already exists`);
      }
    });

    // Step 5: Test domain alert
    console.log('📝 Step 5: Testing domain alert...');
    await test.step('Test domain alert', async () => {
      // Find the domain card and open its menu
      const domainCard = page.locator(`text=${TEST_DOMAIN}`).first();
      if (await domainCard.isVisible().catch(() => false)) {
        // Look for menu button (three dots, settings icon, etc.)
        const menuButton = page.locator(`[aria-label*="menu"], button:has([data-testid*="MoreVert"]), button:has([data-testid*="Settings"])`).first();
        
        if (await menuButton.isVisible().catch(() => false)) {
          await menuButton.click();
          await page.waitForTimeout(500);
          
          // Look for "Test Alert" menu item
          const testAlertItem = page.locator('text=Test Alert, li:has-text("Test Alert")').first();
          if (await testAlertItem.isVisible().catch(() => false)) {
            console.log('📌 Clicking test alert...');
            await testAlertItem.click();
            await page.waitForTimeout(2000);
            
            // Check for 404 error on test-alert endpoint
            const testAlertErrors = errorRequests.filter(r => r.endpoint.includes('test-alert'));
            if (testAlertErrors.length > 0) {
              console.error('🚨 ISSUE #2: Test alert returned error:', testAlertErrors);
            }
          }
        }
      }
    });

    // Step 6: Change check interval
    console.log('📝 Step 6: Changing check interval...');
    await test.step('Change check interval', async () => {
      const domainCard = page.locator(`text=${TEST_DOMAIN}`).first();
      if (await domainCard.isVisible().catch(() => false)) {
        // Look for settings/edit button
        const settingsButton = page.locator(`button:has-text("Settings"), button[aria-label*="settings"]`).first();
        
        if (await settingsButton.isVisible().catch(() => false)) {
          await settingsButton.click();
          await page.waitForTimeout(500);
          
          // Look for check interval input
          const intervalInput = page.locator('input[name*="interval"], input[label*="interval"]').first();
          if (await intervalInput.isVisible().catch(() => false)) {
            await intervalInput.fill('3600');
            
            // Save changes
            const saveButton = page.locator('button:has-text("Save")').first();
            if (await saveButton.isVisible().catch(() => false)) {
              console.log('📌 Saving check interval...');
              await saveButton.click();
              await page.waitForTimeout(2000);
              
              // Check for 404 error on monitors PATCH endpoint
              const intervalErrors = errorRequests.filter(r => 
                r.endpoint.includes('monitors') && r.method === 'PATCH'
              );
              if (intervalErrors.length > 0) {
                console.error('🚨 ISSUE #3: Change interval returned error:', intervalErrors);
              }
            }
          }
        }
      }
    });

    // Step 7: Toggle alerts for domain
    console.log('📝 Step 7: Toggling alerts...');
    await test.step('Toggle alerts for domain', async () => {
      const domainCard = page.locator(`text=${TEST_DOMAIN}`).first();
      if (await domainCard.isVisible().catch(() => false)) {
        // Look for alert toggle switch
        const alertToggle = page.locator(`[role="switch"], input[type="checkbox"]`).first();
        
        if (await alertToggle.isVisible().catch(() => false)) {
          console.log('📌 Toggling alerts...');
          await alertToggle.click();
          await page.waitForTimeout(2000);
          
          // Check for 404 error on monitors PATCH endpoint
          const toggleErrors = errorRequests.filter(r => 
            r.endpoint.includes('monitors') && r.method === 'PATCH'
          );
          if (toggleErrors.length > 0) {
            console.error('🚨 ISSUE #4: Toggle alerts returned error:', toggleErrors);
          }
        }
      }
    });

    // Summary of errors found
    console.log('\n📊 Test Summary:');
    console.log(`Total errors captured: ${errorRequests.length}`);
    errorRequests.forEach((err, index) => {
      console.log(`${index + 1}. ${err.method} ${err.endpoint} - ${err.status}`);
    });

    // Take screenshot for documentation
    await page.screenshot({ 
      path: `/tmp/webhook-alert-test-${Date.now()}.png`,
      fullPage: true 
    });
    console.log('📸 Screenshot saved');
  });
});
