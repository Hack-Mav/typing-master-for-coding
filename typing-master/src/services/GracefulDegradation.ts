/**
 * Graceful Degradation Service
 * Provides fallback mechanisms and feature detection for robust operation
 */

import { getErrorHandler, ErrorType } from './ErrorHandler';
import { getPerformanceMonitor } from './PerformanceMonitor';

// Feature availability status
export interface FeatureStatus {
  available: boolean;
  fallbackActive: boolean;
  lastChecked: number;
  errorCount: number;
  message?: string;
}

// System capabilities
export interface SystemCapabilities {
  webWorkers: FeatureStatus;
  webAssembly: FeatureStatus;
  indexedDB: FeatureStatus;
  localStorage: FeatureStatus;
  serviceWorker: FeatureStatus;
  offlineSupport: FeatureStatus;
  treeSitter: FeatureStatus;
  performanceAPI: FeatureStatus;
  networkConnection: FeatureStatus;
}

// Fallback configuration
export interface FallbackConfig {
  enableFallbacks: boolean;
  maxRetryAttempts: number;
  retryDelay: number;
  featureCheckInterval: number;
  gracefulDegradationMode: 'strict' | 'permissive' | 'disabled';
}

class GracefulDegradationService {
  private capabilities: SystemCapabilities;
  private config: FallbackConfig;
  private errorHandler = getErrorHandler();
  private performanceMonitor = getPerformanceMonitor();
  private featureCheckTimer: NodeJS.Timeout | null = null;

  constructor(config?: Partial<FallbackConfig>) {
    this.config = {
      enableFallbacks: true,
      maxRetryAttempts: 3,
      retryDelay: 1000,
      featureCheckInterval: 30000, // 30 seconds
      gracefulDegradationMode: 'permissive',
      ...config,
    };

    this.capabilities = this.initializeCapabilities();
    this.startFeatureMonitoring();
  }

  /**
   * Initialize system capabilities detection
   */
  private initializeCapabilities(): SystemCapabilities {
    return {
      webWorkers: this.checkWebWorkers(),
      webAssembly: this.checkWebAssembly(),
      indexedDB: this.checkIndexedDB(),
      localStorage: this.checkLocalStorage(),
      serviceWorker: this.checkServiceWorker(),
      offlineSupport: this.checkOfflineSupport(),
      treeSitter: this.checkTreeSitter(),
      performanceAPI: this.checkPerformanceAPI(),
      networkConnection: this.checkNetworkConnection(),
    };
  }

  /**
   * Get current system capabilities
   */
  getCapabilities(): SystemCapabilities {
    return { ...this.capabilities };
  }

  /**
   * Check if a specific feature is available
   */
  isFeatureAvailable(feature: keyof SystemCapabilities): boolean {
    return this.capabilities[feature].available;
  }

  /**
   * Check if fallback is active for a feature
   */
  isFallbackActive(feature: keyof SystemCapabilities): boolean {
    return this.capabilities[feature].fallbackActive;
  }

  /**
   * Force enable fallback for a feature
   */
  enableFallback(feature: keyof SystemCapabilities, reason?: string): void {
    this.capabilities[feature].fallbackActive = true;
    this.capabilities[feature].available = false;
    this.capabilities[feature].message = reason || 'Fallback manually enabled';
    this.capabilities[feature].lastChecked = Date.now();
  }

  /**
   * Try to restore a feature from fallback mode
   */
  async tryRestoreFeature(feature: keyof SystemCapabilities): Promise<boolean> {
    const featureStatus = this.capabilities[feature];

    if (!featureStatus.fallbackActive) {
      return true; // Already working
    }

    // Re-check the feature
    let newStatus: FeatureStatus;
    switch (feature) {
      case 'webWorkers':
        newStatus = this.checkWebWorkers();
        break;
      case 'webAssembly':
        newStatus = this.checkWebAssembly();
        break;
      case 'indexedDB':
        newStatus = this.checkIndexedDB();
        break;
      case 'localStorage':
        newStatus = this.checkLocalStorage();
        break;
      case 'serviceWorker':
        newStatus = this.checkServiceWorker();
        break;
      case 'offlineSupport':
        newStatus = this.checkOfflineSupport();
        break;
      case 'treeSitter':
        newStatus = this.checkTreeSitter();
        break;
      case 'performanceAPI':
        newStatus = this.checkPerformanceAPI();
        break;
      case 'networkConnection':
        newStatus = this.checkNetworkConnection();
        break;
      default:
        return false;
    }

    if (newStatus.available) {
      this.capabilities[feature] = {
        ...newStatus,
        errorCount: 0,
        fallbackActive: false,
      };
      return true;
    }

    return false;
  }

