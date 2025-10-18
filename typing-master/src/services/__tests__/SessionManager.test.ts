import { SessionManager } from '../SessionManager';
import { SessionConfig, SessionMode } from '../../types/session';
import { KeystrokeEvent } from '../../types/typing';

// Mock IndexedDB
jest.mock('../../utils/indexedDB', () => ({
  indexedDBManager: {
    saveSession: jest.fn().mockResolvedValue(undefined),
    getSession: jest.fn().mockResolvedValue(null),
    getUnsyncedSessions: jest.fn().mockResolvedValue([]),
    markSessionSynced: jest.fn().mockResolvedValue(undefined),
    saveLesson: jest.fn().mockResolvedValue(undefined),
    getLesson: jest.fn().mockResolvedValue(null),
    getSnippet: jest.fn().mockResolvedValue(null),
    saveSetting: jest.fn().mockResolvedValue(undefined),
    getSetting: jest.fn().mockResolvedValue(null),
    getAllSettings: jest.fn().mockResolvedValue({}),
  },
}));

describe('SessionManager', () => {
  let sessionManager: SessionManager;

  beforeEach(() => {
    sessionManager = new SessionManager();
    jest.clearAllMocks();
  });

  describe('Session Lifecycle', () => {
    test('should create a new session', async () => {
      const config: SessionConfig = {
        mode: 'zen' as SessionMode,
        languageId: 'javascript',
        settings: {
          showMetrics: false,
          showTimer: false,
          showProgress: false,
          enableSound: false,
          theme: 'dark',
          fontSize: 14,
          lineHeight: 1.5,
          keyboardLayout: 'qwerty',
          tabWidth: 2,
          useSpaces: true,
        },
      };

      const session = await sessionManager.createSession(config);

      expect(session).toBeDefined();
      expect(session.id).toBeTruthy();
      expect(session.config).toEqual(config);
      expect(session.state.status).toBe('created');
      expect(session.events).toEqual([]);
      expect(session.synced).toBe(false);
    });

    test('should start a session', async () => {
      const config: SessionConfig = {
        mode: 'timed-drill' as SessionMode,
        languageId: 'python',
        duration: 300000, // 5 minutes
        settings: {
          showMetrics: true,
          showTimer: true,
          showProgress: true,
          enableSound: false,
          theme: 'light',
          fontSize: 16,
          lineHeight: 1.4,
          keyboardLayout: 'qwerty',
          tabWidth: 4,
          useSpaces: true,
        },
      };

      const session = await sessionManager.createSession(config);
      await sessionManager.startSession(session.id);

      const retrievedSession = await sessionManager.getSession(session.id);
      expect(retrievedSession?.state.status).toBe('active');
    });

    test('should pause and resume a session', async () => {
      const config: SessionConfig = {
        mode: 'tutorial' as SessionMode,
        languageId: 'javascript',
        lessonId: 'lesson-1',
        settings: {
          showMetrics: true,
          showTimer: false,
          showProgress: true,
          enableSound: false,
          theme: 'solarized',
          fontSize: 14,
          lineHeight: 1.5,
          keyboardLayout: 'dvorak',
          tabWidth: 2,
          useSpaces: true,
        },
      };

      const session = await sessionManager.createSession(config);
      await sessionManager.startSession(session.id);

      const pauseTime = Date.now();
      await sessionManager.pauseSession(session.id);

      let retrievedSession = await sessionManager.getSession(session.id);
      expect(retrievedSession?.state.status).toBe('paused');
      expect(retrievedSession?.pausedAt).toBeGreaterThanOrEqual(pauseTime);

      await sessionManager.resumeSession(session.id);

      retrievedSession = await sessionManager.getSession(session.id);
      expect(retrievedSession?.state.status).toBe('active');
      expect(retrievedSession?.resumedAt).toBeDefined();
    });

    test('should finalize a session and calculate metrics', async () => {
      const config: SessionConfig = {
        mode: 'accuracy' as SessionMode,
        languageId: 'rust',
        settings: {
          showMetrics: true,
          showTimer: false,
          showProgress: true,
          enableSound: false,
          theme: 'dark',
          fontSize: 14,
          lineHeight: 1.6,
          keyboardLayout: 'colemak',
          tabWidth: 4,
          useSpaces: true,
        },
      };

      const session = await sessionManager.createSession(config);
      await sessionManager.startSession(session.id);

      // Simulate some typing events
      const keystrokeEvents: KeystrokeEvent[] = [
        {
          key: 'h',
          code: 'KeyH',
          timestamp: Date.now(),
          action: 'keydown',
          cursorPosition: 0,
          modifiers: { ctrl: false, alt: false, shift: false, meta: false },
        },
        {
          key: 'e',
          code: 'KeyE',
          timestamp: Date.now() + 100,
          action: 'keydown',
          cursorPosition: 1,
          modifiers: { ctrl: false, alt: false, shift: false, meta: false },
        },
      ];

      for (const event of keystrokeEvents) {
        await sessionManager.recordKeystroke(session.id, event);
      }

      const result = await sessionManager.finalizeSession(session.id);

      expect(result).toBeDefined();
      expect(result.sessionId).toBe(session.id);
      expect(result.metrics).toBeDefined();
      expect(result.summary).toBeDefined();
      expect(result.summary.mode).toBe('accuracy');
      expect(result.summary.language).toBe('rust');
    });

    test('should abandon a session', async () => {
      const config: SessionConfig = {
        mode: 'custom-snippet' as SessionMode,
        languageId: 'cpp',
        snippetId: 'snippet-1',
        settings: {
          showMetrics: false,
          showTimer: false,
          showProgress: false,
          enableSound: false,
          theme: 'light',
          fontSize: 12,
          lineHeight: 1.3,
          keyboardLayout: 'azerty',
          tabWidth: 3,
          useSpaces: false,
        },
      };

      const session = await sessionManager.createSession(config);
      await sessionManager.startSession(session.id);

      await sessionManager.abandonSession(session.id, 'user_quit');

      const retrievedSession = await sessionManager.getSession(session.id);
      expect(retrievedSession).toBeNull(); // Should be removed from active sessions
    });
  });

  describe('Event Recording', () => {
    test('should record keystroke events', async () => {
      const config: SessionConfig = {
        mode: 'zen' as SessionMode,
        languageId: 'yaml',
        settings: {
          showMetrics: false,
          showTimer: false,
          showProgress: false,
          enableSound: false,
          theme: 'dark',
          fontSize: 14,
          lineHeight: 1.5,
          keyboardLayout: 'qwerty',
          tabWidth: 2,
          useSpaces: true,
        },
      };

      const session = await sessionManager.createSession(config);
      await sessionManager.startSession(session.id);

      const keystrokeEvent: KeystrokeEvent = {
        key: 'a',
        code: 'KeyA',
        timestamp: Date.now(),
        action: 'keydown',
        cursorPosition: 0,
        modifiers: { ctrl: false, alt: false, shift: false, meta: false },
      };

      await sessionManager.recordKeystroke(session.id, keystrokeEvent);

      const retrievedSession = await sessionManager.getSession(session.id);
      expect(retrievedSession?.events).toHaveLength(1);
      expect(retrievedSession?.events[0]).toEqual(keystrokeEvent);
    });

    test('should update progress during typing', async () => {
      const config: SessionConfig = {
        mode: 'timed-drill' as SessionMode,
        languageId: 'javascript',
        duration: 60000, // 1 minute
        settings: {
          showMetrics: true,
          showTimer: true,
          showProgress: true,
          enableSound: false,
          theme: 'light',
          fontSize: 14,
          lineHeight: 1.5,
          keyboardLayout: 'qwerty',
          tabWidth: 2,
          useSpaces: true,
        },
      };

      const session = await sessionManager.createSession(config);
      await sessionManager.startSession(session.id);

      // Record multiple keystrokes
      const keys = ['c', 'o', 'n', 's', 'o', 'l', 'e'];
      for (let i = 0; i < keys.length; i++) {
        const event: KeystrokeEvent = {
          key: keys[i],
          code: `Key${keys[i].toUpperCase()}`,
          timestamp: Date.now() + i * 100,
          action: 'keydown',
          cursorPosition: i,
          modifiers: { ctrl: false, alt: false, shift: false, meta: false },
        };
        await sessionManager.recordKeystroke(session.id, event);
      }

      const retrievedSession = await sessionManager.getSession(session.id);
      expect(retrievedSession?.progress.charactersTyped).toBe(keys.length);
      expect(retrievedSession?.progress.currentSpeed).toBeGreaterThan(0);
    });
  });

  describe('Event Listeners', () => {
    test('should emit session events', async () => {
      const eventListener = jest.fn();
      sessionManager.addEventListener(eventListener);

      const config: SessionConfig = {
        mode: 'assessment' as SessionMode,
        languageId: 'python',
        settings: {
          showMetrics: true,
          showTimer: true,
          showProgress: true,
          enableSound: false,
          theme: 'dark',
          fontSize: 14,
          lineHeight: 1.5,
          keyboardLayout: 'qwerty',
          tabWidth: 4,
          useSpaces: true,
        },
      };

      await sessionManager.createSession(config);

      expect(eventListener).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'SESSION_CREATED',
          payload: expect.any(Object),
        })
      );

      sessionManager.removeEventListener(eventListener);
    });
  });

  describe('Session History', () => {
    test('should retrieve empty session history for new user', async () => {
      const history = await sessionManager.getSessionHistory();

      expect(history.sessions).toEqual([]);
      expect(history.totalSessions).toBe(0);
      expect(history.totalTime).toBe(0);
      expect(history.averageSpeed).toBe(0);
      expect(history.averageAccuracy).toBe(0);
      expect(history.bestSpeed).toBe(0);
      expect(history.bestAccuracy).toBe(0);
      expect(history.streakDays).toBe(0);
      expect(history.lastPracticeDate).toBe(0);
    });
  });
});
