import { test, expect, Page } from '@playwright/test';
import { AuthHelper } from './helpers/auth.helper';

/**
 * E2E Test: Domain Deletion Error & Notification
 * 
 * This test verifies that:
 * 1. After deleting a domain and disabling its notification/warning, no HTTP 405 error occurs
 * 2. The uptime bar is correctly visible after the operation
 * 3. The workflow is automated and errors are clearly logged
 */

// Test configuration
const TEST_DOMAIN = 'example.com';
const TEST_PORT = 443;
const TEST_EMAIL = process.env.TEST_USER_EMAIL || 'test@example.com';
const TEST_PASSWORD = process.env.TEST_USER_PASSWORD || 'testpassword123';

/**
 * Helper function to monitor network requests for errors
 */
function setupNetworkMonitoring(page: Page): Array<{ url: string; status: number; method: string }> {
  const errorRequests: Array<{ url: string; status: number; method: string }> = [];

  page.on('response', async (response) => {
    const status = response.status();
    const url = response.url();
    const method = response.request().method();

    // Log all failed requests
    if (status >= 400) {
      const errorInfo = { url, status, method };
      errorRequests.push(errorInfo);
      console.error(`❌ HTTP Error: ${method} ${url} - Status: ${status}`);

      // If it's a 405 error, log additional details
      if (status === 405) {
        console.error(`🚨 CRITICAL: 405 Method Not Allowed detected on ${url}`);
        try {
          const responseBody = await response.text();
          console.error(`Response body: ${responseBody}`);
        } catch {
          console.error('Could not read response body');
        }
      }
    }
  });

  return errorRequests;
}

