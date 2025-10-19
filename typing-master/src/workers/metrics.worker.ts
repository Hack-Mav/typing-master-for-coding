/**
 * Web Worker for metrics computation
 * Offloads heavy metrics calculations to prevent UI blocking
 */

// Make this file a module for TypeScript's --isolatedModules
export {};

interface KeystrokeEvent {
  timestamp: number;
  key: string;
  action: 'down' | 'up';
  cursorPosition: number;
  errorFlag: boolean;
  expectedToken?: string;
}

interface MetricsMessage {
  type: 'calculate' | 'analyze';
  id?: string;
  events?: KeystrokeEvent[];
  input?: string;
  expected?: string;
  durationMs?: number;
  language?: string;
}

interface MetricsResponse {
  type: 'metrics' | 'analysis' | 'error';
  id?: string;
  result?: any;
  error?: string;
}

// Calculate speed metrics
function calculateSpeedMetrics(
  events: KeystrokeEvent[],
  durationMs: number
): any {
  if (events.length === 0 || durationMs === 0) {
    return {
      cpm: 0,
      wpm: 0,
      kps: 0,
      avgKeystrokeInterval: 0,
    };
  }

  const durationMinutes = durationMs / 60000;
  const durationSeconds = durationMs / 1000;

  // Count actual character inputs (excluding backspace)
  const characterEvents = events.filter(
    e => e.action === 'down' && e.key !== 'Backspace'
  );
  const characterCount = characterEvents.length;

  // CPM (Characters Per Minute)
  const cpm = characterCount / durationMinutes;

  // WPM (Words Per Minute) - assuming average word length of 5
  const wpm = cpm / 5;

  // KPS (Keystrokes Per Second)
  const kps = events.length / durationSeconds;

  // Average keystroke interval
  const intervals: number[] = [];
  for (let i = 1; i < events.length; i++) {
    intervals.push(events[i].timestamp - events[i - 1].timestamp);
  }
  const avgKeystrokeInterval =
    intervals.length > 0
      ? intervals.reduce((a, b) => a + b, 0) / intervals.length
      : 0;

  return {
    cpm: Math.round(cpm * 100) / 100,
    wpm: Math.round(wpm * 100) / 100,
    kps: Math.round(kps * 100) / 100,
    avgKeystrokeInterval: Math.round(avgKeystrokeInterval * 100) / 100,
  };
}

// Calculate accuracy metrics
function calculateAccuracyMetrics(input: string, expected: string): any {
  if (expected.length === 0) {
    return {
      rawAccuracy: 100,
      characterAccuracy: 100,
      errorCount: 0,
      errorRate: 0,
    };
  }

  let correctChars = 0;
  const minLength = Math.min(input.length, expected.length);

  for (let i = 0; i < minLength; i++) {
    if (input[i] === expected[i]) {
      correctChars++;
    }
  }

  const rawAccuracy = (correctChars / expected.length) * 100;
  const errorCount = expected.length - correctChars;
  const errorRate = (errorCount / expected.length) * 100;

  return {
    rawAccuracy: Math.round(rawAccuracy * 100) / 100,
    characterAccuracy: Math.round((correctChars / minLength) * 100 * 100) / 100,
    errorCount,
    errorRate: Math.round(errorRate * 100) / 100,
    completeness:
      Math.round((input.length / expected.length) * 100 * 100) / 100,
  };
}

// Analyze error patterns
function analyzeErrorPatterns(events: KeystrokeEvent[]): any {
  const backspaceEvents = events.filter(e => e.key === 'Backspace');
  const errorEvents = events.filter(e => e.errorFlag);

  // Calculate backspace rate
  const backspaceRate =
    events.length > 0 ? (backspaceEvents.length / events.length) * 100 : 0;

  // Error clustering - find consecutive errors
  const errorClusters: number[][] = [];
  let currentCluster: number[] = [];

  events.forEach((event, index) => {
    if (event.errorFlag) {
      currentCluster.push(index);
    } else if (currentCluster.length > 0) {
      errorClusters.push([...currentCluster]);
      currentCluster = [];
    }
  });

  if (currentCluster.length > 0) {
    errorClusters.push(currentCluster);
  }

  // Error hotspots - positions with most errors
  const errorPositions: Map<number, number> = new Map();
  errorEvents.forEach(event => {
    const pos = event.cursorPosition;
    errorPositions.set(pos, (errorPositions.get(pos) || 0) + 1);
  });

  const hotspots = Array.from(errorPositions.entries())
    .sort((a, b) => b[1] - a[1])
    .slice(0, 10)
    .map(([position, count]) => ({ position, count }));

  // Correction latency - time between error and backspace
  const correctionLatencies: number[] = [];
  for (let i = 0; i < events.length - 1; i++) {
    if (events[i].errorFlag && events[i + 1].key === 'Backspace') {
      correctionLatencies.push(events[i + 1].timestamp - events[i].timestamp);
    }
  }

  const avgCorrectionLatency =
    correctionLatencies.length > 0
      ? correctionLatencies.reduce((a, b) => a + b, 0) /
        correctionLatencies.length
      : 0;

  return {
    backspaceRate: Math.round(backspaceRate * 100) / 100,
    backspaceCount: backspaceEvents.length,
    errorCount: errorEvents.length,
    errorClusters: errorClusters.length,
    avgClusterSize:
      errorClusters.length > 0
        ? Math.round(
            (errorClusters.reduce((sum, cluster) => sum + cluster.length, 0) /
              errorClusters.length) *
              100
          ) / 100
        : 0,
    hotspots,
    avgCorrectionLatency: Math.round(avgCorrectionLatency * 100) / 100,
  };
}

