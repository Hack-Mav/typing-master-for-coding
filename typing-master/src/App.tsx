import React, { useState, Component, ErrorInfo, ReactNode } from 'react';
import ZenMode from './components/ZenMode';
import TimedDrillMode from './components/TimedDrillMode';
import SyntaxTutorialMode from './components/SyntaxTutorialMode';
import AccuracyMode from './components/AccuracyMode';
import CustomSnippetsMode from './components/CustomSnippetsMode';
import { AssessmentMode } from './components/AssessmentMode';
import { SessionResult } from './types/session';
import { AssessmentResult } from './types/assessment';
import AccessibilityProvider from './components/AccessibilitySettings';
import { Language } from './types/parser';
import './App.css';

type AppMode =
  | 'menu'
  | 'zen'
  | 'timed-drill'
  | 'syntax-tutorial'
  | 'accuracy'
  | 'custom-snippets'
  | 'assessment';

interface AppState {
  currentMode: AppMode;
  selectedLanguage: string;
  timedDrillDuration: number;
}

// Error Boundary Component
interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
}

class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="error-boundary">
          <div className="error-content">
            <h2>Oops! Something went wrong</h2>
            <p>
              There was an error loading the application. Please refresh the
              page to try again.
            </p>
            <button
              onClick={() => {
                if (typeof window !== 'undefined' && window.location) {
                  window.location.reload();
                } else {
                  // Fallback for environments without window
                  console.log('Please refresh the page to continue');
                }
              }}
              className="error-retry-btn"
            >
              Refresh Page
            </button>
            {typeof process !== 'undefined' &&
              process.env &&
              process.env.NODE_ENV === 'development' &&
              this.state.error && (
                <details className="error-details">
                  <summary>Error Details (Development)</summary>
                  <pre>{this.state.error.toString()}</pre>
                </details>
              )}
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

function App() {
  return (
    <AccessibilityProvider>
      <AppContent />
    </AccessibilityProvider>
  );
}

