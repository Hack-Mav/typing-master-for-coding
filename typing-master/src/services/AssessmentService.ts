import { 
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
  AssessmentAnalytics
} from '../types/assessment';
import { Language } from '../types/parser';
import { SessionMetrics } from '../types/metrics';

// Define Snippet interface locally to avoid import issues
interface Snippet {
  id: string;
  languageId: string;
  title: string;
  sourceCode: string;
  tags: string[];
  difficulty: number;
  estimatedTime: number;
  checksum: string;
  accessibilityTags: Record<string, any>;
  createdAt: Date;
}

/**
 * Assessment Service - Manages standardized assessments and advanced scoring
 */
class AssessmentService {
  private baseUrl: string;
  private defaultScoringConfig: AdvancedScoringConfig;

  constructor() {
    this.baseUrl = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';
    this.defaultScoringConfig = this.getDefaultScoringConfig();
  }

  /**
   * Get available assessment blueprints for a language
   */
  async getAssessmentBlueprints(language: Language): Promise<AssessmentBlueprint[]> {
    try {
      const response = await fetch(`${this.baseUrl}/assessments/blueprints?language=${language}`);
      if (!response.ok) {
        throw new Error(`Failed to fetch assessment blueprints: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error fetching assessment blueprints:', error);
      // Return default blueprints for development
      return this.getDefaultBlueprints(language);
    }
  }

  /**
   * Get a specific assessment blueprint
   */
  async getAssessmentBlueprint(blueprintId: string): Promise<AssessmentBlueprint> {
    try {
      const response = await fetch(`${this.baseUrl}/assessments/blueprints/${blueprintId}`);
      if (!response.ok) {
        throw new Error(`Failed to fetch assessment blueprint: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error fetching assessment blueprint:', error);
      throw error;
    }
  }

  /**
   * Create a new assessment session
   */
  async createAssessmentSession(blueprintId: string, userId?: string): Promise<AssessmentSession> {
    try {
      const response = await fetch(`${this.baseUrl}/assessments/sessions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          blueprintId,
          userId,
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to create assessment session: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error creating assessment session:', error);
      // Return mock session for development
      return this.createMockSession(blueprintId, userId);
    }
  }

  /**
   * Get assessment snippet by ID
   */
  async getAssessmentSnippet(blueprintId: string, snippetId: string): Promise<Snippet> {
    try {
      const response = await fetch(`${this.baseUrl}/assessments/blueprints/${blueprintId}/snippets/${snippetId}`);
      if (!response.ok) {
        throw new Error(`Failed to fetch assessment snippet: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error fetching assessment snippet:', error);
      // Return mock snippet for development
      return this.getMockSnippet(snippetId);
    }
  }

  /**
   * Record snippet result
   */
  async recordSnippetResult(sessionId: string, result: AssessmentSnippetResult): Promise<void> {
    try {
      const response = await fetch(`${this.baseUrl}/assessments/sessions/${sessionId}/snippets`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(result),
      });

      if (!response.ok) {
        throw new Error(`Failed to record snippet result: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Error recording snippet result:', error);
      // In development, just log the result
      console.log('Mock snippet result recorded:', result);
    }
  }

  /**
   * Finalize assessment and calculate final result
   */
  async finalizeAssessment(sessionId: string, timeExpired: boolean = false): Promise<AssessmentResult> {
    try {
      const response = await fetch(`${this.baseUrl}/assessments/sessions/${sessionId}/finalize`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ timeExpired }),
      });

      if (!response.ok) {
        throw new Error(`Failed to finalize assessment: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error finalizing assessment:', error);
      // Return mock result for development
      return this.calculateMockResult(sessionId, timeExpired);
    }
  }

  /**
   * Calculate advanced scoring with structural penalties
   */
  calculateAdvancedScore(
    metrics: SessionMetrics,
    structuralScore: StructuralConformityScore,
    criteria: AssessmentCriteria,
    config: AdvancedScoringConfig = this.defaultScoringConfig
  ): AssessmentScoreBreakdown {
    const weights = config.baseWeights;
    const penalties = config.penaltyMultipliers;
    const bonuses = config.bonusMultipliers;

    // Base scores (0-100 each)
    const speedScore = Math.min(100, (metrics.speed.twpm / criteria.minimumSpeed) * 100);
    const accuracyScore = metrics.accuracy.rawAccuracy * 100;
    const structuralScoreValue = structuralScore.overallConformity * 100;
    const syntaxScore = structuralScore.syntaxValidationScore * 100;
    
    // Consistency score based on typing rhythm
    const consistencyScore = this.calculateConsistencyScore(metrics);
    
    // Error recovery score based on correction efficiency
    const errorRecoveryScore = this.calculateErrorRecoveryScore(metrics);

    // Calculate penalties
    let totalPenalties = 0;
    
    // Syntax error penalties
    const syntaxPenalties = structuralScore.structuralPenalties.reduce((sum, penalty) => {
      return sum + penalty.penaltyPoints * penalties.syntaxError;
    }, 0);
    totalPenalties += syntaxPenalties;

    // Structural mismatch penalties
    const structuralPenalties = (1 - structuralScore.astSimilarity) * 50 * penalties.structuralMismatch;
    totalPenalties += structuralPenalties;

    // Excessive backspace penalty
    if (metrics.errors.backspaceRate > 0.1) { // More than 10% backspaces
      totalPenalties += (metrics.errors.backspaceRate - 0.1) * 100 * penalties.excessiveBackspace;
    }

    // Calculate bonuses
    let bonusPoints = 0;
    
    // Perfect accuracy bonus
    if (metrics.accuracy.rawAccuracy >= 0.99) {
      bonusPoints += 20 * bonuses.perfectAccuracy;
    }
    
    // Speed bonus for exceeding minimum
    if (metrics.speed.twpm > criteria.minimumSpeed * 1.5) {
      bonusPoints += 15 * bonuses.speedBonus;
    }
    
    // Consistency bonus
    if (consistencyScore > 80) {
      bonusPoints += 10 * bonuses.consistencyBonus;
    }

    return {
      speedScore: speedScore * weights.speed,
      accuracyScore: accuracyScore * weights.accuracy,
      structuralScore: structuralScoreValue * weights.structuralConformity,
      syntaxScore: syntaxScore * weights.syntaxCorrectness,
      consistencyScore: consistencyScore * weights.consistency,
      errorRecoveryScore: errorRecoveryScore * weights.errorRecovery,
      totalPenalties,
      bonusPoints,
    };
  }

  /**
   * Determine assessment grade based on score
   */
  calculateGrade(totalScore: number, config: AdvancedScoringConfig = this.defaultScoringConfig): AssessmentGrade {
    const scale = config.gradingScale;
    
    for (const [grade, range] of Object.entries(scale)) {
      if (totalScore >= range.minScore && totalScore <= range.maxScore) {
        return grade as AssessmentGrade;
      }
    }
    
    return 'F'; // Default to F if no match
  }

  /**
   * Evaluate if a snippet result passes the criteria
   */
  evaluateSnippetPassing(
    metrics: SessionMetrics,
    structuralScore: StructuralConformityScore,
    criteria: AssessmentCriteria
  ): boolean {
    return (
      metrics.accuracy.rawAccuracy >= criteria.minimumAccuracy &&
      metrics.speed.twpm >= criteria.minimumSpeed &&
      (metrics.errors.errorRate || 0) <= criteria.maximumErrorRate &&
      structuralScore.overallConformity >= criteria.structuralAccuracyWeight
    );
  }

  /**
   * Get assessment analytics
   */
  async getAssessmentAnalytics(language?: Language): Promise<AssessmentAnalytics> {
    try {
      const url = language 
        ? `${this.baseUrl}/assessments/analytics?language=${language}`
        : `${this.baseUrl}/assessments/analytics`;
      
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`Failed to fetch assessment analytics: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error fetching assessment analytics:', error);
      return this.getMockAnalytics();
    }
  }

  /**
   * Schedule periodic assessment
   */
  async scheduleAssessment(userId: string, assessmentId: string, scheduledAt: Date): Promise<AssessmentSchedule> {
    try {
      const response = await fetch(`${this.baseUrl}/assessments/schedule`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          userId,
          assessmentId,
          scheduledAt: scheduledAt.toISOString(),
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to schedule assessment: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error scheduling assessment:', error);
      throw error;
    }
  }

  /**
   * Get earned badges for user
   */
  async getUserBadges(userId: string): Promise<AssessmentBadge[]> {
    try {
      const response = await fetch(`${this.baseUrl}/assessments/badges?userId=${userId}`);
      if (!response.ok) {
        throw new Error(`Failed to fetch user badges: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error fetching user badges:', error);
      return [];
    }
  }

  // Private helper methods

  private calculateConsistencyScore(metrics: SessionMetrics): number {
    // Calculate typing rhythm consistency
    // This would analyze keystroke timing patterns
    // For now, return a mock score based on accuracy
    return Math.max(0, Math.min(100, metrics.accuracy.rawAccuracy * 100 - (metrics.errors.backspaceRate * 50)));
  }

  private calculateErrorRecoveryScore(metrics: SessionMetrics): number {
    // Calculate how efficiently errors are corrected
    // For now, return inverse of backspace rate
    return Math.max(0, Math.min(100, 100 - (metrics.errors.backspaceRate * 100)));
  }

  private getDefaultScoringConfig(): AdvancedScoringConfig {
    return {
      baseWeights: {
        speed: 0.25,
        accuracy: 0.30,
        structuralConformity: 0.20,
        syntaxCorrectness: 0.15,
        consistency: 0.05,
        errorRecovery: 0.05,
      },
      penaltyMultipliers: {
        syntaxError: 2.0,
        structuralMismatch: 1.5,
        timeOverrun: 1.2,
        excessiveBackspace: 1.0,
      },
      bonusMultipliers: {
        perfectAccuracy: 1.5,
        speedBonus: 1.2,
        consistencyBonus: 1.1,
        earlyCompletion: 1.0,
      },
      gradingScale: {
        'A+': { minScore: 950, maxScore: 1000 },
        'A': { minScore: 900, maxScore: 949 },
        'A-': { minScore: 850, maxScore: 899 },
        'B+': { minScore: 800, maxScore: 849 },
        'B': { minScore: 750, maxScore: 799 },
        'B-': { minScore: 700, maxScore: 749 },
        'C+': { minScore: 650, maxScore: 699 },
        'C': { minScore: 600, maxScore: 649 },
        'C-': { minScore: 550, maxScore: 599 },
        'D': { minScore: 500, maxScore: 549 },
        'F': { minScore: 0, maxScore: 499 },
      },
    };
  }

  private getDefaultBlueprints(language: Language): AssessmentBlueprint[] {
    return [
      {
        id: `${language}-basic-assessment`,
        name: `${language.toUpperCase()} Basic Assessment`,
        description: `Standardized assessment for ${language} syntax and typing proficiency`,
        language,
        difficulty: 3,
        estimatedDuration: 15,
        snippetIds: [`${language}-snippet-1`, `${language}-snippet-2`, `${language}-snippet-3`],
        passingCriteria: {
          minimumAccuracy: 0.85,
          minimumSpeed: 30,
          maximumErrorRate: 5,
          structuralAccuracyWeight: 0.8,
          syntaxPenaltyMultiplier: 2.0,
          timeLimit: 20,
        },
        scoringWeights: this.defaultScoringConfig.baseWeights,
        version: 1,
        createdAt: new Date(),
      },
    ];
  }

  private createMockSession(blueprintId: string, userId?: string): AssessmentSession {
    return {
      id: `session-${Date.now()}`,
      blueprintId,
      userId,
      startedAt: new Date(),
      status: 'not_started',
      currentSnippetIndex: 0,
      snippetResults: [],
      metadata: {
        language: 'javascript' as Language,
        difficulty: 3,
        totalSnippets: 3,
        estimatedDuration: 15,
        environment: {
          userAgent: navigator.userAgent,
          screenResolution: `${screen.width}x${screen.height}`,
          keyboardLayout: 'QWERTY',
        },
        settings: {
          fontSize: 14,
          theme: 'light',
          soundEnabled: false,
        },
      },
    };
  }

  private getMockSnippet(snippetId: string): Snippet {
    const mockSnippets: Record<string, string> = {
      'javascript-snippet-1': `function fibonacci(n) {
  if (n <= 1) return n;
  return fibonacci(n - 1) + fibonacci(n - 2);
}`,
      'javascript-snippet-2': `const users = [
  { id: 1, name: 'Alice', active: true },
  { id: 2, name: 'Bob', active: false }
];

const activeUsers = users.filter(user => user.active);`,
      'javascript-snippet-3': `class Calculator {
  constructor() {
    this.result = 0;
  }
  
  add(value) {
    this.result += value;
    return this;
  }
  
  multiply(value) {
    this.result *= value;
    return this;
  }
}`,
    };

    return {
      id: snippetId,
      languageId: 'javascript',
      title: `Assessment Snippet ${snippetId.split('-').pop()}`,
      sourceCode: mockSnippets[snippetId] || 'console.log("Hello, World!");',
      tags: ['assessment', 'basic'],
      difficulty: 3,
      estimatedTime: 300,
      checksum: 'mock-checksum',
      accessibilityTags: {},
      createdAt: new Date(),
    };
  }

  private calculateMockResult(sessionId: string, timeExpired: boolean): AssessmentResult {
    const mockScore = timeExpired ? 650 : 850;
    const grade = this.calculateGrade(mockScore);
    
    return {
      overallScore: mockScore,
      passed: mockScore >= 700,
      grade,
      breakdown: {
        speedScore: 75,
        accuracyScore: 85,
        structuralScore: 80,
        syntaxScore: 90,
        consistencyScore: 70,
        errorRecoveryScore: 75,
        totalPenalties: 25,
        bonusPoints: 10,
      },
      recommendations: [
        'Focus on improving typing consistency',
        'Practice more complex syntax patterns',
        'Work on error correction efficiency',
      ],
      certificateEligible: mockScore >= 800,
      retakeAllowed: mockScore < 700,
      nextAssessmentSuggestion: mockScore >= 800 ? 'advanced-assessment' : undefined,
    };
  }

  private getMockAnalytics(): AssessmentAnalytics {
    return {
      totalAttempts: 1250,
      passRate: 0.72,
      averageScore: 745,
      averageDuration: 18.5,
      commonFailurePoints: [
        {
          snippetId: 'javascript-snippet-2',
          position: 45,
          errorType: 'syntax_error',
          frequency: 0.35,
          averageRecoveryTime: 2500,
        },
      ],
      difficultyDistribution: {
        1: 0.05,
        2: 0.15,
        3: 0.40,
        4: 0.30,
        5: 0.10,
      },
      languagePerformance: {
        javascript: {
          averageScore: 745,
          passRate: 0.72,
          commonErrors: ['missing_semicolon', 'bracket_mismatch'],
          strengthAreas: ['function_syntax', 'variable_declaration'],
        },
        python: {
          averageScore: 720,
          passRate: 0.68,
          commonErrors: ['indentation_error', 'colon_missing'],
          strengthAreas: ['list_comprehension', 'function_definition'],
        },
        cpp: {
          averageScore: 680,
          passRate: 0.62,
          commonErrors: ['pointer_syntax', 'template_syntax'],
          strengthAreas: ['basic_syntax', 'control_flow'],
        },
        rust: {
          averageScore: 665,
          passRate: 0.58,
          commonErrors: ['lifetime_syntax', 'ownership_concepts'],
          strengthAreas: ['pattern_matching', 'basic_syntax'],
        },
        yaml: {
          averageScore: 780,
          passRate: 0.78,
          commonErrors: ['indentation_error', 'key_value_syntax'],
          strengthAreas: ['basic_structure', 'list_syntax'],
        },
      },
    };
  }
}

export { AssessmentService };
export const assessmentService = new AssessmentService();