import { test, expect } from '@playwright/test';

test.describe('Language Selection E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  const languages = ['C++', 'Rust', 'Python', 'JavaScript', 'YAML'];

  for (const language of languages) {
    test(`should display tutorials for ${language} when selected`, async ({ page }) => {
      await page.click('text=Practice');
      await page.click(`text=${language}`);
      
      // Verify that the tutorial list for the selected language is visible
      await expect(page.locator(`[data-testid="lesson-list-${language.toLowerCase()}"]`)).toBeVisible();
      
      // Verify that at least one lesson is displayed
      await expect(page.locator(`[data-testid^="lesson-intro-"]`)).toBeVisible();
    });
  }

  test('should navigate back to language selection from tutorial list', async ({ page }) => {
    await page.click('text=Practice');
    await page.click('text=Python');
    
    // Click on a back button or similar element to go back to language selection
    // Assuming there's a back button with data-testid="back-button"
    await page.click('[data-testid="back-button"]');
    
    // Verify that the language selection is visible again
    await expect(page.locator('text=Select a Language')).toBeVisible();
  });
});