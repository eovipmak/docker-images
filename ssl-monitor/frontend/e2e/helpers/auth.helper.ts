import { Page, expect } from '@playwright/test';

/**
 * Helper class for authentication in e2e tests
 */
export class AuthHelper {
  constructor(private page: Page) {}

  /**
   * Login to the application with the provided credentials
   */
  async login(email: string, password: string): Promise<void> {
    // Navigate to login page
    await this.page.goto('/login');
    
    // Wait for the login form to be visible
    await this.page.waitForLoadState('networkidle');
    
    // Fill in the email/username field
    const emailInput = this.page.locator('input[type="email"], input[name="email"], input[name="username"]').first();
    await emailInput.fill(email);
    
    // Fill in the password field
    const passwordInput = this.page.locator('input[type="password"]').first();
    await passwordInput.fill(password);
    
    // Click the login button
    const loginButton = this.page.locator('button[type="submit"]').first();
    await loginButton.click();
    
    // Wait for navigation to complete
    await this.page.waitForURL((url) => !url.pathname.includes('/login'), {
      timeout: 10000,
    });
    
    // Verify we're logged in by checking for auth token in localStorage
    const token = await this.page.evaluate(() => localStorage.getItem('auth_token'));
    expect(token).toBeTruthy();
  }

  /**
   * Logout from the application
   */
  async logout(): Promise<void> {
    await this.page.evaluate(() => {
      localStorage.removeItem('auth_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('user');
    });
  }

  /**
   * Set authentication tokens directly (useful for bypassing login UI)
   */
  async setAuthTokens(accessToken: string, refreshToken?: string): Promise<void> {
    await this.page.evaluate(
      ({ access, refresh }) => {
        localStorage.setItem('auth_token', access);
        if (refresh) {
          localStorage.setItem('refresh_token', refresh);
        }
      },
      { access: accessToken, refresh: refreshToken }
    );
  }

  /**
   * Check if user is authenticated
   */
  async isAuthenticated(): Promise<boolean> {
    const token = await this.page.evaluate(() => localStorage.getItem('auth_token'));
    return !!token;
  }
}
