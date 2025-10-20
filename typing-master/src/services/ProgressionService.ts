import { authService } from './AuthService';
import { LessonData } from '../utils/indexedDB';

/**
 * ProgressionService - Handles lesson progression and user progress tracking
 */
export interface LessonProgress {
  lessonId: string;
  userId: string;
  status: 'not_started' | 'in_progress' | 'completed' | 'mastered';
  attempts: number;
  bestScore: number;
  lastAttemptAt?: Date;
  timeSpent: number; // in seconds
  mistakes: number;
  hintsUsed: number;
  completedAt?: Date;
}

export interface UserProgressSummary {
  totalLessons: number;
  completedLessons: number;
  masteredLessons: number;
  totalTimeSpent: number;
  averageScore: number;
  currentStreak: number;
  longestStreak: number;
  strengths: string[];
  areasForImprovement: string[];
  recentAchievements: string[];
}

export interface LessonPrerequisite {
  lessonId: string;
  prerequisiteIds: string[];
  requiredScore?: number;
  estimatedTime: number;
}

export interface LessonProgressionFlow {
  lessonId: string;
  prerequisites: string[];
  nextLessons: string[];
  relatedLessons: string[];
  difficulty: number;
  estimatedTime: number;
  learningPath: string[];
}

class ProgressionService {
  private apiBaseUrl: string;

  constructor() {
    this.apiBaseUrl = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';
  }

  /**
   * Get lesson progress for a specific lesson
   */
  async getLessonProgress(lessonId: string): Promise<LessonProgress | null> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/lessons/${lessonId}/progress`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        if (response.status === 404) {
          return null;
        }
        throw new Error(`Failed to fetch lesson progress: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching lesson progress:', error);
      return null;
    }
  }

  /**
   * Update lesson progress
   */
  async updateLessonProgress(
    lessonId: string,
    progress: Partial<LessonProgress>
  ): Promise<LessonProgress> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/lessons/${lessonId}/progress`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          ...authService.getAuthHeader(),
        },
        body: JSON.stringify(progress),
      });

      if (!response.ok) {
        throw new Error(`Failed to update lesson progress: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error updating lesson progress:', error);
      throw error;
    }
  }

  /**
   * Check if lesson prerequisites are met
   */
  async checkLessonPrerequisites(lessonId: string): Promise<{
    canAccess: boolean;
    missingPrerequisites: string[];
    recommendations: string[];
  }> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/lessons/${lessonId}/prerequisites`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(`Failed to check lesson prerequisites: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error checking lesson prerequisites:', error);
      return {
        canAccess: false,
        missingPrerequisites: [],
        recommendations: [],
      };
    }
  }

  /**
   * Get lesson progression flow
   */
  async getLessonProgressionFlow(lessonId: string): Promise<LessonProgressionFlow> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/lessons/${lessonId}/flow`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch lesson progression flow: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching lesson progression flow:', error);
      throw error;
    }
  }

  /**
   * Get user progression summary
   */
  async getUserProgressionSummary(): Promise<UserProgressSummary> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/progression/summary`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch progression summary: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching progression summary:', error);
      throw error;
    }
  }

  /**
   * Get next recommended lessons based on progress
   */
  async getRecommendedLessons(limit: number = 5): Promise<LessonData[]> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/progression/recommendations?limit=${limit}`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch recommended lessons: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching recommended lessons:', error);
      return [];
    }
  }

  /**
   * Mark lesson as completed
   */
  async completeLesson(
    lessonId: string,
    score: number,
    timeSpent: number,
    mistakes: number
  ): Promise<LessonProgress> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/lessons/${lessonId}/complete`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...authService.getAuthHeader(),
        },
        body: JSON.stringify({
          score,
          timeSpent,
          mistakes,
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to complete lesson: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error completing lesson:', error);
      throw error;
    }
  }

  /**
   * Get learning path for a specific skill or topic
   */
  async getLearningPath(topic: string): Promise<LessonData[]> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/progression/path?topic=${encodeURIComponent(topic)}`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch learning path: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching learning path:', error);
      return [];
    }
  }

  /**
   * Get user's current skill level for each programming concept
   */
  async getSkillLevels(): Promise<Record<string, {
    level: number;
    proficiency: 'beginner' | 'intermediate' | 'advanced' | 'expert';
    lessonsCompleted: number;
    averageScore: number;
  }>> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/progression/skill-levels`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch skill levels: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching skill levels:', error);
      return {};
    }
  }

  /**
   * Get detailed progress analytics
   */
  async getProgressAnalytics(timeRange: 'week' | 'month' | 'quarter' | 'year' = 'month'): Promise<{
    totalPracticeTime: number;
    lessonsCompleted: number;
    averageScore: number;
    improvementRate: number;
    consistencyScore: number;
    dailyProgress: { date: string; lessons: number; timeSpent: number }[];
    skillProgression: { skill: string; progress: number; trend: 'up' | 'down' | 'stable' }[];
  }> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/progression/analytics?timeRange=${timeRange}`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch progress analytics: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching progress analytics:', error);
      throw error;
    }
  }

  /**
   * Reset progress for a specific lesson (for testing/restarting)
   */
  async resetLessonProgress(lessonId: string): Promise<void> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/lessons/${lessonId}/progress`, {
        method: 'DELETE',
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(`Failed to reset lesson progress: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Error resetting lesson progress:', error);
      throw error;
    }
  }

  /**
   * Get lesson completion certificate (if applicable)
   */
  async getLessonCertificate(lessonId: string): Promise<{
    certificateId: string;
    lessonName: string;
    completedAt: Date;
    score: number;
    certificateUrl?: string;
  } | null> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/lessons/${lessonId}/certificate`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        if (response.status === 404) {
          return null;
        }
        throw new Error(`Failed to fetch lesson certificate: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching lesson certificate:', error);
      return null;
    }
  }
}

// Export singleton instance
export const progressionService = new ProgressionService();
export default progressionService;
