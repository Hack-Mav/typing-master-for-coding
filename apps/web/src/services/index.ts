// Parser services exports
export { ParserManager, parserManager } from './ParserManager';
export {
  TokenizationService,
  tokenizationService,
} from './TokenizationService';

// Content management exports
export { ContentService, contentService } from './ContentService';

// Session management exports
export { SessionManager, sessionManager } from './SessionManager';

// Metrics exports
export { MetricsCalculator, metricsCalculator } from './MetricsCalculator';

// Assessment exports
export { AssessmentService, assessmentService } from './AssessmentService';
export { StructuralAnalyzer, structuralAnalyzer } from './StructuralAnalyzer';

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

// Leaderboard and tournament exports
export { leaderboardService } from './LeaderboardService';
export type {
  LeaderboardEntry,
  Tournament,
  TournamentParticipant,
  TournamentResult,
} from './LeaderboardService';

// Admin panel exports
export { adminService } from './AdminService';
export type { AdminDashboard, ABTest } from './AdminService';

// Lesson progression exports
export { progressionService } from './ProgressionService';
export type {
  LessonProgress,
  UserProgressSummary,
  LessonPrerequisite,
  LessonProgressionFlow,
} from './ProgressionService';

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

export type {
  AssessmentBlueprint,
  AssessmentSession,
  AssessmentResult,
  AssessmentSnippetResult,
  AssessmentCriteria,
  StructuralConformityScore,
  AssessmentGrade,
  AdvancedScoringConfig,
  AssessmentWeights,
  AssessmentScoreBreakdown,
  AssessmentBadge,
  AssessmentSchedule,
  AssessmentAnalytics,
  ASTConformityAnalysis,
  ASTNodeDifference,
  StructuralPenalty,
  StructuralPenaltyType,
} from '../types/assessment';

export type {
  LanguageEntity,
  Lesson,
  Snippet,
  Playlist,
  ContentVersion,
  ContentValidation,
  ValidationError,
  ValidationWarning,
} from '../types/content';
