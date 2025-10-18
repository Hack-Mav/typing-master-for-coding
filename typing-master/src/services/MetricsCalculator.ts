import { tokenizationService } from './TokenizationService';
import { KeystrokeEvent } from '../types/typing';
import {
  SpeedMetrics,
  AccuracyMetrics,
  ErrorMetrics,
  TimingMetrics,
  CompositeScore,
  ScoreWeights,
  SessionMetrics,
  PerformanceInsights,
  BurstAnalysis,
  ErrorCluster,
} from '../types/metrics';

/**
 * MetricsCalculator provides comprehensive typing performance analysis
 * including CPM, Token WPM, KPS, accuracy metrics, and composite scoring.
 */
export class MetricsCalculator {
  private static instance: MetricsCalculator;

  // Default scoring weights (configurable)
  private defaultWeights: ScoreWeights = {
    alpha: 0.4, // tWPM weight
    beta: 0.25, // Raw accuracy weight
    gamma: 0.2, // Syntax accuracy weight
    delta: 0.05, // Error penalty weight
    epsilon: 0.05, // Backspace penalty weight
    zeta: 0.05, // Idle time penalty weight
  };

  // Constants for calculations
  private readonly AVERAGE_WORD_LENGTH = 5; // Standard WPM calculation
  private readonly BURST_THRESHOLD_MS = 2000; // 2 seconds defines a burst boundary
  private readonly IDLE_THRESHOLD_MS = 2000; // 2 seconds defines idle time
  private readonly MIN_BURST_CHARS = 5; // Minimum characters for a valid burst

  private constructor() {}

  public static getInstance(): MetricsCalculator {
    if (!MetricsCalculator.instance) {
      MetricsCalculator.instance = new MetricsCalculator();
    }
    return MetricsCalculator.instance;
  }

  /**
   * Calculate real-time metrics for live feedback during typing
   */
  public calculateRealtimeMetrics(
    keystrokeEvents: KeystrokeEvent[],
    expectedText: string,
    actualText: string
  ): { speed: SpeedMetrics; accuracy: AccuracyMetrics } {
    if (keystrokeEvents.length === 0) {
      return {
        speed: { cpm: 0, wpm: 0, twpm: 0, kps: 0, netWpm: 0, burstSpeed: 0 },
        accuracy: {
          rawAccuracy: 1,
          tokenAccuracy: 1,
          syntaxAccuracy: 1,
          whitespaceAccuracy: 1,
          overallAccuracy: 1,
        },
      };
    }

    // Calculate basic speed metrics without tokenization for performance
    const startTime = keystrokeEvents[0].timestamp;
    const endTime = keystrokeEvents[keystrokeEvents.length - 1].timestamp;
    const durationMinutes = (endTime - startTime) / (1000 * 60);
    const durationSeconds = (endTime - startTime) / 1000;

    const cpm = durationMinutes > 0 ? actualText.length / durationMinutes : 0;
    const wpm = cpm / this.AVERAGE_WORD_LENGTH;
    const totalKeystrokes = keystrokeEvents.filter(
      e => e.action === 'keydown'
    ).length;
    const kps = durationSeconds > 0 ? totalKeystrokes / durationSeconds : 0;

    const speed: SpeedMetrics = {
      cpm: Math.round(cpm * 100) / 100,
      wpm: Math.round(wpm * 100) / 100,
      twpm: Math.round(wpm * 100) / 100, // Simplified for real-time
      kps: Math.round(kps * 100) / 100,
      netWpm: Math.round(wpm * 100) / 100,
      burstSpeed: 0,
    };

    // Calculate basic accuracy
    const rawAccuracy = this.calculateRawAccuracy(expectedText, actualText);
    const accuracy: AccuracyMetrics = {
      rawAccuracy,
      tokenAccuracy: rawAccuracy, // Simplified for real-time
      syntaxAccuracy: rawAccuracy,
      whitespaceAccuracy: rawAccuracy,
      overallAccuracy: rawAccuracy,
    };

    return { speed, accuracy };
  }

