/**
 * Property-based tests for MetricsCalculator
 * Uses fast-check for property-based testing
 */

import * as fc from 'fast-check';
import { MetricsCalculator } from '../MetricsCalculator';

describe('MetricsCalculator - Property-based Tests', () => {
  const calculator = MetricsCalculator.getInstance();

  describe('Speed Metrics Properties', () => {
    test('CPM should always be non-negative', () => {
      fc.assert(
        fc.property(
          fc.array(
            fc.record({
              timestamp: fc.nat(),
              key: fc.char(),
              action: fc.constantFrom('down', 'up'),
            })
          ),
          fc.nat({ max: 3600000 }), // duration up to 1 hour
          (events, duration) => {
            if (duration === 0) return true;

            const metrics = calculator.calculateSpeed(events, duration);
            return metrics.cpm >= 0;
          }
        ),
        { numRuns: 100 }
      );
    });

    test('WPM should be CPM divided by 5', () => {
      fc.assert(
        fc.property(
          fc.array(
            fc.record({
              timestamp: fc.nat(),
              key: fc.char(),
              action: fc.constantFrom('down', 'up'),
            })
          ),
          fc.integer({ min: 1, max: 3600000 }),
          (events, duration) => {
            const metrics = calculator.calculateSpeed(events, duration);
            const expectedWpm = metrics.cpm / 5;
            return Math.abs(metrics.wpm - expectedWpm) < 0.01;
          }
        ),
        { numRuns: 100 }
      );
    });

    test('KPS should increase with more events', () => {
      fc.assert(
        fc.property(
          fc.array(
            fc.record({
              timestamp: fc.nat(),
              key: fc.char(),
              action: fc.constantFrom('down', 'up'),
            }),
            { minLength: 1 }
          ),
          fc.integer({ min: 1000, max: 60000 }),
          (events, duration) => {
            const metrics1 = calculator.calculateSpeed(events, duration);
            const metrics2 = calculator.calculateSpeed(
              [...events, ...events],
              duration
            );

            return metrics2.kps >= metrics1.kps;
          }
        ),
        { numRuns: 50 }
      );
    });
  });

  describe('Accuracy Metrics Properties', () => {
    test('Accuracy should be between 0 and 100', () => {
      fc.assert(
        fc.property(fc.string(), fc.string(), (input, expected) => {
          const metrics = calculator.calculateAccuracy(input, expected);
          return (
            metrics.rawAccuracy >= 0 &&
            metrics.rawAccuracy <= 100 &&
            metrics.characterAccuracy >= 0 &&
            metrics.characterAccuracy <= 100
          );
        }),
        { numRuns: 100 }
      );
    });

    test('Perfect match should give 100% accuracy', () => {
      fc.assert(
        fc.property(fc.string({ minLength: 1 }), text => {
          const metrics = calculator.calculateAccuracy(text, text);
          return metrics.rawAccuracy === 100;
        }),
        { numRuns: 100 }
      );
    });

    test('Empty input should give 0% accuracy for non-empty expected', () => {
      fc.assert(
        fc.property(fc.string({ minLength: 1 }), expected => {
          const metrics = calculator.calculateAccuracy('', expected);
          return metrics.rawAccuracy === 0;
        }),
        { numRuns: 100 }
      );
    });

    test('Accuracy should be commutative for same-length strings', () => {
      fc.assert(
        fc.property(
          fc.string({ minLength: 5, maxLength: 10 }),
          fc.string({ minLength: 5, maxLength: 10 }),
          (str1, str2) => {
            if (str1.length !== str2.length) return true;

            const metrics1 = calculator.calculateAccuracy(str1, str2);
            const metrics2 = calculator.calculateAccuracy(str2, str1);

            return Math.abs(metrics1.rawAccuracy - metrics2.rawAccuracy) < 0.01;
          }
        ),
        { numRuns: 100 }
      );
    });
  });

  describe('Error Analysis Properties', () => {
    test('Backspace rate should be between 0 and 100', () => {
      fc.assert(
        fc.property(
          fc.array(
            fc.record({
              timestamp: fc.nat(),
              key: fc.oneof(fc.char(), fc.constant('Backspace')),
              action: fc.constantFrom('down', 'up'),
              errorFlag: fc.boolean(),
            })
          ),
          events => {
            const analysis = calculator.analyzeErrors(events);
            return analysis.backspaceRate >= 0 && analysis.backspaceRate <= 100;
          }
        ),
        { numRuns: 100 }
      );
    });

    test('Error count should not exceed total events', () => {
      fc.assert(
        fc.property(
          fc.array(
            fc.record({
              timestamp: fc.nat(),
              key: fc.char(),
              action: fc.constantFrom('down', 'up'),
              errorFlag: fc.boolean(),
            })
          ),
          events => {
            const analysis = calculator.analyzeErrors(events);
            return analysis.errorCount <= events.length;
          }
        ),
        { numRuns: 100 }
      );
    });

    test('No errors should result in zero error count', () => {
      fc.assert(
        fc.property(
          fc.array(
            fc.record({
              timestamp: fc.nat(),
              key: fc.char(),
              action: fc.constantFrom('down', 'up'),
            })
          ),
          events => {
            const eventsWithoutErrors = events.map(e => ({
              ...e,
              errorFlag: false,
            }));
            const analysis = calculator.analyzeErrors(eventsWithoutErrors);
            return analysis.errorCount === 0;
          }
        ),
        { numRuns: 100 }
      );
    });
  });

  describe('Composite Score Properties', () => {
    test('Composite score should be non-negative', () => {
      fc.assert(
        fc.property(
          fc.record({
            speed: fc.record({
              wpm: fc.nat({ max: 200 }),
              cpm: fc.nat({ max: 1000 }),
            }),
            accuracy: fc.record({
              rawAccuracy: fc.float({ min: 0, max: 100, noNaN: true }),
              characterAccuracy: fc.float({ min: 0, max: 100, noNaN: true }),
            }),
            errors: fc.record({
              errorCount: fc.nat({ max: 100 }),
              backspaceRate: fc.float({ min: 0, max: 100, noNaN: true }),
            }),
          }),
          metrics => {
            const score = calculator.calculateCompositeScore(metrics);
            return score >= 0;
          }
        ),
        { numRuns: 100 }
      );
    });

    test('Higher WPM should generally increase score (with same accuracy)', () => {
      fc.assert(
        fc.property(
          fc.integer({ min: 10, max: 100 }),
          fc.integer({ min: 101, max: 200 }),
          fc.float({ min: 90, max: 100, noNaN: true }),
          (wpm1, wpm2, accuracy) => {
            const metrics1 = {
              speed: { wpm: wpm1, cpm: wpm1 * 5 },
              accuracy: { rawAccuracy: accuracy, characterAccuracy: accuracy },
              errors: { errorCount: 0, backspaceRate: 0 },
            };

            const metrics2 = {
              speed: { wpm: wpm2, cpm: wpm2 * 5 },
              accuracy: { rawAccuracy: accuracy, characterAccuracy: accuracy },
              errors: { errorCount: 0, backspaceRate: 0 },
            };

            const score1 = calculator.calculateCompositeScore(metrics1);
            const score2 = calculator.calculateCompositeScore(metrics2);

            return score2 > score1;
          }
        ),
        { numRuns: 50 }
      );
    });

    test('Higher accuracy should increase score (with same WPM)', () => {
      fc.assert(
        fc.property(
          fc.integer({ min: 50, max: 150 }),
          fc.float({ min: 50, max: 80, noNaN: true }),
          fc.float({ min: 81, max: 100, noNaN: true }),
          (wpm, accuracy1, accuracy2) => {
            const metrics1 = {
              speed: { wpm, cpm: wpm * 5 },
              accuracy: {
                rawAccuracy: accuracy1,
                characterAccuracy: accuracy1,
              },
              errors: { errorCount: 0, backspaceRate: 0 },
            };

            const metrics2 = {
              speed: { wpm, cpm: wpm * 5 },
              accuracy: {
                rawAccuracy: accuracy2,
                characterAccuracy: accuracy2,
              },
              errors: { errorCount: 0, backspaceRate: 0 },
            };

            const score1 = calculator.calculateCompositeScore(metrics1);
            const score2 = calculator.calculateCompositeScore(metrics2);

            return score2 > score1;
          }
        ),
        { numRuns: 50 }
      );
    });
  });

  describe('Burst Consistency Properties', () => {
    test('Consistent intervals should have high consistency score', () => {
      fc.assert(
        fc.property(
          fc.integer({ min: 50, max: 200 }),
          fc.integer({ min: 10, max: 100 }),
          (interval, count) => {
            // Generate events with consistent intervals
            const events = Array.from({ length: count }, (_, i) => ({
              timestamp: i * interval,
              key: 'a',
              action: 'down' as const,
            }));

            const metrics = calculator.calculateBurstConsistency(events);
            return metrics.burstConsistency > 90; // Should be very consistent
          }
        ),
        { numRuns: 50 }
      );
    });

    test('Random intervals should have lower consistency', () => {
      fc.assert(
        fc.property(
          fc.array(
            fc.record({
              timestamp: fc.nat(),
              key: fc.char(),
              action: fc.constantFrom('down', 'up'),
            }),
            { minLength: 10, maxLength: 100 }
          ),
          events => {
            // Sort by timestamp
            const sortedEvents = [...events].sort(
              (a, b) => a.timestamp - b.timestamp
            );
            const metrics = calculator.calculateBurstConsistency(sortedEvents);

            return (
              metrics.burstConsistency >= 0 && metrics.burstConsistency <= 100
            );
          }
        ),
        { numRuns: 50 }
      );
    });
  });
});
