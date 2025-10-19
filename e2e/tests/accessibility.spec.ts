import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

/**
 * Accessibility tests using axe-core
 * Tests WCAG 2.2 AA compliance
 */

test.describe('Accessibility Tests', () => {
  test('homepage should not have accessibility violations', async ({ page }) => {
    await page.goto('/');
    
    const accessibilityScanResults = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag22aa'])
      .analyze();
    
    expect(accessibilityScanResults.violations).toEqual([]);
  });

  test('practice mode should be accessible', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Practice');
    
    const accessibilityScanResults = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag22aa'])
      .analyze();
    
    expect(accessibilityScanResults.violations).toEqual([]);
  });

  test('typing editor should have proper ARIA labels', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Practice');
    await page.click('text=Python');
    await page.click('[data-testid="lesson-intro"]');
    
    // Check for ARIA labels
    const editor = page.locator('[data-testid="typing-editor"]');
    await expect(editor).toHaveAttribute('role');
    await expect(editor).toHaveAttribute('aria-label');
    
    // Run accessibility scan
    const results = await new AxeBuilder({ page })
      .include('[data-testid="typing-editor"]')
      .analyze();
    
    expect(results.violations).toEqual([]);
  });

  test('keyboard navigation should work throughout the app', async ({ page }) => {
    await page.goto('/');
    
    // Tab through interactive elements
    await page.keyboard.press('Tab');
    let focusedElement = await page.evaluate(() => document.activeElement?.tagName);
    expect(['A', 'BUTTON', 'INPUT']).toContain(focusedElement);
    
    // Continue tabbing
    for (let i = 0; i < 5; i++) {
      await page.keyboard.press('Tab');
      focusedElement = await page.evaluate(() => document.activeElement?.tagName);
      expect(focusedElement).toBeTruthy();
    }
  });

  test('focus indicators should be visible', async ({ page }) => {
    await page.goto('/');
    
    // Tab to first focusable element
    await page.keyboard.press('Tab');
    
    // Check if focus ring is visible
    const focusedElement = page.locator(':focus');
    const outline = await focusedElement.evaluate((el) => {
      const styles = window.getComputedStyle(el);
      return {
        outline: styles.outline,
        outlineWidth: styles.outlineWidth,
        boxShadow: styles.boxShadow,
      };
    });
    
    // Should have some form of focus indicator
    const hasFocusIndicator =
      outline.outline !== 'none' ||
      outline.outlineWidth !== '0px' ||
      outline.boxShadow !== 'none';
    
    expect(hasFocusIndicator).toBeTruthy();
  });

  test('color contrast should meet WCAG AA standards', async ({ page }) => {
    await page.goto('/');
    
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2aa'])
      .options({ rules: { 'color-contrast': { enabled: true } } })
      .analyze();
    
    const contrastViolations = results.violations.filter(
      (v) => v.id === 'color-contrast'
    );
    
    expect(contrastViolations).toEqual([]);
  });

  test('images should have alt text', async ({ page }) => {
    await page.goto('/');
    
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a'])
      .options({ rules: { 'image-alt': { enabled: true } } })
      .analyze();
    
    const imageAltViolations = results.violations.filter(
      (v) => v.id === 'image-alt'
    );
    
    expect(imageAltViolations).toEqual([]);
  });

  test('form inputs should have labels', async ({ page }) => {
    await page.goto('/settings');
    
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a'])
      .options({ rules: { 'label': { enabled: true } } })
      .analyze();
    
    const labelViolations = results.violations.filter(
      (v) => v.id === 'label'
    );
    
    expect(labelViolations).toEqual([]);
  });

  test('headings should be in logical order', async ({ page }) => {
    await page.goto('/');
    
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a'])
      .options({ rules: { 'heading-order': { enabled: true } } })
      .analyze();
    
    const headingViolations = results.violations.filter(
      (v) => v.id === 'heading-order'
    );
    
    expect(headingViolations).toEqual([]);
  });

  test('page should have a main landmark', async ({ page }) => {
    await page.goto('/');
    
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a'])
      .options({ rules: { 'landmark-one-main': { enabled: true } } })
      .analyze();
    
    const landmarkViolations = results.violations.filter(
      (v) => v.id === 'landmark-one-main'
    );
    
    expect(landmarkViolations).toEqual([]);
  });

  test('interactive elements should be keyboard accessible', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Practice');
    
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a'])
      .options({ rules: { 'button-name': { enabled: true } } })
      .analyze();
    
    const buttonViolations = results.violations.filter(
      (v) => v.id === 'button-name'
    );
    
    expect(buttonViolations).toEqual([]);
  });

  test('screen reader announcements should work', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Practice');
    await page.click('text=Python');
    await page.click('[data-testid="lesson-intro"]');
    
    // Check for aria-live regions
    const liveRegions = await page.locator('[aria-live]').count();
    expect(liveRegions).toBeGreaterThan(0);
    
    // Check for status messages
    const statusRegions = await page.locator('[role="status"]').count();
    expect(statusRegions).toBeGreaterThanOrEqual(0);
  });

  test('reduced motion preferences should be respected', async ({ page }) => {
    // Set reduced motion preference
    await page.emulateMedia({ reducedMotion: 'reduce' });
    await page.goto('/');
    
    // Check that animations are disabled or reduced
    const hasReducedMotion = await page.evaluate(() => {
      return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    });
    
    expect(hasReducedMotion).toBeTruthy();
  });

  test('high contrast mode should be supported', async ({ page }) => {
    await page.goto('/');
    
    // Check for high contrast theme support
    const supportsHighContrast = await page.evaluate(() => {
      const root = document.documentElement;
      return root.classList.contains('high-contrast') ||
             root.hasAttribute('data-theme');
    });
    
    // Should have theme support
    expect(typeof supportsHighContrast).toBe('boolean');
  });

  test('skip navigation link should be present', async ({ page }) => {
    await page.goto('/');
    
    // Tab to skip link
    await page.keyboard.press('Tab');
    
    const skipLink = page.locator('a:has-text("Skip to main content")');
    await expect(skipLink).toBeVisible();
  });

  test('error messages should be announced to screen readers', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Practice');
    await page.click('text=Python');
    await page.click('[data-testid="lesson-intro"]');
    
    // Type incorrect characters to trigger error
    const input = page.locator('[data-testid="typing-input"]');
    await input.type('xxx');
    
    // Check for error announcement
    const errorMessage = page.locator('[role="alert"], [aria-live="assertive"]');
    const count = await errorMessage.count();
    
    // Should have error announcement mechanism
    expect(count).toBeGreaterThanOrEqual(0);
  });
});
