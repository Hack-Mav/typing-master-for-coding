// Session management service for handling typing session lifecycle

import {
  TypingSession,
  SessionConfig,
  SessionResult,
  SessionHistory,
  EventBatch,
  OfflineQueue,
  SessionEvent as SessionEventType,
} from '../types/session';
import { KeystrokeEvent } from '../types/typing';
import { SessionMetrics } from '../types/metrics';
import { indexedDBManager } from '../utils/indexedDB';
import { metricsCalculator } from './MetricsCalculator';

export class SessionManager {
  private activeSessions: Map<string, TypingSession> = new Map();
  private eventBatches: Map<string, KeystrokeEvent[]> = new Map();
  private offlineQueue: OfflineQueue = {
    batches: [],
    pendingSessions: [],
    syncInProgress: false,
  };
  private metricsCalculator = metricsCalculator;
  private eventListeners: Array<(event: SessionEventType) => void> = [];
  private batchSize = 50; // Events per batch
  private batchTimeout = 5000; // 5 seconds
  private syncRetryDelay = 30000; // 30 seconds

  constructor() {
    this.initializeOfflineHandling();
  }

  // Session lifecycle management
  async createSession(config: SessionConfig): Promise<TypingSession> {
    const sessionId = this.generateSessionId();

    const session: TypingSession = {
      id: sessionId,
      userId: this.getCurrentUserId(),
      config,
      state: {
        status: 'created',
        currentPosition: 0,
        completedCharacters: 0,
        totalCharacters: 0,
        errorCount: 0,
        correctionCount: 0,
        lastActivity: Date.now(),
        isValid: true,
      },
      events: [],
      startedAt: Date.now(),
      totalPauseTime: 0,
      currentText: '',
      targetText: await this.loadTargetText(config),
      progress: {
        percentComplete: 0,
        charactersTyped: 0,
        wordsTyped: 0,
        tokensTyped: 0,
        errorsCount: 0,
        currentSpeed: 0,
        currentAccuracy: 100,
        timeElapsed: 0,
        timeRemaining: config.duration,
      },
      isOffline: !navigator.onLine,
      synced: false,
    };

    // Update total characters based on target text
    session.state.totalCharacters = session.targetText.length;

    this.activeSessions.set(sessionId, session);
    this.eventBatches.set(sessionId, []);

    // Save to IndexedDB for persistence
    await this.persistSession(session);

    this.emitEvent({
      type: 'SESSION_CREATED',
      payload: session,
    });

    return session;
  }

  async startSession(sessionId: string): Promise<void> {
    const session = this.activeSessions.get(sessionId);
    if (!session) {
      throw new Error(`Session ${sessionId} not found`);
    }

    session.state.status = 'active';
    session.state.lastActivity = Date.now();
    session.startedAt = Date.now();

    await this.persistSession(session);

    this.emitEvent({
      type: 'SESSION_STARTED',
      payload: { sessionId },
    });
  }

  async pauseSession(sessionId: string): Promise<void> {
    const session = this.activeSessions.get(sessionId);
    if (!session || session.state.status !== 'active') {
      return;
    }

    const now = Date.now();
    session.state.status = 'paused';
    session.pausedAt = now;

    await this.persistSession(session);

    this.emitEvent({
      type: 'SESSION_PAUSED',
      payload: { sessionId, timestamp: now },
    });
  }

  async resumeSession(sessionId: string): Promise<void> {
    const session = this.activeSessions.get(sessionId);
    if (!session || session.state.status !== 'paused') {
      return;
    }

    const now = Date.now();
    if (session.pausedAt) {
      session.totalPauseTime += now - session.pausedAt;
    }

    session.state.status = 'active';
    session.resumedAt = now;
    session.state.lastActivity = now;

    await this.persistSession(session);

    this.emitEvent({
      type: 'SESSION_RESUMED',
      payload: { sessionId, timestamp: now },
    });
  }

