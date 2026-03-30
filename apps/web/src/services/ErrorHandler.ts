/**
 * Frontend Error Handling and Logging Service
 * Provides structured error handling, user-friendly messages, and graceful degradation
 */

import { getPerformanceMonitor } from './PerformanceMonitor';

// Error types for categorization
export enum ErrorType {
  NETWORK_ERROR = 'NETWORK_ERROR',
  PARSING_ERROR = 'PARSING_ERROR',
  VALIDATION_ERROR = 'VALIDATION_ERROR',
  AUTHENTICATION_ERROR = 'AUTHENTICATION_ERROR',
  AUTHORIZATION_ERROR = 'AUTHORIZATION_ERROR',
  NOT_FOUND_ERROR = 'NOT_FOUND_ERROR',
  TIMEOUT_ERROR = 'TIMEOUT_ERROR',
  STORAGE_ERROR = 'STORAGE_ERROR',
  WORKER_ERROR = 'WORKER_ERROR',
  METRICS_ERROR = 'METRICS_ERROR',
  SESSION_ERROR = 'SESSION_ERROR',
  UNKNOWN_ERROR = 'UNKNOWN_ERROR',
}

// Error severity levels
export enum ErrorSeverity {
  LOW = 'LOW',
  MEDIUM = 'MEDIUM',
  HIGH = 'HIGH',
  CRITICAL = 'CRITICAL',
}

// Structured error interface
export interface AppError {
  id: string;
  type: ErrorType;
  severity: ErrorSeverity;
  message: string;
  userMessage: string;
  technicalDetails?: Record<string, any>;
  stackTrace?: string;
  timestamp: number;
  userId?: string;
  sessionId?: string;
  requestId?: string;
  retryable: boolean;
  retryAfter?: number;
  suggestedAction?: string;
  context?: {
    component?: string;
    action?: string;
    url?: string;
    userAgent?: string;
  };
}

// Error log entry for storage
export interface ErrorLogEntry {
  id: string;
  error: AppError;
  resolved: boolean;
  reportedToServer: boolean;
  occurrenceCount: number;
  firstOccurrence: number;
  lastOccurrence: number;
}

// User-friendly error messages
const USER_MESSAGES: Record<ErrorType, string> = {
  [ErrorType.NETWORK_ERROR]:
    'Connection issue detected. Please check your internet connection.',
  [ErrorType.PARSING_ERROR]:
    'There was an issue processing the code. Please try a different snippet.',
  [ErrorType.VALIDATION_ERROR]:
    'Invalid input detected. Please check your entries and try again.',
  [ErrorType.AUTHENTICATION_ERROR]:
    'Authentication required. Please log in to continue.',
  [ErrorType.AUTHORIZATION_ERROR]:
    "You don't have permission to perform this action.",
  [ErrorType.NOT_FOUND_ERROR]: 'The requested resource could not be found.',
  [ErrorType.TIMEOUT_ERROR]: 'The operation took too long. Please try again.',
  [ErrorType.STORAGE_ERROR]:
    'Unable to save your progress. Your session will continue in memory.',
  [ErrorType.WORKER_ERROR]:
    'Background processing error. Some features may be limited.',
  [ErrorType.METRICS_ERROR]:
    'Unable to calculate metrics. Your typing session will continue.',
  [ErrorType.SESSION_ERROR]:
    'Session error occurred. Your progress has been saved.',
  [ErrorType.UNKNOWN_ERROR]: 'An unexpected error occurred. Please try again.',
};

// Suggested actions for different error types
const SUGGESTED_ACTIONS: Record<ErrorType, string> = {
  [ErrorType.NETWORK_ERROR]:
    'Check your connection and try again, or continue in offline mode',
  [ErrorType.PARSING_ERROR]:
    'Try selecting a different language or code snippet',
  [ErrorType.VALIDATION_ERROR]: 'Review your input and correct any errors',
  [ErrorType.AUTHENTICATION_ERROR]:
    'Please log in or continue in anonymous mode',
  [ErrorType.AUTHORIZATION_ERROR]:
    'Contact support if you believe this is an error',
  [ErrorType.NOT_FOUND_ERROR]: 'Go back and try a different selection',
  [ErrorType.TIMEOUT_ERROR]: 'Wait a moment and try again',
  [ErrorType.STORAGE_ERROR]:
    'Continue practicing - your session will work in memory',
  [ErrorType.WORKER_ERROR]: 'Refresh the page to restore full functionality',
  [ErrorType.METRICS_ERROR]:
    'Continue typing - metrics will be calculated when possible',
  [ErrorType.SESSION_ERROR]: 'Start a new session or refresh the page',
  [ErrorType.UNKNOWN_ERROR]:
    'Refresh the page or contact support if the issue persists',
};