  /**
   * Execute operation with fallback support
   */
  async withFallback<T>(
    primaryOperation: () => Promise<T>,
    fallbackOperation: () => Promise<T>,
    feature: keyof SystemCapabilities,
    context?: string
  ): Promise<T> {
    if (!this.config.enableFallbacks) {
      return primaryOperation();
    }

    const featureStatus = this.capabilities[feature];

    // If feature is known to be unavailable, use fallback immediately
    if (!featureStatus.available && featureStatus.fallbackActive) {
      try {
        return await fallbackOperation();
      } catch (error) {
        throw this.errorHandler.createError(
          ErrorType.UNKNOWN_ERROR,
          `Both primary and fallback operations failed for ${feature}`,
          {
            cause: error as Error,
            context: {
              component: 'GracefulDegradation',
              action: context || 'fallback_operation',
            },
          }
        );
      }
    }

    // Try primary operation first
    try {
      const result = await primaryOperation();

      // Reset error count on success
      if (featureStatus.errorCount > 0) {
        featureStatus.errorCount = 0;
        featureStatus.fallbackActive = false;
      }

      return result;
    } catch (error) {
      featureStatus.errorCount++;

      // Activate fallback if error threshold reached
      if (featureStatus.errorCount >= this.config.maxRetryAttempts) {
        featureStatus.fallbackActive = true;
        featureStatus.available = false;

        this.errorHandler.handleError(
          this.errorHandler.createError(
            ErrorType.UNKNOWN_ERROR,
            `Feature ${feature} degraded to fallback mode`,
            {
              cause: error as Error,
              technicalDetails: { errorCount: featureStatus.errorCount },
              context: {
                component: 'GracefulDegradation',
                action: context || 'feature_degradation',
              },
            }
          )
        );
      }

      // Try fallback operation
      try {
        return await fallbackOperation();
      } catch (fallbackError) {
        throw this.errorHandler.createError(
          ErrorType.UNKNOWN_ERROR,
          `Both primary and fallback operations failed for ${feature}`,
          {
            cause: fallbackError as Error,
            technicalDetails: {
              primaryError: (error as Error).message,
              fallbackError: (fallbackError as Error).message,
            },
            context: {
              component: 'GracefulDegradation',
              action: context || 'complete_failure',
            },
          }
        );
      }
    }
  }

  /**
   * Parser fallback mechanisms with multiple fallback levels
   */
  async parseWithFallback(
    code: string,
    language: string,
    primaryParser: () => Promise<any>,
    context?: string
  ): Promise<any> {
    // Level 1: Try primary parser (Tree-sitter)
    if (
      this.capabilities.treeSitter.available &&
      !this.capabilities.treeSitter.fallbackActive
    ) {
      try {
        return await primaryParser();
      } catch (error) {
        this.capabilities.treeSitter.errorCount++;

        // If too many errors, enable fallback
        if (
          this.capabilities.treeSitter.errorCount >=
          this.config.maxRetryAttempts
        ) {
          this.capabilities.treeSitter.fallbackActive = true;
          this.capabilities.treeSitter.available = false;
        }

        this.errorHandler.handleError(
          this.errorHandler.createError(
            ErrorType.PARSING_ERROR,
            `Primary parser failed for ${language}`,
            {
              cause: error as Error,
              technicalDetails: { language, codeLength: code.length },
              context: {
                component: 'GracefulDegradation',
                action: context || 'primary_parsing',
              },
            }
          )
        );
      }
    }

    // Level 2: Try regex-based parsing
    try {
      return await this.regexBasedParsing(code, language);
    } catch (error) {
      this.errorHandler.handleError(
        this.errorHandler.createError(
          ErrorType.PARSING_ERROR,
          `Regex parser failed for ${language}`,
          {
            cause: error as Error,
            technicalDetails: { language, codeLength: code.length },
            context: {
              component: 'GracefulDegradation',
              action: context || 'regex_parsing',
            },
          }
        )
      );
    }

    // Level 3: Basic text parsing (always works)
    try {
      return await this.basicTextParsing(code, language);
    } catch (error) {
      // This should never fail, but just in case
      return this.emergencyFallbackParsing(code);
    }
  }

  /**
   * Regex-based parsing as intermediate fallback
   */
  private async regexBasedParsing(
    code: string,
    language: string
  ): Promise<any> {
    const tokens: any[] = [];
    const patterns = this.getLanguagePatterns(language);

    const lines = code.split('\n');

    for (let lineIndex = 0; lineIndex < lines.length; lineIndex++) {
      const line = lines[lineIndex];
      let position = 0;

      // Try to match patterns in order of priority
      while (position < line.length) {
        let matched = false;

        for (const [type, pattern] of patterns) {
          const regex = new RegExp(pattern, 'g');
          regex.lastIndex = position;
          const match = regex.exec(line);

          if (match && match.index === position) {
            tokens.push({
              type,
              value: match[0],
              start: { line: lineIndex, column: position },
              end: { line: lineIndex, column: position + match[0].length },
            });

            position += match[0].length;
            matched = true;
            break;
          }
        }

        if (!matched) {
          // Skip unknown character
          position++;
        }
      }
    }

    return { tokens };
  }