// Calculate burst consistency
function calculateBurstConsistency(events: KeystrokeEvent[]): any {
  if (events.length < 10) {
    return {
      burstConsistency: 100,
      variance: 0,
      standardDeviation: 0,
    };
  }

  // Calculate intervals between keystrokes
  const intervals: number[] = [];
  for (let i = 1; i < events.length; i++) {
    intervals.push(events[i].timestamp - events[i - 1].timestamp);
  }

  // Calculate mean
  const mean = intervals.reduce((a, b) => a + b, 0) / intervals.length;

  // Calculate variance
  const variance =
    intervals.reduce((sum, interval) => sum + Math.pow(interval - mean, 2), 0) /
    intervals.length;

  // Calculate standard deviation
  const stdDev = Math.sqrt(variance);

  // Consistency score (lower variance = higher consistency)
  const burstConsistency = Math.max(0, 100 - (stdDev / mean) * 100);

  return {
    burstConsistency: Math.round(burstConsistency * 100) / 100,
    variance: Math.round(variance * 100) / 100,
    standardDeviation: Math.round(stdDev * 100) / 100,
    meanInterval: Math.round(mean * 100) / 100,
  };
}

// Calculate composite score
function calculateCompositeScore(metrics: any): number {
  // Configurable weights
  const weights = {
    alpha: 0.4, // tWPM weight
    beta: 0.25, // Raw accuracy weight
    gamma: 0.2, // Syntax accuracy weight (use raw accuracy if not available)
    delta: 0.1, // Error weight
    epsilon: 0.03, // Backspace rate weight
    zeta: 0.02, // Idle time penalty weight
  };

  const wpm = metrics.speed?.wpm || 0;
  const rawAccuracy = metrics.accuracy?.rawAccuracy || 0;
  const syntaxAccuracy = metrics.accuracy?.rawAccuracy || 0; // Fallback to raw
  const errorRate = metrics.errors?.errorCount || 0;
  const backspaceRate = metrics.errors?.backspaceRate || 0;

  const score =
    weights.alpha * wpm +
    weights.beta * rawAccuracy +
    weights.gamma * syntaxAccuracy -
    weights.delta * errorRate -
    weights.epsilon * backspaceRate;

  return Math.max(0, Math.round(score * 100) / 100);
}

// Main message handler
globalThis.onmessage = async (event: MessageEvent<MetricsMessage>) => {
  const { type, id, events, input, expected, durationMs } = event.data;

  try {
    switch (type) {
      case 'calculate': {
        if (!events || !input || !expected || durationMs === undefined) {
          throw new Error(
            'Missing required parameters for metrics calculation'
          );
        }

        const speedMetrics = calculateSpeedMetrics(events, durationMs);
        const accuracyMetrics = calculateAccuracyMetrics(input, expected);
        const errorAnalysis = analyzeErrorPatterns(events);
        const burstMetrics = calculateBurstConsistency(events);

        const allMetrics = {
          speed: speedMetrics,
          accuracy: accuracyMetrics,
          errors: errorAnalysis,
          burst: burstMetrics,
        };

        const compositeScore = calculateCompositeScore(allMetrics);

        postMessage({
          type: 'metrics',
          id,
          result: {
            ...allMetrics,
            compositeScore,
          },
        } as MetricsResponse);
        break;
      }

      case 'analyze': {
        if (!events) {
          throw new Error('Events required for analysis');
        }

        const errorAnalysis = analyzeErrorPatterns(events);
        const burstMetrics = calculateBurstConsistency(events);

        postMessage({
          type: 'analysis',
          id,
          result: {
            errors: errorAnalysis,
            burst: burstMetrics,
          },
        } as MetricsResponse);
        break;
      }

      default:
        throw new Error(`Unknown message type: ${type}`);
    }
  } catch (error) {
    postMessage({
      type: 'error',
      id,
      error: error instanceof Error ? error.message : String(error),
    } as MetricsResponse);
  }
};
