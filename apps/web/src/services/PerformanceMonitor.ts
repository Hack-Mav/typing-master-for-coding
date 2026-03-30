/**
 * Performance Monitoring Service
 * Tracks and reports performance metrics using OpenTelemetry-compatible format
 */

interface PerformanceMetric {
  name: string;
  value: number;
  timestamp: number;
  tags?: Record<string, string>;
  unit?: string;
}

interface PerformanceTrace {
  name: string;
  startTime: number;
  endTime?: number;
  duration?: number;
  tags?: Record<string, string>;
  spans?: PerformanceSpan[];
}

interface PerformanceSpan {
  name: string;
  startTime: number;
  endTime: number;
  duration: number;
  tags?: Record<string, string>;
}

class PerformanceMonitor {
  private metrics: PerformanceMetric[] = [];
  private traces: Map<string, PerformanceTrace> = new Map();
  private readonly MAX_METRICS = 1000;
  private readonly MAX_TRACES = 100;
  private readonly FLUSH_INTERVAL = 60000; // 1 minute
  private flushTimer: NodeJS.Timeout | null = null;
  private enabled: boolean = true;

  constructor() {
    this.startAutoFlush();
    this.setupPerformanceObserver();
  }

  /**
   * Setup Performance Observer for Web Vitals
   */
  private setupPerformanceObserver(): void {
    if (typeof window === 'undefined' || !window.PerformanceObserver) {
      return;
    }

    try {
      // Observe Long Tasks
      const longTaskObserver = new PerformanceObserver(list => {
        for (const entry of list.getEntries()) {
          this.recordMetric({
            name: 'long_task',
            value: entry.duration,
            timestamp: entry.startTime,
            tags: {
              type: entry.entryType,
            },
            unit: 'ms',
          });
        }
      });
      longTaskObserver.observe({ entryTypes: ['longtask'] });

      // Observe Layout Shifts
      const layoutShiftObserver = new PerformanceObserver(list => {
        for (const entry of list.getEntries()) {
          const layoutShift = entry as any;
          this.recordMetric({
            name: 'layout_shift',
            value: layoutShift.value || 0,
            timestamp: entry.startTime,
            tags: {
              hadRecentInput: String(layoutShift.hadRecentInput || false),
            },
            unit: 'score',
          });
        }
      });
      layoutShiftObserver.observe({ entryTypes: ['layout-shift'] });

      // Observe Largest Contentful Paint
      const lcpObserver = new PerformanceObserver(list => {
        const entries = list.getEntries();
        const lastEntry = entries[entries.length - 1];
        this.recordMetric({
          name: 'largest_contentful_paint',
          value: lastEntry.startTime,
          timestamp: Date.now(),
          unit: 'ms',
        });
      });
      lcpObserver.observe({ entryTypes: ['largest-contentful-paint'] });

      // Observe First Input Delay
      const fidObserver = new PerformanceObserver(list => {
        for (const entry of list.getEntries()) {
          const firstInput = entry as any;
          this.recordMetric({
            name: 'first_input_delay',
            value: firstInput.processingStart - firstInput.startTime,
            timestamp: entry.startTime,
            unit: 'ms',
          });
        }
      });
      fidObserver.observe({ entryTypes: ['first-input'] });
    } catch (error) {
      console.warn('Performance Observer setup failed:', error);
    }
  }

  /**
   * Record a performance metric
   */
  recordMetric(metric: PerformanceMetric): void {
    if (!this.enabled) return;

    this.metrics.push({
      ...metric,
      timestamp: metric.timestamp || Date.now(),
    });

    // Limit metrics array size
    if (this.metrics.length > this.MAX_METRICS) {
      this.metrics = this.metrics.slice(-this.MAX_METRICS);
    }
  }

  /**
   * Start a performance trace
   */
  startTrace(name: string, tags?: Record<string, string>): string {
    if (!this.enabled) return name;

    const traceId = `${name}_${Date.now()}_${Math.random()}`;
    this.traces.set(traceId, {
      name,
      startTime: performance.now(),
      tags,
      spans: [],
    });

    return traceId;
  }

  /**
   * End a performance trace
   */
  endTrace(traceId: string, tags?: Record<string, string>): void {
    if (!this.enabled) return;

    const trace = this.traces.get(traceId);
    if (!trace) return;

    trace.endTime = performance.now();
    trace.duration = trace.endTime - trace.startTime;
    if (tags) {
      trace.tags = { ...trace.tags, ...tags };
    }

    // Record as metric
    this.recordMetric({
      name: `trace_${trace.name}`,
      value: trace.duration,
      timestamp: Date.now(),
      tags: trace.tags,
      unit: 'ms',
    });

    // Limit traces map size
    if (this.traces.size > this.MAX_TRACES) {
      const firstKey = this.traces.keys().next().value;
      this.traces.delete(firstKey);
    }
  }

  /**
   * Add a span to an existing trace
   */
  addSpan(
    traceId: string,
    spanName: string,
    duration: number,
    tags?: Record<string, string>
  ): void {
    if (!this.enabled) return;

    const trace = this.traces.get(traceId);
    if (!trace || !trace.spans) return;

    trace.spans.push({
      name: spanName,
      startTime: performance.now() - duration,
      endTime: performance.now(),
      duration,
      tags,
    });
  }

