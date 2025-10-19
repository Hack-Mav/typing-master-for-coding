import React, { useState, useEffect, useCallback, useRef } from 'react';
import { assessmentService } from '../services/AssessmentService';
import { sessionManager } from '../services/SessionManager';
import { typingValidationEngine } from '../services/TypingValidationEngine';
import { metricsCalculator } from '../services/MetricsCalculator';

// Import StructuralAnalyzer types and create a mock implementation
import { StructuralConformityScore } from '../types/assessment';

// Mock structural analyzer for now
const structuralAnalyzer = {
  analyzeStructuralConformity: async (
    expectedCode: string,
    actualCode: string,
    language: string
  ): Promise<StructuralConformityScore> => {
    // Simple mock implementation
    const similarity = expectedCode === actualCode ? 1.0 : 0.8;
    return {
      astSimilarity: similarity,
      tokenSequenceAccuracy: similarity,
      syntaxValidationScore: similarity,
      structuralPenalties: [],
      overallConformity: similarity,
    };
  },
};
import { 
  AssessmentSession, 
  AssessmentBlueprint, 
  AssessmentStatus,
  AssessmentResult 
} from '../types/assessment';
import { KeystrokeEvent } from '../types/typing';
import { Language } from '../types/parser';
import './AssessmentMode.css';

interface AssessmentModeProps {
  language: Language;
  onComplete: (result: AssessmentResult) => void;
  onExit: () => void;
}