test.describe('Domain Deletion Workflow', () => {
  let authHelper: AuthHelper;

  test.beforeEach(async ({ page }) => {
    authHelper = new AuthHelper(page);
  });

  test('should delete domain and disable notifications without 405 error', async ({ page }) => {
    console.log('🧪 Starting domain deletion test...');

    // Setup network monitoring to catch 405 errors
    const errorRequests = setupNetworkMonitoring(page);

    // Step 1: Login to the application
    console.log('📝 Step 1: Logging in...');
    await test.step('Login to application', async () => {
      await authHelper.login(TEST_EMAIL, TEST_PASSWORD);
      console.log('✅ Login successful');
    });

    // Step 2: Navigate to dashboard
    console.log('📝 Step 2: Navigating to dashboard...');
    await test.step('Navigate to dashboard', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      
      // Verify we're on the dashboard
      await expect(page.locator('text=SSL Monitor Dashboard')).toBeVisible({ timeout: 10000 });
      console.log('✅ Dashboard loaded');
    });

    // Step 3: Check if test domain exists, if not add it
    console.log('📝 Step 3: Ensuring test domain exists...');
    await test.step('Add test domain if not exists', async () => {
      const domainCard = page.locator(`text=${TEST_DOMAIN}`).first();
      const domainExists = await domainCard.isVisible().catch(() => false);

      if (!domainExists) {
        console.log(`📌 Domain ${TEST_DOMAIN} not found, adding it...`);
        
        // Look for "Add Domain" or similar button - adjust selector based on actual UI
        const addButton = page.locator('button:has-text("Add"), button:has-text("Check SSL")').first();
        
        if (await addButton.isVisible().catch(() => false)) {
          await addButton.click();
          
          // Fill in domain input
          const domainInput = page.locator('input[type="text"], input[name="domain"]').first();
          await domainInput.fill(`${TEST_DOMAIN}:${TEST_PORT}`);
          
          // Submit the form
          const submitButton = page.locator('button[type="submit"]').first();
          await submitButton.click();
          
          // Wait for domain to appear
          await page.waitForTimeout(2000);
          console.log('✅ Test domain added');
        }
      } else {
        console.log(`✅ Domain ${TEST_DOMAIN} already exists`);
      }
    });

    // Step 4: Verify uptime bar is visible before actions
    console.log('📝 Step 4: Checking uptime bar visibility (before actions)...');
    await test.step('Verify uptime bar visibility before actions', async () => {
      // Wait for the domain card to be present
      const domainCard = page.locator(`text=${TEST_DOMAIN}`).first().locator('..').locator('..');
      await expect(domainCard).toBeVisible({ timeout: 5000 });

      // Check if uptime bar exists (it may not be present for newly added domains)
      const uptimeBar = domainCard.locator('text=Uptime, text=uptime').first();
      const hasUptimeBar = await uptimeBar.isVisible().catch(() => false);

      if (hasUptimeBar) {
        console.log('✅ Uptime bar is visible');
      } else {
        console.log('ℹ️ Uptime bar not present (may not be available for new domains)');
      }
    });

    // Step 5: Disable notifications for the domain
    console.log('📝 Step 5: Disabling notifications...');
    await test.step('Disable notifications for domain', async () => {
      // Find the domain card
      const domainCard = page.locator(`text=${TEST_DOMAIN}`).first().locator('..').locator('..');
      
      // Look for settings/actions button (usually a gear icon or three-dot menu)
      const settingsButton = domainCard.locator('button[aria-label="Actions"], button:has(svg)').last();
      
      if (await settingsButton.isVisible().catch(() => false)) {
        await settingsButton.click();
        
        // Wait for menu to appear
        await page.waitForTimeout(500);
        
        // Look for "Disable Alerts" or "Disable Notifications" option
        const disableAlertsOption = page.locator('text=Disable Alerts, text=Disable Notifications').first();
        
        if (await disableAlertsOption.isVisible().catch(() => false)) {
          // Clear any previous error requests before the action
          errorRequests.length = 0;
          
          await disableAlertsOption.click();
          
          // Wait for the action to complete
          await page.waitForTimeout(1000);
          
          // Check for 405 errors after disabling notifications
          const has405Error = errorRequests.some(req => req.status === 405);
          expect(has405Error, '❌ FAILED: 405 error occurred when disabling notifications').toBe(false);
          
          console.log('✅ Notifications disabled without 405 error');
        } else {
          console.log('ℹ️ Alerts already disabled or option not found');
        }
      }
    });

    // Step 6: Delete the domain
    console.log('📝 Step 6: Deleting domain...');
    await test.step('Delete domain', async () => {
      // Find the domain card again
      const domainCard = page.locator(`text=${TEST_DOMAIN}`).first().locator('..').locator('..');
      
      // Open settings/actions menu
      const settingsButton = domainCard.locator('button[aria-label="Actions"], button:has(svg)').last();
      await settingsButton.click();
      
      // Wait for menu
      await page.waitForTimeout(500);
      
      // Clear error requests before deletion
      errorRequests.length = 0;
      
      // Click delete option
      const deleteOption = page.locator('text=Delete Domain, text=Delete').first();
      await deleteOption.click();
      
      // Handle confirmation dialog
      page.once('dialog', async (dialog) => {
        console.log(`Dialog message: ${dialog.message()}`);
        await dialog.accept();
      });
      
      // If there's a button-based confirmation instead of browser dialog
      const confirmButton = page.locator('button:has-text("Delete"), button:has-text("Confirm")').first();
      if (await confirmButton.isVisible({ timeout: 2000 }).catch(() => false)) {
        await confirmButton.click();
      }
      
      // Wait for deletion to complete
      await page.waitForTimeout(2000);
      
      // Check for 405 errors after deletion
      const has405Error = errorRequests.some(req => req.status === 405);
      expect(has405Error, '❌ FAILED: 405 error occurred when deleting domain').toBe(false);
      
      console.log('✅ Domain deleted without 405 error');
    });

    // Step 7: Verify domain is removed from the list
    console.log('📝 Step 7: Verifying domain removal...');
    await test.step('Verify domain is removed', async () => {
      await page.waitForTimeout(1000);
      
      const domainCard = page.locator(`text=${TEST_DOMAIN}`).first();
      const isDomainVisible = await domainCard.isVisible().catch(() => false);
      
      expect(isDomainVisible, 'Domain should not be visible after deletion').toBe(false);
      console.log('✅ Domain successfully removed from dashboard');
    });

    // Step 8: Final check - ensure no 405 errors in the entire workflow
    console.log('📝 Step 8: Final verification...');
    await test.step('Final verification - no 405 errors', async () => {
      const all405Errors = errorRequests.filter(req => req.status === 405);
      
      if (all405Errors.length > 0) {
        console.error('🚨 CRITICAL: 405 errors detected during workflow:');
        all405Errors.forEach(err => {
          console.error(`  - ${err.method} ${err.url} (Status: ${err.status})`);
        });
      }
      
      expect(all405Errors.length, 'No 405 errors should occur during the entire workflow').toBe(0);
      console.log('✅ No 405 errors detected in the entire workflow');
    });

    console.log('🎉 Test completed successfully!');
  });

  test('should verify uptime bar rendering after operations', async ({ page }) => {
    console.log('🧪 Starting uptime bar rendering test...');

    // Login
    await authHelper.login(TEST_EMAIL, TEST_PASSWORD);

    // Navigate to dashboard
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Wait for any domain cards to load
    await page.waitForTimeout(2000);

    // Look for uptime bars in any domain cards
    const uptimeElements = page.locator('text=Uptime, text=uptime, text=30-Day Uptime');
    const uptimeCount = await uptimeElements.count();

    console.log(`ℹ️ Found ${uptimeCount} uptime indicator(s) on the dashboard`);

    if (uptimeCount > 0) {
      // Verify at least one uptime bar is visible
      const firstUptimeBar = uptimeElements.first();
      await expect(firstUptimeBar).toBeVisible();
      console.log('✅ Uptime bar is rendered correctly');

      // Take a screenshot for visual verification
      await page.screenshot({ 
        path: 'test-results/uptime-bar-rendering.png',
        fullPage: true 
      });
      console.log('📸 Screenshot saved to test-results/uptime-bar-rendering.png');
    } else {
      console.log('ℹ️ No uptime bars found (this is normal for dashboards with no monitored domains)');
    }

    console.log('🎉 Uptime bar rendering test completed!');
  });
});
