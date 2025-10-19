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

// Authentication and user management exports
export { authService } from './AuthService';
export type {
  User,
  TokenPair,
  AuthResponse,
  RegisterRequest,
  LoginRequest,
  AnonymousSessionRequest,
} from './AuthService';

// Privacy and GDPR compliance exports
export { privacyService } from './PrivacyService';
export type {
  ConsentStatus,
  PrivacySettings,
  DataExportOptions,
  TelemetryEvent,
} from './PrivacyService';

// Local storage for anonymous mode exports
export { localStorageService } from './LocalStorageService';
export type {
  LocalSession,
  LocalResult,
  LocalProgress,
  LocalSettings,
} from './LocalStorageService';

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
