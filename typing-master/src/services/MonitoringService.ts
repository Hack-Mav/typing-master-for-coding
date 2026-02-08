/**
 * Frontend Monitoring Service
 * Integrates with backend monitoring and provides real-time error tracking
 */

import {
  getErrorHandler,
  AppError,
  ErrorType,
  ErrorSeverity,
} from './ErrorHandler';
import { getPerformanceMonitor } from './PerformanceMonitor';

export interface MonitoringConfig {
  enableRealTimeReporting: boolean;
  batchSize: number;
  flushInterval: number;
  maxRetries: number;
  enablePerformanceTracking: boolean;
  enableUserSessionTracking: boolean;
  apiEndpoint: string;
}

export interface SessionMetrics {
  sessionId: string;
  userId?: string;
  startTime: number;
  endTime?: number;
  pageViews: number;
  errorCount: number;
  performanceMetrics: {
    avgResponseTime: number;
    slowestOperation: string;
    fastestOperation: string;
  };
  userActions: {
    clicks: number;
    keystrokes: number;
    scrolls: number;
  };
}

export interface SystemHealth {
  status: 'healthy' | 'degraded' | 'critical';
  uptime: number;
  errorRate: number;
  performanceScore: number;
  featureAvailability: Record<string, boolean>;
  lastUpdated: number;
}

class MonitoringService {
  private config: MonitoringConfig;
  private errorHandler = getErrorHandler();
  private performanceMonitor = getPerformanceMonitor();
  private sessionMetrics: SessionMetrics;
  private errorQueue: AppError[] = [];
  private metricsQueue: any[] = [];
  private flushTimer: number | null = null;
  private systemHealth: SystemHealth;

  constructor(config?: Partial<MonitoringConfig>) {
    this.config = {
      enableRealTimeReporting: true,
      batchSize: 10,
      flushInterval: 30000, // 30 seconds
      maxRetries: 3,
      enablePerformanceTracking: true,
      enableUserSessionTracking: true,
      apiEndpoint: '/api/monitoring',
      ...config,
    };

    this.sessionMetrics = this.initializeSessionMetrics();
    this.systemHealth = this.initializeSystemHealth();
    this.setupEventListeners();
    this.startPeriodicFlush();
  }

  /**
   * Initialize session metrics tracking
   */
  private initializeSessionMetrics(): SessionMetrics {
    return {
      sessionId: this.generateSessionId(),
      startTime: Date.now(),
      pageViews: 1,
      errorCount: 0,
      performanceMetrics: {
        avgResponseTime: 0,
        slowestOperation: '',
        fastestOperation: '',
      },
      userActions: {
        clicks: 0,
        keystrokes: 0,
        scrolls: 0,
      },
    };
  }

  /**
   * Initialize system health monitoring
   */
  private initializeSystemHealth(): SystemHealth {
    return {
      status: 'healthy',
      uptime: Date.now(),
      errorRate: 0,
      performanceScore: 100,
      featureAvailability: {},
      lastUpdated: Date.now(),
    };
  }

