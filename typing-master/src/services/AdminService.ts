import { authService } from './AuthService';
import { LanguageEntity, Lesson, Snippet } from '../types/content';

/**
 * AdminService - Handles admin panel functionality
 */
export interface AdminDashboard {
  totalUsers: number;
  activeUsers: number;
  totalSessions: number;
  averageSessionDuration: number;
  popularLanguages: { language: string; count: number }[];
  recentActivity: {
    userId: string;
    action: string;
    timestamp: Date;
    details?: string;
  }[];
  systemHealth: {
    database: 'healthy' | 'warning' | 'error';
    cache: 'healthy' | 'warning' | 'error';
    api: 'healthy' | 'warning' | 'error';
  };
}

export interface ABTest {
  id: string;
  name: string;
  description: string;
  status: 'draft' | 'running' | 'completed' | 'paused';
  variants: {
    id: string;
    name: string;
    trafficPercentage: number;
    config: Record<string, any>;
  }[];
  startDate?: Date;
  endDate?: Date;
  results?: {
    variantId: string;
    participants: number;
    conversions: number;
    conversionRate: number;
    statisticalSignificance: number;
  }[];
}

class AdminService {
  private apiBaseUrl: string;

  constructor() {
    this.apiBaseUrl =
      process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';
  }

  /**
   * Get admin dashboard data
   */
  async getDashboard(): Promise<AdminDashboard> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/admin/dashboard`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch dashboard: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching admin dashboard:', error);
      throw error;
    }
  }

  // Content Management

  /**
   * Get all languages (admin view)
   */
  async getAdminLanguages(): Promise<LanguageEntity[]> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/admin/languages`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(
          `Failed to fetch admin languages: ${response.statusText}`
        );
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching admin languages:', error);
      return [];
    }
  }

  /**
   * Create a new language
   */
  async createLanguage(
    language: Omit<LanguageEntity, 'id' | 'createdAt'>
  ): Promise<LanguageEntity> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/admin/languages`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...authService.getAuthHeader(),
        },
        body: JSON.stringify(language),
      });

      if (!response.ok) {
        throw new Error(`Failed to create language: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error creating language:', error);
      throw error;
    }
  }

  /**
   * Update a language
   */
  async updateLanguage(
    languageId: string,
    updates: Partial<LanguageEntity>
  ): Promise<LanguageEntity> {
    try {
      const response = await fetch(
        `${this.apiBaseUrl}/admin/languages/${languageId}`,
        {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
            ...authService.getAuthHeader(),
          },
          body: JSON.stringify(updates),
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to update language: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error updating language:', error);
      throw error;
    }
  }

  /**
   * Delete a language
   */
  async deleteLanguage(languageId: string): Promise<void> {
    try {
      const response = await fetch(
        `${this.apiBaseUrl}/admin/languages/${languageId}`,
        {
          method: 'DELETE',
          headers: authService.getAuthHeader(),
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to delete language: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Error deleting language:', error);
      throw error;
    }
  }

  /**
   * Get all lessons (admin view)
   */
  async getAdminLessons(): Promise<Lesson[]> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/admin/lessons`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(
          `Failed to fetch admin lessons: ${response.statusText}`
        );
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching admin lessons:', error);
      return [];
    }
  }

  /**
   * Create a new lesson
   */
  async createLesson(
    lesson: Omit<Lesson, 'id' | 'createdAt'>
  ): Promise<Lesson> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/admin/lessons`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...authService.getAuthHeader(),
        },
        body: JSON.stringify(lesson),
      });

      if (!response.ok) {
        throw new Error(`Failed to create lesson: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error creating lesson:', error);
      throw error;
    }
  }

  /**
   * Update a lesson
   */
  async updateLesson(
    lessonId: string,
    updates: Partial<Lesson>
  ): Promise<Lesson> {
    try {
      const response = await fetch(
        `${this.apiBaseUrl}/admin/lessons/${lessonId}`,
        {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
            ...authService.getAuthHeader(),
          },
          body: JSON.stringify(updates),
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to update lesson: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error updating lesson:', error);
      throw error;
    }
  }

  /**
   * Delete a lesson
   */
  async deleteLesson(lessonId: string): Promise<void> {
    try {
      const response = await fetch(
        `${this.apiBaseUrl}/admin/lessons/${lessonId}`,
        {
          method: 'DELETE',
          headers: authService.getAuthHeader(),
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to delete lesson: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Error deleting lesson:', error);
      throw error;
    }
  }

  /**
   * Get all snippets (admin view)
   */
  async getAdminSnippets(): Promise<Snippet[]> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/admin/snippets`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(
          `Failed to fetch admin snippets: ${response.statusText}`
        );
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching admin snippets:', error);
      return [];
    }
  }

  /**
   * Create a new snippet
   */
  async createSnippet(
    snippet: Omit<Snippet, 'id' | 'createdAt'>
  ): Promise<Snippet> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/admin/snippets`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...authService.getAuthHeader(),
        },
        body: JSON.stringify(snippet),
      });

      if (!response.ok) {
        throw new Error(`Failed to create snippet: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error creating snippet:', error);
      throw error;
    }
  }

  /**
   * Update a snippet
   */
  async updateSnippet(
    snippetId: string,
    updates: Partial<Snippet>
  ): Promise<Snippet> {
    try {
      const response = await fetch(
        `${this.apiBaseUrl}/admin/snippets/${snippetId}`,
        {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
            ...authService.getAuthHeader(),
          },
          body: JSON.stringify(updates),
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to update snippet: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error updating snippet:', error);
      throw error;
    }
  }

  /**
   * Delete a snippet
   */
  async deleteSnippet(snippetId: string): Promise<void> {
    try {
      const response = await fetch(
        `${this.apiBaseUrl}/admin/snippets/${snippetId}`,
        {
          method: 'DELETE',
          headers: authService.getAuthHeader(),
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to delete snippet: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Error deleting snippet:', error);
      throw error;
    }
  }

  // A/B Testing Management

  /**
   * Get all A/B tests
   */
  async getABTests(): Promise<ABTest[]> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/admin/ab-tests`, {
        headers: authService.getAuthHeader(),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch A/B tests: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching A/B tests:', error);
      return [];
    }
  }

  /**
   * Create a new A/B test
   */
  async createABTest(test: Omit<ABTest, 'id'>): Promise<ABTest> {
    try {
      const response = await fetch(`${this.apiBaseUrl}/admin/ab-tests`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...authService.getAuthHeader(),
        },
        body: JSON.stringify(test),
      });

      if (!response.ok) {
        throw new Error(`Failed to create A/B test: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error creating A/B test:', error);
      throw error;
    }
  }

  /**
   * Update an A/B test
   */
  async updateABTest(
    testId: string,
    updates: Partial<ABTest>
  ): Promise<ABTest> {
    try {
      const response = await fetch(
        `${this.apiBaseUrl}/admin/ab-tests/${testId}`,
        {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
            ...authService.getAuthHeader(),
          },
          body: JSON.stringify(updates),
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to update A/B test: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error updating A/B test:', error);
      throw error;
    }
  }

  /**
   * Delete an A/B test
   */
  async deleteABTest(testId: string): Promise<void> {
    try {
      const response = await fetch(
        `${this.apiBaseUrl}/admin/ab-tests/${testId}`,
        {
          method: 'DELETE',
          headers: authService.getAuthHeader(),
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to delete A/B test: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Error deleting A/B test:', error);
      throw error;
    }
  }

  /**
   * Get A/B test results
   */
  async getABTestResults(testId: string): Promise<ABTest['results']> {
    try {
      const response = await fetch(
        `${this.apiBaseUrl}/admin/ab-tests/${testId}/results`,
        {
          headers: authService.getAuthHeader(),
        }
      );

      if (!response.ok) {
        throw new Error(
          `Failed to fetch A/B test results: ${response.statusText}`
        );
      }

      return await response.json();
    } catch (error) {
      console.error('Error fetching A/B test results:', error);
      return [];
    }
  }
}

// Export singleton instance
export const adminService = new AdminService();
export default adminService;
