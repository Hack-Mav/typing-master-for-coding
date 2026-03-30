# Testing Guide - Typing Master for Coding

This guide covers all testing strategies implemented for the Typing Master application.

## Table of Contents
1. [Unit Tests](#unit-tests)
2. [Property-Based Tests](#property-based-tests)
3. [End-to-End Tests](#end-to-end-tests)
4. [Accessibility Tests](#accessibility-tests)
5. [Load Tests](#load-tests)
6. [Running All Tests](#running-all-tests)

## Prerequisites

### Frontend Testing
```bash
cd typing-master
npm install
```

### E2E Testing
```bash
cd e2e
npm install
npx playwright install
```

### Load Testing
```bash
# Install K6
# Windows (using Chocolatey)
choco install k6

# macOS
brew install k6

# Linux
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

## Unit Tests

### Running Unit Tests

```bash
cd typing-master

# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run with coverage
npm test -- --coverage

# Run specific test file
npm test -- MetricsCalculator.test
```

### Test Coverage

Unit tests cover:
- MetricsCalculator
- ParserManager
- SessionManager
- TypingValidationEngine
- AuthService
- ContentService
- And more...

**Coverage Goals:**
- Line coverage: >80%
- Branch coverage: >75%
- Function coverage: >85%

## Property-Based Tests

Property-based tests use `fast-check` to validate algorithmic invariants.

### Running Property-Based Tests

```bash
cd typing-master

# Run only property-based tests
npm run test:property

# Run specific property test
npm test -- MetricsCalculator.property.test
```

### What's Tested

**MetricsCalculator Properties:**
- CPM is always non-negative
- WPM = CPM / 5
- Accuracy is between 0-100%
- Perfect match gives 100% accuracy
- Composite score properties
- Error analysis invariants

**ParserManager Properties:**
- Parsing idempotency
- Token ordering
- AST structure validity
- Parent-child relationships
- Error detection

### Adding New Property Tests

```typescript
import * as fc from 'fast-check';

test('your property', () => {
  fc.assert(
    fc.property(
      fc.string(), // Generator
      (input) => {
        // Your test logic
        return true; // Property should hold
      }
    ),
    { numRuns: 100 } // Number of test cases
  );
});
```

## End-to-End Tests

E2E tests use Playwright to test complete user workflows.

### Running E2E Tests

```bash
cd e2e

# Run all E2E tests
npm test

# Run in headed mode (see browser)
npm run test:headed

# Run in UI mode (interactive)
npm run test:ui

# Run specific browser
npm run test:chromium
npm run test:firefox
npm run test:webkit

# Run mobile tests
npm run test:mobile

# Debug mode
npm run test:debug

# View report
npm run report
```

### Test Scenarios

1. **Typing Session Tests** (`typing-session.spec.ts`)
   - Homepage loading
   - Starting practice sessions
   - Typing input handling
   - Metrics calculation
   - Error tracking
   - Zen Mode
   - Timed Drill mode
   - Offline functionality
   - Keyboard shortcuts
   - Keystroke latency measurement

2. **Leaderboard Tests** (`leaderboard.spec.ts`)
   - Leaderboard display
   - Filtering by language
   - Filtering by time period
   - User rank highlighting
   - Pagination

### Writing New E2E Tests

```typescript
import { test, expect } from '@playwright/test';

test('my test', async ({ page }) => {
  await page.goto('/');
  await page.click('text=Button');
  await expect(page.locator('[data-testid="result"]')).toBeVisible();
});
```

## Accessibility Tests

Accessibility tests ensure WCAG 2.2 AA compliance using axe-core.

### Running Accessibility Tests

```bash
cd e2e

# Run all accessibility tests
npm test tests/accessibility.spec.ts

# Run in headed mode
npm run test:headed tests/accessibility.spec.ts
```

### What's Tested

- No accessibility violations
- Proper ARIA labels
- Keyboard navigation
- Visible focus indicators
- Color contrast (4.5:1 ratio)
- Image alt text
- Form input labels
- Logical heading order
- Landmark regions
- Screen reader support
- Reduced motion support
- High contrast mode

### Accessibility Checklist

- [ ] All interactive elements are keyboard accessible
- [ ] Focus indicators are visible
- [ ] Color contrast meets WCAG AA standards
- [ ] All images have alt text
- [ ] Form inputs have labels
- [ ] Headings are in logical order
- [ ] Page has main landmark
- [ ] Screen reader announcements work
- [ ] Reduced motion is respected
- [ ] Skip navigation link is present

## Load Tests

Load tests verify the application can handle production traffic.

### Running Load Tests

```bash
cd load-tests

# Run all load test scenarios
./run-load-tests.sh

# Run specific test
k6 run k6-load-test.js

# Run with custom parameters
k6 run --vus 1000 --duration 5m k6-load-test.js

# Run with environment variables
BASE_URL=https://staging.typing-master.app k6 run k6-load-test.js
```

### Test Scenarios

1. **Smoke Test**: Quick validation (1 user, 30s)
2. **Load Test**: Normal capacity (5k concurrent users)
3. **Stress Test**: Beyond capacity (15k users)
4. **Spike Test**: Sudden traffic spike
5. **Soak Test**: Extended duration (1 hour)

### Performance Targets

- **Concurrent Users**: 5,000+
- **Events/sec/user**: 500
- **API Response Time (p95)**: <200ms
- **Keystroke Latency (p95)**: <8ms
- **Error Rate**: <1%

### Interpreting Results

```bash
# View summary
cat results/load_*.summary.json

# View detailed metrics
cat results/load_*.json | jq '.metrics'

# Check thresholds
cat results/load_*.json | jq '.root_group.checks'
```

## Running All Tests

### Complete Test Suite

```bash
# 1. Frontend unit tests
cd typing-master
npm test

# 2. Property-based tests
npm run test:property

# 3. E2E tests
cd ../e2e
npm test

# 4. Accessibility tests
npm test tests/accessibility.spec.ts

# 5. Load tests (requires running backend)
cd ../load-tests
./run-load-tests.sh
```

### CI/CD Pipeline

The complete test suite runs automatically on:
- Pull requests
- Commits to main branch
- Scheduled nightly builds

**Pipeline Stages:**
1. Lint and type check
2. Unit tests with coverage
3. Property-based tests
4. Build application
5. E2E tests (Chromium only in CI)
6. Accessibility tests
7. Deploy to staging
8. Smoke tests on staging
9. Deploy to production (manual approval)

### GitHub Actions Example

```yaml
name: Test Suite

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Install dependencies
        run: |
          cd typing-master && npm ci
          cd ../e2e && npm ci
      
      - name: Run unit tests
        run: cd typing-master && npm test -- --coverage
      
      - name: Run property tests
        run: cd typing-master && npm run test:property
      
      - name: Build application
        run: cd typing-master && npm run build
      
      - name: Install Playwright
        run: cd e2e && npx playwright install --with-deps
      
      - name: Run E2E tests
        run: cd e2e && npm test
      
      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: test-results
          path: |
            typing-master/coverage/
            e2e/playwright-report/
            e2e/test-results/
```

## Test Data Management

### Test Fixtures

Test fixtures are located in:
- `typing-master/src/services/__tests__/fixtures/`
- `e2e/fixtures/`

### Mock Data

Mock data for tests:
```typescript
// Example mock session
const mockSession = {
  id: 'test-session-1',
  userId: 'test-user',
  mode: 'practice',
  languageId: 'python',
  lessonId: 'intro',
};

// Example mock events
const mockEvents = [
  { timestamp: 0, key: 'a', action: 'down' },
  { timestamp: 50, key: 'a', action: 'up' },
];
```

## Debugging Tests

### Unit Tests

```bash
# Run with Node debugger
node --inspect-brk node_modules/.bin/jest --runInBand

# Use Chrome DevTools
chrome://inspect
```

### E2E Tests

```bash
# Debug mode (pauses on failure)
npm run test:debug

# UI mode (interactive)
npm run test:ui

# Headed mode (see browser)
npm run test:headed

# Trace viewer
npx playwright show-trace trace.zip
```

### Load Tests

```bash
# Verbose output
k6 run --verbose k6-load-test.js

# Real-time metrics
k6 run --out influxdb=http://localhost:8086/k6 k6-load-test.js
```

## Performance Benchmarks

### Expected Test Durations

- Unit tests: ~30 seconds
- Property-based tests: ~2 minutes
- E2E tests (all browsers): ~10 minutes
- E2E tests (single browser): ~3 minutes
- Accessibility tests: ~2 minutes
- Load tests (full suite): ~30 minutes
- Load tests (single scenario): ~5-10 minutes

### Optimization Tips

1. **Run tests in parallel**: Use `--workers` flag
2. **Use test sharding**: Split tests across CI jobs
3. **Cache dependencies**: Cache `node_modules`
4. **Skip unnecessary tests**: Use test tags
5. **Optimize fixtures**: Reuse test data

## Troubleshooting

### Common Issues

**Issue: Tests timeout**
- Increase timeout in test configuration
- Check for infinite loops
- Verify network connectivity

**Issue: Flaky E2E tests**
- Add explicit waits
- Use `waitForSelector` instead of `sleep`
- Check for race conditions

**Issue: Load tests fail**
- Verify backend is running
- Check resource limits
- Monitor system resources

**Issue: Accessibility violations**
- Run axe DevTools extension
- Check ARIA attributes
- Verify color contrast

## Best Practices

1. **Write tests first** (TDD)
2. **Keep tests independent**
3. **Use descriptive test names**
4. **Test one thing at a time**
5. **Mock external dependencies**
6. **Clean up after tests**
7. **Use data-testid attributes**
8. **Avoid testing implementation details**
9. **Keep tests fast**
10. **Review test coverage regularly**

## Resources

- [Jest Documentation](https://jestjs.io/)
- [Playwright Documentation](https://playwright.dev/)
- [fast-check Documentation](https://fast-check.dev/)
- [axe-core Documentation](https://github.com/dequelabs/axe-core)
- [K6 Documentation](https://k6.io/docs/)
- [WCAG 2.2 Guidelines](https://www.w3.org/WAI/WCAG22/quickref/)

## Support

For issues or questions:
1. Check this guide
2. Review test logs
3. Search existing issues
4. Create new issue with reproduction steps