  /**
   * Measure keystroke latency
   */
  measureKeystrokeLatency(
    latency: number,
    tags?: Record<string, string>
  ): void {
    this.recordMetric({
      name: 'keystroke_latency',
      value: latency,
      timestamp: Date.now(),
      tags,
      unit: 'ms',
    });

    // Alert if latency exceeds target
    if (latency > 8) {
      console.warn(
        `Keystroke latency exceeded target: ${latency.toFixed(2)}ms`
      );
    }
  }

  /**
   * Measure parsing performance
   */
  measureParsingPerformance(
    language: string,
    codeLength: number,
    duration: number
  ): void {
    this.recordMetric({
      name: 'parsing_duration',
      value: duration,
      timestamp: Date.now(),
      tags: {
        language,
        codeLength: String(codeLength),
      },
      unit: 'ms',
    });
  }

  /**
   * Measure metrics calculation performance
   */
  measureMetricsCalculation(eventCount: number, duration: number): void {
    this.recordMetric({
      name: 'metrics_calculation_duration',
      value: duration,
      timestamp: Date.now(),
      tags: {
        eventCount: String(eventCount),
      },
      unit: 'ms',
    });
  }

  /**
   * Measure API request performance
   */
  measureApiRequest(
    endpoint: string,
    method: string,
    duration: number,
    statusCode: number
  ): void {
    this.recordMetric({
      name: 'api_request_duration',
      value: duration,
      timestamp: Date.now(),
      tags: {
        endpoint,
        method,
        statusCode: String(statusCode),
      },
      unit: 'ms',
    });
  }

  /**
   * Get performance summary
   */
  getSummary(): any {
    const summary: any = {
      totalMetrics: this.metrics.length,
      activeTraces: this.traces.size,
      metricsByName: {},
      averages: {},
    };

    // Group metrics by name
    this.metrics.forEach(metric => {
      if (!summary.metricsByName[metric.name]) {
        summary.metricsByName[metric.name] = [];
      }
      summary.metricsByName[metric.name].push(metric.value);
    });

    // Calculate averages
    Object.keys(summary.metricsByName).forEach(name => {
      const values = summary.metricsByName[name];
      const avg =
        values.reduce((a: number, b: number) => a + b, 0) / values.length;
      const min = Math.min(...values);
      const max = Math.max(...values);
      const p95 = this.calculatePercentile(values, 95);
      const p99 = this.calculatePercentile(values, 99);

      summary.averages[name] = {
        avg: Math.round(avg * 100) / 100,
        min: Math.round(min * 100) / 100,
        max: Math.round(max * 100) / 100,
        p95: Math.round(p95 * 100) / 100,
        p99: Math.round(p99 * 100) / 100,
        count: values.length,
      };
    });

    return summary;
  }

  /**
   * Calculate percentile
   */
  private calculatePercentile(values: number[], percentile: number): number {
    if (values.length === 0) return 0;

    const sorted = [...values].sort((a, b) => a - b);
    const index = Math.ceil((percentile / 100) * sorted.length) - 1;
    return sorted[Math.max(0, index)];
  }

  /**
   * Export metrics in OpenTelemetry format
   */
  exportMetrics(): any {
    return {
      resourceMetrics: [
        {
          resource: {
            attributes: [
              {
                key: 'service.name',
                value: { stringValue: 'typing-master-frontend' },
              },
              { key: 'service.version', value: { stringValue: '1.0.0' } },
            ],
          },
          scopeMetrics: [
            {
              scope: {
                name: 'typing-master-performance',
                version: '1.0.0',
              },
              metrics: this.metrics.map(metric => ({
                name: metric.name,
                unit: metric.unit || '',
                gauge: {
                  dataPoints: [
                    {
                      timeUnixNano: metric.timestamp * 1000000,
                      asDouble: metric.value,
                      attributes: Object.entries(metric.tags || {}).map(
                        ([key, value]) => ({
                          key,
                          value: { stringValue: value },
                        })
                      ),
                    },
                  ],
                },
              })),
            },
          ],
        },
      ],
    };
  }

  /**
   * Flush metrics to backend
   */
  async flush(): Promise<void> {
    if (this.metrics.length === 0) return;

    try {
      const data = this.exportMetrics();

      // Send to backend (if configured)
      const endpoint = process.env.REACT_APP_TELEMETRY_ENDPOINT;
      if (endpoint && typeof endpoint === 'string') {
        await fetch(endpoint, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(data),
        });
      }

      // Log summary to console in development
      if (process.env.NODE_ENV === 'development') {
        console.log('Performance Summary:', this.getSummary());
      }

      // Clear metrics after flush
      this.metrics = [];
    } catch (error) {
      console.error('Failed to flush metrics:', error);
    }
  }

  /**
   * Start auto-flush timer
   */
  private startAutoFlush(): void {
    this.flushTimer = setInterval(() => {
      this.flush();
    }, this.FLUSH_INTERVAL);
  }

  /**
   * Stop auto-flush timer
   */
  stopAutoFlush(): void {
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
      this.flushTimer = null;
    }
  }

  /**
   * Enable/disable monitoring
   */
  setEnabled(enabled: boolean): void {
    this.enabled = enabled;
  }

  /**
   * Clear all metrics and traces
   */
  clear(): void {
    this.metrics = [];
    this.traces.clear();
  }
}

// Singleton instance
let performanceMonitorInstance: PerformanceMonitor | null = null;

export function getPerformanceMonitor(): PerformanceMonitor {
  if (!performanceMonitorInstance) {
    performanceMonitorInstance = new PerformanceMonitor();
  }
  return performanceMonitorInstance;
}

export default PerformanceMonitor;
