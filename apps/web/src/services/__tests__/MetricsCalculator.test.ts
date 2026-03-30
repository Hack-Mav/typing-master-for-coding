import { metricsCalculator } from '../MetricsCalculator';

// Mock the tokenization service
jest.mock('../TokenizationService', () => ({
  tokenizationService: {
    tokenizeWithMetadata: jest.fn().mockResolvedValue([]),
    compareTokenSequences: jest.fn().mockReturnValue({
      matches: [],
      accuracy: 1.0,
      tokenAccuracy: 1.0,
      firstMismatchIndex: null,
    }),
    extractWhitespaceTokens: jest.fn().mockReturnValue([]),
  },
}));

describe('MetricsCalculator', () => {
  test('should exist and be defined', () => {
    expect(metricsCalculator).toBeDefined();
  });

  test('should calculate speed metrics with empty input', () => {
    const result = metricsCalculator.calculateSpeedMetrics([], 'test', '', []);

    expect(result.cpm).toBe(0);
    expect(result.wpm).toBe(0);
    expect(result.twpm).toBe(0);
    expect(result.kps).toBe(0);
  });

  test('should calculate accuracy metrics with empty input', () => {
    const result = metricsCalculator.calculateAccuracyMetrics('', '', [], []);

    expect(result.rawAccuracy).toBe(1);
    expect(result.tokenAccuracy).toBe(1);
    expect(result.syntaxAccuracy).toBe(1);
    expect(result.whitespaceAccuracy).toBe(1);
  });

  test('should calculate timing metrics with empty input', () => {
    const result = metricsCalculator.calculateTimingMetrics([]);

    expect(result.totalDuration).toBe(0);
    expect(result.activeDuration).toBe(0);
    expect(result.idleTime).toBe(0);
    expect(result.averageKeystrokeInterval).toBe(0);
  });
});