export const AssessmentMode: React.FC<AssessmentModeProps> = ({
  language,
  onComplete,
  onExit,
}) => {
  const [assessmentSession, setAssessmentSession] = useState<AssessmentSession | null>(null);
  const [blueprint, setBlueprint] = useState<AssessmentBlueprint | null>(null);
  const [currentSnippet, setCurrentSnippet] = useState<string>('');
  const [userInput, setUserInput] = useState<string>('');
  const [status, setStatus] = useState<AssessmentStatus>('not_started');
  const [timeRemaining, setTimeRemaining] = useState<number>(0);
  const [currentSnippetIndex, setCurrentSnippetIndex] = useState<number>(0);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const timerRef = useRef<NodeJS.Timeout | null>(null);
  const keystrokeEvents = useRef<KeystrokeEvent[]>([]);
  const snippetStartTime = useRef<number>(0);

  // Initialize assessment
  useEffect(() => {
    const initializeAssessment = async () => {
      try {
        setIsLoading(true);
        
        // Get assessment blueprint for the language
        const availableBlueprints = await assessmentService.getAssessmentBlueprints(language);
        if (availableBlueprints.length === 0) {
          throw new Error(`No assessments available for ${language}`);
        }
        
        // Select appropriate blueprint (for now, use the first one)
        const selectedBlueprint = availableBlueprints[0];
        setBlueprint(selectedBlueprint);
        
        // Create assessment session
        const session = await assessmentService.createAssessmentSession(selectedBlueprint.id);
        setAssessmentSession(session);
        
        // Load first snippet
        await loadSnippet(session, 0);
        
        setStatus('not_started');
        setIsLoading(false);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to initialize assessment');
        setIsLoading(false);
      }
    };

    initializeAssessment();
  }, [language]);

  // Load snippet for current index
  const loadSnippet = async (session: AssessmentSession, index: number) => {
    try {
      const snippet = await assessmentService.getAssessmentSnippet(
        session.blueprintId,
        session.snippetResults[index]?.snippetId || 
        (await assessmentService.getAssessmentBlueprint(session.blueprintId)).snippetIds[index]
      );
      setCurrentSnippet(snippet.sourceCode);
      setCurrentSnippetIndex(index);
      
      // Initialize typing validation engine
      typingValidationEngine.initialize(snippet.sourceCode, language);
    } catch (err) {
      setError('Failed to load assessment snippet');
    }
  };

  // Start assessment
  const startAssessment = useCallback(() => {
    if (!assessmentSession || !blueprint) return;
    
    setStatus('in_progress');
    setUserInput('');
    keystrokeEvents.current = [];
    snippetStartTime.current = Date.now();
    
    // Start timer if there's a time limit
    if (blueprint.passingCriteria.timeLimit) {
      setTimeRemaining(blueprint.passingCriteria.timeLimit * 60 * 1000); // Convert to milliseconds
      
      timerRef.current = setInterval(() => {
        setTimeRemaining(prev => {
          if (prev <= 1000) {
            handleTimeExpired();
            return 0;
          }
          return prev - 1000;
        });
      }, 1000);
    }
    
    // Focus input
    inputRef.current?.focus();
  }, [assessmentSession, blueprint]);

  // Handle time expiration
  const handleTimeExpired = useCallback(async () => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
    }
    
    setStatus('expired');
    
    if (assessmentSession) {
      const result = await assessmentService.finalizeAssessment(assessmentSession.id, true);
      onComplete(result);
    }
  }, [assessmentSession, onComplete]);

  // Handle keystroke events
  const handleKeyDown = useCallback((event: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (status !== 'in_progress') return;

    // Handle escape key for pause
    if (event.key === 'Escape') {
      event.preventDefault();
      pauseAssessment();
      return;
    }

    // Record keystroke event
    const keystrokeEvent: KeystrokeEvent = {
      key: event.key,
      code: event.code,
      timestamp: Date.now(),
      action: 'keydown',
      cursorPosition: userInput.length,
      modifiers: {
        ctrl: event.ctrlKey,
        shift: event.shiftKey,
        alt: event.altKey,
        meta: event.metaKey,
      },
    };
    
    keystrokeEvents.current.push(keystrokeEvent);
    
    // Process validation
    const validationResult = typingValidationEngine.processKeystroke(keystrokeEvent);
    
    // Update UI based on validation (could add visual feedback here)
    if (!validationResult.isValid) {
      // Could add error highlighting or feedback
    }
  }, [status, userInput]);

  // Handle input change
  const handleInputChange = useCallback((event: React.ChangeEvent<HTMLTextAreaElement>) => {
    if (status !== 'in_progress') return;
    
    const newValue = event.target.value;
    setUserInput(newValue);
    
    // Check if snippet is completed
    if (newValue === currentSnippet) {
      completeCurrentSnippet();
    }
  }, [status, currentSnippet]);

  // Complete current snippet
  const completeCurrentSnippet = useCallback(async () => {
    if (!assessmentSession || !blueprint) return;
    
    try {
      const endTime = Date.now();
      const timeSpent = endTime - snippetStartTime.current;
      
      // Calculate metrics for this snippet
      const metrics = metricsCalculator.calculateSessionMetrics(
        keystrokeEvents.current,
        currentSnippet,
        userInput
      );
      
      // Perform structural analysis
      const structuralScore = await structuralAnalyzer.analyzeStructuralConformity(
        currentSnippet,
        userInput,
        language
      );
      
      // Record snippet result
      await assessmentService.recordSnippetResult(assessmentSession.id, {
        snippetId: blueprint.snippetIds[currentSnippetIndex],
        startedAt: new Date(snippetStartTime.current),
        completedAt: new Date(endTime),
        expectedText: currentSnippet,
        actualText: userInput,
        keystrokeEvents: keystrokeEvents.current,
        metrics,
        structuralScore,
        passed: assessmentService.evaluateSnippetPassing(metrics, structuralScore, blueprint.passingCriteria),
        timeSpent,
      });
      
      // Move to next snippet or complete assessment
      if (currentSnippetIndex < blueprint.snippetIds.length - 1) {
        await loadSnippet(assessmentSession, currentSnippetIndex + 1);
        setUserInput('');
        keystrokeEvents.current = [];
        snippetStartTime.current = Date.now();
      } else {
        await completeAssessment();
      }
    } catch (err) {
      setError('Failed to process snippet completion');
    }
  }, [assessmentSession, blueprint, currentSnippetIndex, currentSnippet, userInput, language]);

  // Complete entire assessment
  const completeAssessment = useCallback(async () => {
    if (!assessmentSession) return;
    
    if (timerRef.current) {
      clearInterval(timerRef.current);
    }
    
    setStatus('completed');
    
    try {
      const result = await assessmentService.finalizeAssessment(assessmentSession.id);
      onComplete(result);
    } catch (err) {
      setError('Failed to finalize assessment');
    }
  }, [assessmentSession, onComplete]);

  // Pause assessment
  const pauseAssessment = useCallback(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
    }
    setStatus('paused');
  }, []);

  // Resume assessment
  const resumeAssessment = useCallback(() => {
    setStatus('in_progress');
    
    if (blueprint?.passingCriteria.timeLimit && timeRemaining > 0) {
      timerRef.current = setInterval(() => {
        setTimeRemaining(prev => {
          if (prev <= 1000) {
            handleTimeExpired();
            return 0;
          }
          return prev - 1000;
        });
      }, 1000);
    }
    
    inputRef.current?.focus();
  }, [blueprint, timeRemaining, handleTimeExpired]);

  // Format time display
  const formatTime = (milliseconds: number): string => {
    const minutes = Math.floor(milliseconds / 60000);
    const seconds = Math.floor((milliseconds % 60000) / 1000);
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  };

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
    };
  }, []);

  if (isLoading) {
    return (
      <div className="assessment-loading">
        <div className="assessment-loading-spinner" />
        <p>Preparing your assessment...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="assessment-error">
        <h3>Assessment Error</h3>
        <p>{error}</p>
        <button onClick={onExit} className="assessment-exit-btn">
          Return to Menu
        </button>
      </div>
    );
  }

  if (status === 'not_started') {
    return (
      <div className="assessment-intro">
        <div className="assessment-header">
          <h2>Assessment Mode</h2>
          <p>Language: {language.toUpperCase()}</p>
        </div>
        
        {blueprint && (
          <div className="assessment-info">
            <h3>{blueprint.name}</h3>
            <p>{blueprint.description}</p>
            
            <div className="assessment-details">
              <div className="detail-item">
                <span className="label">Difficulty:</span>
                <span className="value">{blueprint.difficulty}/5</span>
              </div>
              <div className="detail-item">
                <span className="label">Estimated Duration:</span>
                <span className="value">{blueprint.estimatedDuration} minutes</span>
              </div>
              <div className="detail-item">
                <span className="label">Snippets:</span>
                <span className="value">{blueprint.snippetIds.length}</span>
              </div>
              {blueprint.passingCriteria.timeLimit && (
                <div className="detail-item">
                  <span className="label">Time Limit:</span>
                  <span className="value">{blueprint.passingCriteria.timeLimit} minutes</span>
                </div>
              )}
            </div>
            
            <div className="passing-criteria">
              <h4>Passing Criteria</h4>
              <ul>
                <li>Minimum Accuracy: {(blueprint.passingCriteria.minimumAccuracy * 100).toFixed(1)}%</li>
                <li>Minimum Speed: {blueprint.passingCriteria.minimumSpeed} tWPM</li>
                <li>Maximum Error Rate: {blueprint.passingCriteria.maximumErrorRate} errors/min</li>
              </ul>
            </div>
          </div>
        )}
        
        <div className="assessment-actions">
          <button onClick={startAssessment} className="assessment-start-btn">
            Start Assessment
          </button>
          <button onClick={onExit} className="assessment-exit-btn">
            Cancel
          </button>
        </div>
      </div>
    );
  }

  if (status === 'paused') {
    return (
      <div className="assessment-paused">
        <h3>Assessment Paused</h3>
        <p>Your progress has been saved.</p>
        <div className="assessment-actions">
          <button onClick={resumeAssessment} className="assessment-resume-btn">
            Resume Assessment
          </button>
          <button onClick={onExit} className="assessment-exit-btn">
            Exit Assessment
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="assessment-mode">
      <div className="assessment-hud">
        <div className="assessment-progress">
          <span>Snippet {currentSnippetIndex + 1} of {blueprint?.snippetIds.length || 1}</span>
          <div className="progress-bar">
            <div 
              className="progress-fill" 
              style={{ 
                width: `${((currentSnippetIndex + 1) / (blueprint?.snippetIds.length || 1)) * 100}%` 
              }}
            />
          </div>
        </div>
        
        {timeRemaining > 0 && (
          <div className="assessment-timer">
            <span className="timer-label">Time Remaining:</span>
            <span className={`timer-value ${timeRemaining < 60000 ? 'warning' : ''}`}>
              {formatTime(timeRemaining)}
            </span>
          </div>
        )}
        
        <div className="assessment-controls">
          <button onClick={pauseAssessment} className="assessment-pause-btn">
            Pause
          </button>
          <button onClick={onExit} className="assessment-exit-btn">
            Exit
          </button>
        </div>
      </div>
      
      <div className="assessment-content">
        <div className="assessment-target">
          <h4>Type the following code:</h4>
          <pre className="target-code">
            <code>{currentSnippet}</code>
          </pre>
        </div>
        
        <div className="assessment-input">
          <textarea
            ref={inputRef}
            value={userInput}
            onChange={handleInputChange}
            onKeyDown={handleKeyDown}
            className="assessment-textarea"
            placeholder="Start typing here..."
            spellCheck={false}
            autoComplete="off"
            autoCorrect="off"
            autoCapitalize="off"
          />
        </div>
        
        <div className="assessment-feedback">
          <div className="progress-indicator">
            Progress: {Math.round((userInput.length / currentSnippet.length) * 100)}%
          </div>
        </div>
      </div>
    </div>
  );
};