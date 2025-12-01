/**
 * Error Recovery Service
 * Provides automatic error recovery and user guidance for system failures
 */

import {
  getErrorHandler,
  AppError,
  ErrorType,
  ErrorSeverity,
} from './ErrorHandler';
import { getGracefulDegradationService } from './GracefulDegradation';
import { getMonitoringService } from './MonitoringService';

export interface RecoveryStrategy {
  name: string;
  description: string;
  automatic: boolean;
  execute: () => Promise<boolean>;
  rollback?: () => Promise<void>;
  estimatedTime?: number; // in milliseconds
  successRate?: number; // 0-1
}

export interface RecoverySession {
  id: string;
  startTime: number;
  endTime?: number;
  originalError: AppError;
  strategiesAttempted: Array<{
    strategy: string;
    success: boolean;
    duration: number;
    error?: Error;
  }>;
  finalOutcome: 'recovered' | 'partial' | 'failed';
  userInterventionRequired: boolean;
}

class ErrorRecoveryService {
  private errorHandler = getErrorHandler();
  private gracefulDegradation = getGracefulDegradationService();
  private monitoringService = getMonitoringService();
  private activeRecoveries = new Map<string, RecoverySession>();
  private recoveryStrategies = new Map<ErrorType, RecoveryStrategy[]>();

  constructor() {
    this.initializeRecoveryStrategies();
    this.setupEventListeners();
  }

  /**
   * Initialize recovery strategies for different error types
   */
  private initializeRecoveryStrategies(): void {
    // Network error recovery strategies
    this.recoveryStrategies.set(ErrorType.NETWORK_ERROR, [
      {
        name: 'retry_with_backoff',
        description: 'Retry the operation with exponential backoff',
        automatic: true,
        execute: async () => {
          // This would be implemented by the calling code
          return false; // Placeholder
        },
        estimatedTime: 5000,
        successRate: 0.7,
      },
      {
        name: 'switch_to_offline',
        description: 'Switch to offline mode and use cached data',
        automatic: true,
        execute: async () => {
          this.gracefulDegradation.enableFallback(
            'networkConnection',
            'Network error recovery'
          );
          return true;
        },
        estimatedTime: 1000,
        successRate: 0.9,
      },
      {
        name: 'reduce_request_size',
        description: 'Reduce request payload size and retry',
        automatic: true,
        execute: async () => {
          // Implementation would depend on the specific operation
          return false; // Placeholder
        },
        estimatedTime: 2000,
        successRate: 0.5,
      },
    ]);

    // Parsing error recovery strategies
    this.recoveryStrategies.set(ErrorType.PARSING_ERROR, [
      {
        name: 'fallback_parser',
        description: 'Use fallback parsing method',
        automatic: true,
        execute: async () => {
          this.gracefulDegradation.enableFallback(
            'treeSitter',
            'Parsing error recovery'
          );
          return true;
        },
        estimatedTime: 500,
        successRate: 0.8,
      },
      {
        name: 'simplify_content',
        description: 'Simplify content and retry parsing',
        automatic: true,
        execute: async () => {
          // This would involve content preprocessing
          return false; // Placeholder
        },
        estimatedTime: 1000,
        successRate: 0.6,
      },
      {
        name: 'skip_problematic_sections',
        description: 'Skip problematic code sections',
        automatic: false,
        execute: async () => {
          // This would require user confirmation
          return false; // Placeholder
        },
        estimatedTime: 2000,
        successRate: 0.9,
      },
    ]);

    // Storage error recovery strategies
    this.recoveryStrategies.set(ErrorType.STORAGE_ERROR, [
      {
        name: 'clear_old_data',
        description: 'Clear old data to free up storage space',
        automatic: true,
        execute: async () => {
          try {
            this.clearOldStorageData();
            return true;
          } catch (error) {
            return false;
          }
        },
        estimatedTime: 2000,
        successRate: 0.7,
      },
      {
        name: 'switch_to_memory',
        description: 'Use in-memory storage as fallback',
        automatic: true,
        execute: async () => {
          this.gracefulDegradation.enableFallback(
            'indexedDB',
            'Storage error recovery'
          );
          return true;
        },
        estimatedTime: 500,
        successRate: 0.9,
      },
      {
        name: 'compress_data',
        description: 'Compress data before storage',
        automatic: true,
        execute: async () => {
          // Implementation would involve data compression
          return false; // Placeholder
        },
        estimatedTime: 1500,
        successRate: 0.6,
      },
    ]);

    // Worker error recovery strategies
    this.recoveryStrategies.set(ErrorType.WORKER_ERROR, [
      {
        name: 'restart_worker',
        description: 'Restart the web worker',
        automatic: true,
        execute: async () => {
          try {
            // This would integrate with WorkerManager
            return false; // Placeholder
          } catch (error) {
            return false;
          }
        },
        estimatedTime: 3000,
        successRate: 0.8,
      },
      {
        name: 'fallback_to_main_thread',
        description: 'Execute operations on main thread',
        automatic: true,
        execute: async () => {
          this.gracefulDegradation.enableFallback(
            'webWorkers',
            'Worker error recovery'
          );
          return true;
        },
        estimatedTime: 1000,
        successRate: 0.9,
      },
    ]);

    // Session error recovery strategies
    this.recoveryStrategies.set(ErrorType.SESSION_ERROR, [
      {
        name: 'restore_from_backup',
        description: 'Restore session from backup data',
        automatic: true,
        execute: async () => {
          try {
            // This would integrate with SessionManager
            return false; // Placeholder
          } catch (error) {
            return false;
          }
        },
        estimatedTime: 2000,
        successRate: 0.7,
      },
      {
        name: 'create_new_session',
        description: 'Create a new session with current state',
        automatic: false,
        execute: async () => {
          // This would require user confirmation
          return false; // Placeholder
        },
        estimatedTime: 1000,
        successRate: 0.9,
      },
    ]);
  }

