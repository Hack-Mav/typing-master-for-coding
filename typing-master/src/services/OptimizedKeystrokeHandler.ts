/**
 * Optimized Keystroke Handler
 * Achieves <8ms latency through various optimization techniques
 */

import { getPerformanceMonitor } from './PerformanceMonitor';

interface KeystrokeEvent {
  timestamp: number;
  key: string;
  code: string;
  action: 'down' | 'up';
  cursorPosition: number;
  modifiers: {
    ctrl: boolean;
    shift: boolean;
    alt: boolean;
    meta: boolean;
  };
}

interface KeystrokeCallback {
  (event: KeystrokeEvent): void;
}

class OptimizedKeystrokeHandler {
  private callbacks: Set<KeystrokeCallback> = new Set();
  private eventBuffer: KeystrokeEvent[] = [];
  private rafId: number | null = null;
  private lastProcessTime: number = 0;
  private readonly BUFFER_SIZE = 100;
  private readonly PROCESS_INTERVAL = 16; // ~60fps
  private isProcessing = false;
  private performanceMonitor = getPerformanceMonitor();

  // Event pooling for memory efficiency
  private eventPool: KeystrokeEvent[] = [];
  private readonly POOL_SIZE = 50;

  constructor() {
    this.initializeEventPool();
  }

  /**
   * Initialize event object pool
   */
  private initializeEventPool(): void {
    for (let i = 0; i < this.POOL_SIZE; i++) {
      this.eventPool.push(this.createEmptyEvent());
    }
  }

  /**
   * Create empty event object
   */
  private createEmptyEvent(): KeystrokeEvent {
    return {
      timestamp: 0,
      key: '',
      code: '',
      action: 'down',
      cursorPosition: 0,
      modifiers: {
        ctrl: false,
        shift: false,
        alt: false,
        meta: false,
      },
    };
  }

  /**
   * Get event from pool or create new one
   */
  private acquireEvent(): KeystrokeEvent {
    return this.eventPool.pop() || this.createEmptyEvent();
  }

  /**
   * Return event to pool
   */
  private releaseEvent(event: KeystrokeEvent): void {
    if (this.eventPool.length < this.POOL_SIZE) {
      // Reset event
      event.timestamp = 0;
      event.key = '';
      event.code = '';
      event.action = 'down';
      event.cursorPosition = 0;
      event.modifiers.ctrl = false;
      event.modifiers.shift = false;
      event.modifiers.alt = false;
      event.modifiers.meta = false;

      this.eventPool.push(event);
    }
  }

  /**
   * Register a callback for keystroke events
   */
  subscribe(callback: KeystrokeCallback): () => void {
    this.callbacks.add(callback);
    return () => this.callbacks.delete(callback);
  }

  /**
   * Handle keyboard event with minimal latency
   */
  handleKeyEvent(nativeEvent: KeyboardEvent, cursorPosition: number): void {
    const startTime = performance.now();

    // Prevent default for typing keys only
    if (this.isTypingKey(nativeEvent)) {
      nativeEvent.preventDefault();
    }

    // Acquire event from pool
    const event = this.acquireEvent();
    event.timestamp = startTime;
    event.key = nativeEvent.key;
    event.code = nativeEvent.code;
    event.action = nativeEvent.type === 'keydown' ? 'down' : 'up';
    event.cursorPosition = cursorPosition;
    event.modifiers.ctrl = nativeEvent.ctrlKey;
    event.modifiers.shift = nativeEvent.shiftKey;
    event.modifiers.alt = nativeEvent.altKey;
    event.modifiers.meta = nativeEvent.metaKey;

    // Process immediately for critical events
    if (this.isCriticalEvent(event)) {
      this.processEventImmediate(event);
      const latency = performance.now() - startTime;
      this.performanceMonitor.measureKeystrokeLatency(latency, {
        key: event.key,
        critical: 'true',
      });
    } else {
      // Buffer non-critical events
      this.bufferEvent(event);
      this.scheduleProcessing();
    }
  }