  /**
   * Get regex patterns for different languages
   */
  private getLanguagePatterns(language: string): [string, string][] {
    const commonPatterns: [string, string][] = [
      ['comment', '//.*$|/\\*[\\s\\S]*?\\*/'],
      ['string', '"(?:[^"\\\\]|\\\\.)*"|\'(?:[^\'\\\\]|\\\\.)*\''],
      ['number', '\\b\\d+(?:\\.\\d+)?(?:[eE][+-]?\\d+)?\\b'],
      [
        'keyword',
        '\\b(?:' + this.getLanguageKeywords(language).join('|') + ')\\b',
      ],
      ['identifier', '\\b[a-zA-Z_][a-zA-Z0-9_]*\\b'],
      ['operator', '[+\\-*/%=<>!&|^~]|\\+\\+|--|==|!=|<=|>=|&&|\\|\\||<<|>>'],
      ['punctuation', '[{}()\\[\\];,.]'],
      ['whitespace', '\\s+'],
    ];

    // Language-specific patterns
    const languageSpecific: Record<string, [string, string][]> = {
      python: [
        ['decorator', '@[a-zA-Z_][a-zA-Z0-9_]*'],
        ['f_string', 'f"[^"]*"|f\'[^\']*\''],
      ],
      javascript: [
        ['template_literal', '`[^`]*`'],
        ['arrow_function', '=>'],
      ],
      cpp: [
        ['preprocessor', '#[a-zA-Z_][a-zA-Z0-9_]*'],
        ['scope_resolution', '::'],
      ],
      rust: [
        ['lifetime', "'[a-zA-Z_][a-zA-Z0-9_]*"],
        ['macro', '[a-zA-Z_][a-zA-Z0-9_]*!'],
      ],
    };

    return [...(languageSpecific[language] || []), ...commonPatterns];
  }

  /**
   * Emergency fallback parsing (minimal tokenization)
   */
  private emergencyFallbackParsing(code: string): any {
    const tokens = code.split(/(\s+)/).map((token, index) => ({
      type: token.trim() ? 'unknown' : 'whitespace',
      value: token,
      start: { line: 0, column: index },
      end: { line: 0, column: index + token.length },
    }));

    return { tokens };
  }

  /**
   * Storage fallback mechanisms with multiple storage options
   */
  async storeWithFallback<T>(
    key: string,
    data: T,
    primaryStorage: () => Promise<void>,
    context?: string
  ): Promise<void> {
    // Level 1: Try primary storage (IndexedDB)
    if (
      this.capabilities.indexedDB.available &&
      !this.capabilities.indexedDB.fallbackActive
    ) {
      try {
        await primaryStorage();
        return;
      } catch (error) {
        this.capabilities.indexedDB.errorCount++;

        if (
          this.capabilities.indexedDB.errorCount >= this.config.maxRetryAttempts
        ) {
          this.capabilities.indexedDB.fallbackActive = true;
          this.capabilities.indexedDB.available = false;
        }

        this.errorHandler.handleError(
          this.errorHandler.createError(
            ErrorType.STORAGE_ERROR,
            `Primary storage failed for key: ${key}`,
            {
              cause: error as Error,
              technicalDetails: { key, dataSize: JSON.stringify(data).length },
              context: {
                component: 'GracefulDegradation',
                action: context || 'primary_storage',
              },
            }
          )
        );
      }
    }

    // Level 2: Try localStorage
    if (this.capabilities.localStorage.available) {
      try {
        await this.localStorageFallback(key, data);
        return;
      } catch (error) {
        this.capabilities.localStorage.errorCount++;

        if (
          this.capabilities.localStorage.errorCount >=
          this.config.maxRetryAttempts
        ) {
          this.capabilities.localStorage.fallbackActive = true;
          this.capabilities.localStorage.available = false;
        }

        this.errorHandler.handleError(
          this.errorHandler.createError(
            ErrorType.STORAGE_ERROR,
            `localStorage failed for key: ${key}`,
            {
              cause: error as Error,
              technicalDetails: { key, dataSize: JSON.stringify(data).length },
              context: {
                component: 'GracefulDegradation',
                action: context || 'localStorage_fallback',
              },
            }
          )
        );
      }
    }

    // Level 3: Memory storage (always works)
    await this.memoryStorage(key, data);
  }

