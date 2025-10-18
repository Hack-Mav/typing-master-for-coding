import React, { useState, Component, ErrorInfo, ReactNode } from 'react';
import ZenMode from './components/ZenMode';
import TimedDrillMode from './components/TimedDrillMode';
import { SessionResult } from './types/session';
import AccessibilityProvider from './components/AccessibilitySettings';
import './App.css';

type AppMode = 'menu' | 'zen' | 'timed-drill';

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

  const handleSessionComplete = (result: SessionResult) => {
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
          <div className="mode-card" onClick={() => handleModeSelect('zen')}>
            <h3>Zen Mode</h3>
            <p>
              Distraction-free coding practice with real-time syntax validation
            </p>
            <div className="mode-features">
              <span>• Minimal UI</span>
              <span>• Syntax highlighting</span>
              <span>• No time pressure</span>
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
        </div>
      </div>

      <div className="language-selection">
        <h3>Select Programming Language</h3>
        <div className="language-buttons">
          {['javascript', 'python', 'typescript', 'java', 'cpp', 'rust'].map(
            lang => (
              <button
                key={lang}
                className={`language-btn ${state.selectedLanguage === lang ? 'active' : ''}`}
                onClick={() => handleLanguageSelect(lang)}
              >
                {lang.toUpperCase()}
              </button>
            )
          )}
        </div>
      </div>

      {state.currentMode === 'timed-drill' && (
        <div className="duration-selection">
          <h3>Select Duration</h3>
          <div className="duration-buttons">
            {[30000, 60000, 180000, 300000].map(duration => (
              <button
                key={duration}
                className={`duration-btn ${state.timedDrillDuration === duration ? 'active' : ''}`}
                onClick={() => handleDurationSelect(duration)}
              >
                {duration === 30000
                  ? '30s'
                  : duration === 60000
                    ? '1m'
                    : duration === 180000
                      ? '3m'
                      : '5m'}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );

  // Render current mode
  switch (state.currentMode) {
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

    default:
      return <ErrorBoundary>{renderMenu()}</ErrorBoundary>;
  }
}
export default App;