  async finalizeSession(sessionId: string): Promise<SessionResult> {
    const session = this.activeSessions.get(sessionId);
    if (!session) {
      throw new Error(`Session ${sessionId} not found`);
    }

    // Process any remaining events in batch
    await this.flushEventBatch(sessionId);

    session.state.status = 'completed';
    session.endedAt = Date.now();

    // Calculate final metrics
    const metrics = this.metricsCalculator.calculateSessionMetrics(
      session.events,
      session.targetText,
      session.currentText
    );

    const result: SessionResult = {
      sessionId,
      metrics,
      summary: {
        mode: session.config.mode,
        language: session.config.languageId,
        duration: this.getActiveDuration(session),
        completionRate: session.progress.percentComplete,
        finalSpeed: metrics.speed.twpm,
        finalAccuracy: metrics.accuracy.overallAccuracy * 100,
        errorCount: session.state.errorCount,
        improvementAreas: this.identifyImprovementAreas(metrics),
        strengths: this.identifyStrengths(metrics),
      },
    };

    // Mark session as completed and persist
    await this.persistSession(session);
    await this.persistSessionResult(result);

    // Remove from active sessions
    this.activeSessions.delete(sessionId);
    this.eventBatches.delete(sessionId);

    // Queue for sync if offline
    if (session.isOffline || !navigator.onLine) {
      this.offlineQueue.pendingSessions.push(session);
      await this.persistOfflineQueue();
    }

    this.emitEvent({
      type: 'SESSION_COMPLETED',
      payload: { sessionId, result },
    });

    return result;
  }

  async abandonSession(
    sessionId: string,
    reason: string = 'user_abandoned'
  ): Promise<void> {
    const session = this.activeSessions.get(sessionId);
    if (!session) {
      return;
    }

    session.state.status = 'abandoned';
    session.endedAt = Date.now();

    await this.persistSession(session);

    this.activeSessions.delete(sessionId);
    this.eventBatches.delete(sessionId);

    this.emitEvent({
      type: 'SESSION_ABANDONED',
      payload: { sessionId, reason },
    });
  }

  // Event recording and batching
  async recordKeystroke(
    sessionId: string,
    event: KeystrokeEvent
  ): Promise<void> {
    const session = this.activeSessions.get(sessionId);
    if (!session || session.state.status !== 'active') {
      return;
    }

    // Update session state
    session.state.lastActivity = Date.now();
    session.events.push(event);

    // Add to batch
    const batch = this.eventBatches.get(sessionId) || [];
    batch.push(event);
    this.eventBatches.set(sessionId, batch);

    // Update progress
    await this.updateProgress(session, event);

    // Flush batch if it reaches size limit
    if (batch.length >= this.batchSize) {
      await this.flushEventBatch(sessionId);
    }

    this.emitEvent({
      type: 'KEYSTROKE_RECORDED',
      payload: { sessionId, event },
    });
  }

  private async updateProgress(
    session: TypingSession,
    event: KeystrokeEvent
  ): Promise<void> {
    // Update current text based on keystroke
    if (event.action === 'keydown') {
      if (event.key === 'Backspace') {
        if (session.currentText.length > 0) {
          session.currentText = session.currentText.slice(0, -1);
          session.state.correctionCount++;
        }
      } else if (event.key.length === 1) {
        session.currentText += event.key;
        session.state.completedCharacters = session.currentText.length;
      }
    }

    // Calculate progress metrics
    const progress = session.progress;
    progress.charactersTyped = session.currentText.length;
    progress.percentComplete = Math.min(
      (progress.charactersTyped / session.state.totalCharacters) * 100,
      100
    );

    // Calculate real-time metrics
    if (session.events.length > 0) {
      const realtimeMetrics = this.metricsCalculator.calculateRealtimeMetrics(
        session.events,
        session.targetText,
        session.currentText
      );

      progress.currentSpeed = realtimeMetrics.speed.cpm;
      progress.currentAccuracy = realtimeMetrics.accuracy.overallAccuracy * 100;
      progress.errorsCount = session.state.errorCount;
    }

    progress.timeElapsed = this.getActiveDuration(session);

    if (session.config.duration) {
      progress.timeRemaining = Math.max(
        0,
        session.config.duration - progress.timeElapsed
      );
    }

    await this.persistSession(session);

    this.emitEvent({
      type: 'PROGRESS_UPDATED',
      payload: { sessionId: session.id, progress },
    });
  }

  private async flushEventBatch(sessionId: string): Promise<void> {
    const batch = this.eventBatches.get(sessionId);
    if (!batch || batch.length === 0) {
      return;
    }

    const eventBatch: EventBatch = {
      sessionId,
      events: [...batch],
      batchId: this.generateBatchId(),
      timestamp: Date.now(),
      synced: false,
    };

    // Clear the batch
    this.eventBatches.set(sessionId, []);

    // Queue for sync
    this.offlineQueue.batches.push(eventBatch);
    await this.persistOfflineQueue();

    this.emitEvent({
      type: 'BATCH_QUEUED',
      payload: { batch: eventBatch },
    });

    // Attempt sync if online
    if (navigator.onLine && !this.offlineQueue.syncInProgress) {
      this.attemptSync();
    }
  }

  // Session retrieval and history
  async getSession(sessionId: string): Promise<TypingSession | null> {
    // Check active sessions first
    const activeSession = this.activeSessions.get(sessionId);
    if (activeSession) {
      return activeSession;
    }

    // For now, we only return active sessions since stored sessions have different structure
    // In a full implementation, you'd convert SessionData back to TypingSession
    return null;
  }