  /**
   * Retrieve data with storage fallbacks
   */
  async retrieveWithFallback<T>(
    key: string,
    primaryRetrieval: () => Promise<T | null>,
    context?: string
  ): Promise<T | null> {
    // Level 1: Try primary storage (IndexedDB)
    if (
      this.capabilities.indexedDB.available &&
      !this.capabilities.indexedDB.fallbackActive
    ) {
      try {
        const result = await primaryRetrieval();
        if (result !== null) {
          return result;
        }
      } catch (error) {
        this.errorHandler.handleError(
          this.errorHandler.createError(
            ErrorType.STORAGE_ERROR,
            `Primary retrieval failed for key: ${key}`,
            {
              cause: error as Error,
              context: {
                component: 'GracefulDegradation',
                action: context || 'primary_retrieval',
              },
            }
          )
        );
      }
    }

    // Level 2: Try localStorage
    if (this.capabilities.localStorage.available) {
      try {
        const result = this.localStorageRetrieve<T>(key);
        if (result !== null) {
          return result;
        }
      } catch (error) {
        this.errorHandler.handleError(
          this.errorHandler.createError(
            ErrorType.STORAGE_ERROR,
            `localStorage retrieval failed for key: ${key}`,
            {
              cause: error as Error,
              context: {
                component: 'GracefulDegradation',
                action: context || 'localStorage_retrieval',
              },
            }
          )
        );
      }
    }

    // Level 3: Try memory storage
    return this.memoryStorageRetrieve<T>(key);
  }

  /**
   * localStorage fallback implementation
   */
  private async localStorageFallback<T>(key: string, data: T): Promise<void> {
    try {
      const serialized = JSON.stringify({
        data,
        timestamp: Date.now(),
        version: '1.0',
      });

      localStorage.setItem(`typing_master_${key}`, serialized);
      console.warn(
        `Data stored in localStorage (key: ${key}). Limited storage capacity.`
      );
    } catch (error) {
      if (error instanceof Error && error.name === 'QuotaExceededError') {
        // Try to clear old data and retry
        this.clearOldLocalStorageData();

        try {
          const serialized = JSON.stringify({ data, timestamp: Date.now() });
          localStorage.setItem(`typing_master_${key}`, serialized);
        } catch (retryError) {
          throw new Error(
            `localStorage quota exceeded and cleanup failed: ${retryError}`
          );
        }
      } else {
        throw error;
      }
    }
  }

  /**
   * Retrieve from localStorage
   */
  private localStorageRetrieve<T>(key: string): T | null {
    try {
      const item = localStorage.getItem(`typing_master_${key}`);
      if (!item) return null;

      const parsed = JSON.parse(item);
      return parsed.data || parsed; // Handle both new and legacy formats
    } catch (error) {
      console.warn(
        `Failed to retrieve from localStorage (key: ${key}):`,
        error
      );
      return null;
    }
  }

  /**
   * Retrieve from memory storage
   */
  private memoryStorageRetrieve<T>(key: string): T | null {
    return this.memoryStorageMap.get(key) || null;
  }