  /**
   * Calculate comprehensive session metrics from keystroke events
   */
  public calculateSessionMetrics(
    keystrokeEvents: KeystrokeEvent[],
    expectedText: string,
    actualText: string
  ): SessionMetrics {
    const weights = { ...this.defaultWeights };

    // For now, use simplified tokenization without language-specific parsing
    const expectedTokens: any[] = [];
    const actualTokens: any[] = [];

    // Calculate individual metric components
    const speed = this.calculateSpeedMetrics(
      keystrokeEvents,
      expectedText,
      actualText,
      expectedTokens
    );
    const accuracy = this.calculateAccuracyMetrics(
      expectedText,
      actualText,
      expectedTokens,
      actualTokens
    );
    const errors = this.calculateErrorMetrics(
      keystrokeEvents,
      expectedText,
      actualText
    );
    const timing = this.calculateTimingMetrics(keystrokeEvents);
    const composite = this.calculateCompositeScore(
      speed,
      accuracy,
      errors,
      timing,
      weights
    );

    return {
      speed,
      accuracy,
      errors,
      timing,
      composite,
      metadata: {
        language: 'javascript', // Default for now
        mode: 'practice',
        snippetLength: expectedText.length,
        tokenCount: expectedTokens.length,
        difficulty: 1,
        timestamp: Date.now(),
      },
    };
  }

  /**
   * Calculate speed metrics (CPM, WPM, tWPM, KPS)
   */
  public calculateSpeedMetrics(
    keystrokeEvents: KeystrokeEvent[],
    expectedText: string,
    actualText: string,
    expectedTokens: any[]
  ): SpeedMetrics {
    if (keystrokeEvents.length === 0) {
      return { cpm: 0, wpm: 0, twpm: 0, kps: 0, netWpm: 0, burstSpeed: 0 };
    }

    const startTime = keystrokeEvents[0].timestamp;
    const endTime = keystrokeEvents[keystrokeEvents.length - 1].timestamp;
    const durationMinutes = (endTime - startTime) / (1000 * 60);
    const durationSeconds = (endTime - startTime) / 1000;

    // Characters per minute (CPM)
    const cpm = durationMinutes > 0 ? actualText.length / durationMinutes : 0;

    // Traditional words per minute (WPM) - 5 characters = 1 word
    const wpm = cpm / this.AVERAGE_WORD_LENGTH;

    // Token words per minute (tWPM) - code-aware metric
    const completedTokens = this.countCompletedTokens(
      actualText,
      expectedTokens
    );
    const twpm = durationMinutes > 0 ? completedTokens / durationMinutes : 0;

    // Keystrokes per second (KPS)
    const totalKeystrokes = keystrokeEvents.filter(
      e => e.action === 'keydown'
    ).length;
    const kps = durationSeconds > 0 ? totalKeystrokes / durationSeconds : 0;

    // Net WPM (accounting for errors)
    const errorCount = this.countErrors(expectedText, actualText);
    const netWpm = Math.max(0, wpm - errorCount / durationMinutes);

    // Burst speed analysis
    const burstAnalysis = this.analyzeBursts(keystrokeEvents);
    const burstSpeed = burstAnalysis.peakBurstSpeed;

    return {
      cpm: Math.round(cpm * 100) / 100,
      wpm: Math.round(wpm * 100) / 100,
      twpm: Math.round(twpm * 100) / 100,
      kps: Math.round(kps * 100) / 100,
      netWpm: Math.round(netWpm * 100) / 100,
      burstSpeed: Math.round(burstSpeed * 100) / 100,
    };
  }

  /**
   * Calculate accuracy metrics (raw, token, syntax, whitespace)
   */
  public calculateAccuracyMetrics(
    expectedText: string,
    actualText: string,
    expectedTokens: any[],
    actualTokens: any[]
  ): AccuracyMetrics {
    // Raw character-level accuracy
    const rawAccuracy = this.calculateRawAccuracy(expectedText, actualText);

    // Token-level accuracy
    const tokenComparison = tokenizationService.compareTokenSequences(
      expectedTokens,
      actualTokens
    );
    const tokenAccuracy =
      tokenComparison?.tokenAccuracy || (expectedTokens.length === 0 ? 1 : 0);

    // Syntax accuracy (structural correctness)
    const syntaxAccuracy = this.calculateSyntaxAccuracy(
      expectedTokens,
      actualTokens
    );

    // Whitespace accuracy
    const whitespaceAccuracy = this.calculateWhitespaceAccuracy(
      expectedText,
      actualText
    );

    // Overall composite accuracy
    const overallAccuracy =
      rawAccuracy * 0.4 +
      tokenAccuracy * 0.3 +
      syntaxAccuracy * 0.2 +
      whitespaceAccuracy * 0.1;

    return {
      rawAccuracy: Math.round(rawAccuracy * 10000) / 10000,
      tokenAccuracy: Math.round(tokenAccuracy * 10000) / 10000,
      syntaxAccuracy: Math.round(syntaxAccuracy * 10000) / 10000,
      whitespaceAccuracy: Math.round(whitespaceAccuracy * 10000) / 10000,
      overallAccuracy: Math.round(overallAccuracy * 10000) / 10000,
    };
  }