  /**
   * Set up event listeners for automatic recovery
   */
  private setupEventListeners(): void {
    // Listen for errors that might need recovery
    window.addEventListener('app-error', ((event: CustomEvent<AppError>) => {
      this.handleErrorForRecovery(event.detail);
    }) as EventListener);

    // Listen for network status changes
    window.addEventListener('online', () => {
      this.handleNetworkRestore();
    });

    window.addEventListener('offline', () => {
      this.handleNetworkLoss();
    });
  }

  /**
   * Handle an error and attempt recovery
   */
  async handleErrorForRecovery(
    error: AppError
  ): Promise<RecoverySession | null> {
    // Don't attempt recovery for low-severity errors
    if (error.severity === ErrorSeverity.LOW) {
      return null;
    }

    // Check if we already have an active recovery for this error type
    const existingRecovery = Array.from(this.activeRecoveries.values()).find(
      session => session.originalError.type === error.type && !session.endTime
    );

    if (existingRecovery) {
      return existingRecovery; // Recovery already in progress
    }

    const strategies = this.recoveryStrategies.get(error.type);
    if (!strategies || strategies.length === 0) {
      return null; // No recovery strategies available
    }

    return this.executeRecoverySession(error, strategies);
  }

  /**
   * Execute a recovery session
   */
  private async executeRecoverySession(
    error: AppError,
    strategies: RecoveryStrategy[]
  ): Promise<RecoverySession> {
    const session: RecoverySession = {
      id: this.generateRecoveryId(),
      startTime: Date.now(),
      originalError: error,
      strategiesAttempted: [],
      finalOutcome: 'failed',
      userInterventionRequired: false,
    };

    this.activeRecoveries.set(session.id, session);

    // Sort strategies by success rate and automatic execution
    const sortedStrategies = [...strategies].sort((a, b) => {
      if (a.automatic !== b.automatic) {
        return a.automatic ? -1 : 1; // Automatic strategies first
      }
      return (b.successRate || 0) - (a.successRate || 0); // Higher success rate first
    });

    let recovered = false;

    for (const strategy of sortedStrategies) {
      if (!strategy.automatic && !session.userInterventionRequired) {
        session.userInterventionRequired = true;
        // For now, skip manual strategies in automatic recovery
        continue;
      }

      const startTime = Date.now();
      let success = false;
      let strategyError: Error | undefined;

      try {
        this.monitoringService.getSystemHealth(); // Log recovery attempt
        success = await strategy.execute();
      } catch (err) {
        strategyError = err as Error;
        success = false;
      }

      const duration = Date.now() - startTime;

      session.strategiesAttempted.push({
        strategy: strategy.name,
        success,
        duration,
        error: strategyError,
      });

      if (success) {
        recovered = true;
        break;
      }

      // Wait a bit before trying the next strategy
      await this.delay(500);
    }

    // Finalize session
    session.endTime = Date.now();
    session.finalOutcome = recovered
      ? 'recovered'
      : session.strategiesAttempted.some(s => s.success)
        ? 'partial'
        : 'failed';

    // Log recovery results
    this.logRecoverySession(session);

    return session;
  }