class ErrorHandler {
  private errorLog: Map<string, ErrorLogEntry> = new Map();
  private errorCounts: Map<ErrorType, number> = new Map();
  private lastCleanup: number = Date.now();
  private readonly MAX_LOG_ENTRIES = 100;
  private readonly CLEANUP_INTERVAL = 60000; // 1 minute
  private readonly MAX_RETRY_ATTEMPTS = 3;
  private readonly RETRY_DELAYS = [1000, 3000, 5000]; // Progressive delays
  private performanceMonitor = getPerformanceMonitor();

  constructor() {
    this.setupGlobalErrorHandlers();
    this.startPeriodicCleanup();
  }

  /**
   * Create a structured error from various input types
   */
  createError(
    type: ErrorType,
    message: string,
    options: {
      cause?: Error;
      severity?: ErrorSeverity;
      userMessage?: string;
      technicalDetails?: Record<string, any>;
      context?: AppError['context'];
      retryable?: boolean;
      retryAfter?: number;
    } = {}
  ): AppError {
    const {
      cause,
      severity = this.getSeverityForType(type),
      userMessage = USER_MESSAGES[type],
      technicalDetails = {},
      context = {},
      retryable = this.isRetryableByDefault(type),
      retryAfter,
    } = options;

    const error: AppError = {
      id: this.generateErrorId(),
      type,
      severity,
      message,
      userMessage,
      technicalDetails: {
        ...technicalDetails,
        originalError: cause?.message,
        originalStack: cause?.stack,
      },
      stackTrace: cause?.stack || new Error().stack,
      timestamp: Date.now(),
      userId: this.getCurrentUserId(),
      sessionId: this.getCurrentSessionId(),
      requestId: this.generateRequestId(),
      retryable,
      retryAfter,
      suggestedAction: SUGGESTED_ACTIONS[type],
      context: {
        ...context,
        url: window.location.href,
        userAgent: navigator.userAgent,
      },
    };

    return error;
  }

  /**
   * Handle an error with logging and user notification
   */
  handleError(
    error: Error | AppError,
    context?: AppError['context']
  ): AppError {
    let appError: AppError;

    if (this.isAppError(error)) {
      appError = error;
    } else {
      // Convert regular Error to AppError
      const errorType = this.classifyError(error);
      appError = this.createError(errorType, error.message, {
        cause: error,
        context,
      });
    }

    // Log the error
    this.logError(appError);

    // Track error metrics
    this.trackErrorMetrics(appError);

    // Store in error log
    this.storeErrorLog(appError);

    // Report to server if configured
    this.reportErrorToServer(appError);

    return appError;
  }

  /**
   * Handle network errors with retry logic
   */
  async handleNetworkError(
    operation: () => Promise<any>,
    context?: AppError['context']
  ): Promise<any> {
    let lastError: Error | null = null;

    for (let attempt = 1; attempt <= this.MAX_RETRY_ATTEMPTS; attempt++) {
      try {
        return await operation();
      } catch (error) {
        lastError = error as Error;

        const appError = this.createError(
          ErrorType.NETWORK_ERROR,
          `Network operation failed (attempt ${attempt}/${this.MAX_RETRY_ATTEMPTS})`,
          {
            cause: lastError,
            context: {
              ...context,
              component: context?.component || 'NetworkHandler',
              action: context?.action || 'retry_attempt',
              url: context?.url,
              userAgent: context?.userAgent,
            },
            technicalDetails: {
              attempt: attempt.toString(),
            },
            retryable: attempt < this.MAX_RETRY_ATTEMPTS,
            retryAfter: this.RETRY_DELAYS[attempt - 1],
          }
        );

        this.logError(appError);

        // Don't retry on last attempt
        if (attempt === this.MAX_RETRY_ATTEMPTS) {
          throw this.handleError(lastError, context);
        }

        // Wait before retry
        await this.delay(this.RETRY_DELAYS[attempt - 1]);
      }
    }

    throw this.handleError(lastError!, context);
  }