  /**
   * Calculate error metrics and analysis
   */
  public calculateErrorMetrics(
    keystrokeEvents: KeystrokeEvent[],
    expectedText: string,
    actualText: string
  ): ErrorMetrics {
    const totalErrors = this.countErrors(expectedText, actualText);
    const backspaceCount = keystrokeEvents.filter(
      e => e.key === 'Backspace'
    ).length;

    const durationMinutes =
      keystrokeEvents.length > 0
        ? (keystrokeEvents[keystrokeEvents.length - 1].timestamp -
            keystrokeEvents[0].timestamp) /
          (1000 * 60)
        : 0;

    const errorRate = durationMinutes > 0 ? totalErrors / durationMinutes : 0;
    const backspaceRate =
      durationMinutes > 0 ? backspaceCount / durationMinutes : 0;

    // Calculate correction latency
    const correctionLatency = this.calculateCorrectionLatency(keystrokeEvents);

    // Analyze error types
    const errorTypes = this.analyzeErrorTypes(expectedText, actualText);

    // Find error clusters
    const errorClusters = this.findErrorClusters(expectedText, actualText);

    return {
      totalErrors,
      errorRate: Math.round(errorRate * 100) / 100,
      backspaceRate: Math.round(backspaceRate * 100) / 100,
      correctionLatency: Math.round(correctionLatency),
      errorTypes,
      errorClusters,
    };
  }

  /**
   * Calculate timing metrics
   */
  public calculateTimingMetrics(
    keystrokeEvents: KeystrokeEvent[]
  ): TimingMetrics {
    if (keystrokeEvents.length < 2) {
      return {
        totalDuration: 0,
        activeDuration: 0,
        idleTime: 0,
        averageKeystrokeInterval: 0,
        keystrokeVariability: 0,
        pauseCount: 0,
      };
    }

    const startTime = keystrokeEvents[0].timestamp;
    const endTime = keystrokeEvents[keystrokeEvents.length - 1].timestamp;
    const totalDuration = endTime - startTime;

    // Calculate intervals between keystrokes
    const intervals: number[] = [];
    let idleTime = 0;
    let pauseCount = 0;

    for (let i = 1; i < keystrokeEvents.length; i++) {
      const interval =
        keystrokeEvents[i].timestamp - keystrokeEvents[i - 1].timestamp;
      intervals.push(interval);

      if (interval > this.IDLE_THRESHOLD_MS) {
        idleTime += interval;
        pauseCount++;
      }
    }

    const activeDuration = totalDuration - idleTime;
    const averageKeystrokeInterval =
      intervals.length > 0
        ? intervals.reduce((a, b) => a + b, 0) / intervals.length
        : 0;

    // Calculate keystroke variability (standard deviation)
    const keystrokeVariability =
      intervals.length > 0 ? this.calculateStandardDeviation(intervals) : 0;

    return {
      totalDuration,
      activeDuration,
      idleTime,
      averageKeystrokeInterval: Math.round(averageKeystrokeInterval),
      keystrokeVariability: Math.round(keystrokeVariability),
      pauseCount,
    };
  }