  async getSessionHistory(
    userId?: string,
    limit: number = 50
  ): Promise<SessionHistory> {
    // This would typically fetch from server, but for offline-first we use IndexedDB
    const sessions = await this.getStoredSessions(userId, limit);

    if (sessions.length === 0) {
      return {
        sessions: [],
        totalSessions: 0,
        totalTime: 0,
        averageSpeed: 0,
        averageAccuracy: 0,
        bestSpeed: 0,
        bestAccuracy: 0,
        streakDays: 0,
        lastPracticeDate: 0,
      };
    }

    const totalTime = sessions.reduce(
      (sum, s) => sum + (s.summary?.duration || 0),
      0
    );
    const avgSpeed =
      sessions.reduce((sum, s) => sum + (s.summary?.finalSpeed || 0), 0) /
      sessions.length;
    const avgAccuracy =
      sessions.reduce((sum, s) => sum + (s.summary?.finalAccuracy || 0), 0) /
      sessions.length;
    const bestSpeed = Math.max(
      ...sessions.map(s => s.summary?.finalSpeed || 0)
    );
    const bestAccuracy = Math.max(
      ...sessions.map(s => s.summary?.finalAccuracy || 0)
    );

    return {
      sessions,
      totalSessions: sessions.length,
      totalTime,
      averageSpeed: avgSpeed,
      averageAccuracy: avgAccuracy,
      bestSpeed,
      bestAccuracy,
      streakDays: this.calculateStreakDays(sessions),
      lastPracticeDate: Math.max(
        ...sessions.map(s => s.metrics?.metadata?.timestamp || 0)
      ),
    };
  }

  // Offline handling and sync
  private async initializeOfflineHandling(): Promise<void> {
    // Load offline queue from IndexedDB
    const storedQueue = await indexedDBManager.getSetting('offlineQueue');
    if (storedQueue) {
      this.offlineQueue = storedQueue;
    }

    // Set up periodic sync attempts
    setInterval(() => {
      if (navigator.onLine && !this.offlineQueue.syncInProgress) {
        this.attemptSync();
      }
    }, this.syncRetryDelay);

    // Listen for online events
    window.addEventListener('online', () => {
      this.attemptSync();
    });
  }

  private async attemptSync(): Promise<void> {
    if (this.offlineQueue.syncInProgress) {
      return;
    }

    this.offlineQueue.syncInProgress = true;
    this.offlineQueue.lastSyncAttempt = Date.now();

    this.emitEvent({
      type: 'SYNC_STARTED',
      payload: { timestamp: Date.now() },
    });

    try {
      let syncedCount = 0;

      // Sync pending sessions
      for (const session of this.offlineQueue.pendingSessions) {
        if (await this.syncSession(session)) {
          syncedCount++;
        }
      }

      // Sync event batches
      for (const batch of this.offlineQueue.batches) {
        if (await this.syncEventBatch(batch)) {
          syncedCount++;
        }
      }

      // Remove synced items
      this.offlineQueue.pendingSessions =
        this.offlineQueue.pendingSessions.filter(s => !s.synced);
      this.offlineQueue.batches = this.offlineQueue.batches.filter(
        b => !b.synced
      );

      await this.persistOfflineQueue();

      this.emitEvent({
        type: 'SYNC_COMPLETED',
        payload: { syncedCount, timestamp: Date.now() },
      });
    } catch (error) {
      console.error('Sync failed:', error);
      this.emitEvent({
        type: 'SYNC_FAILED',
        payload: {
          error: error instanceof Error ? error.message : 'Unknown error',
          timestamp: Date.now(),
        },
      });
    } finally {
      this.offlineQueue.syncInProgress = false;
    }
  }

  private async syncSession(session: TypingSession): Promise<boolean> {
    try {
      // This would make an API call to sync the session
      // For now, we'll just mark it as synced
      session.synced = true;
      await this.persistSession(session);
      return true;
    } catch (error) {
      console.error('Failed to sync session:', error);
      return false;
    }
  }

  private async syncEventBatch(batch: EventBatch): Promise<boolean> {
    try {
      // This would make an API call to sync the event batch
      // For now, we'll just mark it as synced
      batch.synced = true;
      return true;
    } catch (error) {
      console.error('Failed to sync event batch:', error);
      return false;
    }
  }

