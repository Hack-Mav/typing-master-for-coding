import { test, expect } from '@playwright/test';

test.describe('Typing Session E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should load homepage successfully', async ({ page }) => {
    await expect(page).toHaveTitle(/Typing Master/);
    await expect(page.locator('h1')).toBeVisible();
  });

  test('should start a practice session', async ({ page }) => {
    // Navigate to practice mode
    await page.click('text=Practice');
    
    // Select a language
    await page.click('text=Python');
    
    // Select a lesson
    await page.click('[data-testid="lesson-intro"]');
    
    // Wait for editor to load
    await expect(page.locator('[data-testid="typing-editor"]')).toBeVisible();
    
    // Verify target code is displayed
    const targetCode = await page.locator('[data-testid="target-code"]').textContent();
    expect(targetCode).toBeTruthy();
  });

  test('should handle typing input correctly', async ({ page }) => {
    // Start a session
    await page.click('text=Practice');
    await page.click('text=Python');
    await page.click('[data-testid="lesson-intro"]');
    
    // Wait for editor
    await page.waitForSelector('[data-testid="typing-input"]');
    
    // Type some characters
    const input = page.locator('[data-testid="typing-input"]');
    await input.type('def hello():');
    
    // Verify input is captured
    const inputValue = await input.inputValue();
    expect(inputValue).toBe('def hello():');
  });

  test('should calculate metrics after session', async ({ page }) => {
    // Start and complete a short session
    await page.click('text=Practice');
    await page.click('text=Python');
    await page.click('[data-testid="lesson-intro"]');
    
    // Type the target code
    const targetCode = await page.locator('[data-testid="target-code"]').textContent();
    const input = page.locator('[data-testid="typing-input"]');
    
    if (targetCode) {
      await input.type(targetCode.substring(0, 50)); // Type first 50 chars
    }
    
    // Complete session
    await page.click('[data-testid="finish-session"]');
    
    // Verify results page
    await expect(page.locator('[data-testid="results-page"]')).toBeVisible();
    
    // Check metrics are displayed
    await expect(page.locator('[data-testid="metric-cpm"]')).toBeVisible();
    await expect(page.locator('[data-testid="metric-accuracy"]')).toBeVisible();
    await expect(page.locator('[data-testid="metric-score"]')).toBeVisible();
  });

  test('should handle backspace correctly', async ({ page }) => {
    await page.click('text=Practice');
    await page.click('text=Python');
    await page.click('[data-testid="lesson-intro"]');
    
    const input = page.locator('[data-testid="typing-input"]');
    
    // Type and delete
    await input.type('hello');
    await input.press('Backspace');
    await input.press('Backspace');
    
    const value = await input.inputValue();
    expect(value).toBe('hel');
  });

  test('should track errors correctly', async ({ page }) => {
    await page.click('text=Practice');
    await page.click('text=Python');
    await page.click('[data-testid="lesson-intro"]');
    
    // Type incorrect characters
    const input = page.locator('[data-testid="typing-input"]');
    await input.type('xxx'); // Assuming this is wrong
    
    // Verify error indicator
    await expect(page.locator('[data-testid="error-indicator"]')).toBeVisible();
  });

  test('should support Zen Mode', async ({ page }) => {
    await page.click('text=Zen Mode');
    await page.click('text=Python');
    await page.click('[data-testid="lesson-intro"]');
    
    // Verify no metrics are displayed during typing
    await expect(page.locator('[data-testid="metrics-hud"]')).not.toBeVisible();
    
    // Verify minimal UI
    await expect(page.locator('[data-testid="zen-mode-indicator"]')).toBeVisible();
  });

  test('should support Timed Drill mode', async ({ page }) => {
    await page.click('text=Timed Drill');
    await page.click('text=Python');
    
    // Select duration
    await page.click('[data-testid="duration-1min"]');
    
    // Start drill
    await page.click('[data-testid="start-drill"]');
    
    // Verify timer is running
    await expect(page.locator('[data-testid="timer"]')).toBeVisible();
    
    // Verify real-time metrics
    await expect(page.locator('[data-testid="realtime-cpm"]')).toBeVisible();
  });

  test('should save session offline', async ({ page, context }) => {
    // Go offline
    await context.setOffline(true);
    
    // Start session
    await page.click('text=Practice');
    await page.click('text=Python');
    await page.click('[data-testid="lesson-intro"]');
    
    // Type some content
    const input = page.locator('[data-testid="typing-input"]');
    await input.type('def test():');
    
    // Finish session
    await page.click('[data-testid="finish-session"]');
    
    // Verify offline indicator
    await expect(page.locator('[data-testid="offline-indicator"]')).toBeVisible();
    
    // Go back online
    await context.setOffline(false);
    
    // Verify sync happens
    await page.waitForSelector('[data-testid="sync-complete"]', { timeout: 10000 });
  });

  test('should handle keyboard shortcuts', async ({ page }) => {
    await page.click('text=Practice');
    await page.click('text=Python');
    await page.click('[data-testid="lesson-intro"]');
    
    // Test Escape to pause
    await page.keyboard.press('Escape');
    await expect(page.locator('[data-testid="pause-menu"]')).toBeVisible();
    
    // Resume
    await page.keyboard.press('Escape');
    await expect(page.locator('[data-testid="pause-menu"]')).not.toBeVisible();
  });

  test('should measure keystroke latency', async ({ page }) => {
    await page.click('text=Practice');
    await page.click('text=Python');
    await page.click('[data-testid="lesson-intro"]');
    
    const input = page.locator('[data-testid="typing-input"]');
    
    // Measure latency for multiple keystrokes
    const latencies: number[] = [];
    
    for (let i = 0; i < 10; i++) {
      const start = Date.now();
      await input.type('a');
      const latency = Date.now() - start;
      latencies.push(latency);
    }
    
    // Calculate average latency
    const avgLatency = latencies.reduce((a, b) => a + b, 0) / latencies.length;
    
    // Verify latency is under target (8ms + network overhead)
    // In E2E tests, we allow more time due to network and rendering
    expect(avgLatency).toBeLessThan(50);
  });
});