  /**
   * Handle parsing errors with fallback
   */
  handleParsingError(
    error: Error,
    language: string,
    code: string,
    fallbackFn?: () => any
  ): any {
    const appError = this.createError(
      ErrorType.PARSING_ERROR,
      `Failed to parse ${language} code`,
      {
        cause: error,
        technicalDetails: {
          language,
          codeLength: code.length,
          codePreview: code.substring(0, 100),
        },
        context: {
          component: 'ParserManager',
          action: 'parse',
        },
      }
    );

    this.handleError(appError);

    // Try fallback if available
    if (fallbackFn) {
      try {
        console.warn('Using fallback parsing method');
        return fallbackFn();
      } catch (fallbackError) {
        console.error('Fallback parsing also failed:', fallbackError);
      }
    }

    // Return basic tokenization as last resort
    return this.basicTokenization(code);
  }

  /**
   * Handle storage errors with graceful degradation
   */
  handleStorageError(
    error: Error,
    operation: string,
    fallbackFn?: () => void
  ): void {
    const appError = this.createError(
      ErrorType.STORAGE_ERROR,
      `Storage operation failed: ${operation}`,
      {
        cause: error,
        technicalDetails: {
          operation,
          storageAvailable: this.checkStorageAvailability(),
        },
        context: {
          component: 'StorageService',
          action: operation,
        },
      }
    );

    this.handleError(appError);

    // Execute fallback if provided
    if (fallbackFn) {
      try {
        fallbackFn();
      } catch (fallbackError) {
        console.error('Storage fallback failed:', fallbackError);
      }
    }
  }

  /**
   * Handle session errors with recovery
   */
  handleSessionError(
    error: Error,
    sessionId: string,
    recoveryFn?: () => void
  ): void {
    const appError = this.createError(
      ErrorType.SESSION_ERROR,
      `Session error occurred`,
      {
        cause: error,
        technicalDetails: {
          sessionId,
          sessionState: this.getSessionState(sessionId),
        },
        context: {
          component: 'SessionManager',
          action: 'session_operation',
        },
      }
    );

    this.handleError(appError);

    // Attempt recovery
    if (recoveryFn) {
      try {
        recoveryFn();
      } catch (recoveryError) {
        console.error('Session recovery failed:', recoveryError);
      }
    }
  }

  /**
   * Get user-friendly error message for display
   */
  getUserMessage(error: AppError): string {
    return (
      error.userMessage ||
      USER_MESSAGES[error.type] ||
      USER_MESSAGES[ErrorType.UNKNOWN_ERROR]
    );
  }

  /**
   * Get suggested action for error
   */
  getSuggestedAction(error: AppError): string {
    return (
      error.suggestedAction ||
      SUGGESTED_ACTIONS[error.type] ||
      SUGGESTED_ACTIONS[ErrorType.UNKNOWN_ERROR]
    );
  }

  /**
   * Check if error is retryable
   */
  isRetryable(error: AppError): boolean {
    return error.retryable && error.type !== ErrorType.VALIDATION_ERROR;
  }

  /**
   * Get error statistics
   */
  getErrorStats(): {
    totalErrors: number;
    errorsByType: Record<ErrorType, number>;
    recentErrors: AppError[];
    criticalErrors: AppError[];
  } {
    const recentErrors: AppError[] = [];
    const criticalErrors: AppError[] = [];
    const errorsByType: Record<ErrorType, number> = {} as Record<
      ErrorType,
      number
    >;

    // Initialize counts
    Object.values(ErrorType).forEach(type => {
      errorsByType[type] = 0;
    });

    // Process error log
    this.errorLog.forEach(entry => {
      errorsByType[entry.error.type] += entry.occurrenceCount;

      if (entry.lastOccurrence > Date.now() - 300000) {
        // Last 5 minutes
        recentErrors.push(entry.error);
      }

      if (entry.error.severity === ErrorSeverity.CRITICAL) {
        criticalErrors.push(entry.error);
      }
    });

    return {
      totalErrors: Array.from(this.errorLog.values()).reduce(
        (sum, entry) => sum + entry.occurrenceCount,
        0
      ),
      errorsByType,
      recentErrors: recentErrors.slice(0, 10),
      criticalErrors: criticalErrors.slice(0, 5),
    };
  }

