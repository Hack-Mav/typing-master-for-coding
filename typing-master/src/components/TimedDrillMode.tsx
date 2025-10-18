import React, { useState, useEffect, useRef, useCallback } from 'react';
import { sessionManager } from '../services/SessionManager';
import { TypingSession, SessionConfig, SessionResult } from '../types/session';
import { KeystrokeEvent } from '../types/typing';
import './TimedDrillMode.css';

interface TimedDrillModeProps {
  languageId: string;
  lessonId?: string;
  snippetId?: string;
  duration: number; // Duration in milliseconds (1/3/5/10 minutes)
  onComplete?: (result: SessionResult) => void;
  onExit?: () => void;
}

interface TimedDrillState {
  session: TypingSession | null;
  targetText: string;
  currentText: string;
  isActive: boolean;
  isPaused: boolean;
  timeRemaining: number;
  showResults: boolean;
  sessionResult: SessionResult | null;
  realTimeMetrics: {
    cpm: number;
    twpm: number;
    accuracy: number;
    errorCount: number;
  };
}

const TimedDrillMode: React.FC<TimedDrillModeProps> = ({
  languageId,
  lessonId,
  snippetId,
  duration,
  onComplete,
  onExit,
}) => {
  const [state, setState] = useState<TimedDrillState>({
    session: null,
    targetText: '',
    currentText: '',
    isActive: false,
    isPaused: false,
    timeRemaining: duration,
    showResults: false,
    sessionResult: null,
    realTimeMetrics: {
      cpm: 0,
      twpm: 0,
      accuracy: 100,
      errorCount: 0,
    },
  });

  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const timerRef = useRef<NodeJS.Timeout | null>(null);
  const startTimeRef = useRef<number>(0);
  const pausedTimeRef = useRef<number>(0);

  // Initialize session
  useEffect(() => {
    const initializeSession = async () => {
      const config: SessionConfig = {
        mode: 'timed-drill',
        languageId,
        lessonId,
        snippetId,
        duration,
        settings: {
          showMetrics: true,
          showTimer: true,
          showProgress: true,
          enableSound: false,
          theme: 'dark',
          fontSize: 16,
          lineHeight: 1.6,
          keyboardLayout: 'qwerty',
          tabWidth: 2,
          useSpaces: true,
        },
      };

      try {
        const session = await sessionManager.createSession(config);
        if (session) {
          setState(prev => ({
            ...prev,
            session,
            targetText: session.targetText || 'console.log("Hello, World!");',
            timeRemaining: duration,
          }));
        }
      } catch (error) {
        console.error('Failed to create timed drill session:', error);
      }
    };

    initializeSession();
  }, [languageId, lessonId, snippetId, duration]);

  const pauseTimer = useCallback(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
    pausedTimeRef.current = Date.now() - startTimeRef.current;
  }, []);

  // Handle time up
  const handleTimeUp = useCallback(async () => {
    if (!state.session || !state.isActive) return;

    try {
      pauseTimer();
      const result = await sessionManager.finalizeSession(state.session.id);
      setState(prev => ({
        ...prev,
        isActive: false,
        showResults: true,
        sessionResult: result,
      }));

      if (onComplete) {
        onComplete(result);
      }
    } catch (error) {
      console.error('Failed to complete timed drill session:', error);
    }
  }, [state.session, state.isActive, pauseTimer, onComplete]);

  // Timer management
  const startTimer = useCallback(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
    }

    startTimeRef.current = Date.now() - pausedTimeRef.current;

    timerRef.current = setInterval(() => {
      const elapsed = Date.now() - startTimeRef.current;
      const remaining = Math.max(0, duration - elapsed);

      setState(prev => ({
        ...prev,
        timeRemaining: remaining,
      }));

      if (remaining <= 0) {
        handleTimeUp();
      }
    }, 100); // Update every 100ms for smooth countdown
  }, [duration, handleTimeUp]);

  const resetTimer = useCallback(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
    startTimeRef.current = 0;
    pausedTimeRef.current = 0;
    setState(prev => ({
      ...prev,
      timeRemaining: duration,
    }));
  }, [duration]);

  // Start session when ready
  useEffect(() => {
    if (state.session && !state.isActive && !state.showResults) {
      const startSession = async () => {
        try {
          await sessionManager.startSession(state.session!.id);
          setState(prev => ({ ...prev, isActive: true }));
          startTimer();

          // Focus the textarea
          if (textareaRef.current) {
            textareaRef.current.focus();
          }
        } catch (error) {
          console.error('Failed to start timed drill session:', error);
        }
      };

      startSession();
    }
  }, [state.session, state.isActive, state.showResults, startTimer]);

  // Handle pause/resume
  const handlePause = useCallback(async () => {
    if (!state.session) return;

    if (state.isActive && !state.isPaused) {
      // Pause
      try {
        await sessionManager.pauseSession(state.session.id);
        pauseTimer();
        setState(prev => ({ ...prev, isPaused: true }));
      } catch (error) {
        console.error('Failed to pause session:', error);
      }
    } else if (state.isPaused) {
      // Resume
      try {
        await sessionManager.resumeSession(state.session.id);
        startTimer();
        setState(prev => ({ ...prev, isPaused: false }));

        // Refocus textarea
        if (textareaRef.current) {
          textareaRef.current.focus();
        }
      } catch (error) {
        console.error('Failed to resume session:', error);
      }
    }
  }, [state.session, state.isActive, state.isPaused, pauseTimer, startTimer]);

  // Handle keystroke events
  const handleKeyDown = useCallback(
    async (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if (!state.session || !state.isActive || state.isPaused) return;

      // Handle special keys
      if (event.key === 'Escape') {
        await handlePause();
        return;
      }

      // Create keystroke event
      const keystrokeEvent: KeystrokeEvent = {
        key: event.key,
        code: event.code,
        timestamp: Date.now(),
        action: 'keydown',
        cursorPosition: state.currentText.length,
        modifiers: {
          ctrl: event.ctrlKey,
          alt: event.altKey,
          shift: event.shiftKey,
          meta: event.metaKey,
        },
      };

      try {
        await sessionManager.recordKeystroke(state.session.id, keystrokeEvent);

        // Update current text based on keystroke
        let newText = state.currentText;

        if (event.key === 'Backspace') {
          newText = newText.slice(0, -1);
        } else if (event.key.length === 1) {
          newText += event.key;
        }

        // Update real-time metrics
        const session = await sessionManager.getSession(state.session.id);
        if (session) {
          const progress = session.progress;
          setState(prev => ({
            ...prev,
            currentText: newText,
            realTimeMetrics: {
              cpm: progress.currentSpeed,
              twpm: Math.round(progress.currentSpeed / 5), // Simplified tWPM calculation
              accuracy: progress.currentAccuracy,
              errorCount: progress.errorsCount,
            },
          }));
        }

        // Check if text is complete
        if (newText === state.targetText) {
          await handleTimeUp();
        }
      } catch (error) {
        console.error('Failed to record keystroke:', error);
      }
    },
    [
      state.session,
      state.isActive,
      state.isPaused,
      state.currentText,
      state.targetText,
      handleTimeUp,
      handlePause,
    ]
  );

  // Handle exit
  const handleExit = useCallback(async () => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
    }

    if (state.session && state.isActive) {
      try {
        await sessionManager.abandonSession(state.session.id, 'user_exit');
      } catch (error) {
        console.error('Failed to abandon session:', error);
      }
    }

    if (onExit) {
      onExit();
    }
  }, [state.session, state.isActive, onExit]);

  // Handle restart
  const handleRestart = useCallback(async () => {
    if (state.session) {
      try {
        await sessionManager.abandonSession(state.session.id, 'user_restart');
      } catch (error) {
        console.error('Failed to abandon session for restart:', error);
      }
    }

    // Reset state and reinitialize
    setState({
      session: null,
      targetText: '',
      currentText: '',
      isActive: false,
      isPaused: false,
      timeRemaining: duration,
      showResults: false,
      sessionResult: null,
      realTimeMetrics: {
        cpm: 0,
        twpm: 0,
        accuracy: 100,
        errorCount: 0,
      },
    });

    resetTimer();
  }, [state.session, duration, resetTimer]);

  // Format time display
  const formatTime = (milliseconds: number): string => {
    const totalSeconds = Math.ceil(milliseconds / 1000);
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  };

  // Calculate progress percentage
  const progressPercentage =
    state.targetText.length > 0
      ? (state.currentText.length / state.targetText.length) * 100
      : 0;

  // Render character with styling
  const renderCharacter = (char: string, index: number) => {
    const isTyped = index < state.currentText.length;
    const isCorrect = isTyped && state.currentText[index] === char;
    const isCurrent = index === state.currentText.length;
    const isError = isTyped && !isCorrect;

    let className = 'drill-char';

    if (isCurrent && state.isActive && !state.isPaused) {
      className += ' drill-char-current';
    } else if (isError) {
      className += ' drill-char-error';
    } else if (isCorrect) {
      className += ' drill-char-correct';
    } else {
      className += ' drill-char-pending';
    }

    return (
      <span key={index} className={className}>
        {char === '\n' ? '↵\n' : char === ' ' ? '·' : char}
      </span>
    );
  };

  if (!state.session) {
    return (
      <div className="drill-loading">
        <div className="drill-loading-spinner"></div>
        <p>Preparing your timed drill...</p>
      </div>
    );
  }

  if (state.showResults && state.sessionResult) {
    return (
      <div className="drill-results">
        <div className="drill-results-container">
          <h2>Drill Complete!</h2>

          <div className="drill-results-stats">
            <div className="drill-stat-card">
              <div className="drill-stat-value">
                {Math.round(state.sessionResult.summary.finalSpeed)}
              </div>
              <div className="drill-stat-label">tWPM</div>
            </div>
            <div className="drill-stat-card">
              <div className="drill-stat-value">
                {Math.round(state.sessionResult.summary.finalAccuracy)}%
              </div>
              <div className="drill-stat-label">Accuracy</div>
            </div>
            <div className="drill-stat-card">
              <div className="drill-stat-value">
                {state.sessionResult.summary.errorCount}
              </div>
              <div className="drill-stat-label">Errors</div>
            </div>
            <div className="drill-stat-card">
              <div className="drill-stat-value">
                {Math.round(state.sessionResult.summary.completionRate)}%
              </div>
              <div className="drill-stat-label">Complete</div>
            </div>
          </div>

          <div className="drill-results-actions">
            <button
              className="drill-btn drill-btn-primary"
              onClick={handleRestart}
            >
              Try Again
            </button>
            <button
              className="drill-btn drill-btn-secondary"
              onClick={handleExit}
            >
              Exit
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="drill-mode">
      {/* Header with timer and controls */}
      <div className="drill-header">
        <div className="drill-timer">
          <div
            className={`drill-timer-display ${state.timeRemaining < 30000 ? 'drill-timer-warning' : ''}`}
          >
            {formatTime(state.timeRemaining)}
          </div>
          <div className="drill-timer-label">Time Remaining</div>
        </div>

        <div className="drill-controls">
          <button
            className="drill-btn drill-btn-icon"
            onClick={handlePause}
            title={state.isPaused ? 'Resume (Esc)' : 'Pause (Esc)'}
          >
            {state.isPaused ? '▶️' : '⏸️'}
          </button>
          <button
            className="drill-btn drill-btn-icon"
            onClick={handleRestart}
            title="Restart"
          >
            🔄
          </button>
          <button
            className="drill-btn drill-btn-icon"
            onClick={handleExit}
            title="Exit"
          >
            ✕
          </button>
        </div>
      </div>

      {/* Real-time HUD */}
      <div className="drill-hud">
        <div className="drill-metric">
          <div className="drill-metric-value">{state.realTimeMetrics.cpm}</div>
          <div className="drill-metric-label">CPM</div>
        </div>
        <div className="drill-metric">
          <div className="drill-metric-value">{state.realTimeMetrics.twpm}</div>
          <div className="drill-metric-label">tWPM</div>
        </div>
        <div className="drill-metric">
          <div className="drill-metric-value">
            {Math.round(state.realTimeMetrics.accuracy)}%
          </div>
          <div className="drill-metric-label">Accuracy</div>
        </div>
        <div className="drill-metric">
          <div className="drill-metric-value">
            {state.realTimeMetrics.errorCount}
          </div>
          <div className="drill-metric-label">Errors</div>
        </div>
      </div>

      {/* Progress bar */}
      <div className="drill-progress-container">
        <div className="drill-progress-bar">
          <div
            className="drill-progress-fill"
            style={{ width: `${Math.min(progressPercentage, 100)}%` }}
          ></div>
        </div>
        <div className="drill-progress-text">
          {Math.round(progressPercentage)}% Complete
        </div>
      </div>

      {/* Target text display */}
      <div className="drill-content">
        <div className="drill-target-text">
          {state.targetText
            .split('')
            .map((char, index) => renderCharacter(char, index))}
        </div>

        {/* Hidden textarea for input capture */}
        <textarea
          ref={textareaRef}
          className="drill-input"
          value={state.currentText}
          onChange={() => {}} // Controlled by keydown handler
          onKeyDown={handleKeyDown}
          disabled={!state.isActive || state.isPaused}
          spellCheck={false}
          autoComplete="off"
          autoCorrect="off"
          autoCapitalize="off"
          style={{ opacity: 0, position: 'absolute', left: '-9999px' }}
        />
      </div>

      {/* Pause overlay */}
      {state.isPaused && (
        <div className="drill-pause-overlay">
          <div className="drill-pause-content">
            <h3>Paused</h3>
            <p>
              Press <kbd>Esc</kbd> or click Resume to continue
            </p>
            <button
              className="drill-btn drill-btn-primary"
              onClick={handlePause}
            >
              Resume
            </button>
          </div>
        </div>
      )}

      {/* Error markers */}
      <div className="drill-error-markers">
        {state.currentText.split('').map((char, index) => {
          const expectedChar = state.targetText[index];
          const isError = char !== expectedChar;

          if (!isError) return null;

          return (
            <div
              key={index}
              className="drill-error-marker"
              style={{ left: `${(index / state.targetText.length) * 100}%` }}
            >
              ⚠️
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default TimedDrillMode;