  /**
   * Calculate composite score using configurable weights
   */
  public calculateCompositeScore(
    speed: SpeedMetrics,
    accuracy: AccuracyMetrics,
    errors: ErrorMetrics,
    timing: TimingMetrics,
    weights: ScoreWeights
  ): CompositeScore {
    // Base components (0-1000 scale)
    const speedComponent = speed.twpm * 10; // tWPM * 10 for scale
    const accuracyComponent = accuracy.rawAccuracy * 1000;
    const syntaxComponent = accuracy.syntaxAccuracy * 1000;
    const consistencyComponent = Math.max(
      0,
      1000 - timing.keystrokeVariability
    );

    // Penalties
    const errorPenalty = errors.errorRate * 10;
    const backspacePenalty = errors.backspaceRate * 5;
    const idlePenalty = (timing.idleTime / timing.totalDuration) * 100;
    const syntaxPenalty = (1 - accuracy.syntaxAccuracy) * 200;

    // Apply weights and calculate final score
    const score = Math.max(
      0,
      weights.alpha * speedComponent +
        weights.beta * accuracyComponent +
        weights.gamma * syntaxComponent +
        0.1 * consistencyComponent -
        weights.delta * errorPenalty -
        weights.epsilon * backspacePenalty -
        weights.zeta * idlePenalty -
        0.1 * syntaxPenalty
    );

    return {
      score: Math.round(score),
      breakdown: {
        speedComponent: Math.round(speedComponent),
        accuracyComponent: Math.round(accuracyComponent),
        syntaxComponent: Math.round(syntaxComponent),
        consistencyComponent: Math.round(consistencyComponent),
      },
      penalties: {
        errorPenalty: Math.round(errorPenalty),
        backspacePenalty: Math.round(backspacePenalty),
        idlePenalty: Math.round(idlePenalty),
        syntaxPenalty: Math.round(syntaxPenalty),
      },
      weights,
    };
  }

  /**
   * Generate performance insights and recommendations
   */
  public generateInsights(
    currentMetrics: SessionMetrics,
    historicalMetrics: SessionMetrics[] = []
  ): PerformanceInsights {
    const strengths: string[] = [];
    const weaknesses: string[] = [];
    const recommendations: string[] = [];

    // Analyze current performance
    if (currentMetrics.accuracy.overallAccuracy > 0.95) {
      strengths.push('Excellent accuracy');
    } else if (currentMetrics.accuracy.overallAccuracy < 0.85) {
      weaknesses.push('Accuracy needs improvement');
      recommendations.push('Focus on accuracy over speed');
    }

    if (currentMetrics.speed.twpm > 40) {
      strengths.push('Good typing speed');
    } else if (currentMetrics.speed.twpm < 20) {
      weaknesses.push('Typing speed below average');
      recommendations.push('Practice common code patterns');
    }

    if (currentMetrics.errors.errorRate < 2) {
      strengths.push('Low error rate');
    } else {
      weaknesses.push('High error rate');
      recommendations.push('Slow down and focus on precision');
    }

    // Analyze trends if historical data available
    const trends = this.analyzeTrends(currentMetrics, historicalMetrics);

    // Set target goals
    const targetMetrics = {
      nextSpeedGoal: Math.ceil(currentMetrics.speed.twpm * 1.1),
      nextAccuracyGoal: Math.min(
        1.0,
        currentMetrics.accuracy.overallAccuracy + 0.05
      ),
      estimatedTimeToGoal: this.estimateTimeToGoal(
        currentMetrics,
        historicalMetrics
      ),
    };

    return {
      strengths,
      weaknesses,
      recommendations,
      progressIndicators: trends,
      targetMetrics,
    };
  }

  /**
   * Analyze typing bursts for consistency measurement
   */
  public analyzeBursts(keystrokeEvents: KeystrokeEvent[]): BurstAnalysis {
    const bursts: BurstAnalysis['bursts'] = [];
    let currentBurst: { startTime: number; events: KeystrokeEvent[] } | null =
      null;

    for (const event of keystrokeEvents) {
      if (event.action !== 'keydown') continue;

      if (!currentBurst) {
        currentBurst = { startTime: event.timestamp, events: [event] };
      } else {
        const timeSinceLastEvent =
          event.timestamp -
          currentBurst.events[currentBurst.events.length - 1].timestamp;

        if (timeSinceLastEvent > this.BURST_THRESHOLD_MS) {
          // End current burst and start new one
          if (currentBurst.events.length >= this.MIN_BURST_CHARS) {
            bursts.push(this.processBurst(currentBurst));
          }
          currentBurst = { startTime: event.timestamp, events: [event] };
        } else {
          currentBurst.events.push(event);
        }
      }
    }

    // Process final burst
    if (currentBurst && currentBurst.events.length >= this.MIN_BURST_CHARS) {
      bursts.push(this.processBurst(currentBurst));
    }

    const speeds = bursts.map(b => b.speed);
    const averageBurstSpeed =
      speeds.length > 0 ? speeds.reduce((a, b) => a + b, 0) / speeds.length : 0;
    const peakBurstSpeed = speeds.length > 0 ? Math.max(...speeds) : 0;
    const burstConsistency =
      speeds.length > 1
        ? 1 - this.calculateStandardDeviation(speeds) / averageBurstSpeed
        : 1;

    return {
      bursts,
      averageBurstSpeed: Math.round(averageBurstSpeed * 100) / 100,
      peakBurstSpeed: Math.round(peakBurstSpeed * 100) / 100,
      burstConsistency: Math.max(0, Math.min(1, burstConsistency)),
    };
  }

