import { test, expect } from '@playwright/test';

test.describe('Leaderboard E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display global leaderboard', async ({ page }) => {
    await page.click('text=Leaderboard');
    
    // Wait for leaderboard to load
    await expect(page.locator('[data-testid="leaderboard-table"]')).toBeVisible();
    
    // Verify columns
    await expect(page.locator('text=Rank')).toBeVisible();
    await expect(page.locator('text=User')).toBeVisible();
    await expect(page.locator('text=Score')).toBeVisible();
  });

  test('should filter leaderboard by language', async ({ page }) => {
    await page.click('text=Leaderboard');
    
    // Select Python filter
    await page.click('[data-testid="filter-language"]');
    await page.click('text=Python');
    
    // Verify filter is applied
    await expect(page.locator('[data-testid="active-filter-python"]')).toBeVisible();
  });

  test('should filter leaderboard by time period', async ({ page }) => {
    await page.click('text=Leaderboard');
    
    // Select weekly filter
    await page.click('[data-testid="filter-period"]');
    await page.click('text=Weekly');
    
    // Verify filter is applied
    await expect(page.locator('[data-testid="active-filter-weekly"]')).toBeVisible();
  });

  test('should display user rank if logged in', async ({ page }) => {
    // Login first (assuming auth is implemented)
    await page.click('text=Login');
    await page.fill('[data-testid="email-input"]', 'test@example.com');
    await page.fill('[data-testid="password-input"]', 'password123');
    await page.click('[data-testid="login-button"]');
    
    // Navigate to leaderboard
    await page.click('text=Leaderboard');
    
    // Verify user's rank is highlighted
    await expect(page.locator('[data-testid="user-rank-highlight"]')).toBeVisible();
  });

  test('should paginate leaderboard results', async ({ page }) => {
    await page.click('text=Leaderboard');
    
    // Wait for initial results
    await page.waitForSelector('[data-testid="leaderboard-row"]');
    
    // Click next page
    await page.click('[data-testid="next-page"]');
    
    // Verify page changed
    await expect(page.locator('[data-testid="page-indicator"]')).toContainText('2');
  });
});
