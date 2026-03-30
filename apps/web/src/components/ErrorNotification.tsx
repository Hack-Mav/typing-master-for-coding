import React, { useState, useEffect, useCallback } from 'react';
import {
  getErrorHandler,
  AppError,
  ErrorSeverity,
} from '../services/ErrorHandler';
import { getGracefulDegradationService } from '../services/GracefulDegradation';
import './ErrorComponents.css';

interface ErrorNotificationProps {
  maxNotifications?: number;
  autoHideDelay?: number;
  position?: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left';
}

interface NotificationItem {
  id: string;
  error: AppError;
  timestamp: number;
  dismissed: boolean;
  autoHide: boolean;
}

const ErrorNotification: React.FC<ErrorNotificationProps> = ({
  maxNotifications = 5,
  autoHideDelay = 5000,
  position = 'top-right',
}) => {
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const errorHandler = getErrorHandler();

  // Listen for new errors
  useEffect(() => {
    const handleNewError = (error: AppError) => {
      // Don't show notifications for low severity errors unless they're user-facing
      if (error.severity === ErrorSeverity.LOW && !error.userMessage) {
        return;
      }

      const notification: NotificationItem = {
        id: error.id,
        error,
        timestamp: Date.now(),
        dismissed: false,
        autoHide: error.severity !== ErrorSeverity.CRITICAL,
      };

      setNotifications(prev => {
        // Remove oldest notifications if we exceed the limit
        const updated = [notification, ...prev].slice(0, maxNotifications);
        return updated;
      });
    };

    // This would be connected to the error handler's event system
    // For now, we'll simulate it with a custom event
    const handleErrorEvent = (event: CustomEvent<AppError>) => {
      handleNewError(event.detail);
    };

    window.addEventListener('app-error', handleErrorEvent as EventListener);
    return () => {
      window.removeEventListener(
        'app-error',
        handleErrorEvent as EventListener
      );
    };
  }, [maxNotifications]);

  // Auto-hide notifications
  useEffect(() => {
    const timer = setInterval(() => {
      setNotifications(prev =>
        prev.map(notification => {
          if (
            notification.autoHide &&
            !notification.dismissed &&
            Date.now() - notification.timestamp > autoHideDelay
          ) {
            return { ...notification, dismissed: true };
          }
          return notification;
        })
      );
    }, 1000);

    return () => clearInterval(timer);
  }, [autoHideDelay]);

  // Remove dismissed notifications after animation
  useEffect(() => {
    const timer = setTimeout(() => {
      setNotifications(prev => prev.filter(n => !n.dismissed));
    }, 300); // Match CSS animation duration

    return () => clearTimeout(timer);
  }, [notifications]);

  const dismissNotification = useCallback((id: string) => {
    setNotifications(prev =>
      prev.map(n => (n.id === id ? { ...n, dismissed: true } : n))
    );
  }, []);

  const retryAction = useCallback(
    (error: AppError) => {
      if (error.retryable) {
        // Emit a retry event that components can listen to
        window.dispatchEvent(
          new CustomEvent('error-retry', { detail: { errorId: error.id } })
        );
        dismissNotification(error.id);
      }
    },
    [dismissNotification]
  );

  const getNotificationIcon = (severity: ErrorSeverity): string => {
    switch (severity) {
      case ErrorSeverity.CRITICAL:
        return '🚨';
      case ErrorSeverity.HIGH:
        return '❌';
      case ErrorSeverity.MEDIUM:
        return '⚠️';
      case ErrorSeverity.LOW:
      default:
        return 'ℹ️';
    }
  };

  const getNotificationClass = (severity: ErrorSeverity): string => {
    const baseClass = 'error-notification';
    switch (severity) {
      case ErrorSeverity.CRITICAL:
        return `${baseClass} critical`;
      case ErrorSeverity.HIGH:
        return `${baseClass} high`;
      case ErrorSeverity.MEDIUM:
        return `${baseClass} medium`;
      case ErrorSeverity.LOW:
      default:
        return `${baseClass} low`;
    }
  };

  if (notifications.length === 0) {
    return null;
  }

  return (
    <div className={`error-notifications ${position}`}>
      {notifications.map(notification => (
        <div
          key={notification.id}
          className={`${getNotificationClass(notification.error.severity)} ${
            notification.dismissed ? 'dismissed' : ''
          }`}
        >
          <div className="notification-content">
            <div className="notification-header">
              <span className="notification-icon">
                {getNotificationIcon(notification.error.severity)}
              </span>
              <span className="notification-title">
                {notification.error.severity === ErrorSeverity.CRITICAL
                  ? 'Critical Error'
                  : notification.error.severity === ErrorSeverity.HIGH
                    ? 'Error'
                    : notification.error.severity === ErrorSeverity.MEDIUM
                      ? 'Warning'
                      : 'Notice'}
              </span>
              <button
                className="notification-close"
                onClick={() => dismissNotification(notification.id)}
                aria-label="Dismiss notification"
              >
                ×
              </button>
            </div>

            <div className="notification-message">
              {errorHandler.getUserMessage(notification.error)}
            </div>

            {notification.error.suggestedAction && (
              <div className="notification-suggestion">
                {errorHandler.getSuggestedAction(notification.error)}
              </div>
            )}

            <div className="notification-actions">
              {notification.error.retryable && (
                <button
                  className="notification-action retry"
                  onClick={() => retryAction(notification.error)}
                >
                  Retry
                </button>
              )}

              <button
                className="notification-action dismiss"
                onClick={() => dismissNotification(notification.id)}
              >
                Dismiss
              </button>
            </div>

            {!notification.autoHide && (
              <div className="notification-persistent">
                This error requires attention
              </div>
            )}
          </div>

          {notification.autoHide && (
            <div
              className="notification-progress"
              style={{
                animationDuration: `${autoHideDelay}ms`,
              }}
            />
          )}
        </div>
      ))}

      {/* System status indicator */}
      <SystemStatusIndicator />
    </div>
  );
};