  // Helper methods

  private countCompletedTokens(
    actualText: string,
    expectedTokens: any[]
  ): number {
    let completedTokens = 0;
    let currentPosition = 0;

    for (const token of expectedTokens) {
      const tokenEnd = token.startIndex + token.value.length;
      if (currentPosition >= tokenEnd && actualText.length >= tokenEnd) {
        const actualTokenText = actualText.substring(
          token.startIndex,
          tokenEnd
        );
        if (actualTokenText === token.value) {
          completedTokens++;
        }
      }
      currentPosition = tokenEnd;
    }

    return completedTokens;
  }

  private countErrors(expectedText: string, actualText: string): number {
    let errors = 0;
    const minLength = Math.min(expectedText.length, actualText.length);

    for (let i = 0; i < minLength; i++) {
      if (expectedText[i] !== actualText[i]) {
        errors++;
      }
    }

    // Add errors for length differences
    errors += Math.abs(expectedText.length - actualText.length);

    return errors;
  }

  private calculateRawAccuracy(
    expectedText: string,
    actualText: string
  ): number {
    if (expectedText.length === 0) return 1;

    const errors = this.countErrors(expectedText, actualText);
    const totalChars = Math.max(expectedText.length, actualText.length);

    return Math.max(0, (totalChars - errors) / totalChars);
  }

  private calculateSyntaxAccuracy(
    expectedTokens: any[],
    actualTokens: any[]
  ): number {
    if (expectedTokens.length === 0) return 1;

    const comparison = tokenizationService.compareTokenSequences(
      expectedTokens,
      actualTokens
    );
    return comparison?.tokenAccuracy || 1;
  }

  private calculateWhitespaceAccuracy(
    expectedText: string,
    actualText: string
  ): number {
    const expectedWhitespace =
      tokenizationService.extractWhitespaceTokens(expectedText) || [];
    const actualWhitespace =
      tokenizationService.extractWhitespaceTokens(actualText) || [];

    if (expectedWhitespace.length === 0) return 1;

    let correctWhitespace = 0;
    const minLength = Math.min(
      expectedWhitespace.length,
      actualWhitespace.length
    );

    for (let i = 0; i < minLength; i++) {
      if (expectedWhitespace[i].value === actualWhitespace[i].value) {
        correctWhitespace++;
      }
    }

    return correctWhitespace / expectedWhitespace.length;
  }

  private calculateCorrectionLatency(
    keystrokeEvents: KeystrokeEvent[]
  ): number {
    const corrections: number[] = [];
    let lastErrorTime: number | null = null;

    for (const event of keystrokeEvents) {
      if (event.key === 'Backspace' && lastErrorTime) {
        corrections.push(event.timestamp - lastErrorTime);
        lastErrorTime = null;
      } else if (event.action === 'keydown' && event.key !== 'Backspace') {
        // Assume this might be an error that needs correction
        lastErrorTime = event.timestamp;
      }
    }

    return corrections.length > 0
      ? corrections.reduce((a, b) => a + b, 0) / corrections.length
      : 0;
  }

  private analyzeErrorTypes(
    expectedText: string,
    actualText: string
  ): Record<string, number> {
    const errorTypes: Record<string, number> = {
      wrong_character: 0,
      missing_character: 0,
      extra_character: 0,
      transposition: 0,
    };

    const minLength = Math.min(expectedText.length, actualText.length);

    for (let i = 0; i < minLength; i++) {
      if (expectedText[i] !== actualText[i]) {
        errorTypes.wrong_character++;
      }
    }

    if (expectedText.length > actualText.length) {
      errorTypes.missing_character += expectedText.length - actualText.length;
    } else if (actualText.length > expectedText.length) {
      errorTypes.extra_character += actualText.length - expectedText.length;
    }

    return errorTypes;
  }

