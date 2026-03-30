/**
 * LocalStorageService - Manages local-only data storage for anonymous mode
 * All data stays on the device and is never transmitted to the server
 */

export interface LocalSession {
  id: string;
  mode: string;
  languageId: string;
  lessonId?: string;
  snippetId?: string;
  startedAt: string;
  endedAt?: string;
  durationMs?: number;
  settings: Record<string, any>;
}

export interface LocalResult {
  sessionId: string;
  cpm: number;
  twpm: number;
  rawAccuracy: number;
  tokenAccuracy: number;
  syntaxAccuracy: number;
  backspaceRate: number;
  compositeScore: number;
  breakdown: Record<string, any>;
  createdAt: string;
}

export interface LocalProgress {
  languageId: string;
  lessonId: string;
  completed: boolean;
  lastPracticed: string;
  attempts: number;
  bestScore: number;
}

export interface LocalSettings {
  theme: string;
  fontSize: number;
  keyboardLayout: string;
  locale: string;
  soundEnabled: boolean;
  showMetrics: boolean;
  privacyMode: boolean;
}

class LocalStorageService {
  private readonly STORAGE_PREFIX = 'typing_master_';
  private readonly SESSIONS_KEY = `${this.STORAGE_PREFIX}sessions`;
  private readonly RESULTS_KEY = `${this.STORAGE_PREFIX}results`;
  private readonly PROGRESS_KEY = `${this.STORAGE_PREFIX}progress`;
  private readonly SETTINGS_KEY = `${this.STORAGE_PREFIX}settings`;
  private readonly MAX_SESSIONS = 100;
  private readonly MAX_RESULTS = 500;

  /**
   * Save a session to local storage
   */
  saveSession(session: LocalSession): void {
    const sessions = this.getSessions();

    // Check if session already exists
    const existingIndex = sessions.findIndex(s => s.id === session.id);
    if (existingIndex >= 0) {
      sessions[existingIndex] = session;
    } else {
      sessions.push(session);
    }

    // Limit number of stored sessions
    if (sessions.length > this.MAX_SESSIONS) {
      sessions.sort(
        (a, b) =>
          new Date(b.startedAt).getTime() - new Date(a.startedAt).getTime()
      );
      sessions.splice(this.MAX_SESSIONS);
    }

    this.setItem(this.SESSIONS_KEY, sessions);
  }

  /**
   * Get all sessions from local storage
   */
  getSessions(): LocalSession[] {
    return this.getItem<LocalSession[]>(this.SESSIONS_KEY) || [];
  }

  /**
   * Get a specific session by ID
   */
  getSession(sessionId: string): LocalSession | null {
    const sessions = this.getSessions();
    return sessions.find(s => s.id === sessionId) || null;
  }

  /**
   * Delete a session
   */
  deleteSession(sessionId: string): void {
    const sessions = this.getSessions().filter(s => s.id !== sessionId);
    this.setItem(this.SESSIONS_KEY, sessions);
  }

  /**
   * Save a result to local storage
   */
  saveResult(result: LocalResult): void {
    const results = this.getResults();

    // Check if result already exists
    const existingIndex = results.findIndex(
      r => r.sessionId === result.sessionId
    );
    if (existingIndex >= 0) {
      results[existingIndex] = result;
    } else {
      results.push(result);
    }

    // Limit number of stored results
    if (results.length > this.MAX_RESULTS) {
      results.sort(
        (a, b) =>
          new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      );
      results.splice(this.MAX_RESULTS);
    }

    this.setItem(this.RESULTS_KEY, results);
  }

  /**
   * Get all results from local storage
   */
  getResults(): LocalResult[] {
    return this.getItem<LocalResult[]>(this.RESULTS_KEY) || [];
  }

  /**
   * Get a specific result by session ID
   */
  getResult(sessionId: string): LocalResult | null {
    const results = this.getResults();
    return results.find(r => r.sessionId === sessionId) || null;
  }

  /**
   * Delete a result
   */
  deleteResult(sessionId: string): void {
    const results = this.getResults().filter(r => r.sessionId !== sessionId);
    this.setItem(this.RESULTS_KEY, results);
  }

  /**
   * Save lesson progress
   */
  saveProgress(progress: LocalProgress): void {
    const allProgress = this.getProgress();
    const key = `${progress.languageId}_${progress.lessonId}`;

    const existingIndex = allProgress.findIndex(
      p => `${p.languageId}_${p.lessonId}` === key
    );

    if (existingIndex >= 0) {
      allProgress[existingIndex] = progress;
    } else {
      allProgress.push(progress);
    }

    this.setItem(this.PROGRESS_KEY, allProgress);
  }

  /**
   * Get all lesson progress
   */
  getProgress(): LocalProgress[] {
    return this.getItem<LocalProgress[]>(this.PROGRESS_KEY) || [];
  }