function AppContent() {
  const [state, setState] = useState<AppState>({
    currentMode: 'menu',
    selectedLanguage: 'javascript',
    timedDrillDuration: 60000, // 1 minute default
  });

  const handleModeSelect = (mode: AppMode) => {
    setState(prev => ({ ...prev, currentMode: mode }));
  };

  const handleLanguageSelect = (language: string) => {
    setState(prev => ({ ...prev, selectedLanguage: language }));
  };

  const handleDurationSelect = (duration: number) => {
    setState(prev => ({ ...prev, timedDrillDuration: duration }));
  };

  const handleSessionComplete = (result: SessionResult | AssessmentResult) => {
    console.log('Session completed:', result);
    // Return to menu after completion
    setState(prev => ({ ...prev, currentMode: 'menu' }));
  };

  const handleExit = () => {
    setState(prev => ({ ...prev, currentMode: 'menu' }));
  };

  // Main menu component
  const renderMenu = () => (
    <div className="app-menu">
      <div className="app-header">
        <h1>Typing Master for Coding</h1>
        <p>Improve your coding skills with syntax-aware typing practice</p>
      </div>

      <div className="mode-selection">
        <h2>Choose Your Practice Mode</h2>

        <div className="mode-cards">
          <div
            className="mode-card"
            onClick={() => handleModeSelect('syntax-tutorial')}
          >
            <h3>Syntax Tutorials</h3>
            <p>Structured lessons from basics to advanced patterns</p>
            <div className="mode-features">
              <span>• Progressive learning</span>
              <span>• Language-specific</span>
              <span>• Guided practice</span>
            </div>
          </div>

          <div
            className="mode-card"
            onClick={() => handleModeSelect('timed-drill')}
          >
            <h3>Timed Drill</h3>
            <p>Speed-focused practice with metrics and time challenges</p>
            <div className="mode-features">
              <span>• Real-time metrics</span>
              <span>• Time pressure</span>
              <span>• Performance tracking</span>
            </div>
          </div>

          <div
            className="mode-card"
            onClick={() => handleModeSelect('accuracy')}
          >
            <h3>Accuracy Mode</h3>
            <p>Focus on error-free typing with precision scoring</p>
            <div className="mode-features">
              <span>• Error penalties</span>
              <span>• Syntax matching</span>
              <span>• Precision focus</span>
            </div>
          </div>

          <div className="mode-card" onClick={() => handleModeSelect('zen')}>
            <h3>Zen Mode</h3>
            <p>Distraction-free coding practice without metrics</p>
            <div className="mode-features">
              <span>• Minimal UI</span>
              <span>• No pressure</span>
              <span>• Optional summary</span>
            </div>
          </div>

          <div
            className="mode-card"
            onClick={() => handleModeSelect('custom-snippets')}
          >
            <h3>Custom Snippets</h3>
            <p>Practice with your own code snippets</p>
            <div className="mode-features">
              <span>• Upload code</span>
              <span>• Personal practice</span>
              <span>• Any language</span>
            </div>
          </div>

          <div
            className="mode-card"
            onClick={() => handleModeSelect('assessment')}
          >
            <h3>Assessment</h3>
            <p>Structured evaluation of your typing skills</p>
            <div className="mode-features">
              <span>• Skill testing</span>
              <span>• Pass/fail criteria</span>
              <span>• Certification</span>
            </div>
          </div>
        </div>
      </div>

      <div className="language-selection">
        <h3>Select Programming Language</h3>
        <div className="language-buttons">
          {['javascript', 'python', 'cpp', 'rust', 'yaml'].map(lang => (
            <button
              key={lang}
              className={`language-btn ${state.selectedLanguage === lang ? 'active' : ''}`}
              onClick={() => handleLanguageSelect(lang)}
            >
              {lang.toUpperCase()}
            </button>
          ))}
        </div>
      </div>

      {state.currentMode === 'timed-drill' && (
        <div className="duration-selection">
          <h3>Select Duration</h3>
          <div className="duration-buttons">
            {[60000, 180000, 300000, 600000].map(duration => (
              <button
                key={duration}
                className={`duration-btn ${state.timedDrillDuration === duration ? 'active' : ''}`}
                onClick={() => handleDurationSelect(duration)}
              >
                {duration === 60000
                  ? '1m'
                  : duration === 180000
                    ? '3m'
                    : duration === 300000
                      ? '5m'
                      : '10m'}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );

  // Render current mode
  switch (state.currentMode) {
    case 'syntax-tutorial':
      return (
        <ErrorBoundary>
          <SyntaxTutorialMode
            languageId={state.selectedLanguage as Language}
            onComplete={handleSessionComplete}
            onExit={handleExit}
          />
        </ErrorBoundary>
      );

    case 'timed-drill':
      return (
        <ErrorBoundary>
          <TimedDrillMode
            languageId={state.selectedLanguage}
            duration={state.timedDrillDuration}
            onComplete={handleSessionComplete}
            onExit={handleExit}
          />
        </ErrorBoundary>
      );

    case 'accuracy':
      return (
        <ErrorBoundary>
          <AccuracyMode
            languageId={state.selectedLanguage}
            onComplete={handleSessionComplete}
            onExit={handleExit}
          />
        </ErrorBoundary>
      );

    case 'zen':
      return (
        <ErrorBoundary>
          <ZenMode
            languageId={state.selectedLanguage}
            onComplete={handleSessionComplete}
            onExit={handleExit}
          />
        </ErrorBoundary>
      );

    case 'custom-snippets':
      return (
        <ErrorBoundary>
          <CustomSnippetsMode
            languageId={state.selectedLanguage as Language}
            onComplete={handleSessionComplete}
            onExit={handleExit}
          />
        </ErrorBoundary>
      );

    case 'assessment':
      return (
        <ErrorBoundary>
          <AssessmentMode
            language={state.selectedLanguage as Language}
            onComplete={handleSessionComplete}
            onExit={handleExit}
          />
        </ErrorBoundary>
      );

    default:
      return <ErrorBoundary>{renderMenu()}</ErrorBoundary>;
  }
}
export default App;