  // Utility methods
  private generateSessionId(): string {
    return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  private generateBatchId(): string {
    return `batch_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  private getCurrentUserId(): string | undefined {
    // This would get the current user ID from auth context
    return undefined; // Anonymous mode by default
  }

  private async loadTargetText(config: SessionConfig): Promise<string> {
    if (config.lessonId) {
      const lesson = await indexedDBManager.getLesson(config.lessonId);
      return lesson?.content || '';
    }

    if (config.snippetId) {
      const snippet = await indexedDBManager.getSnippet(config.snippetId);
      return snippet?.sourceCode || '';
    }

    // Default fallback text
    return 'console.log("Hello, World!");';
  }

  private getActiveDuration(session: TypingSession): number {
    const now = Date.now();
    const endTime = session.endedAt || now;
    const totalTime = endTime - session.startedAt;
    return Math.max(0, totalTime - session.totalPauseTime);
  }

  private identifyImprovementAreas(metrics: SessionMetrics): string[] {
    const areas: string[] = [];

    if (metrics.accuracy.overallAccuracy < 0.95) {
      areas.push('accuracy');
    }

    if (metrics.speed.twpm < 30) {
      areas.push('speed');
    }

    if (metrics.errors.backspaceRate > 10) {
      areas.push('error_correction');
    }

    if (metrics.timing.keystrokeVariability > 200) {
      areas.push('consistency');
    }

    return areas;
  }

  private identifyStrengths(metrics: SessionMetrics): string[] {
    const strengths: string[] = [];

    if (metrics.accuracy.overallAccuracy >= 0.98) {
      strengths.push('high_accuracy');
    }

    if (metrics.speed.twpm >= 50) {
      strengths.push('fast_typing');
    }

    if (metrics.timing.keystrokeVariability < 100) {
      strengths.push('consistent_rhythm');
    }

    return strengths;
  }

  private calculateStreakDays(sessions: SessionResult[]): number {
    // Calculate consecutive days of practice
    const dates = sessions
      .map(s => new Date(s.metrics?.metadata?.timestamp || 0))
      .map(d => d.toDateString())
      .filter((date, index, arr) => arr.indexOf(date) === index)
      .sort();

    let streak = 0;

    for (let i = dates.length - 1; i >= 0; i--) {
      const date = new Date(dates[i]);
      const expectedDate = new Date();
      expectedDate.setDate(expectedDate.getDate() - streak);

      if (date.toDateString() === expectedDate.toDateString()) {
        streak++;
      } else {
        break;
      }
    }

    return streak;
  }

  private async persistSession(session: TypingSession): Promise<void> {
    await indexedDBManager.saveSession({
      id: session.id,
      userId: session.userId,
      mode: session.config.mode,
      languageId: session.config.languageId,
      lessonId: session.config.lessonId,
      snippetId: session.config.snippetId,
      startedAt: session.startedAt,
      endedAt: session.endedAt,
      events: session.events.map(e => ({
        timestampMs: e.timestamp,
        keyPressed: e.key,
        action: e.action === 'keydown' ? 'down' : 'up',
        cursorPosition: e.cursorPosition,
        errorFlag: false, // This would be determined by validation
        expectedToken: '',
        metadata: { modifiers: e.modifiers },
      })),
      settings: session.config.settings,
      synced: session.synced,
    });
  }

  private async persistSessionResult(result: SessionResult): Promise<void> {
    await indexedDBManager.saveSetting(`result_${result.sessionId}`, result);
  }

  private async persistOfflineQueue(): Promise<void> {
    await indexedDBManager.saveSetting('offlineQueue', this.offlineQueue);
  }

  private async getStoredSessions(
    userId?: string,
    limit: number = 50
  ): Promise<SessionResult[]> {
    // This is a simplified implementation - in practice you'd have proper querying
    const allSettings = await indexedDBManager.getAllSettings();
    const results: SessionResult[] = [];

    if (allSettings) {
      for (const [key, value] of Object.entries(allSettings)) {
        if (key.startsWith('result_') && results.length < limit) {
          results.push(value as SessionResult);
        }
      }
    }

    return results.sort(
      (a, b) =>
        (b.metrics?.metadata?.timestamp || 0) -
        (a.metrics?.metadata?.timestamp || 0)
    );
  }

  // Event system
  addEventListener(listener: (event: SessionEventType) => void): void {
    this.eventListeners.push(listener);
  }

  removeEventListener(listener: (event: SessionEventType) => void): void {
    const index = this.eventListeners.indexOf(listener);
    if (index > -1) {
      this.eventListeners.splice(index, 1);
    }
  }

  private emitEvent(event: SessionEventType): void {
    this.eventListeners.forEach(listener => {
      try {
        listener(event);
      } catch (error) {
        console.error('Error in session event listener:', error);
      }
    });
  }
}

export const sessionManager = new SessionManager();
