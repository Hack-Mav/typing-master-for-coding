// Session management related types

import { KeystrokeEvent } from './typing';
import { SessionMetrics } from './metrics';

export interface SessionConfig {
  mode: SessionMode;
  languageId: string;
  lessonId?: string;
  snippetId?: string;
  duration?: number; // For timed modes (in milliseconds)
  settings: SessionSettings;
}

export type SessionMode =
  | 'tutorial'
  | 'timed-drill'
  | 'accuracy'
  | 'zen'
  | 'custom-snippet'
  | 'assessment';

export interface SessionSettings {
  showMetrics: boolean;
  showTimer: boolean;
  showProgress: boolean;
  enableSound: boolean;
  theme: 'light' | 'dark' | 'solarized';
  fontSize: number;
  lineHeight: number;
  keyboardLayout: string;
  tabWidth: number;
  useSpaces: boolean;
}

export interface TypingSession {
  id: string;
  userId?: string;
  config: SessionConfig;
  state: SessionState;
  events: KeystrokeEvent[];
  startedAt: number;
  endedAt?: number;
  pausedAt?: number;
  resumedAt?: number;
  totalPauseTime: number;
  currentText: string;
  targetText: string;
  progress: SessionProgress;
  isOffline: boolean;
  synced: boolean;
}

export interface SessionState {
  status: 'created' | 'active' | 'paused' | 'completed' | 'abandoned';
  currentPosition: number;
  completedCharacters: number;
  totalCharacters: number;
  errorCount: number;
  correctionCount: number;
  lastActivity: number;
  isValid: boolean;
}

export interface SessionProgress {
  percentComplete: number;
  charactersTyped: number;
  wordsTyped: number;
  tokensTyped: number;
  errorsCount: number;
  currentSpeed: number; // Real-time CPM
  currentAccuracy: number; // Real-time accuracy
  timeElapsed: number; // Active typing time in ms
  timeRemaining?: number; // For timed modes
}

export interface SessionResult {
  sessionId: string;
  metrics: SessionMetrics;
  summary: SessionSummary;
  achievements?: Achievement[];
  recommendations?: string[];
}

export interface SessionSummary {
  mode: SessionMode;
  language: string;
  duration: number;
  completionRate: number;
  finalSpeed: number;
  finalAccuracy: number;
  errorCount: number;
  improvementAreas: string[];
  strengths: string[];
}

export interface Achievement {
  id: string;
  title: string;
  description: string;
  type: 'speed' | 'accuracy' | 'consistency' | 'milestone';
  earnedAt: number;
  value?: number;
}

export interface SessionHistory {
  sessions: SessionResult[];
  totalSessions: number;
  totalTime: number; // Total practice time in ms
  averageSpeed: number;
  averageAccuracy: number;
  bestSpeed: number;
  bestAccuracy: number;
  streakDays: number;
  lastPracticeDate: number;
}

export interface EventBatch {
  sessionId: string;
  events: KeystrokeEvent[];
  batchId: string;
  timestamp: number;
  synced: boolean;
}

export interface OfflineQueue {
  batches: EventBatch[];
  pendingSessions: TypingSession[];
  lastSyncAttempt?: number;
  syncInProgress: boolean;
}

// Session lifecycle events
export type SessionEvent =
  | { type: 'SESSION_CREATED'; payload: TypingSession }
  | { type: 'SESSION_STARTED'; payload: { sessionId: string } }
  | {
      type: 'SESSION_PAUSED';
      payload: { sessionId: string; timestamp: number };
    }
  | {
      type: 'SESSION_RESUMED';
      payload: { sessionId: string; timestamp: number };
    }
  | {
      type: 'SESSION_COMPLETED';
      payload: { sessionId: string; result: SessionResult };
    }
  | {
      type: 'SESSION_ABANDONED';
      payload: { sessionId: string; reason: string };
    }
  | {
      type: 'KEYSTROKE_RECORDED';
      payload: { sessionId: string; event: KeystrokeEvent };
    }
  | {
      type: 'PROGRESS_UPDATED';
      payload: { sessionId: string; progress: SessionProgress };
    }
  | { type: 'BATCH_QUEUED'; payload: { batch: EventBatch } }
  | { type: 'SYNC_STARTED'; payload: { timestamp: number } }
  | {
      type: 'SYNC_COMPLETED';
      payload: { syncedCount: number; timestamp: number };
    }
  | { type: 'SYNC_FAILED'; payload: { error: string; timestamp: number } }
  | { 
      type: 'AUTHENTICATION_CHOICE_REQUIRED'; 
      payload: { sessionId: string; options: string[] } 
    }
  | { 
      type: 'AUTHENTICATION_REQUIRED'; 
      payload: { sessionId: string; reason: string } 
    };
