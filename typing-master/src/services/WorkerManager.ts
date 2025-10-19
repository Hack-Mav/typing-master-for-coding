/**
 * Worker Manager
 * Manages Web Workers for parsing and metrics computation
 */

type WorkerType = 'parsing' | 'metrics';

interface PendingRequest {
  resolve: (value: any) => void;
  reject: (error: Error) => void;
  timeout: NodeJS.Timeout;
}

class WorkerManager {
  private workers: Map<WorkerType, Worker> = new Map();
  private pendingRequests: Map<string, PendingRequest> = new Map();
  private requestIdCounter = 0;
  private readonly REQUEST_TIMEOUT = 30000; // 30 seconds

  constructor() {
    this.initializeWorkers();
  }

  /**
   * Initialize all workers
   */
  private initializeWorkers(): void {
    try {
      // Initialize parsing worker
      const parsingWorker = new Worker(
        new URL('../workers/parsing.worker.ts', import.meta.url),
        { type: 'module' }
      );
      parsingWorker.onmessage = this.handleParsingMessage.bind(this);
      parsingWorker.onerror = this.handleWorkerError.bind(this, 'parsing');
      this.workers.set('parsing', parsingWorker);

      // Initialize metrics worker
      const metricsWorker = new Worker(
        new URL('../workers/metrics.worker.ts', import.meta.url),
        { type: 'module' }
      );
      metricsWorker.onmessage = this.handleMetricsMessage.bind(this);
      metricsWorker.onerror = this.handleWorkerError.bind(this, 'metrics');
      this.workers.set('metrics', metricsWorker);
    } catch (error) {
      console.error('Failed to initialize workers:', error);
    }
  }

  /**
   * Generate unique request ID
   */
  private generateRequestId(): string {
    return `req_${++this.requestIdCounter}_${Date.now()}`;
  }

  /**
   * Handle parsing worker messages
   */
  private handleParsingMessage(event: MessageEvent): void {
    const { type, id, result, error } = event.data;

    if (type === 'ready') {
      console.log('Parsing worker ready');
      return;
    }

    if (!id) return;

    const pending = this.pendingRequests.get(id);
    if (!pending) return;

    clearTimeout(pending.timeout);
    this.pendingRequests.delete(id);

    if (type === 'error') {
      pending.reject(new Error(error || 'Unknown parsing error'));
    } else {
      pending.resolve(result);
    }
  }

  /**
   * Handle metrics worker messages
   */
  private handleMetricsMessage(event: MessageEvent): void {
    const { type, id, result, error } = event.data;

    if (!id) return;

    const pending = this.pendingRequests.get(id);
    if (!pending) return;

    clearTimeout(pending.timeout);
    this.pendingRequests.delete(id);

    if (type === 'error') {
      pending.reject(new Error(error || 'Unknown metrics error'));
    } else {
      pending.resolve(result);
    }
  }

  /**
   * Handle worker errors
   */
  private handleWorkerError(workerType: WorkerType, error: ErrorEvent): void {
    console.error(`${workerType} worker error:`, error);
  }

  /**
   * Send message to worker with timeout
   */
  private sendMessage(workerType: WorkerType, message: any): Promise<any> {
    return new Promise((resolve, reject) => {
      const worker = this.workers.get(workerType);
      if (!worker) {
        reject(new Error(`${workerType} worker not initialized`));
        return;
      }

      const id = this.generateRequestId();
      const timeout = setTimeout(() => {
        this.pendingRequests.delete(id);
        reject(new Error(`${workerType} worker request timeout`));
      }, this.REQUEST_TIMEOUT);

      this.pendingRequests.set(id, { resolve, reject, timeout });

      worker.postMessage({ ...message, id });
    });
  }

  /**
   * Parse code using parsing worker
   */
  async parseCode(language: string, code: string): Promise<any> {
    const startTime = performance.now();
    try {
      const result = await this.sendMessage('parsing', {
        type: 'parse',
        language,
        code,
      });
      const duration = performance.now() - startTime;
      console.log(`Parsing completed in ${duration.toFixed(2)}ms`);
      return result;
    } catch (error) {
      console.error('Parse error:', error);
      throw error;
    }
  }

  /**
   * Tokenize code using parsing worker
   */
  async tokenizeCode(language: string, code: string): Promise<any[]> {
    const startTime = performance.now();
    try {
      const result = await this.sendMessage('parsing', {
        type: 'tokenize',
        language,
        code,
      });
      const duration = performance.now() - startTime;
      console.log(`Tokenization completed in ${duration.toFixed(2)}ms`);
      return result;
    } catch (error) {
      console.error('Tokenize error:', error);
      throw error;
    }
  }

  /**
   * Calculate metrics using metrics worker
   */
  async calculateMetrics(params: {
    events: any[];
    input: string;
    expected: string;
    durationMs: number;
    language?: string;
  }): Promise<any> {
    const startTime = performance.now();
    try {
      const result = await this.sendMessage('metrics', {
        type: 'calculate',
        ...params,
      });
      const duration = performance.now() - startTime;
      console.log(`Metrics calculation completed in ${duration.toFixed(2)}ms`);
      return result;
    } catch (error) {
      console.error('Metrics calculation error:', error);
      throw error;
    }
  }

  /**
   * Analyze error patterns using metrics worker
   */
  async analyzeErrors(events: any[]): Promise<any> {
    const startTime = performance.now();
    try {
      const result = await this.sendMessage('metrics', {
        type: 'analyze',
        events,
      });
      const duration = performance.now() - startTime;
      console.log(`Error analysis completed in ${duration.toFixed(2)}ms`);
      return result;
    } catch (error) {
      console.error('Error analysis error:', error);
      throw error;
    }
  }

  /**
   * Terminate all workers
   */
  terminate(): void {
    this.workers.forEach(worker => worker.terminate());
    this.workers.clear();
    this.pendingRequests.forEach(({ timeout }) => clearTimeout(timeout));
    this.pendingRequests.clear();
  }
}

// Singleton instance
let workerManagerInstance: WorkerManager | null = null;

export function getWorkerManager(): WorkerManager {
  if (!workerManagerInstance) {
    workerManagerInstance = new WorkerManager();
  }
  return workerManagerInstance;
}

export function terminateWorkers(): void {
  if (workerManagerInstance) {
    workerManagerInstance.terminate();
    workerManagerInstance = null;
  }
}

export default WorkerManager;
