import React, { useState, useEffect, useRef, useCallback } from 'react';
import MonacoTypingInterface from './MonacoTypingInterface';
import { sessionManager } from '../services/SessionManager';
import { TypingSession, SessionConfig, SessionResult } from '../types/session';
import './AccuracyMode.css';

interface AccuracyModeProps {
  languageId: string;
  lessonId?: string;
  snippetId?: string;
  onComplete?: (result: SessionResult) => void;
  onExit?: () => void;
}

interface AccuracyState {
  session: TypingSession | null;
  targetText: string;
  currentText: string;
  isActive: boolean;
  showResults: boolean;
  sessionResult: SessionResult | null;
  realTimeMetrics: {
    accuracy: number;
    errorCount: number;
    correctionCount: number;
    syntaxErrors: number;
  };
  errorPenalty: number;
}

const AccuracyMode: React.FC<AccuracyModeProps> = ({
  languageId,
  lessonId,
  snippetId,
  onComplete,
  onExit,
}) => {
  const [state, setState] = useState<AccuracyState>({
    session: null,
    targetText: '',
    currentText: '',
    isActive: false,
    showResults: false,
    sessionResult: null,
    realTimeMetrics: {
      accuracy: 100,
      errorCount: 0,
      correctionCount: 0,
      syntaxErrors: 0,
    },
    errorPenalty: 0,
  });

  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Initialize session
  useEffect(() => {
    const initializeSession = async () => {
      const config: SessionConfig = {
        mode: 'accuracy',
        languageId,
        lessonId,
        snippetId,
        settings: {
          showMetrics: true,
          showTimer: false,
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
          }));
        }
      } catch (error) {
        console.error('Failed to create accuracy session:', error);
      }
    };

    initializeSession();
  }, [languageId, lessonId, snippetId]);

  // Start session when ready
  useEffect(() => {
    if (state.session && !state.isActive) {
      const startSession = async () => {
        try {
          await sessionManager.startSession(state.session!.id);
          setState(prev => ({ ...prev, isActive: true }));

          if (textareaRef.current) {
            textareaRef.current.focus();
          }
        } catch (error) {
          console.error('Failed to start accuracy session:', error);
        }
      };

      startSession();
    }
  }, [state.session, state.isActive]);

  // Handle session completion
  const handleComplete = useCallback(
    async (result?: SessionResult) => {
      if (!state.session) return;

      try {
        let finalResult: SessionResult;

        if (result) {
          // We have a result from session manager
          finalResult = result;
        } else {
          // Monaco called us without a result, finalize the session
          finalResult = await sessionManager.finalizeSession(state.session.id);
        }

        // Apply accuracy-specific scoring adjustments
        const accuracyScore = calculateAccuracyScore(
          finalResult,
          state.errorPenalty
        );
        const adjustedResult = {
          ...finalResult,
          metrics: {
            ...finalResult.metrics,
            composite: {
              ...finalResult.metrics.composite,
              score: accuracyScore,
            },
          },
        };

        setState(prev => ({
          ...prev,
          isActive: false,
          showResults: true,
          sessionResult: adjustedResult,
        }));

        if (onComplete) {
          onComplete(adjustedResult);
        }
      } catch (error) {
        console.error('Failed to complete accuracy session:', error);
      }
    },
    [state.session, state.errorPenalty, onComplete]
  );

  // Calculate accuracy-focused score with heavy error penalties
  const calculateAccuracyScore = (
    result: SessionResult,
    errorPenalty: number
  ): number => {
    const baseScore = result.metrics.composite.score;
    const accuracyBonus = result.metrics.accuracy.overallAccuracy * 500;
    const syntaxBonus = result.metrics.accuracy.syntaxAccuracy * 300;
    const errorMalus = errorPenalty * 50; // Heavy penalty for errors

    return Math.max(0, baseScore + accuracyBonus + syntaxBonus - errorMalus);
  };

  if (state.showResults && state.sessionResult) {
    return (
      <div className="accuracy-results">
        <div className="results-header">
          <h2>Accuracy Mode Results</h2>
          <div className="accuracy-grade">
            {state.sessionResult.metrics.accuracy.overallAccuracy >= 0.98 ? (
              <span className="grade-excellent">A+</span>
            ) : state.sessionResult.metrics.accuracy.overallAccuracy >= 0.95 ? (
              <span className="grade-good">A</span>
            ) : state.sessionResult.metrics.accuracy.overallAccuracy >= 0.9 ? (
              <span className="grade-ok">B</span>
            ) : (
              <span className="grade-poor">C</span>
            )}
          </div>
        </div>

        <div className="results-metrics">
          <div className="metric-card">
            <span className="metric-label">Overall Accuracy</span>
            <span className="metric-value">
              {(
                state.sessionResult.metrics.accuracy.overallAccuracy * 100
              ).toFixed(2)}
              %
            </span>
          </div>
          <div className="metric-card">
            <span className="metric-label">Syntax Accuracy</span>
            <span className="metric-value">
              {(
                state.sessionResult.metrics.accuracy.syntaxAccuracy * 100
              ).toFixed(2)}
              %
            </span>
          </div>
          <div className="metric-card">
            <span className="metric-label">Error Count</span>
            <span className="metric-value">
              {state.sessionResult.metrics.errors.totalErrors}
            </span>
          </div>
          <div className="metric-card">
            <span className="metric-label">Corrections</span>
            <span className="metric-value">
              {state.realTimeMetrics.correctionCount}
            </span>
          </div>
          <div className="metric-card">
            <span className="metric-label">Accuracy Score</span>
            <span className="metric-value">
              {state.sessionResult.metrics.composite.score}
            </span>
          </div>
        </div>

        <div className="results-actions">
          <button
            onClick={() => window.location.reload()}
            className="btn-retry"
          >
            Try Again
          </button>
          <button onClick={onExit} className="btn-exit">
            Exit
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="accuracy-mode">
      <div className="accuracy-hud">
        <div className="accuracy-info">
          <h3>Accuracy Mode</h3>
          <p>Focus on precision - every error counts!</p>
        </div>

        <div className="accuracy-metrics">
          <div className="metric">
            <span className="metric-label">Accuracy</span>
            <span className="metric-value">
              {state.realTimeMetrics.accuracy.toFixed(1)}%
            </span>
          </div>
          <div className="metric">
            <span className="metric-label">Errors</span>
            <span className="metric-value error-count">
              {state.realTimeMetrics.errorCount}
            </span>
          </div>
          <div className="metric">
            <span className="metric-label">Corrections</span>
            <span className="metric-value">
              {state.realTimeMetrics.correctionCount}
            </span>
          </div>
          <div className="metric">
            <span className="metric-label">Penalty</span>
            <span className="metric-value penalty">-{state.errorPenalty}</span>
          </div>
        </div>

        <button onClick={onExit} className="btn-exit-small">
          Exit
        </button>
      </div>

      <MonacoTypingInterface
        targetText={state.targetText}
        languageId={languageId}
        currentText={state.currentText}
        onTextChange={text =>
          setState(prev => ({ ...prev, currentText: text }))
        }
        onComplete={() => {
          // MonacoTypingInterface calls onComplete when typing is complete
          // We need to get the final result and pass it to handleComplete
          handleComplete();
        }}
        theme="dark"
        fontSize={16}
        lineHeight={1.6}
        keyboardLayout="qwerty"
        className="accuracy-monaco-editor"
      />
    </div>
  );
};

export default AccuracyMode;