  /**
   * Clear error log
   */
  clearErrorLog(): void {
    this.errorLog.clear();
    this.errorCounts.clear();
  }

  // Private methods

  private setupGlobalErrorHandlers(): void {
    // Handle unhandled promise rejections
    window.addEventListener('unhandledrejection', event => {
      const error = this.createError(
        ErrorType.UNKNOWN_ERROR,
        'Unhandled promise rejection',
        {
          cause:
            event.reason instanceof Error
              ? event.reason
              : new Error(String(event.reason)),
          severity: ErrorSeverity.HIGH,
          context: {
            component: 'GlobalHandler',
            action: 'unhandled_rejection',
          },
        }
      );

      this.handleError(error);
      event.preventDefault(); // Prevent console error
    });

    // Handle global errors
    window.addEventListener('error', event => {
      const error = this.createError(
        ErrorType.UNKNOWN_ERROR,
        'Global error occurred',
        {
          cause: event.error || new Error(event.message),
          severity: ErrorSeverity.HIGH,
          technicalDetails: {
            filename: event.filename,
            lineno: event.lineno,
            colno: event.colno,
          },
          context: {
            component: 'GlobalHandler',
            action: 'global_error',
          },
        }
      );

      this.handleError(error);
    });
  }

  private isAppError(error: any): error is AppError {
    return (
      error &&
      typeof error === 'object' &&
      'type' in error &&
      'severity' in error
    );
  }

  private classifyError(error: Error): ErrorType {
    const message = error.message.toLowerCase();

    if (message.includes('network') || message.includes('fetch')) {
      return ErrorType.NETWORK_ERROR;
    }
    if (message.includes('parse') || message.includes('syntax')) {
      return ErrorType.PARSING_ERROR;
    }
    if (message.includes('timeout')) {
      return ErrorType.TIMEOUT_ERROR;
    }
    if (message.includes('storage') || message.includes('quota')) {
      return ErrorType.STORAGE_ERROR;
    }
    if (message.includes('worker')) {
      return ErrorType.WORKER_ERROR;
    }
    if (message.includes('session')) {
      return ErrorType.SESSION_ERROR;
    }
    if (
      message.includes('unauthorized') ||
      message.includes('authentication')
    ) {
      return ErrorType.AUTHENTICATION_ERROR;
    }
    if (message.includes('forbidden') || message.includes('permission')) {
      return ErrorType.AUTHORIZATION_ERROR;
    }
    if (message.includes('not found')) {
      return ErrorType.NOT_FOUND_ERROR;
    }

    return ErrorType.UNKNOWN_ERROR;
  }

  private getSeverityForType(type: ErrorType): ErrorSeverity {
    switch (type) {
      case ErrorType.AUTHENTICATION_ERROR:
      case ErrorType.AUTHORIZATION_ERROR:
        return ErrorSeverity.HIGH;
      case ErrorType.NETWORK_ERROR:
      case ErrorType.TIMEOUT_ERROR:
      case ErrorType.SESSION_ERROR:
        return ErrorSeverity.MEDIUM;
      case ErrorType.PARSING_ERROR:
      case ErrorType.STORAGE_ERROR:
      case ErrorType.WORKER_ERROR:
      case ErrorType.METRICS_ERROR:
        return ErrorSeverity.LOW;
      case ErrorType.VALIDATION_ERROR:
      case ErrorType.NOT_FOUND_ERROR:
        return ErrorSeverity.LOW;
      default:
        return ErrorSeverity.MEDIUM;
    }
  }

  private isRetryableByDefault(type: ErrorType): boolean {
    return [
      ErrorType.NETWORK_ERROR,
      ErrorType.TIMEOUT_ERROR,
      ErrorType.WORKER_ERROR,
    ].includes(type);
  }

  private logError(error: AppError): void {
    const logLevel = this.getLogLevel(error.severity);
    const logMessage = `[${error.type}] ${error.message}`;

    switch (logLevel) {
      case 'error':
        console.error(logMessage, error);
        break;
      case 'warn':
        console.warn(logMessage, error);
        break;
      default:
        console.log(logMessage, error);
    }
  }

  private getLogLevel(severity: ErrorSeverity): 'error' | 'warn' | 'log' {
    switch (severity) {
      case ErrorSeverity.CRITICAL:
      case ErrorSeverity.HIGH:
        return 'error';
      case ErrorSeverity.MEDIUM:
        return 'warn';
      default:
        return 'log';
    }
  }