  /**
   * Handle network restoration
   */
  private async handleNetworkRestore(): Promise<void> {
    // Try to restore network-dependent features
    const networkRestored =
      await this.gracefulDegradation.tryRestoreFeature('networkConnection');

    if (networkRestored) {
      // Attempt to restore other features that depend on network
      await this.gracefulDegradation.tryRestoreFeature('serviceWorker');

      // Notify user of restoration
      const restorationEvent = new CustomEvent('network-restored', {
        detail: { timestamp: Date.now() },
      });
      window.dispatchEvent(restorationEvent);
    }
  }

  /**
   * Handle network loss
   */
  private handleNetworkLoss(): void {
    // Enable offline mode
    this.gracefulDegradation.enableFallback(
      'networkConnection',
      'Network connection lost'
    );

    // Notify user of offline mode
    const offlineEvent = new CustomEvent('network-lost', {
      detail: { timestamp: Date.now() },
    });
    window.dispatchEvent(offlineEvent);
  }

  /**
   * Get recovery suggestions for an error
   */
  getRecoverySuggestions(error: AppError): Array<{
    title: string;
    description: string;
    action: () => Promise<void>;
    automatic: boolean;
  }> {
    const strategies = this.recoveryStrategies.get(error.type) || [];

    return strategies.map(strategy => ({
      title: strategy.name
        .replace(/_/g, ' ')
        .replace(/\b\w/g, l => l.toUpperCase()),
      description: strategy.description,
      action: async (): Promise<void> => {
        await strategy.execute();
      },
      automatic: strategy.automatic,
    }));
  }

  /**
   * Get active recovery sessions
   */
  getActiveRecoveries(): RecoverySession[] {
    return Array.from(this.activeRecoveries.values()).filter(
      session => !session.endTime
    );
  }

  /**
   * Get recovery history
   */
  getRecoveryHistory(limit: number = 10): RecoverySession[] {
    return Array.from(this.activeRecoveries.values())
      .filter(session => session.endTime)
      .sort((a, b) => (b.endTime || 0) - (a.endTime || 0))
      .slice(0, limit);
  }

  /**
   * Clear old storage data
   */
  private clearOldStorageData(): void {
    // Clear old localStorage entries
    const cutoffTime = Date.now() - 7 * 24 * 60 * 60 * 1000; // 7 days
    const keysToRemove: string[] = [];

    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key && key.startsWith('typing_master_')) {
        try {
          const item = localStorage.getItem(key);
          if (item) {
            const parsed = JSON.parse(item);
            if (parsed.timestamp && parsed.timestamp < cutoffTime) {
              keysToRemove.push(key);
            }
          }
        } catch (error) {
          keysToRemove.push(key);
        }
      }
    }

    keysToRemove.forEach(key => localStorage.removeItem(key));
  }

  /**
   * Log recovery session results
   */
  private logRecoverySession(session: RecoverySession): void {
    const duration = (session.endTime || Date.now()) - session.startTime;

    console.group(`🔄 Recovery Session ${session.id}`);
    console.log(
      'Original Error:',
      session.originalError.type,
      session.originalError.message
    );
    console.log('Duration:', `${duration}ms`);
    console.log('Outcome:', session.finalOutcome);
    console.log('Strategies Attempted:', session.strategiesAttempted.length);

    session.strategiesAttempted.forEach((attempt, index) => {
      console.log(
        `  ${index + 1}. ${attempt.strategy}: ${attempt.success ? '✅' : '❌'} (${attempt.duration}ms)`
      );
      if (attempt.error) {
        console.log(`     Error: ${attempt.error.message}`);
      }
    });

    console.groupEnd();
  }

  /**
   * Generate unique recovery ID
   */
  private generateRecoveryId(): string {
    return `recovery_${Date.now()}_${Math.random().toString(36).substring(2, 11)}`;
  }

  /**
   * Delay utility
   */
  private async delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Cleanup completed recovery sessions
   */
  cleanup(): void {
    const cutoffTime = Date.now() - 60 * 60 * 1000; // 1 hour ago

    for (const [id, session] of this.activeRecoveries.entries()) {
      if (session.endTime && session.endTime < cutoffTime) {
        this.activeRecoveries.delete(id);
      }
    }
  }
}

// Singleton instance
let errorRecoveryServiceInstance: ErrorRecoveryService | null = null;

export function getErrorRecoveryService(): ErrorRecoveryService {
  if (!errorRecoveryServiceInstance) {
    errorRecoveryServiceInstance = new ErrorRecoveryService();
  }
  return errorRecoveryServiceInstance;
}

export default ErrorRecoveryService;
