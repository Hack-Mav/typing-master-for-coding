import React, { Component, ErrorInfo, ReactNode } from 'react';
import {
  getErrorHandler,
  ErrorType,
  ErrorSeverity,
} from '../services/ErrorHandler';
import { getGracefulDegradationService } from '../services/GracefulDegradation';
import './ErrorComponents.css';

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: (error: Error, errorInfo: ErrorInfo) => ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
  level?: 'page' | 'component' | 'feature';
}

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
  errorInfo?: ErrorInfo;
  errorId?: string;
  retryCount: number;
}

class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  private errorHandler = getErrorHandler();
  private gracefulDegradation = getGracefulDegradationService();
  private maxRetries = 3;

  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      retryCount: 0,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return {
      hasError: true,
      error,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Handle the error with our error handler
    const appError = this.errorHandler.createError(
      this.classifyError(error),
      error.message,
      {
        cause: error,
        severity: this.getSeverityForLevel(this.props.level || 'component'),
        technicalDetails: {
          componentStack: errorInfo.componentStack,
          errorBoundaryLevel: this.props.level || 'component',
          retryCount: this.state.retryCount,
        },
        context: {
          component: 'ErrorBoundary',
          action: 'component_error',
        },
      }
    );

    const handledError = this.errorHandler.handleError(appError);

    this.setState({
      errorInfo,
      errorId: handledError.id,
    });

    // Call custom error handler if provided
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }

    // Log to console in development
    if (
      (window as any).__DEV__ ||
      !window.location.hostname.includes('production')
    ) {
      console.group('🚨 Error Boundary Caught Error');
      console.error('Error:', error);
      console.error('Error Info:', errorInfo);
      console.error('App Error:', handledError);
      console.groupEnd();
    }
  }

  private classifyError(error: Error): ErrorType {
    const message = error.message.toLowerCase();

    if (message.includes('chunk') || message.includes('loading')) {
      return ErrorType.NETWORK_ERROR;
    }
    if (message.includes('parse') || message.includes('syntax')) {
      return ErrorType.PARSING_ERROR;
    }
    if (message.includes('worker')) {
      return ErrorType.WORKER_ERROR;
    }
    if (message.includes('storage') || message.includes('quota')) {
      return ErrorType.STORAGE_ERROR;
    }

    return ErrorType.UNKNOWN_ERROR;
  }

  private getSeverityForLevel(level: string): ErrorSeverity {
    switch (level) {
      case 'page':
        return ErrorSeverity.HIGH;
      case 'feature':
        return ErrorSeverity.MEDIUM;
      case 'component':
      default:
        return ErrorSeverity.LOW;
    }
  }

  private handleRetry = () => {
    if (this.state.retryCount < this.maxRetries) {
      this.setState(prevState => ({
        hasError: false,
        error: undefined,
        errorInfo: undefined,
        errorId: undefined,
        retryCount: prevState.retryCount + 1,
      }));
    }
  };

  private handleRefresh = () => {
    window.location.reload();
  };

  private handleReportError = () => {
    if (this.state.errorId) {
      // This would open a support ticket or feedback form
      console.log('Reporting error:', this.state.errorId);
      // You could integrate with a support system here
    }
  };

  private renderFallbackUI(): ReactNode {
    const { error, errorInfo, retryCount } = this.state;
    const { level = 'component' } = this.props;

    // Use custom fallback if provided
    if (this.props.fallback && error && errorInfo) {
      return this.props.fallback(error, errorInfo);
    }

    // Get degradation report
    const degradationReport = this.gracefulDegradation.getDegradationReport();
    const canRetry = retryCount < this.maxRetries;

    // Different UI based on error level
    if (level === 'page') {
      return (
        <div className="error-boundary-page">
          <div className="error-content">
            <div className="error-icon">⚠️</div>
            <h1>Oops! Something went wrong</h1>
            <p className="error-message">
              We encountered an unexpected error. Don't worry, your progress has
              been saved.
            </p>

            {degradationReport.overallHealth !== 'healthy' && (
              <div className="degradation-notice">
                <h3>System Status: {degradationReport.overallHealth}</h3>
                <ul>
                  {degradationReport.recommendations.map((rec, index) => (
                    <li key={index}>{rec}</li>
                  ))}
                </ul>
              </div>
            )}

            <div className="error-actions">
              {canRetry && (
                <button onClick={this.handleRetry} className="btn-primary">
                  Try Again ({this.maxRetries - retryCount} attempts left)
                </button>
              )}
              <button onClick={this.handleRefresh} className="btn-secondary">
                Refresh Page
              </button>
              <button onClick={this.handleReportError} className="btn-outline">
                Report Issue
              </button>
            </div>

            {((window as any).__DEV__ ||
              !window.location.hostname.includes('production')) &&
              error && (
                <details className="error-details">
                  <summary>Error Details (Development)</summary>
                  <pre className="error-stack">
                    {error.toString()}
                    {errorInfo?.componentStack}
                  </pre>
                </details>
              )}
          </div>
        </div>
      );
    }

    if (level === 'feature') {
      return (
        <div className="error-boundary-feature">
          <div className="error-content">
            <div className="error-icon">⚠️</div>
            <h3>Feature Temporarily Unavailable</h3>
            <p>
              This feature is experiencing issues. You can continue using other
              parts of the application.
            </p>

            <div className="error-actions">
              {canRetry && (
                <button onClick={this.handleRetry} className="btn-small">
                  Retry
                </button>
              )}
              <button
                onClick={this.handleReportError}
                className="btn-small btn-outline"
              >
                Report
              </button>
            </div>
          </div>
        </div>
      );
    }

    // Component level error (minimal UI)
    return (
      <div className="error-boundary-component">
        <div className="error-content">
          <span className="error-icon">⚠️</span>
          <span className="error-text">Component error</span>
          {canRetry && (
            <button onClick={this.handleRetry} className="btn-tiny">
              Retry
            </button>
          )}
        </div>
      </div>
    );
  }

  render() {
    if (this.state.hasError) {
      return this.renderFallbackUI();
    }

    return this.props.children;
  }
}

// Higher-order component for easy wrapping
export function withErrorBoundary<P extends object>(
  Component: React.ComponentType<P>,
  errorBoundaryProps?: Omit<ErrorBoundaryProps, 'children'>
) {
  const WrappedComponent = (props: P) => (
    <ErrorBoundary {...errorBoundaryProps}>
      <Component {...props} />
    </ErrorBoundary>
  );

  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name})`;
  return WrappedComponent;
}

// Hook for manual error reporting
export function useErrorHandler() {
  const errorHandler = getErrorHandler();

  const reportError = React.useCallback(
    (error: Error, context?: any) => {
      return errorHandler.handleError(error, context);
    },
    [errorHandler]
  );

  const reportUserError = React.useCallback(
    (userMessage: string, technicalMessage: string, error?: Error) => {
      const appError = errorHandler.createError(
        ErrorType.UNKNOWN_ERROR,
        technicalMessage,
        {
          cause: error,
          userMessage,
          context: {
            component: 'UserReported',
            action: 'manual_report',
          },
        }
      );
      return errorHandler.handleError(appError);
    },
    [errorHandler]
  );

  return {
    reportError,
    reportUserError,
    getErrorStats: errorHandler.getErrorStats.bind(errorHandler),
    clearErrorLog: errorHandler.clearErrorLog.bind(errorHandler),
  };
}

export default ErrorBoundary;
