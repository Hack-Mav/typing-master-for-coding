import React, { useState, useEffect, useCallback } from 'react';
import MonacoTypingInterface from './MonacoTypingInterface';
import { contentService } from '../services/ContentService';
import { SessionResult } from '../types/session';
import { Language } from '../types/parser';
import './SyntaxTutorialMode.css';

interface SyntaxTutorialModeProps {
  languageId: Language;
  onComplete?: (result: SessionResult) => void;
  onExit?: () => void;
}

type TutorialLevel =
  | 'intro'
  | 'core-syntax'
  | 'idioms'
  | 'advanced'
  | 'mixed-review';

interface LessonProgress {
  level: TutorialLevel;
  lessonIndex: number;
  completed: number[];
}

const TUTORIAL_LEVELS: Array<{
  id: TutorialLevel;
  name: string;
  description: string;
}> = [
  {
    id: 'intro',
    name: 'Introduction',
    description: 'Basic syntax and structure',
  },
  {
    id: 'core-syntax',
    name: 'Core Syntax',
    description: 'Essential language constructs',
  },
  {
    id: 'idioms',
    name: 'Idioms',
    description: 'Common patterns and practices',
  },
  {
    id: 'advanced',
    name: 'Advanced Patterns',
    description: 'Complex language features',
  },
  {
    id: 'mixed-review',
    name: 'Mixed Review',
    description: 'Comprehensive practice',
  },
];

const SyntaxTutorialMode: React.FC<SyntaxTutorialModeProps> = ({
  languageId,
  onComplete,
  onExit,
}) => {
  const [progress, setProgress] = useState<LessonProgress>({
    level: 'intro',
    lessonIndex: 0,
    completed: [],
  });
  const [currentLesson, setCurrentLesson] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [showLevelSelect, setShowLevelSelect] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Load lesson for current progress
  const loadLesson = useCallback(
    async (level: TutorialLevel, index: number) => {
      try {
        setIsLoading(true);
        setError(null);

        // Get lessons for the current level
        const lessons = await contentService.getLessonsByCategory(
          languageId,
          `tutorial-${level}`
        );

        if (!lessons || lessons.length === 0) {
          throw new Error(`No lessons found for ${level} level`);
        }

        const lesson = lessons[index] || lessons[0];
        setCurrentLesson(lesson);
        setShowLevelSelect(false);
        setIsLoading(false);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load lesson');
        setIsLoading(false);
      }
    },
    [languageId]
  );

  // Initialize with first lesson
  useEffect(() => {
    if (!showLevelSelect) {
      loadLesson(progress.level, progress.lessonIndex);
    }
  }, [showLevelSelect, progress.level, progress.lessonIndex, loadLesson]);

  const handleLevelSelect = (level: TutorialLevel) => {
    setProgress({ level, lessonIndex: 0, completed: [] });
    setShowLevelSelect(false);
  };

  const handleLessonComplete = async (result?: SessionResult) => {
    // Mark lesson as completed
    const newCompleted = [...progress.completed, progress.lessonIndex];
    setProgress(prev => ({ ...prev, completed: newCompleted }));

    // Check if we should advance to next lesson or level
    const lessons = await contentService.getLessonsByCategory(
      languageId,
      `tutorial-${progress.level}`
    );

    if (progress.lessonIndex < lessons.length - 1) {
      // More lessons in this level
      setProgress(prev => ({ ...prev, lessonIndex: prev.lessonIndex + 1 }));
    } else {
      // Level completed, show level select
      setShowLevelSelect(true);
    }

    if (onComplete && result) {
      onComplete(result);
    }
  };

  const handleBackToLevels = () => {
    setShowLevelSelect(true);
  };

  if (error) {
    return (
      <div className="syntax-tutorial-error">
        <h3>Error Loading Tutorial</h3>
        <p>{error}</p>
        <button onClick={onExit} className="btn-exit">
          Return to Menu
        </button>
      </div>
    );
  }

  if (showLevelSelect) {
    return (
      <div className="syntax-tutorial-menu">
        <div className="tutorial-header">
          <h2>Syntax Tutorials - {languageId.toUpperCase()}</h2>
          <p>Choose your learning path</p>
          <button onClick={onExit} className="btn-exit-small">
            Exit
          </button>
        </div>

        <div className="tutorial-levels">
          {TUTORIAL_LEVELS.map((level, index) => {
            const isUnlocked = index === 0 || progress.completed.length > 0;
            return (
              <div
                key={level.id}
                className={`tutorial-level-card ${!isUnlocked ? 'locked' : ''}`}
                onClick={() => isUnlocked && handleLevelSelect(level.id)}
              >
                <div className="level-number">{index + 1}</div>
                <div className="level-content">
                  <h3>{level.name}</h3>
                  <p>{level.description}</p>
                  {!isUnlocked && (
                    <span className="locked-badge">🔒 Locked</span>
                  )}
                </div>
              </div>
            );
          })}
        </div>

        <div className="tutorial-info">
          <h4>About Syntax Tutorials</h4>
          <p>
            Progress through structured lessons designed to teach you the syntax
            and idioms of {languageId}. Each level builds on the previous one,
            taking you from basics to advanced patterns.
          </p>
        </div>
      </div>
    );
  }

  if (isLoading || !currentLesson) {
    return (
      <div className="syntax-tutorial-loading">
        <div className="loading-spinner" />
        <p>Loading lesson...</p>
      </div>
    );
  }

  return (
    <div className="syntax-tutorial-mode">
      <div className="tutorial-hud">
        <div className="tutorial-progress">
          <span className="level-badge">
            {TUTORIAL_LEVELS.find(l => l.id === progress.level)?.name}
          </span>
          <span className="lesson-number">
            Lesson {progress.lessonIndex + 1}
          </span>
        </div>
        <div className="tutorial-controls">
          <button onClick={handleBackToLevels} className="btn-back">
            ← Levels
          </button>
          <button onClick={onExit} className="btn-exit">
            Exit
          </button>
        </div>
      </div>

      <div className="tutorial-content">
        <div className="lesson-info">
          <h3>{currentLesson.title}</h3>
          <p>{currentLesson.description}</p>
          {currentLesson.objectives && (
            <div className="lesson-objectives">
              <h4>Learning Objectives:</h4>
              <ul>
                {currentLesson.objectives.map((obj: string, i: number) => (
                  <li key={i}>{obj}</li>
                ))}
              </ul>
            </div>
          )}
        </div>

        <MonacoTypingInterface
          targetText={currentLesson.content}
          languageId={languageId}
          currentText=""
          onTextChange={() => {
            // Monaco will handle text changes internally
          }}
          onComplete={() => {
            // Monaco calls onComplete when typing is complete
            // We need to get the final result and pass it to handleLessonComplete
            handleLessonComplete();
          }}
          theme="dark"
          fontSize={16}
          lineHeight={1.6}
          keyboardLayout="qwerty"
          className="syntax-tutorial-editor"
        />
      </div>
    </div>
  );
};

export default SyntaxTutorialMode;