  private findErrorClusters(
    expectedText: string,
    actualText: string
  ): ErrorCluster[] {
    const clusters: ErrorCluster[] = [];
    const minLength = Math.min(expectedText.length, actualText.length);

    let currentCluster: Partial<ErrorCluster> | null = null;

    for (let i = 0; i < minLength; i++) {
      const isError = expectedText[i] !== actualText[i];

      if (isError) {
        if (!currentCluster) {
          currentCluster = {
            startPosition: i,
            errorCount: 1,
            dominantErrorType: 'wrong_character',
            affectedTokens: [],
          };
        } else {
          currentCluster.errorCount!++;
        }
      } else if (currentCluster) {
        // End current cluster
        currentCluster.endPosition = i - 1;
        currentCluster.severity =
          currentCluster.errorCount! > 3
            ? 'high'
            : currentCluster.errorCount! > 1
              ? 'medium'
              : 'low';
        clusters.push(currentCluster as ErrorCluster);
        currentCluster = null;
      }
    }

    // Handle final cluster
    if (currentCluster) {
      currentCluster.endPosition = minLength - 1;
      currentCluster.severity =
        currentCluster.errorCount! > 3
          ? 'high'
          : currentCluster.errorCount! > 1
            ? 'medium'
            : 'low';
      clusters.push(currentCluster as ErrorCluster);
    }

    return clusters;
  }

  private calculateSnippetDifficulty(tokens: any[]): number {
    if (tokens.length === 0) return 1;

    const avgDifficulty =
      tokens.reduce((sum, token) => sum + token.difficulty, 0) / tokens.length;
    return Math.round(avgDifficulty * 100) / 100;
  }

  private calculateStandardDeviation(values: number[]): number {
    if (values.length === 0) return 0;

    const mean = values.reduce((a, b) => a + b, 0) / values.length;
    const squaredDiffs = values.map(value => Math.pow(value - mean, 2));
    const avgSquaredDiff =
      squaredDiffs.reduce((a, b) => a + b, 0) / squaredDiffs.length;

    return Math.sqrt(avgSquaredDiff);
  }

  private processBurst(burst: {
    startTime: number;
    events: KeystrokeEvent[];
  }): BurstAnalysis['bursts'][0] {
    const endTime = burst.events[burst.events.length - 1].timestamp;
    const durationMinutes = (endTime - burst.startTime) / (1000 * 60);
    const characterCount = burst.events.length;
    const speed =
      durationMinutes > 0
        ? characterCount / this.AVERAGE_WORD_LENGTH / durationMinutes
        : 0;

    return {
      startTime: burst.startTime,
      endTime,
      speed: Math.round(speed * 100) / 100,
      accuracy: 1, // Would need actual vs expected text to calculate
      characterCount,
    };
  }

  private analyzeTrends(
    current: SessionMetrics,
    historical: SessionMetrics[]
  ): PerformanceInsights['progressIndicators'] {
    if (historical.length < 2) {
      return {
        speedTrend: 'stable',
        accuracyTrend: 'stable',
        consistencyTrend: 'stable',
      };
    }

    const recent = historical.slice(-5); // Last 5 sessions
    const avgSpeed =
      recent.reduce((sum, m) => sum + m.speed.twpm, 0) / recent.length;
    const avgAccuracy =
      recent.reduce((sum, m) => sum + m.accuracy.overallAccuracy, 0) /
      recent.length;
    const avgConsistency =
      recent.reduce(
        (sum, m) => sum + (1 - m.timing.keystrokeVariability / 1000),
        0
      ) / recent.length;

    return {
      speedTrend:
        current.speed.twpm > avgSpeed * 1.05
          ? 'improving'
          : current.speed.twpm < avgSpeed * 0.95
            ? 'declining'
            : 'stable',
      accuracyTrend:
        current.accuracy.overallAccuracy > avgAccuracy * 1.02
          ? 'improving'
          : current.accuracy.overallAccuracy < avgAccuracy * 0.98
            ? 'declining'
            : 'stable',
      consistencyTrend:
        1 - current.timing.keystrokeVariability / 1000 > avgConsistency * 1.02
          ? 'improving'
          : 1 - current.timing.keystrokeVariability / 1000 <
              avgConsistency * 0.98
            ? 'declining'
            : 'stable',
    };
  }

  private estimateTimeToGoal(
    _current: SessionMetrics,
    _historical: SessionMetrics[]
  ): number {
    // Simple estimation based on current improvement rate
    // In a real implementation, this would use more sophisticated modeling
    return 30; // Default 30 minutes of practice
  }
}

// Export singleton instance
export const metricsCalculator = MetricsCalculator.getInstance();