  private trackErrorMetrics(error: AppError): void {
    // Track error occurrence
    const currentCount = this.errorCounts.get(error.type) || 0;
    this.errorCounts.set(error.type, currentCount + 1);

    // Record performance metric
    this.performanceMonitor.recordMetric({
      name: 'error_occurred',
      value: 1,
      timestamp: Date.now(),
      tags: {
        error_type: error.type,
        severity: error.severity,
        component: error.context?.component || 'unknown',
      },
    });
  }

  private storeErrorLog(error: AppError): void {
    const existingEntry = this.errorLog.get(error.id);

    if (existingEntry) {
      existingEntry.occurrenceCount++;
      existingEntry.lastOccurrence = error.timestamp;
    } else {
      const entry: ErrorLogEntry = {
        id: error.id,
        error,
        resolved: false,
        reportedToServer: false,
        occurrenceCount: 1,
        firstOccurrence: error.timestamp,
        lastOccurrence: error.timestamp,
      };

      this.errorLog.set(error.id, entry);
    }

    // Cleanup old entries if needed
    if (this.errorLog.size > this.MAX_LOG_ENTRIES) {
      this.cleanupOldEntries();
    }
  }

  private async reportErrorToServer(error: AppError): Promise<void> {
    try {
      const endpoint =
        (window as any).__ERROR_REPORTING_ENDPOINT__ ||
        '/api/monitoring/log-error';
      if (!endpoint) return;

      await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          error: {
            ...error,
            stackTrace: undefined, // Don't send stack trace to server
          },
          userAgent: navigator.userAgent,
          url: window.location.href,
          timestamp: Date.now(),
        }),
      });

      // Mark as reported
      const entry = this.errorLog.get(error.id);
      if (entry) {
        entry.reportedToServer = true;
      }
    } catch (reportingError) {
      console.warn('Failed to report error to server:', reportingError);
    }
  }

  private startPeriodicCleanup(): void {
    setInterval(() => {
      if (Date.now() - this.lastCleanup > this.CLEANUP_INTERVAL) {
        this.cleanupOldEntries();
        this.lastCleanup = Date.now();
      }
    }, this.CLEANUP_INTERVAL);
  }

  private cleanupOldEntries(): void {
    const cutoffTime = Date.now() - 24 * 60 * 60 * 1000; // 24 hours ago

    for (const [id, entry] of this.errorLog.entries()) {
      if (entry.lastOccurrence < cutoffTime) {
        this.errorLog.delete(id);
      }
    }
  }

  private basicTokenization(code: string): any[] {
    // Basic fallback tokenization
    return code.split(/\s+/).map((token, index) => ({
      type: 'unknown',
      value: token,
      start: 0,
      end: token.length,
      index,
    }));
  }

  private checkStorageAvailability(): Record<string, boolean> {
    return {
      localStorage: this.isStorageAvailable('localStorage'),
      sessionStorage: this.isStorageAvailable('sessionStorage'),
      indexedDB: 'indexedDB' in window,
    };
  }

  private isStorageAvailable(type: 'localStorage' | 'sessionStorage'): boolean {
    try {
      const storage = window[type];
      const test = '__storage_test__';
      storage.setItem(test, test);
      storage.removeItem(test);
      return true;
    } catch {
      return false;
    }
  }

  private getSessionState(sessionId: string): any {
    // This would integrate with SessionManager
    return {
      sessionId,
      active: true, // Placeholder
    };
  }

  private generateErrorId(): string {
    return `error_${Date.now()}_${Math.random().toString(36).substring(2, 11)}`;
  }

  private generateRequestId(): string {
    return `req_${Date.now()}_${Math.random().toString(36).substring(2, 11)}`;
  }

  private getCurrentUserId(): string | undefined {
    // This would integrate with AuthService
    return undefined; // Placeholder
  }

  private getCurrentSessionId(): string | undefined {
    // This would integrate with SessionManager
    return undefined; // Placeholder
  }

  private async delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

// Singleton instance
let errorHandlerInstance: ErrorHandler | null = null;

export function getErrorHandler(): ErrorHandler {
  if (!errorHandlerInstance) {
    errorHandlerInstance = new ErrorHandler();
  }
  return errorHandlerInstance;
}

export default ErrorHandler;