  /**
   * Clear old localStorage data to free up space
   */
  private clearOldLocalStorageData(): void {
    const cutoffTime = Date.now() - 7 * 24 * 60 * 60 * 1000; // 7 days ago
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
          // Invalid JSON, remove it
          keysToRemove.push(key);
        }
      }
    }

    keysToRemove.forEach(key => localStorage.removeItem(key));
    console.log(`Cleared ${keysToRemove.length} old localStorage entries`);
  }

  /**
   * Network operation with offline fallback and retry logic
   */
  async networkWithFallback<T>(
    networkOperation: () => Promise<T>,
    offlineOperation: () => Promise<T>,
    context?: string,
    retryOptions?: {
      maxRetries?: number;
      retryDelay?: number;
      exponentialBackoff?: boolean;
    }
  ): Promise<T> {
    const options = {
      maxRetries: 3,
      retryDelay: 1000,
      exponentialBackoff: true,
      ...retryOptions,
    };

    // If we know network is down, go straight to offline
    if (!this.capabilities.networkConnection.available) {
      try {
        return await offlineOperation();
      } catch (error) {
        throw this.errorHandler.createError(
          ErrorType.NETWORK_ERROR,
          'Network unavailable and offline operation failed',
          {
            cause: error as Error,
            context: {
              component: 'GracefulDegradation',
              action: context || 'offline_fallback',
            },
          }
        );
      }
    }

    // Try network operation with retries
    let lastError: Error | null = null;
    for (let attempt = 1; attempt <= options.maxRetries; attempt++) {
      try {
        return await networkOperation();
      } catch (error) {
        lastError = error as Error;

        // Check if this is a network-related error
        const isNetworkError = this.isNetworkError(error as Error);

        if (isNetworkError) {
          // Update network status
          this.capabilities.networkConnection.available = false;
          this.capabilities.networkConnection.fallbackActive = true;
          this.capabilities.networkConnection.errorCount++;
        }

        // Don't retry on last attempt
        if (attempt === options.maxRetries) {
          break;
        }

        // Calculate delay with optional exponential backoff
        const delay = options.exponentialBackoff
          ? options.retryDelay * Math.pow(2, attempt - 1)
          : options.retryDelay;

        await this.delay(delay);
      }
    }

    // All network attempts failed, try offline operation
    try {
      return await offlineOperation();
    } catch (offlineError) {
      throw this.errorHandler.createError(
        ErrorType.NETWORK_ERROR,
        'Both network and offline operations failed',
        {
          cause: offlineError as Error,
          technicalDetails: {
            networkError: lastError?.message,
            offlineError: (offlineError as Error).message,
            attempts: options.maxRetries,
          },
          context: {
            component: 'GracefulDegradation',
            action: context || 'complete_network_failure',
          },
        }
      );
    }
  }

  /**
   * Check if an error is network-related
   */
  private isNetworkError(error: Error): boolean {
    const message = error.message.toLowerCase();
    const networkKeywords = [
      'network',
      'fetch',
      'connection',
      'timeout',
      'offline',
      'cors',
      'dns',
      'unreachable',
      'refused',
      'aborted',
    ];

    return (
      networkKeywords.some(keyword => message.includes(keyword)) ||
      error.name === 'NetworkError' ||
      (error.name === 'TypeError' && message.includes('fetch'))
    );
  }

  /**
   * Delay utility function
   */
  private async delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Web Worker with main thread fallback
   */
  async workerWithFallback<T>(
    workerOperation: () => Promise<T>,
    mainThreadOperation: () => Promise<T>,
    context?: string
  ): Promise<T> {
    return this.withFallback(
      workerOperation,
      mainThreadOperation,
      'webWorkers',
      context || 'worker_operation'
    );
  }

  /**
   * Performance monitoring with graceful degradation
   */
  measureWithFallback<T>(
    operation: () => T,
    operationName: string,
    fallbackMeasurement?: () => void
  ): T {
    if (this.capabilities.performanceAPI.available) {
      const startTime = performance.now();
      try {
        const result = operation();
        const duration = performance.now() - startTime;

        this.performanceMonitor.recordMetric({
          name: operationName,
          value: duration,
          timestamp: Date.now(),
          unit: 'ms',
        });

        return result;
      } catch (error) {
        const duration = performance.now() - startTime;
        this.performanceMonitor.recordMetric({
          name: `${operationName}_error`,
          value: duration,
          timestamp: Date.now(),
          unit: 'ms',
          tags: { error: 'true' },
        });
        throw error;
      }
    } else {
      // Fallback measurement
      if (fallbackMeasurement) {
        fallbackMeasurement();
      }
      return operation();
    }
  }

  /**
   * Get degradation status report with detailed information
   */
  getDegradationReport(): {
    overallHealth: 'healthy' | 'degraded' | 'critical';
    availableFeatures: number;
    totalFeatures: number;
    activeFallbacks: string[];
    recommendations: string[];
    featureDetails: Record<
      string,
      {
        available: boolean;
        fallbackActive: boolean;
        errorCount: number;
        message: string;
        impact: string;
        userAction?: string;
      }
    >;
    performanceImpact: {
      level: 'none' | 'low' | 'medium' | 'high';
      description: string;
    };
  } {
    const features = Object.keys(
      this.capabilities
    ) as (keyof SystemCapabilities)[];
    const availableFeatures = features.filter(
      f => this.capabilities[f].available
    ).length;
    const activeFallbacks = features.filter(
      f => this.capabilities[f].fallbackActive
    );

    let overallHealth: 'healthy' | 'degraded' | 'critical';
    if (availableFeatures === features.length) {
      overallHealth = 'healthy';
    } else if (availableFeatures >= features.length * 0.7) {
      overallHealth = 'degraded';
    } else {
      overallHealth = 'critical';
    }

    const recommendations: string[] = [];
    const featureDetails: Record<string, any> = {};

    // Analyze each feature
    features.forEach(feature => {
      const capability = this.capabilities[feature];
      const impact = this.getFeatureImpact(feature);
      const userAction = this.getFeatureUserAction(feature);

      featureDetails[feature] = {
        available: capability.available,
        fallbackActive: capability.fallbackActive,
        errorCount: capability.errorCount,
        message: capability.message || 'No additional information',
        impact,
        userAction,
      };

      // Generate recommendations
      if (!capability.available) {
        const recommendation = this.getFeatureRecommendation(feature);
        if (recommendation) {
          recommendations.push(recommendation);
        }
      }
    });

    // Calculate performance impact
    const performanceImpact = this.calculatePerformanceImpact(activeFallbacks);

    return {
      overallHealth,
      availableFeatures,
      totalFeatures: features.length,
      activeFallbacks,
      recommendations,
      featureDetails,
      performanceImpact,
    };
  }

  /**
   * Get the impact description for a feature
   */
  private getFeatureImpact(feature: keyof SystemCapabilities): string {
    const impacts: Record<keyof SystemCapabilities, string> = {
      webWorkers: 'Parsing and metrics calculation may be slower',
      webAssembly:
        'Advanced parsing features unavailable, using basic tokenization',
      indexedDB: 'Session data stored in memory only, lost on page refresh',
      localStorage: 'Limited offline storage, some preferences may not persist',
      serviceWorker: 'No offline functionality, requires internet connection',
      offlineSupport:
        'Cannot detect network status, may show incorrect connection state',
      treeSitter: 'Advanced syntax highlighting and validation unavailable',
      performanceAPI: 'Performance metrics may be inaccurate or unavailable',
      networkConnection: 'Operating in offline mode, some features unavailable',
    };

    return impacts[feature];
  }

  /**
   * Get user action suggestion for a feature
   */
  private getFeatureUserAction(
    feature: keyof SystemCapabilities
  ): string | undefined {
    const actions: Partial<Record<keyof SystemCapabilities, string>> = {
      webWorkers: 'Update to a modern browser that supports Web Workers',
      webAssembly: 'Enable WebAssembly in your browser settings',
      indexedDB: 'Allow storage permissions for this site',
      localStorage: 'Clear browser storage or allow storage permissions',
      serviceWorker: 'Enable service workers in browser settings',
      networkConnection: 'Check your internet connection and try again',
    };

    return actions[feature];
  }

  /**
   * Get recommendation for a feature
   */
  private getFeatureRecommendation(
    feature: keyof SystemCapabilities
  ): string | undefined {
    const recommendations: Partial<Record<keyof SystemCapabilities, string>> = {
      webWorkers: 'Consider using a modern browser for better performance',
      webAssembly: 'Enable WebAssembly for advanced parsing features',
      indexedDB: 'Enable browser storage for offline functionality',
      localStorage: 'Allow storage permissions to save preferences',
      serviceWorker: 'Enable service workers for offline support',
      networkConnection:
        'Check your internet connection for full functionality',
      treeSitter: 'Advanced syntax features unavailable, using basic parsing',
      performanceAPI: 'Performance monitoring may be limited',
    };

    return recommendations[feature];
  }

  /**
   * Calculate overall performance impact
   */
  private calculatePerformanceImpact(activeFallbacks: string[]): {
    level: 'none' | 'low' | 'medium' | 'high';
    description: string;
  } {
    if (activeFallbacks.length === 0) {
      return {
        level: 'none',
        description: 'All features operating at full performance',
      };
    }

    const highImpactFeatures = ['webWorkers', 'webAssembly', 'treeSitter'];
    const mediumImpactFeatures = ['indexedDB', 'serviceWorker'];

    const hasHighImpact = activeFallbacks.some(f =>
      highImpactFeatures.includes(f)
    );
    const hasMediumImpact = activeFallbacks.some(f =>
      mediumImpactFeatures.includes(f)
    );

    if (hasHighImpact) {
      return {
        level: 'high',
        description:
          'Significant performance degradation due to missing core features',
      };
    } else if (hasMediumImpact || activeFallbacks.length > 3) {
      return {
        level: 'medium',
        description: 'Moderate performance impact from fallback operations',
      };
    } else {
      return {
        level: 'low',
        description: 'Minor performance impact from compatibility mode',
      };
    }
  }

  /**
   * Get user-friendly error message for current state
   */
  getUserFriendlyStatus(): {
    title: string;
    message: string;
    severity: 'info' | 'warning' | 'error';
    actions: Array<{
      label: string;
      action: () => void;
      primary?: boolean;
    }>;
  } {
    const report = this.getDegradationReport();

    switch (report.overallHealth) {
      case 'healthy':
        return {
          title: 'All Systems Operational',
          message: 'All features are working normally.',
          severity: 'info',
          actions: [],
        };

      case 'degraded':
        return {
          title: 'Some Features Limited',
          message: `${report.activeFallbacks.length} feature(s) running in compatibility mode. Performance may be reduced.`,
          severity: 'warning',
          actions: [
            {
              label: 'View Details',
              action: () =>
                console.log('Feature details:', report.featureDetails),
            },
            {
              label: 'Try Restore',
              action: () => this.attemptFeatureRestore(),
              primary: true,
            },
          ],
        };

      case 'critical':
        return {
          title: 'Limited Functionality',
          message: `Multiple features unavailable. ${report.performanceImpact.description}`,
          severity: 'error',
          actions: [
            {
              label: 'Refresh Page',
              action: () => window.location.reload(),
              primary: true,
            },
            {
              label: 'View Help',
              action: () => this.showTroubleshootingHelp(),
            },
          ],
        };

      default:
        return {
          title: 'Unknown Status',
          message: 'Unable to determine system status.',
          severity: 'error',
          actions: [],
        };
    }
  }

  /**
   * Attempt to restore all features from fallback mode
   */
  private async attemptFeatureRestore(): Promise<void> {
    const features = Object.keys(
      this.capabilities
    ) as (keyof SystemCapabilities)[];
    const restoredFeatures: string[] = [];

    for (const feature of features) {
      if (this.capabilities[feature].fallbackActive) {
        const restored = await this.tryRestoreFeature(feature);
        if (restored) {
          restoredFeatures.push(feature);
        }
      }
    }

    if (restoredFeatures.length > 0) {
      console.log(`Restored features: ${restoredFeatures.join(', ')}`);
    } else {
      console.log('No features could be restored at this time');
    }
  }

  /**
   * Show troubleshooting help
   */
  private showTroubleshootingHelp(): void {
    const report = this.getDegradationReport();
    const helpContent = {
      title: 'Troubleshooting Guide',
      sections: [
        {
          title: 'Current Issues',
          items: report.recommendations,
        },
        {
          title: 'Quick Fixes',
          items: [
            'Refresh the page',
            'Clear browser cache and cookies',
            'Disable browser extensions temporarily',
            'Try a different browser',
            'Check your internet connection',
          ],
        },
        {
          title: 'Feature Details',
          items: Object.entries(report.featureDetails)
            .filter(([, details]) => !details.available)
            .map(([feature, details]) => `${feature}: ${details.impact}`),
        },
      ],
    };

    console.group('🔧 Troubleshooting Help');
    helpContent.sections.forEach(section => {
      console.group(section.title);
      section.items.forEach(item => console.log('•', item));
      console.groupEnd();
    });
    console.groupEnd();
  }

  // Feature detection methods

  private checkWebWorkers(): FeatureStatus {
    try {
      const available = typeof Worker !== 'undefined';
      return {
        available,
        fallbackActive: false,
        lastChecked: Date.now(),
        errorCount: 0,
        message: available
          ? 'Web Workers supported'
          : 'Web Workers not available',
      };
    } catch (error) {
      return {
        available: false,
        fallbackActive: true,
        lastChecked: Date.now(),
        errorCount: 1,
        message: 'Web Workers detection failed',
      };
    }
  }

  private checkWebAssembly(): FeatureStatus {
    try {
      const available =
        typeof WebAssembly !== 'undefined' &&
        typeof WebAssembly.instantiate === 'function';
      return {
        available,
        fallbackActive: false,
        lastChecked: Date.now(),
        errorCount: 0,
        message: available
          ? 'WebAssembly supported'
          : 'WebAssembly not available',
      };
    } catch (error) {
      return {
        available: false,
        fallbackActive: true,
        lastChecked: Date.now(),
        errorCount: 1,
        message: 'WebAssembly detection failed',
      };
    }
  }

  private checkIndexedDB(): FeatureStatus {
    try {
      const available = 'indexedDB' in window && indexedDB !== null;
      return {
        available,
        fallbackActive: false,
        lastChecked: Date.now(),
        errorCount: 0,
        message: available ? 'IndexedDB supported' : 'IndexedDB not available',
      };
    } catch (error) {
      return {
        available: false,
        fallbackActive: true,
        lastChecked: Date.now(),
        errorCount: 1,
        message: 'IndexedDB detection failed',
      };
    }
  }

  private checkLocalStorage(): FeatureStatus {
    try {
      const test = '__localStorage_test__';
      localStorage.setItem(test, test);
      localStorage.removeItem(test);
      return {
        available: true,
        fallbackActive: false,
        lastChecked: Date.now(),
        errorCount: 0,
        message: 'LocalStorage supported',
      };
    } catch (error) {
      return {
        available: false,
        fallbackActive: true,
        lastChecked: Date.now(),
        errorCount: 1,
        message: 'LocalStorage not available',
      };
    }
  }

  private checkServiceWorker(): FeatureStatus {
    try {
      const available = 'serviceWorker' in navigator;
      return {
        available,
        fallbackActive: false,
        lastChecked: Date.now(),
        errorCount: 0,
        message: available
          ? 'Service Worker supported'
          : 'Service Worker not available',
      };
    } catch (error) {
      return {
        available: false,
        fallbackActive: true,
        lastChecked: Date.now(),
        errorCount: 1,
        message: 'Service Worker detection failed',
      };
    }
  }

  private checkOfflineSupport(): FeatureStatus {
    try {
      const available = 'onLine' in navigator;
      return {
        available,
        fallbackActive: false,
        lastChecked: Date.now(),
        errorCount: 0,
        message: available
          ? 'Offline detection supported'
          : 'Offline detection not available',
      };
    } catch (error) {
      return {
        available: false,
        fallbackActive: true,
        lastChecked: Date.now(),
        errorCount: 1,
        message: 'Offline support detection failed',
      };
    }
  }

  private checkTreeSitter(): FeatureStatus {
    try {
      // This would check if Tree-sitter WASM is loaded
      const available = this.capabilities.webAssembly.available;
      return {
        available,
        fallbackActive: false,
        lastChecked: Date.now(),
        errorCount: 0,
        message: available
          ? 'Tree-sitter parsing available'
          : 'Tree-sitter not available, using fallback',
      };
    } catch (error) {
      return {
        available: false,
        fallbackActive: true,
        lastChecked: Date.now(),
        errorCount: 1,
        message: 'Tree-sitter detection failed',
      };
    }
  }

  private checkPerformanceAPI(): FeatureStatus {
    try {
      const available =
        'performance' in window && typeof performance.now === 'function';
      return {
        available,
        fallbackActive: false,
        lastChecked: Date.now(),
        errorCount: 0,
        message: available
          ? 'Performance API supported'
          : 'Performance API not available',
      };
    } catch (error) {
      return {
        available: false,
        fallbackActive: true,
        lastChecked: Date.now(),
        errorCount: 1,
        message: 'Performance API detection failed',
      };
    }
  }

  private checkNetworkConnection(): FeatureStatus {
    try {
      const available = navigator.onLine;
      return {
        available,
        fallbackActive: !available,
        lastChecked: Date.now(),
        errorCount: available ? 0 : 1,
        message: available
          ? 'Network connection available'
          : 'Offline mode active',
      };
    } catch (error) {
      return {
        available: false,
        fallbackActive: true,
        lastChecked: Date.now(),
        errorCount: 1,
        message: 'Network status detection failed',
      };
    }
  }

  // Fallback implementations

  private async basicTextParsing(code: string, language: string): Promise<any> {
    // Basic tokenization fallback
    const tokens: any[] = [];
    const lines = code.split('\n');

    for (let lineIndex = 0; lineIndex < lines.length; lineIndex++) {
      const line = lines[lineIndex];
      const words = line.split(/(\s+|[{}();,.])/);

      let columnIndex = 0;
      for (const word of words) {
        if (word.trim()) {
          tokens.push({
            type: this.guessTokenType(word, language),
            value: word,
            start: { line: lineIndex, column: columnIndex },
            end: { line: lineIndex, column: columnIndex + word.length },
          });
        }
        columnIndex += word.length;
      }
    }

    return { tokens };
  }

  private guessTokenType(token: string, language: string): string {
    // Basic token type guessing
    if (/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(token)) {
      const keywords = this.getLanguageKeywords(language);
      return keywords.includes(token) ? 'keyword' : 'identifier';
    }
    if (/^[0-9]+(\.[0-9]+)?$/.test(token)) {
      return 'number';
    }
    if (/^["'].*["']$/.test(token)) {
      return 'string';
    }
    if (/^[{}();,.]$/.test(token)) {
      return 'punctuation';
    }
    return 'unknown';
  }

  private getLanguageKeywords(language: string): string[] {
    const keywords: Record<string, string[]> = {
      javascript: [
        'function',
        'const',
        'let',
        'var',
        'if',
        'else',
        'for',
        'while',
        'return',
      ],
      python: [
        'def',
        'class',
        'if',
        'else',
        'elif',
        'for',
        'while',
        'return',
        'import',
      ],
      cpp: [
        'int',
        'char',
        'float',
        'double',
        'if',
        'else',
        'for',
        'while',
        'return',
        'class',
      ],
      rust: [
        'fn',
        'let',
        'mut',
        'if',
        'else',
        'for',
        'while',
        'return',
        'struct',
        'impl',
      ],
      yaml: ['true', 'false', 'null'],
    };

    return keywords[language] || [];
  }

  private memoryStorageMap = new Map<string, any>();

  private async memoryStorage<T>(key: string, data: T): Promise<void> {
    this.memoryStorageMap.set(key, data);
    console.warn(
      `Data stored in memory only (key: ${key}). Data will be lost on page refresh.`
    );
  }

  private startFeatureMonitoring(): void {
    if (this.featureCheckTimer) {
      clearInterval(this.featureCheckTimer);
    }

    this.featureCheckTimer = setInterval(() => {
      // Re-check critical features
      this.capabilities.networkConnection = this.checkNetworkConnection();

      // Check if any fallbacks can be disabled
      Object.keys(this.capabilities).forEach(feature => {
        const featureKey = feature as keyof SystemCapabilities;
        const status = this.capabilities[featureKey];

        if (
          status.fallbackActive &&
          Date.now() - status.lastChecked > this.config.featureCheckInterval
        ) {
          // Re-check the feature
          switch (featureKey) {
            case 'webWorkers':
              this.capabilities.webWorkers = this.checkWebWorkers();
              break;
            case 'indexedDB':
              this.capabilities.indexedDB = this.checkIndexedDB();
              break;
            case 'networkConnection':
              this.capabilities.networkConnection =
                this.checkNetworkConnection();
              break;
            // Add other features as needed
          }
        }
      });
    }, this.config.featureCheckInterval);
  }

  /**
   * Cleanup resources
   */
  destroy(): void {
    if (this.featureCheckTimer) {
      clearInterval(this.featureCheckTimer);
      this.featureCheckTimer = null;
    }
    this.memoryStorageMap.clear();
  }
}

// Singleton instance
let gracefulDegradationInstance: GracefulDegradationService | null = null;

export function getGracefulDegradationService(
  config?: Partial<FallbackConfig>
): GracefulDegradationService {
  if (!gracefulDegradationInstance) {
    gracefulDegradationInstance = new GracefulDegradationService(config);
  }
  return gracefulDegradationInstance;
}

export default GracefulDegradationService;
