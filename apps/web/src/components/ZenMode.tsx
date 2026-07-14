import React, { useState, useEffect, useRef, useCallback } from 'react';
import MonacoTypingInterface from './MonacoTypingInterface';
import { sessionManager } from '../services/SessionManager';
import { TypingSession, SessionConfig, SessionResult } from '../types/session';
import { KeystrokeEvent } from '../types/typing';
import './ZenMode.css';

interface ZenModeProps {
  languageId: string;
  lessonId?: string;
  snippetId?: string;
  onComplete?: (result: SessionResult) => void;
  onExit?: () => void;
}

interface ZenModeState {
  session: TypingSession | null;
  targetText: string;
  currentText: string;
  isActive: boolean;
  showSummary: boolean;
  sessionResult: SessionResult | null;
}

const ZenMode: React.FC<ZenModeProps> = ({
  languageId,
  lessonId,
  snippetId,
  onComplete,
  onExit,
}) => {
  const [state, setState] = useState<ZenModeState>({
    session: null,
    targetText: '',
    currentText: '',
    isActive: false,
    showSummary: false,
    sessionResult: null,
  });

  const [theme, setTheme] = useState<'light' | 'dark' | 'solarized'>('dark');

  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Initialize session
  useEffect(() => {
    const initializeSession = async () => {
      const config: SessionConfig = {
        mode: 'zen',
        languageId,
        lessonId,
        snippetId,
        settings: {
          showMetrics: false,
          showTimer: false,
          showProgress: false,
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
            targetText: session.targetText || '',
          }));
        }
      } catch (error) {
        console.error('Failed to create zen session:', error);
      }
    };

    initializeSession();
  }, [languageId, lessonId, snippetId]);

  // Start session when component mounts and session is ready
  useEffect(() => {
    if (state.session && !state.isActive) {
      const startSession = async () => {
        try {
          await sessionManager.startSession(state.session!.id);
          setState(prev => ({ ...prev, isActive: true }));

          // Focus the textarea
          if (textareaRef.current) {
            textareaRef.current.focus();
          }
        } catch (error) {
          console.error('Failed to start zen session:', error);
        }
      };

      startSession();
    }
  }, [state.session, state.isActive]);

  // Handle session completion
  const handleComplete = useCallback(async () => {
    if (!state.session) return;

    try {
      const result = await sessionManager.finalizeSession(state.session.id);
      setState(prev => ({
        ...prev,
        isActive: false,
        sessionResult: result,
      }));

      // In zen mode, we don't show summary by default
      // Just call onComplete if provided
      if (onComplete) {
        onComplete(result);
      }
    } catch (error) {
      console.error('Failed to complete zen session:', error);
    }
  }, [state.session, onComplete]);

  // Handle exit
  const handleExit = useCallback(async () => {
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

  // Toggle summary display
  const toggleSummary = useCallback(() => {
    setState(prev => ({ ...prev, showSummary: !prev.showSummary }));
  }, []);

  // Theme toggle handler
  const handleThemeToggle = useCallback(() => {
    setTheme(prev => {
      const themes: ('light' | 'dark' | 'solarized')[] = [
        'light',
        'dark',
        'solarized',
      ];
      const currentIndex = themes.indexOf(prev);
      return themes[(currentIndex + 1) % themes.length];
    });
  }, []);

  // Render target text with Monaco Editor
  const renderTargetText = () => {
    return (
      <MonacoTypingInterface
        targetText={state.targetText}
        languageId={languageId}
        currentText={state.currentText}
        onTextChange={text =>
          setState(prev => ({ ...prev, currentText: text }))
        }
        onComplete={() => handleComplete()}
        theme={theme}
        fontSize={16}
        lineHeight={1.6}
        keyboardLayout="qwerty"
        className="zen-monaco-editor"
        enableFontControls={true}
        enableLineHeightControls={true}
      />
    );
  };

  if (!state.session) {
    return (
      <div className="zen-loading">
        <div className="zen-loading-spinner"></div>
        <p>Preparing your zen session...</p>
      </div>
    );
  }

  return (
    <div className="zen-mode">
      {/* Ambient background */}
      <div className="zen-ambient-bg"></div>

      {/* Main content */}
      <div className="zen-content">
        {/* Header with minimal controls */}
        <div className="zen-header">
          <button
            className="zen-exit-btn"
            onClick={handleExit}
            title="Exit (Esc)"
          >
            ✕
          </button>

          {state.sessionResult && (
            <button
              className="zen-summary-btn"
              onClick={toggleSummary}
              title="Toggle Summary"
            >
              📊
            </button>
          )}

          {/* Theme toggle button */}
          <button
            className="zen-theme-btn"
            onClick={handleThemeToggle}
            title="Toggle Theme"
          >
            🎨
          </button>
        </div>

        {/* Target text with Monaco Editor integration */}
        {renderTargetText()}

        {/* Hidden input for keyboard capture (fallback) */}
        <input
          type="text"
          className="zen-hidden-input"
          value={state.currentText}
          onChange={e => {
            const newText = e.target.value;
            setState(prev => ({ ...prev, currentText: newText }));

            // Check if session is complete
            if (newText === state.targetText) {
              handleComplete();
            }
          }}
          onKeyDown={async e => {
            if (!state.session || !state.isActive) return;

            // Handle special keys
            if (e.key === 'Escape') {
              await handleExit();
              return;
            }

            // Create keystroke event for analytics
            const keystrokeEvent: KeystrokeEvent = {
              key: e.key,
              code: e.code,
              timestamp: Date.now(),
              action: 'keydown',
              cursorPosition: state.currentText.length,
              modifiers: {
                ctrl: e.ctrlKey,
                alt: e.altKey,
                shift: e.shiftKey,
                meta: e.metaKey,
              },
            };

            try {
              await sessionManager.recordKeystroke(
                state.session.id,
                keystrokeEvent
              );
            } catch (error) {
              console.error('Failed to record keystroke:', error);
            }
          }}
          autoFocus
          style={{ opacity: 0, position: 'absolute', left: '-9999px' }}
        />

        {/* Optional summary overlay */}
        {state.showSummary && state.sessionResult && (
          <div className="zen-summary-overlay">
            <div className="zen-summary">
              <h3>Session Complete</h3>
              <div className="zen-summary-stats">
                <div className="zen-stat">
                  <span className="zen-stat-label">Duration</span>
                  <span className="zen-stat-value">
                    {Math.round(state.sessionResult.summary.duration / 1000)}s
                  </span>
                </div>
                <div className="zen-stat">
                  <span className="zen-stat-label">Completion</span>
                  <span className="zen-stat-value">
                    {Math.round(state.sessionResult.summary.completionRate)}%
                  </span>
                </div>
                <div className="zen-stat">
                  <span className="zen-stat-label">Errors</span>
                  <span className="zen-stat-value">
                    {state.sessionResult.summary.errorCount}
                  </span>
                </div>
              </div>
              <div className="zen-summary-actions">
                <button onClick={toggleSummary}>Close</button>
                <button onClick={handleExit}>Exit</button>
              </div>
            </div>
          </div>
        )}

        {/* Breathing indicator for focus */}
        <div className="zen-breathing-indicator">
          <div className="zen-breath-circle"></div>
        </div>
      </div>

      {/* Instructions */}
      <div className="zen-instructions">
        <p>
          Type the code above. Press <kbd>Esc</kbd> to exit.
        </p>
        {state.sessionResult && (
          <p>
            Session complete! Press <kbd>📊</kbd> for summary.
          </p>
        )}
      </div>
    </div>
  );
};

export default ZenMode;
