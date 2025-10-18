// Parser services exports
export { ParserManager, parserManager } from './ParserManager';
export {
  TokenizationService,
  tokenizationService,
} from './TokenizationService';

// Session management exports
export { SessionManager, sessionManager } from './SessionManager';

// Metrics exports
export { MetricsCalculator, metricsCalculator } from './MetricsCalculator';

// Type exports
export type {
  Language,
  Token,
  ASTNode,
  ParseResult,
  ParseError,
  ParserConfig,
  LanguageGrammar,
} from '../types/parser';

export type { EnhancedToken, TokenCategory } from './TokenizationService';

export type {
  TypingSession,
  SessionConfig,
  SessionMode,
  SessionSettings,
  SessionState,
  SessionProgress,
  SessionResult,
  SessionHistory,
  EventBatch,
  OfflineQueue,
  SessionSummary,
  Achievement,
} from '../types/session';

export type {
  SpeedMetrics,
  AccuracyMetrics,
  ErrorMetrics,
  TimingMetrics,
  CompositeScore,
  ScoreWeights,
  PerformanceInsights,
  BurstAnalysis,
  TokenTypePerformance,
  ErrorCluster,
  SessionMetrics,
} from '../types/metrics';
