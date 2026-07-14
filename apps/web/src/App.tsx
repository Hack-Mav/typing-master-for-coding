import React, {
  useState,
  useEffect,
  Component,
  ErrorInfo,
  ReactNode,
} from 'react';
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
import AuthModal from './components/AuthModal';
import { AuthProvider, useAuth } from './services/AuthContext';
import { privacyService } from './services/PrivacyService';
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

const COMMUNITY_FORUM_URL =
  process.env.REACT_APP_COMMUNITY_FORUM_URL ||
  'https://community.typingmaster.dev/forums';

const SUPPORT_PORTAL_URL =
  process.env.REACT_APP_SUPPORT_PORTAL_URL ||
  'https://community.typingmaster.dev/support';

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
    <AuthProvider>
      <AccessibilityProvider>
        <AppContentWithAuth />
      </AccessibilityProvider>
    </AuthProvider>
  );
}

function AppContentWithAuth() {
  const { isAuthenticated, isAnonymous, loading } = useAuth();
  const [showAuthModal, setShowAuthModal] = useState(false);
  const [state, setState] = useState<AppState>({
    currentMode: 'menu',
    selectedLanguage: 'javascript',
    timedDrillDuration: 60000, // 1 minute default
  });

  // Show auth modal if not authenticated and not in loading state
  useEffect(() => {
    if (!loading && !isAuthenticated && !isAnonymous) {
      setShowAuthModal(true);
    }
  }, [loading, isAuthenticated, isAnonymous]);

  const handleModeSelect = (mode: AppMode) => {
    if (!isAuthenticated && !isAnonymous) {
      setShowAuthModal(true);
      return;
    }
    setState(prev => ({ ...prev, currentMode: mode }));
  };

  const handleLanguageSelect = (language: string) => {
    setState(prev => ({ ...prev, selectedLanguage: language }));
  };

  const handleDurationSelect = (duration: number) => {
    setState(prev => ({ ...prev, timedDrillDuration: duration }));
  };

  const handleCommunityClick = (channel: 'forum' | 'support') => {
    const url = channel === 'forum' ? COMMUNITY_FORUM_URL : SUPPORT_PORTAL_URL;

    try {
      privacyService.recordTelemetryEvent('community_navigation', {
        channel,
        location: 'main_menu',
      });
    } catch (error) {
      console.warn('Failed to record community navigation telemetry', error);
    }

    if (typeof window !== 'undefined') {
      window.open(url, '_blank', 'noopener,noreferrer');
    }
  };

  const handleSessionComplete = (result: SessionResult | AssessmentResult) => {
    console.log('Session completed:', result);
    // Return to menu after completion
    setState(prev => ({ ...prev, currentMode: 'menu' }));
  };

  const handleExit = () => {
    setState(prev => ({ ...prev, currentMode: 'menu' }));
  };

  const handleAuthenticated = (isAnonymous?: boolean) => {
    setShowAuthModal(false);
    // Only reload for regular login/register where cookies are set.
    // Anonymous auth sets state in memory, so no reload needed.
    if (!isAnonymous) {
      window.location.reload();
    }
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
            role="button"
            tabIndex={0}
            aria-label="Syntax Tutorials"
            className="mode-card"
            onClick={() => handleModeSelect('syntax-tutorial')}
            onKeyDown={e => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                handleModeSelect('syntax-tutorial');
              }
            }}
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
            role="button"
            tabIndex={0}
            aria-label="Timed Drill"
            className="mode-card"
            onClick={() => handleModeSelect('timed-drill')}
            onKeyDown={e => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                handleModeSelect('timed-drill');
              }
            }}
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
            role="button"
            tabIndex={0}
            aria-label="Accuracy Mode"
            className="mode-card"
            onClick={() => handleModeSelect('accuracy')}
            onKeyDown={e => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                handleModeSelect('accuracy');
              }
            }}
          >
            <h3>Accuracy Mode</h3>
            <p>Focus on error-free typing with precision scoring</p>
            <div className="mode-features">
              <span>• Error penalties</span>
              <span>• Syntax matching</span>
              <span>• Precision focus</span>
            </div>
          </div>

          <div
            role="button"
            tabIndex={0}
            aria-label="Zen Mode"
            className="mode-card"
            onClick={() => handleModeSelect('zen')}
            onKeyDown={e => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                handleModeSelect('zen');
              }
            }}
          >
            <h3>Zen Mode</h3>
            <p>Distraction-free coding practice without metrics</p>
            <div className="mode-features">
              <span>• Minimal UI</span>
              <span>• No pressure</span>
              <span>• Optional summary</span>
            </div>
          </div>

          <div
            role="button"
            tabIndex={0}
            aria-label="Custom Snippets"
            className="mode-card"
            onClick={() => handleModeSelect('custom-snippets')}
            onKeyDown={e => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                handleModeSelect('custom-snippets');
              }
            }}
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
            role="button"
            tabIndex={0}
            aria-label="Assessment"
            className="mode-card"
            onClick={() => handleModeSelect('assessment')}
            onKeyDown={e => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                handleModeSelect('assessment');
              }
            }}
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

      <div className="community-support">
        <h3>Community &amp; Support</h3>
        <p>Get help, share feedback, and connect with other developers.</p>
        <div className="community-buttons">
          <button
            className="community-btn"
            type="button"
            onClick={() => handleCommunityClick('forum')}
          >
            Community Forums
          </button>
          <button
            className="community-btn secondary"
            type="button"
            onClick={() => handleCommunityClick('support')}
          >
            Support &amp; Help Center
          </button>
        </div>
      </div>
    </div>
  );

  // Render current mode
  if (loading) {
    return (
      <div className="loading-screen">
        <div className="loading-spinner"></div>
        <p>Loading...</p>
      </div>
    );
  }

  return (
    <>
      {state.currentMode === 'menu' && (
        <ErrorBoundary>{renderMenu()}</ErrorBoundary>
      )}

      {state.currentMode === 'syntax-tutorial' &&
        (isAuthenticated || isAnonymous) && (
          <ErrorBoundary>
            <SyntaxTutorialMode
              languageId={state.selectedLanguage as Language}
              onComplete={handleSessionComplete}
              onExit={handleExit}
            />
          </ErrorBoundary>
        )}

      {state.currentMode === 'timed-drill' &&
        (isAuthenticated || isAnonymous) && (
          <ErrorBoundary>
            <TimedDrillMode
              languageId={state.selectedLanguage}
              duration={state.timedDrillDuration}
              onComplete={handleSessionComplete}
              onExit={handleExit}
            />
          </ErrorBoundary>
        )}

      {state.currentMode === 'accuracy' && (isAuthenticated || isAnonymous) && (
        <ErrorBoundary>
          <AccuracyMode
            languageId={state.selectedLanguage}
            onComplete={handleSessionComplete}
            onExit={handleExit}
          />
        </ErrorBoundary>
      )}

      {state.currentMode === 'zen' && (isAuthenticated || isAnonymous) && (
        <ErrorBoundary>
          <ZenMode
            languageId={state.selectedLanguage}
            onComplete={handleSessionComplete}
            onExit={handleExit}
          />
        </ErrorBoundary>
      )}

      {state.currentMode === 'custom-snippets' &&
        (isAuthenticated || isAnonymous) && (
          <ErrorBoundary>
            <CustomSnippetsMode
              languageId={state.selectedLanguage as Language}
              onComplete={handleSessionComplete}
              onExit={handleExit}
            />
          </ErrorBoundary>
        )}

      {state.currentMode === 'assessment' &&
        (isAuthenticated || isAnonymous) && (
          <ErrorBoundary>
            <AssessmentMode
              language={state.selectedLanguage as Language}
              onComplete={handleSessionComplete}
              onExit={handleExit}
            />
          </ErrorBoundary>
        )}

      <AuthModal
        isOpen={showAuthModal}
        onClose={() => setShowAuthModal(false)}
        onAuthenticated={handleAuthenticated}
      />
    </>
  );
}

export default App;