// System status component
const SystemStatusIndicator: React.FC = () => {
  const [statusVisible, setStatusVisible] = useState(false);
  const gracefulDegradation = getGracefulDegradationService();

  useEffect(() => {
    const checkStatus = () => {
      const report = gracefulDegradation.getDegradationReport();
      setStatusVisible(report.overallHealth !== 'healthy');
    };

    checkStatus();
    const interval = setInterval(checkStatus, 30000); // Check every 30 seconds

    return () => clearInterval(interval);
  }, [gracefulDegradation]);

  if (!statusVisible) {
    return null;
  }

  const report = gracefulDegradation.getDegradationReport();

  return (
    <div className={`system-status ${report.overallHealth}`}>
      <div className="status-header">
        <span className="status-icon">
          {report.overallHealth === 'critical' ? '🔴' : '🟡'}
        </span>
        <span className="status-text">
          System Status: {report.overallHealth}
        </span>
      </div>

      <div className="status-details">
        <div className="status-metric">
          Features: {report.availableFeatures}/{report.totalFeatures} available
        </div>

        {report.activeFallbacks.length > 0 && (
          <div className="status-fallbacks">
            Fallback mode: {report.activeFallbacks.join(', ')}
          </div>
        )}

        {report.recommendations.length > 0 && (
          <div className="status-recommendations">
            <details>
              <summary>Recommendations</summary>
              <ul>
                {report.recommendations.map((rec, index) => (
                  <li key={index}>{rec}</li>
                ))}
              </ul>
            </details>
          </div>
        )}
      </div>
    </div>
  );
};

// Hook for triggering error notifications
export function useErrorNotification() {
  const triggerNotification = useCallback((error: AppError) => {
    window.dispatchEvent(new CustomEvent('app-error', { detail: error }));
  }, []);

  return { triggerNotification };
}

export default ErrorNotification;