  /**
   * Get progress for a specific lesson
   */
  getLessonProgress(
    languageId: string,
    lessonId: string
  ): LocalProgress | null {
    const allProgress = this.getProgress();
    return (
      allProgress.find(
        p => p.languageId === languageId && p.lessonId === lessonId
      ) || null
    );
  }

  /**
   * Save user settings
   */
  saveSettings(settings: LocalSettings): void {
    this.setItem(this.SETTINGS_KEY, settings);
  }

  /**
   * Get user settings
   */
  getSettings(): LocalSettings | null {
    return this.getItem<LocalSettings>(this.SETTINGS_KEY);
  }

  /**
   * Get default settings
   */
  getDefaultSettings(): LocalSettings {
    return {
      theme: 'dark',
      fontSize: 14,
      keyboardLayout: 'QWERTY',
      locale: 'en-US',
      soundEnabled: true,
      showMetrics: true,
      privacyMode: false,
    };
  }

  /**
   * Clear all local data
   */
  clearAll(): void {
    localStorage.removeItem(this.SESSIONS_KEY);
    localStorage.removeItem(this.RESULTS_KEY);
    localStorage.removeItem(this.PROGRESS_KEY);
    localStorage.removeItem(this.SETTINGS_KEY);
  }

  /**
   * Export all local data
   */
  exportData(): string {
    const data = {
      sessions: this.getSessions(),
      results: this.getResults(),
      progress: this.getProgress(),
      settings: this.getSettings(),
      exportedAt: new Date().toISOString(),
    };
    return JSON.stringify(data, null, 2);
  }

  /**
   * Import data from JSON
   */
  importData(jsonData: string): void {
    try {
      const data = JSON.parse(jsonData);

      if (data.sessions) {
        this.setItem(this.SESSIONS_KEY, data.sessions);
      }
      if (data.results) {
        this.setItem(this.RESULTS_KEY, data.results);
      }
      if (data.progress) {
        this.setItem(this.PROGRESS_KEY, data.progress);
      }
      if (data.settings) {
        this.setItem(this.SETTINGS_KEY, data.settings);
      }
    } catch (error) {
      throw new Error('Invalid data format');
    }
  }

  /**
   * Get storage usage statistics
   */
  getStorageStats(): {
    sessions: number;
    results: number;
    progress: number;
    totalSize: number;
  } {
    const sessions = this.getSessions();
    const results = this.getResults();
    const progress = this.getProgress();

    // Estimate size in bytes
    const totalSize =
      JSON.stringify(sessions).length +
      JSON.stringify(results).length +
      JSON.stringify(progress).length;

    return {
      sessions: sessions.length,
      results: results.length,
      progress: progress.length,
      totalSize,
    };
  }

  /**
   * Check if storage quota is exceeded
   */
  isStorageQuotaExceeded(): boolean {
    try {
      const testKey = `${this.STORAGE_PREFIX}test`;
      localStorage.setItem(testKey, 'test');
      localStorage.removeItem(testKey);
      return false;
    } catch (e) {
      return true;
    }
  }

  /**
   * Generic get item from localStorage
   */
  private getItem<T>(key: string): T | null {
    try {
      const item = localStorage.getItem(key);
      return item ? JSON.parse(item) : null;
    } catch (error) {
      console.error(`Failed to get item ${key}:`, error);
      return null;
    }
  }

  /**
   * Generic set item to localStorage
   */
  private setItem<T>(key: string, value: T): void {
    try {
      localStorage.setItem(key, JSON.stringify(value));
    } catch (error) {
      console.error(`Failed to set item ${key}:`, error);

      // If quota exceeded, try to free up space
      if (this.isStorageQuotaExceeded()) {
        this.cleanupOldData();
        // Retry
        try {
          localStorage.setItem(key, JSON.stringify(value));
        } catch (retryError) {
          throw new Error('Storage quota exceeded');
        }
      }
    }
  }

  /**
   * Cleanup old data to free up space
   */
  private cleanupOldData(): void {
    // Remove oldest sessions
    const sessions = this.getSessions();
    if (sessions.length > this.MAX_SESSIONS / 2) {
      sessions.sort(
        (a, b) =>
          new Date(b.startedAt).getTime() - new Date(a.startedAt).getTime()
      );
      this.setItem(this.SESSIONS_KEY, sessions.slice(0, this.MAX_SESSIONS / 2));
    }

    // Remove oldest results
    const results = this.getResults();
    if (results.length > this.MAX_RESULTS / 2) {
      results.sort(
        (a, b) =>
          new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      );
      this.setItem(this.RESULTS_KEY, results.slice(0, this.MAX_RESULTS / 2));
    }
  }
}

// Export singleton instance
export const localStorageService = new LocalStorageService();
export default localStorageService;