  /**
   * Set up event listeners for monitoring
   */
  private setupEventListeners(): void {
    // Listen for errors from ErrorHandler
    window.addEventListener('app-error', ((event: CustomEvent<AppError>) => {
      this.trackError(event.detail);
    }) as EventListener);

    // Track user interactions
    if (this.config.enableUserSessionTracking) {
      document.addEventListener('click', () => {
        this.sessionMetrics.userActions.clicks++;
      });

      document.addEventListener('keydown', () => {
        this.sessionMetrics.userActions.keystrokes++;
      });

      document.addEventListener('scroll', () => {
        this.sessionMetrics.userActions.scrolls++;
      });

      // Track page visibility changes
      document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'visible') {
          this.sessionMetrics.pageViews++;
        }
      });
    }

    // Track performance metrics
    if (this.config.enablePerformanceTracking) {
      this.setupPerformanceTracking();
    }

    // Handle page unload
    window.addEventListener('beforeunload', () => {
      this.finalizeSession();
    });
  }

  /**
   * Set up performance tracking
   */
  private setupPerformanceTracking(): void {
    // Track navigation timing
    if ('performance' in window && 'getEntriesByType' in performance) {
      const observer = new PerformanceObserver(list => {
        for (const entry of list.getEntries()) {
          this.trackPerformanceEntry(entry);
        }
      });

      try {
        observer.observe({ entryTypes: ['navigation', 'resource', 'measure'] });
      } catch (error) {
        console.warn('Performance observer not supported:', error);
      }
    }

    // Track long tasks
    if ('PerformanceObserver' in window) {
      try {
        const longTaskObserver = new PerformanceObserver(list => {
          for (const entry of list.getEntries()) {
            this.trackLongTask(entry);
          }
        });
        longTaskObserver.observe({ entryTypes: ['longtask'] });
      } catch (error) {
        console.warn('Long task observer not supported:', error);
      }
    }
  }

  /**
   * Track an error occurrence
   */
  trackError(error: AppError): void {
    this.sessionMetrics.errorCount++;
    this.errorQueue.push(error);

    // Update system health
    this.updateSystemHealth();

    // Immediate reporting for critical errors
    if (
      error.severity === ErrorSeverity.CRITICAL &&
      this.config.enableRealTimeReporting
    ) {
      this.reportErrorImmediate(error);
    }

    // Flush if queue is full
    if (this.errorQueue.length >= this.config.batchSize) {
      this.flushErrorQueue();
    }
  }

  /**
   * Track performance entry
   */
  private trackPerformanceEntry(entry: PerformanceEntry): void {
    const metric = {
      name: entry.name,
      type: entry.entryType,
      startTime: entry.startTime,
      duration: entry.duration,
      timestamp: Date.now(),
    };

    this.metricsQueue.push(metric);

    // Update session performance metrics
    this.updateSessionPerformanceMetrics(entry);

    // Flush if queue is full
    if (this.metricsQueue.length >= this.config.batchSize) {
      this.flushMetricsQueue();
    }
  }

  /**
   * Track long tasks that block the main thread
   */
  private trackLongTask(entry: PerformanceEntry): void {
    const longTaskError = this.errorHandler.createError(
      ErrorType.UNKNOWN_ERROR,
      `Long task detected: ${entry.duration}ms`,
      {
        severity: ErrorSeverity.MEDIUM,
        technicalDetails: {
          duration: entry.duration,
          startTime: entry.startTime,
          name: entry.name,
        },
        context: {
          component: 'PerformanceMonitor',
          action: 'long_task_detection',
        },
      }
    );

    this.trackError(longTaskError);
  }

  /**
   * Update session performance metrics
   */
  private updateSessionPerformanceMetrics(entry: PerformanceEntry): void {
    const { performanceMetrics } = this.sessionMetrics;

    // Update average response time
    if (entry.entryType === 'navigation' || entry.entryType === 'resource') {
      const currentAvg = performanceMetrics.avgResponseTime;
      performanceMetrics.avgResponseTime = (currentAvg + entry.duration) / 2;
    }

    // Track slowest and fastest operations
    if (entry.duration > 0) {
      if (
        !performanceMetrics.slowestOperation ||
        entry.duration >
          this.getOperationDuration(performanceMetrics.slowestOperation)
      ) {
        performanceMetrics.slowestOperation = `${entry.name}: ${entry.duration}ms`;
      }

      if (
        !performanceMetrics.fastestOperation ||
        entry.duration <
          this.getOperationDuration(performanceMetrics.fastestOperation)
      ) {
        performanceMetrics.fastestOperation = `${entry.name}: ${entry.duration}ms`;
      }
    }
  }

  /**
   * Extract duration from operation string
   */
  private getOperationDuration(operation: string): number {
    const match = operation.match(/(\d+(?:\.\d+)?)ms/);
    return match ? parseFloat(match[1]) : 0;
  }

  /**
   * Update system health status
   */
  private updateSystemHealth(): void {
    const now = Date.now();
    const timeWindow = 5 * 60 * 1000; // 5 minutes

    // Calculate error rate
    const recentErrors = this.errorQueue.filter(
      error => error.timestamp > now - timeWindow
    );
    this.systemHealth.errorRate = recentErrors.length / (timeWindow / 60000); // errors per minute

    // Update status based on error rate and severity
    const criticalErrors = recentErrors.filter(
      e => e.severity === ErrorSeverity.CRITICAL
    );
    const highErrors = recentErrors.filter(
      e => e.severity === ErrorSeverity.HIGH
    );

    if (criticalErrors.length > 0 || this.systemHealth.errorRate > 10) {
      this.systemHealth.status = 'critical';
      this.systemHealth.performanceScore = Math.max(
        0,
        100 - this.systemHealth.errorRate * 10
      );
    } else if (highErrors.length > 2 || this.systemHealth.errorRate > 5) {
      this.systemHealth.status = 'degraded';
      this.systemHealth.performanceScore = Math.max(
        20,
        100 - this.systemHealth.errorRate * 5
      );
    } else {
      this.systemHealth.status = 'healthy';
      this.systemHealth.performanceScore = Math.max(
        50,
        100 - this.systemHealth.errorRate
      );
    }

    this.systemHealth.lastUpdated = now;
  }

  /**
   * Report critical error immediately
   */
  private async reportErrorImmediate(error: AppError): Promise<void> {
    try {
      await this.sendToBackend('/log-error', {
        error: this.sanitizeError(error),
        session: this.sessionMetrics,
        immediate: true,
      });
    } catch (reportError) {
      console.error('Failed to report critical error:', reportError);
    }
  }

  /**
   * Flush error queue to backend
   */
  private async flushErrorQueue(): Promise<void> {
    if (this.errorQueue.length === 0) return;

    const errors = [...this.errorQueue];
    this.errorQueue = [];

    try {
      await this.sendToBackend('/log-errors', {
        errors: errors.map(e => this.sanitizeError(e)),
        session: this.sessionMetrics,
      });
    } catch (error) {
      console.error('Failed to flush error queue:', error);
      // Re-queue errors for retry
      this.errorQueue.unshift(...errors);
    }
  }

  /**
   * Flush metrics queue to backend
   */
  private async flushMetricsQueue(): Promise<void> {
    if (this.metricsQueue.length === 0) return;

    const metrics = [...this.metricsQueue];
    this.metricsQueue = [];

    try {
      await this.sendToBackend('/log-metrics', {
        metrics,
        session: this.sessionMetrics,
      });
    } catch (error) {
      console.error('Failed to flush metrics queue:', error);
      // Re-queue metrics for retry
      this.metricsQueue.unshift(...metrics);
    }
  }

  /**
   * Send data to backend with retry logic
   */
  private async sendToBackend(endpoint: string, data: any): Promise<void> {
    const url = `${this.config.apiEndpoint}${endpoint}`;

    for (let attempt = 1; attempt <= this.config.maxRetries; attempt++) {
      try {
        const response = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(data),
        });

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        return; // Success
      } catch (error) {
        if (attempt === this.config.maxRetries) {
          throw error;
        }

        // Wait before retry
        await new Promise(resolve => setTimeout(resolve, 1000 * attempt));
      }
    }
  }

  /**
   * Sanitize error for transmission (remove sensitive data)
   */
  private sanitizeError(error: AppError): Partial<AppError> {
    return {
      id: error.id,
      type: error.type,
      severity: error.severity,
      message: error.message,
      userMessage: error.userMessage,
      timestamp: error.timestamp,
      retryable: error.retryable,
      suggestedAction: error.suggestedAction,
      context: error.context,
      // Exclude stack trace and technical details for privacy
    };
  }

  /**
   * Start periodic flush of queues
   */
  private startPeriodicFlush(): void {
    this.flushTimer = window.setInterval(() => {
      this.flushErrorQueue();
      this.flushMetricsQueue();
    }, this.config.flushInterval);
  }

  /**
   * Finalize session on page unload
   */
  private finalizeSession(): void {
    this.sessionMetrics.endTime = Date.now();

    // Send final session data using sendBeacon for reliability
    if ('sendBeacon' in navigator) {
      const data = JSON.stringify({
        session: this.sessionMetrics,
        systemHealth: this.systemHealth,
      });

      navigator.sendBeacon(`${this.config.apiEndpoint}/finalize-session`, data);
    }
  }

  /**
   * Get current session metrics
   */
  getSessionMetrics(): SessionMetrics {
    return { ...this.sessionMetrics };
  }

  /**
   * Get current system health
   */
  getSystemHealth(): SystemHealth {
    return { ...this.systemHealth };
  }

  /**
   * Get monitoring statistics
   */
  getMonitoringStats(): {
    errorsQueued: number;
    metricsQueued: number;
    sessionDuration: number;
    systemHealth: SystemHealth;
  } {
    return {
      errorsQueued: this.errorQueue.length,
      metricsQueued: this.metricsQueue.length,
      sessionDuration: Date.now() - this.sessionMetrics.startTime,
      systemHealth: this.systemHealth,
    };
  }

  /**
   * Manually flush all queues
   */
  async flush(): Promise<void> {
    await Promise.all([this.flushErrorQueue(), this.flushMetricsQueue()]);
  }

  /**
   * Update monitoring configuration
   */
  updateConfig(newConfig: Partial<MonitoringConfig>): void {
    this.config = { ...this.config, ...newConfig };

    // Restart periodic flush if interval changed
    if (newConfig.flushInterval && this.flushTimer) {
      clearInterval(this.flushTimer);
      this.startPeriodicFlush();
    }
  }

  /**
   * Generate unique session ID
   */
  private generateSessionId(): string {
    return `session_${Date.now()}_${Math.random().toString(36).substring(2, 11)}`;
  }

  /**
   * Cleanup resources
   */
  destroy(): void {
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
      this.flushTimer = null;
    }

    // Final flush
    this.flush().catch(console.error);
    this.finalizeSession();
  }
}

// Singleton instance
let monitoringServiceInstance: MonitoringService | null = null;

export function getMonitoringService(
  config?: Partial<MonitoringConfig>
): MonitoringService {
  if (!monitoringServiceInstance) {
    monitoringServiceInstance = new MonitoringService(config);
  }
  return monitoringServiceInstance;
}

export default MonitoringService;