  /**
   * Check if key is a typing key
   */
  private isTypingKey(event: KeyboardEvent): boolean {
    // Allow browser shortcuts
    if (event.ctrlKey || event.metaKey) {
      return false;
    }

    // Check if it's a printable character or backspace
    return (
      event.key.length === 1 ||
      event.key === 'Backspace' ||
      event.key === 'Enter' ||
      event.key === 'Tab'
    );
  }

  /**
   * Check if event is critical (requires immediate processing)
   */
  private isCriticalEvent(event: KeystrokeEvent): boolean {
    // Process character inputs and backspace immediately
    return (
      event.action === 'down' &&
      (event.key.length === 1 || event.key === 'Backspace')
    );
  }

  /**
   * Process event immediately
   */
  private processEventImmediate(event: KeystrokeEvent): void {
    // Notify all callbacks synchronously
    this.callbacks.forEach(callback => {
      try {
        callback(event);
      } catch (error) {
        console.error('Keystroke callback error:', error);
      }
    });

    // Release event back to pool
    this.releaseEvent(event);
  }

  /**
   * Buffer event for batch processing
   */
  private bufferEvent(event: KeystrokeEvent): void {
    this.eventBuffer.push(event);

    // Prevent buffer overflow
    if (this.eventBuffer.length > this.BUFFER_SIZE) {
      const removed = this.eventBuffer.shift();
      if (removed) {
        this.releaseEvent(removed);
      }
    }
  }

  /**
   * Schedule batch processing using RAF
   */
  private scheduleProcessing(): void {
    if (this.rafId !== null) return;

    this.rafId = requestAnimationFrame(() => {
      this.processBuffer();
      this.rafId = null;
    });
  }

  /**
   * Process buffered events
   */
  private processBuffer(): void {
    if (this.isProcessing || this.eventBuffer.length === 0) return;

    const now = performance.now();
    if (now - this.lastProcessTime < this.PROCESS_INTERVAL) {
      this.scheduleProcessing();
      return;
    }

    this.isProcessing = true;
    this.lastProcessTime = now;

    const startTime = performance.now();
    const events = [...this.eventBuffer];
    this.eventBuffer = [];

    // Process events
    events.forEach(event => {
      this.callbacks.forEach(callback => {
        try {
          callback(event);
        } catch (error) {
          console.error('Keystroke callback error:', error);
        }
      });

      // Release event back to pool
      this.releaseEvent(event);
    });

    const duration = performance.now() - startTime;
    this.performanceMonitor.recordMetric({
      name: 'batch_processing_duration',
      value: duration,
      timestamp: Date.now(),
      tags: {
        eventCount: String(events.length),
      },
      unit: 'ms',
    });

    this.isProcessing = false;
  }

  /**
   * Flush all buffered events immediately
   */
  flush(): void {
    if (this.rafId !== null) {
      cancelAnimationFrame(this.rafId);
      this.rafId = null;
    }
    this.processBuffer();
  }

  /**
   * Clear all events and callbacks
   */
  clear(): void {
    this.flush();
    this.callbacks.clear();
    this.eventBuffer.forEach(event => this.releaseEvent(event));
    this.eventBuffer = [];
  }

  /**
   * Get buffer statistics
   */
  getStats(): any {
    return {
      bufferedEvents: this.eventBuffer.length,
      poolSize: this.eventPool.length,
      activeCallbacks: this.callbacks.size,
      isProcessing: this.isProcessing,
    };
  }
}

// Singleton instance
let keystrokeHandlerInstance: OptimizedKeystrokeHandler | null = null;

export function getKeystrokeHandler(): OptimizedKeystrokeHandler {
  if (!keystrokeHandlerInstance) {
    keystrokeHandlerInstance = new OptimizedKeystrokeHandler();
  }
  return keystrokeHandlerInstance;
}

export default OptimizedKeystrokeHandler;
