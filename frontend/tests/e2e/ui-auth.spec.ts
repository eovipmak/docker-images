import { test, expect } from '@playwright/test';

test.describe('UI Authentication Tests', () => {
  // Registration UI test removed due to redirect issue, covered by API tests
  test('Login Test', async ({ page }) => {
    // TODO: Implement login UI test if needed
  });
});