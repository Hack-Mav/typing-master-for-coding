import React, { useState, useEffect } from 'react';
import {
  privacyService,
  ConsentStatus,
  PrivacySettings as PrivacySettingsType,
} from '../../services/PrivacyService';
import { authService } from '../../services/AuthService';
import './Privacy.css';

interface PrivacySettingsProps {
  onClose?: () => void;
}

export const PrivacySettings: React.FC<PrivacySettingsProps> = ({
  onClose,
}) => {
  const [, setConsent] = useState<ConsentStatus | null>(null);
  const [privacyMode, setPrivacyMode] = useState(false);
  const [telemetryConsent, setTelemetryConsent] = useState(false);
  const [dataProcessingConsent, setDataProcessingConsent] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  useEffect(() => {
    loadConsentStatus();
  }, []);

  const loadConsentStatus = async () => {
    try {
      const status = await privacyService.getConsentStatus();
      setConsent(status);
      setPrivacyMode(status.privacyMode);
      setTelemetryConsent(status.telemetryConsent);
      setDataProcessingConsent(status.dataProcessingConsent);
    } catch (err) {
      setError('Failed to load privacy settings');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setError('');
    setSuccess('');
    setSaving(true);

    try {
      const settings: PrivacySettingsType = {
        privacyMode,
        telemetryConsent,
        dataProcessingConsent,
      };

      await privacyService.updatePrivacySettings(settings);
      setSuccess('Privacy settings updated successfully');

      // Reload consent status
      await loadConsentStatus();
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to update settings'
      );
    } finally {
      setSaving(false);
    }
  };

  const handleExportData = async (format: 'json' | 'csv') => {
    setError('');
    setSuccess('');

    try {
      await privacyService.downloadUserData(format);
      setSuccess(`Data exported successfully as ${format.toUpperCase()}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to export data');
    }
  };

  const handleDeleteData = async () => {
    setError('');
    setSuccess('');

    try {
      await privacyService.deleteUserData(true);
      setSuccess('All data deleted successfully. You will be logged out.');

      // Logout after a short delay
      setTimeout(() => {
        window.location.href = '/';
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete data');
    }
  };

  if (loading) {
    return (
      <div className="privacy-settings">
        <div className="loading">Loading privacy settings...</div>
      </div>
    );
  }

  const isAnonymous = authService.isAnonymous();

  return (
    <div className="privacy-settings">
      <div className="privacy-header">
        <h2>Privacy & Data Settings</h2>
        {onClose && (
          <button className="close-button" onClick={onClose} aria-label="Close">
            ×
          </button>
        )}
      </div>

      {error && <div className="error-message">{error}</div>}
      {success && <div className="success-message">{success}</div>}

      {isAnonymous ? (
        <div className="anonymous-notice">
          <h3>Anonymous Mode</h3>
          <p>
            You are using the app in anonymous mode. All your data is stored
            locally on your device and is never sent to our servers.
          </p>
          <p>
            To access cloud sync, leaderboards, and cross-device features,
            please create an account.
          </p>
        </div>
      ) : (
        <>
          <section className="privacy-section">
            <h3>Privacy Controls</h3>

            <div className="setting-item">
              <div className="setting-header">
                <label htmlFor="privacyMode">
                  <strong>Privacy Mode</strong>
                </label>
                <input
                  id="privacyMode"
                  type="checkbox"
                  checked={privacyMode}
                  onChange={e => setPrivacyMode(e.target.checked)}
                  disabled={saving}
                />
              </div>
              <p className="setting-description">
                When enabled, your sessions are stored locally only and not
                synced to the server. Leaderboards and social features will be
                disabled.
              </p>
            </div>

            <div className="setting-item">
              <div className="setting-header">
                <label htmlFor="telemetryConsent">
                  <strong>Telemetry & Analytics</strong>
                </label>
                <input
                  id="telemetryConsent"
                  type="checkbox"
                  checked={telemetryConsent}
                  onChange={e => setTelemetryConsent(e.target.checked)}
                  disabled={saving}
                />
              </div>
              <p className="setting-description">
                Help us improve the app by sending anonymized usage data. This
                includes practice modes used, session durations, and performance
                metrics (no personal information).
              </p>
            </div>

            <div className="setting-item">
              <div className="setting-header">
                <label htmlFor="dataProcessingConsent">
                  <strong>Data Processing</strong>
                </label>
                <input
                  id="dataProcessingConsent"
                  type="checkbox"
                  checked={dataProcessingConsent}
                  onChange={e => setDataProcessingConsent(e.target.checked)}
                  disabled={saving}
                />
              </div>
              <p className="setting-description">
                Required for cloud sync, leaderboards, and personalized
                features. We process your practice data to provide the service.
              </p>
            </div>

            <button
              className="btn-primary"
              onClick={handleSave}
              disabled={saving}
            >
              {saving ? 'Saving...' : 'Save Settings'}
            </button>
          </section>

          <section className="privacy-section">
            <h3>GDPR Data Rights</h3>

            <div className="gdpr-actions">
              <div className="gdpr-item">
                <h4>Export Your Data</h4>
                <p>
                  Download all your data in a portable format. Includes your
                  profile, sessions, results, and progress.
                </p>
                <div className="button-group">
                  <button
                    className="btn-secondary"
                    onClick={() => handleExportData('json')}
                  >
                    Export as JSON
                  </button>
                  <button
                    className="btn-secondary"
                    onClick={() => handleExportData('csv')}
                  >
                    Export as CSV
                  </button>
                </div>
              </div>

              <div className="gdpr-item danger-zone">
                <h4>Delete Your Data</h4>
                <p>
                  Permanently delete all your data from our servers. This action
                  cannot be undone. You will be logged out immediately.
                </p>

                {!showDeleteConfirm ? (
                  <button
                    className="btn-danger"
                    onClick={() => setShowDeleteConfirm(true)}
                  >
                    Delete All Data
                  </button>
                ) : (
                  <div className="delete-confirm">
                    <p className="warning-text">
                      <strong>Are you sure?</strong> This will permanently
                      delete all your data including your account, sessions, and
                      progress.
                    </p>
                    <div className="button-group">
                      <button className="btn-danger" onClick={handleDeleteData}>
                        Yes, Delete Everything
                      </button>
                      <button
                        className="btn-secondary"
                        onClick={() => setShowDeleteConfirm(false)}
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </section>

          <section className="privacy-section">
            <h3>Data We Collect</h3>
            <ul className="data-list">
              <li>
                <strong>Account Information:</strong> Email, username,
                preferences
              </li>
              <li>
                <strong>Practice Data:</strong> Sessions, typing metrics,
                progress
              </li>
              <li>
                <strong>Usage Analytics:</strong> Feature usage, session
                durations (only if telemetry is enabled)
              </li>
            </ul>
            <p className="data-notice">
              We never sell your data. All data is encrypted in transit and at
              rest. For more information, see our Privacy Policy.
            </p>
          </section>
        </>
      )}
    </div>
  );
};

export default PrivacySettings;
