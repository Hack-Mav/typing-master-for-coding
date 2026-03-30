/**
 * Error Handling Integration Tests
 * Tests the structured logging and graceful degradation implementation
 */

import { getErrorHandler, ErrorType, ErrorSeverity } from '../ErrorHandler';
import { getGracefulDegradationService } from '../GracefulDegradation';
import { getMonitoringService } from '../MonitoringService';
import { getErrorRecoveryService } from '../ErrorRecoveryService';

describe('Error Handling and Recovery', () => {
  let errorHandler: any;
  let gracefulDegradation: any;
  let monitoringService: any;
  let errorRecoveryService: any;

  beforeEach(() => {
    errorHandler = getErrorHandler();
    gracefulDegradation = getGracefulDegradationService();
    monitoringService = getMonitoringService();
    errorRecoveryService = getErrorRecoveryService();
  });

  describe('ErrorHandler', () => {
    test('should create structured errors with proper metadata', () => {
      const error = errorHandler.createError(
        ErrorType.NETWORK_ERROR,
        'Test network error',
        {
          severity: ErrorSeverity.HIGH,
          userMessage: 'Connection failed',
          technicalDetails: { endpoint: '/api/test' },
        }
      );

      expect(error.type).toBe(ErrorType.NETWORK_ERROR);
      expect(error.severity).toBe(ErrorSeverity.HIGH);
      expect(error.message).toBe('Test network error');
      expect(error.userMessage).toBe('Connection failed');
      expect(error.technicalDetails.endpoint).toBe('/api/test');
      expect(error.retryable).toBe(true);
    });

    test('should handle network errors with retry logic', async () => {
      let attempts = 0;
      const mockOperation = jest.fn().mockImplementation(() => {
        attempts++;
        if (attempts < 3) {
          throw new Error('Network timeout');
        }
        return Promise.resolve('success');
      });

      const result = await errorHandler.handleNetworkError(mockOperation);
      expect(result).toBe('success');
      expect(attempts).toBe(3);
    });

    test('should classify errors correctly', () => {
      const networkError = new Error('fetch failed');
      const parsingError = new Error('syntax error in code');
      const storageError = new Error('quota exceeded');

      const networkAppError = errorHandler.handleError(networkError);
      const parsingAppError = errorHandler.handleError(parsingError);
      const storageAppError = errorHandler.handleError(storageError);

      expect(networkAppError.type).toBe(ErrorType.NETWORK_ERROR);
      expect(parsingAppError.type).toBe(ErrorType.PARSING_ERROR);
      expect(storageAppError.type).toBe(ErrorType.STORAGE_ERROR);
    });
  });

  describe('GracefulDegradation', () => {
    test('should detect system capabilities', () => {
      const capabilities = gracefulDegradation.getCapabilities();

      expect(capabilities).toHaveProperty('webWorkers');
      expect(capabilities).toHaveProperty('webAssembly');
      expect(capabilities).toHaveProperty('indexedDB');
      expect(capabilities).toHaveProperty('localStorage');
      expect(capabilities).toHaveProperty('serviceWorker');

      // Each capability should have the required properties
      Object.values(capabilities).forEach((capability: any) => {
        expect(capability).toHaveProperty('available');
        expect(capability).toHaveProperty('fallbackActive');
        expect(capability).toHaveProperty('lastChecked');
        expect(capability).toHaveProperty('errorCount');
      });
    });

    test('should provide degradation report', () => {
      const report = gracefulDegradation.getDegradationReport();

      expect(report).toHaveProperty('overallHealth');
      expect(report).toHaveProperty('availableFeatures');
      expect(report).toHaveProperty('totalFeatures');
      expect(report).toHaveProperty('activeFallbacks');
      expect(report).toHaveProperty('recommendations');
      expect(report).toHaveProperty('featureDetails');
      expect(report).toHaveProperty('performanceImpact');

      expect(['healthy', 'degraded', 'critical']).toContain(
        report.overallHealth
      );
      expect(typeof report.availableFeatures).toBe('number');
      expect(typeof report.totalFeatures).toBe('number');
      expect(Array.isArray(report.activeFallbacks)).toBe(true);
      expect(Array.isArray(report.recommendations)).toBe(true);
    });

    test('should handle parsing fallbacks', async () => {
      const code = 'function test() { return "hello"; }';
      const language = 'javascript';

      const mockPrimaryParser = jest
        .fn()
        .mockRejectedValue(new Error('Parser failed'));

      const result = await gracefulDegradation.parseWithFallback(
        code,
        language,
        mockPrimaryParser
      );

      expect(result).toHaveProperty('tokens');
      expect(Array.isArray(result.tokens)).toBe(true);
      expect(result.tokens.length).toBeGreaterThan(0);
    });

    test('should handle storage fallbacks', async () => {
      const key = 'test-key';
      const data = { test: 'data', timestamp: Date.now() };

      const mockPrimaryStorage = jest
        .fn()
        .mockRejectedValue(new Error('Storage failed'));

      // Should not throw error, should fall back to memory storage
      await expect(
        gracefulDegradation.storeWithFallback(key, data, mockPrimaryStorage)
      ).resolves.not.toThrow();
    });

    test('should provide user-friendly status', () => {
      const status = gracefulDegradation.getUserFriendlyStatus();

      expect(status).toHaveProperty('title');
      expect(status).toHaveProperty('message');
      expect(status).toHaveProperty('severity');
      expect(status).toHaveProperty('actions');

      expect(['info', 'warning', 'error']).toContain(status.severity);
      expect(Array.isArray(status.actions)).toBe(true);
    });
  });

  describe('MonitoringService', () => {
    test('should track session metrics', () => {
      const metrics = monitoringService.getSessionMetrics();

      expect(metrics).toHaveProperty('sessionId');
      expect(metrics).toHaveProperty('startTime');
      expect(metrics).toHaveProperty('pageViews');
      expect(metrics).toHaveProperty('errorCount');
      expect(metrics).toHaveProperty('performanceMetrics');
      expect(metrics).toHaveProperty('userActions');

      expect(typeof metrics.sessionId).toBe('string');
      expect(typeof metrics.startTime).toBe('number');
      expect(typeof metrics.pageViews).toBe('number');
      expect(typeof metrics.errorCount).toBe('number');
    });

    test('should track system health', () => {
      const health = monitoringService.getSystemHealth();

      expect(health).toHaveProperty('status');
      expect(health).toHaveProperty('uptime');
      expect(health).toHaveProperty('errorRate');
      expect(health).toHaveProperty('performanceScore');
      expect(health).toHaveProperty('featureAvailability');
      expect(health).toHaveProperty('lastUpdated');

      expect(['healthy', 'degraded', 'critical']).toContain(health.status);
      expect(typeof health.uptime).toBe('number');
      expect(typeof health.errorRate).toBe('number');
      expect(typeof health.performanceScore).toBe('number');
    });

    test('should provide monitoring statistics', () => {
      const stats = monitoringService.getMonitoringStats();

      expect(stats).toHaveProperty('errorsQueued');
      expect(stats).toHaveProperty('metricsQueued');
      expect(stats).toHaveProperty('sessionDuration');
      expect(stats).toHaveProperty('systemHealth');

      expect(typeof stats.errorsQueued).toBe('number');
      expect(typeof stats.metricsQueued).toBe('number');
      expect(typeof stats.sessionDuration).toBe('number');
    });
  });

  describe('ErrorRecoveryService', () => {
    test('should provide recovery suggestions', () => {
      const error = errorHandler.createError(
        ErrorType.NETWORK_ERROR,
        'Network connection failed'
      );

      const suggestions = errorRecoveryService.getRecoverySuggestions(error);

      expect(Array.isArray(suggestions)).toBe(true);
      expect(suggestions.length).toBeGreaterThan(0);

      suggestions.forEach((suggestion: any) => {
        expect(suggestion).toHaveProperty('title');
        expect(suggestion).toHaveProperty('description');
        expect(suggestion).toHaveProperty('action');
        expect(suggestion).toHaveProperty('automatic');

        expect(typeof suggestion.title).toBe('string');
        expect(typeof suggestion.description).toBe('string');
        expect(typeof suggestion.action).toBe('function');
        expect(typeof suggestion.automatic).toBe('boolean');
      });
    });

    test('should track recovery sessions', () => {
      const activeRecoveries = errorRecoveryService.getActiveRecoveries();
      const recoveryHistory = errorRecoveryService.getRecoveryHistory();

      expect(Array.isArray(activeRecoveries)).toBe(true);
      expect(Array.isArray(recoveryHistory)).toBe(true);
    });
  });

  describe('Integration', () => {
    test('should handle error flow from detection to recovery', async () => {
      // Simulate a parsing error
      const originalError = new Error('Parsing failed for complex syntax');

      // Handle the error
      const appError = errorHandler.handleError(originalError, {
        component: 'ParserManager',
        action: 'parse_code',
      });

      expect(appError.type).toBe(ErrorType.PARSING_ERROR);

      // Get recovery suggestions
      const suggestions = errorRecoveryService.getRecoverySuggestions(appError);
      expect(suggestions.length).toBeGreaterThan(0);

      // Check degradation status
      const report = gracefulDegradation.getDegradationReport();
      expect(report).toHaveProperty('overallHealth');

      // Verify monitoring captured the error
      const stats = monitoringService.getMonitoringStats();
      expect(typeof stats.errorsQueued).toBe('number');
    });

    test('should handle network failure scenario', async () => {
      // Simulate network failure
      const networkError = new Error('fetch failed: network timeout');
      const appError = errorHandler.handleError(networkError);

      expect(appError.type).toBe(ErrorType.NETWORK_ERROR);
      expect(appError.retryable).toBe(true);

      // Check if graceful degradation activates
      const capabilities = gracefulDegradation.getCapabilities();
      expect(capabilities.networkConnection).toBeDefined();

      // Verify recovery options are available
      const suggestions = errorRecoveryService.getRecoverySuggestions(appError);
      const offlineSuggestion = suggestions.find(
        s =>
          s.title.toLowerCase().includes('offline') ||
          s.description.toLowerCase().includes('offline')
      );
      expect(offlineSuggestion).toBeDefined();
    });
  });
});
