import { Language } from './parser';
import { SessionMetrics } from './metrics';

/**
 * Assessment mode types for standardized evaluation
 */

export interface AssessmentBlueprint {
  id: string;
  name: string;
  description: string;
  language: Language;
  difficulty: number;
  estimatedDuration: number; // in minutes
  snippetIds: string[];
  passingCriteria: AssessmentCriteria;
  scoringWeights: AssessmentWeights;
  version: number;
  createdAt: Date;
}

export interface AssessmentCriteria {
  minimumAccuracy: number; // 0-1
  minimumSpeed: number; // tWPM
  maximumErrorRate: number; // errors per minute
  structuralAccuracyWeight: number; // 0-1
  syntaxPenaltyMultiplier: number; // penalty for syntax errors
  timeLimit?: number; // optional time limit in minutes
}

export interface AssessmentWeights {
  speed: number; // tWPM weight
  accuracy: number; // overall accuracy weight
  structuralConformity: number; // AST shape conformity weight
  syntaxCorrectness: number; // syntax validation weight
  consistency: number; // typing consistency weight
  errorRecovery: number; // error correction efficiency weight
}

export interface AssessmentSession {
  id: string;
  blueprintId: string;
  userId?: string;
  startedAt: Date;
  completedAt?: Date;
  status: AssessmentStatus;
  currentSnippetIndex: number;
  snippetResults: AssessmentSnippetResult[];
  overallResult?: AssessmentResult;
  metadata: AssessmentMetadata;
}

export type AssessmentStatus = 
  | 'not_started'
  | 'in_progress' 
  | 'paused'
  | 'completed'
  | 'failed'
  | 'expired';

export interface AssessmentSnippetResult {
  snippetId: string;
  startedAt: Date;
  completedAt?: Date;
  expectedText: string;
  actualText: string;
  keystrokeEvents: any[]; // KeystrokeEvent[]
  metrics: SessionMetrics;
  structuralScore: StructuralConformityScore;
  passed: boolean;
  timeSpent: number; // milliseconds
}

export interface StructuralConformityScore {
  astSimilarity: number; // 0-1, how similar the AST structures are
  tokenSequenceAccuracy: number; // 0-1, token-level accuracy
  syntaxValidationScore: number; // 0-1, syntax correctness
  structuralPenalties: StructuralPenalty[];
  overallConformity: number; // 0-1, weighted overall score
}

export interface StructuralPenalty {
  type: StructuralPenaltyType;
  severity: 'low' | 'medium' | 'high';
  position: number;
  description: string;
  penaltyPoints: number;
}

export type StructuralPenaltyType =
  | 'missing_delimiter'
  | 'extra_delimiter'
  | 'mismatched_delimiter'
  | 'incorrect_indentation'
  | 'missing_token'
  | 'extra_token'
  | 'wrong_token_type'
  | 'structural_mismatch';

export interface AssessmentResult {
  overallScore: number; // 0-1000
  passed: boolean;
  grade: AssessmentGrade;
  breakdown: AssessmentScoreBreakdown;
  recommendations: string[];
  certificateEligible: boolean;
  retakeAllowed: boolean;
  nextAssessmentSuggestion?: string;
}

export type AssessmentGrade = 'A+' | 'A' | 'A-' | 'B+' | 'B' | 'B-' | 'C+' | 'C' | 'C-' | 'D' | 'F';

export interface AssessmentScoreBreakdown {
  speedScore: number;
  accuracyScore: number;
  structuralScore: number;
  syntaxScore: number;
  consistencyScore: number;
  errorRecoveryScore: number;
  totalPenalties: number;
  bonusPoints: number;
}

export interface AssessmentMetadata {
  language: Language;
  difficulty: number;
  totalSnippets: number;
  estimatedDuration: number;
  actualDuration?: number;
  environment: {
    userAgent: string;
    screenResolution: string;
    keyboardLayout: string;
  };
  settings: {
    fontSize: number;
    theme: string;
    soundEnabled: boolean;
  };
}

export interface AssessmentSchedule {
  userId: string;
  assessmentId: string;
  scheduledAt: Date;
  reminderSent: boolean;
  completed: boolean;
  rescheduledCount: number;
}

export interface AssessmentBadge {
  id: string;
  name: string;
  description: string;
  iconUrl: string;
  criteria: BadgeCriteria;
  rarity: 'common' | 'uncommon' | 'rare' | 'epic' | 'legendary';
  earnedAt?: Date;
}

export interface BadgeCriteria {
  minimumScore?: number;
  minimumGrade?: AssessmentGrade;
  specificLanguage?: Language;
  consecutivePasses?: number;
  timeConstraint?: number; // complete within X minutes
  perfectAccuracy?: boolean;
  speedThreshold?: number; // minimum tWPM
}

/**
 * AST Shape Conformity Analysis
 */
export interface ASTConformityAnalysis {
  structuralSimilarity: number; // 0-1
  nodeTypeMatches: number;
  nodeTypeMismatches: number;
  depthSimilarity: number; // 0-1
  branchingSimilarity: number; // 0-1
  missingNodes: ASTNodeDifference[];
  extraNodes: ASTNodeDifference[];
  mismatchedNodes: ASTNodeDifference[];
}

export interface ASTNodeDifference {
  expectedType: string;
  actualType?: string;
  position: number;
  severity: 'low' | 'medium' | 'high';
  impact: string;
}

/**
 * Advanced Scoring Configuration
 */
export interface AdvancedScoringConfig {
  baseWeights: AssessmentWeights;
  penaltyMultipliers: {
    syntaxError: number;
    structuralMismatch: number;
    timeOverrun: number;
    excessiveBackspace: number;
  };
  bonusMultipliers: {
    perfectAccuracy: number;
    speedBonus: number;
    consistencyBonus: number;
    earlyCompletion: number;
  };
  gradingScale: {
    [key in AssessmentGrade]: {
      minScore: number;
      maxScore: number;
    };
  };
}

/**
 * Assessment Analytics
 */
export interface AssessmentAnalytics {
  totalAttempts: number;
  passRate: number;
  averageScore: number;
  averageDuration: number;
  commonFailurePoints: FailurePoint[];
  difficultyDistribution: Record<number, number>;
  languagePerformance: Record<Language, LanguagePerformance>;
}

export interface FailurePoint {
  snippetId: string;
  position: number;
  errorType: string;
  frequency: number;
  averageRecoveryTime: number;
}

export interface LanguagePerformance {
  averageScore: number;
  passRate: number;
  commonErrors: string[];
  strengthAreas: string[];
}