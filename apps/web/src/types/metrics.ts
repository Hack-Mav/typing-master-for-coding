// Metrics calculation related types

export interface SpeedMetrics {
  cpm: number; // Characters per minute
  wpm: number; // Words per minute (traditional)
  twpm: number; // Token words per minute (code-aware)
  kps: number; // Keystrokes per second
  netWpm: number; // Net WPM (accounting for errors)
  burstSpeed: number; // Peak typing speed in bursts
}

export interface AccuracyMetrics {
  rawAccuracy: number; // Character-level accuracy (0-1)
  tokenAccuracy: number; // Token-level accuracy (0-1)
  syntaxAccuracy: number; // Syntax structure accuracy (0-1)
  whitespaceAccuracy: number; // Whitespace conformance (0-1)
  overallAccuracy: number; // Composite accuracy score (0-1)
}

export interface ErrorMetrics {
  totalErrors: number;
  errorRate: number; // Errors per minute
  backspaceRate: number; // Backspaces per minute
  correctionLatency: number; // Average time to correct errors (ms)
  errorTypes: Record<string, number>; // Count by error type
  errorClusters: ErrorCluster[]; // Grouped error patterns
}

export interface ErrorCluster {
  startPosition: number;
  endPosition: number;
  errorCount: number;
  dominantErrorType: string;
  affectedTokens: string[];
  severity: 'low' | 'medium' | 'high';
}

export interface TimingMetrics {
  totalDuration: number; // Total session time (ms)
  activeDuration: number; // Time actually typing (ms)
  idleTime: number; // Time spent idle (ms)
  averageKeystrokeInterval: number; // Average time between keystrokes (ms)
  keystrokeVariability: number; // Standard deviation of intervals
  pauseCount: number; // Number of pauses > 2 seconds
}

export interface CompositeScore {
  score: number; // Final composite score (0-1000+)
  breakdown: {
    speedComponent: number;
    accuracyComponent: number;
    syntaxComponent: number;
    consistencyComponent: number;
  };
  penalties: {
    errorPenalty: number;
    backspacePenalty: number;
    idlePenalty: number;
    syntaxPenalty: number;
  };
  weights: ScoreWeights;
}

export interface ScoreWeights {
  alpha: number; // tWPM weight
  beta: number; // Raw accuracy weight
  gamma: number; // Syntax accuracy weight
  delta: number; // Error penalty weight
  epsilon: number; // Backspace penalty weight
  zeta: number; // Idle time penalty weight
}

export interface SessionMetrics {
  speed: SpeedMetrics;
  accuracy: AccuracyMetrics;
  errors: ErrorMetrics;
  timing: TimingMetrics;
  composite: CompositeScore;
  metadata: {
    language: string;
    mode: string;
    snippetLength: number;
    tokenCount: number;
    difficulty: number;
    timestamp: number;
  };
}

export interface PerformanceInsights {
  strengths: string[];
  weaknesses: string[];
  recommendations: string[];
  progressIndicators: {
    speedTrend: 'improving' | 'stable' | 'declining';
    accuracyTrend: 'improving' | 'stable' | 'declining';
    consistencyTrend: 'improving' | 'stable' | 'declining';
  };
  targetMetrics: {
    nextSpeedGoal: number;
    nextAccuracyGoal: number;
    estimatedTimeToGoal: number; // minutes of practice
  };
}

export interface BurstAnalysis {
  bursts: Array<{
    startTime: number;
    endTime: number;
    speed: number;
    accuracy: number;
    characterCount: number;
  }>;
  averageBurstSpeed: number;
  peakBurstSpeed: number;
  burstConsistency: number; // 0-1, how consistent burst speeds are
}

export interface TokenTypePerformance {
  [tokenType: string]: {
    speed: number; // Average typing speed for this token type
    accuracy: number; // Accuracy for this token type
    errorCount: number;
    totalOccurrences: number;
  };
}
