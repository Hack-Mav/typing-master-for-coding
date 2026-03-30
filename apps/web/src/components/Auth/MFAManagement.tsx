import React, { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../../services/AuthContext';

interface MFAManagementProps {
  onClose?: () => void;
}

export const MFAManagement: React.FC<MFAManagementProps> = ({ onClose }) => {
  const { getMFAStatus, disableMFA, regenerateMFABackupCodes } = useAuth();

  const [mfaStatus, setMfaStatus] = useState<any>(null);
  const [showDisableForm, setShowDisableForm] = useState(false);
  const [showBackupCodes, setShowBackupCodes] = useState(false);
  const [disablePassword, setDisablePassword] = useState('');
  const [disableCode, setDisableCode] = useState('');
  const [regenerateCode, setRegenerateCode] = useState('');
  const [newBackupCodes, setNewBackupCodes] = useState<string[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const loadMFAStatus = useCallback(async () => {
    try {
      const status = await getMFAStatus();
      setMfaStatus(status);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to load MFA status'
      );
    }
  }, [getMFAStatus]);

  useEffect(() => {
    loadMFAStatus();
  }, [loadMFAStatus]);

  const handleDisableMFA = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await disableMFA(disablePassword, disableCode);
      setMfaStatus({ enabled: false });
      setShowDisableForm(false);
      setDisablePassword('');
      setDisableCode('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to disable MFA');
    } finally {
      setLoading(false);
    }
  };

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const handleRegenerateBackupCodes = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const result = await regenerateMFABackupCodes(regenerateCode);
      setNewBackupCodes(result.backup_codes);
      setShowBackupCodes(true);
      setRegenerateCode('');
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to regenerate backup codes'
      );
    } finally {
      setLoading(false);
    }
  };

  const downloadBackupCodes = () => {
    const content = newBackupCodes.join('\n');
    const blob = new Blob([content], { type: 'text/plain' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'mfa-backup-codes.txt';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  };

  if (!mfaStatus) {
    return (
      <div className="auth-form">
        <h2>MFA Management</h2>
        {error && <div className="error-message">{error}</div>}
        <div className="loading">Loading MFA status...</div>
      </div>
    );
  }

  if (showBackupCodes && newBackupCodes.length > 0) {
    return (
      <div className="auth-form">
        <h2>New Backup Codes Generated</h2>

        <div className="success-message">
          <p>✅ Your backup codes have been regenerated successfully.</p>
        </div>

        <div className="backup-codes-section">
          <h3>⚠️ Save Your New Backup Codes</h3>
          <p>
            These codes replace your old backup codes. The old codes will no
            longer work.
            <strong> Save them in a safe place!</strong>
          </p>

          <div className="backup-codes">
            {newBackupCodes.map((code, index) => (
              <div key={index} className="backup-code">
                {code}
              </div>
            ))}
          </div>

          <div className="backup-actions">
            <button
              type="button"
              className="btn-primary"
              onClick={downloadBackupCodes}
            >
              Download Backup Codes
            </button>

            <button
              type="button"
              className="btn-secondary"
              onClick={() =>
                navigator.clipboard.writeText(newBackupCodes.join('\n'))
              }
            >
              Copy All Codes
            </button>
          </div>
        </div>

        <div className="form-actions">
          <button
            type="button"
            className="btn-primary"
            onClick={() => setShowBackupCodes(false)}
          >
            Done
          </button>
        </div>
      </div>
    );
  }

  if (showDisableForm) {
    return (
      <div className="auth-form">
        <h2>Disable MFA</h2>

        {error && <div className="error-message">{error}</div>}

        <div className="warning-message">
          <p>
            <strong>Warning:</strong> Disabling MFA will make your account less
            secure. You'll lose the extra protection that MFA provides.
          </p>
        </div>

        <form onSubmit={handleDisableMFA}>
          <div className="form-group">
            <label htmlFor="disable-password">Confirm Password</label>
            <input
              id="disable-password"
              type="password"
              value={disablePassword}
              onChange={e => setDisablePassword(e.target.value)}
              placeholder="Enter your password"
              required
              disabled={loading}
              autoComplete="current-password"
            />
          </div>

          <div className="form-group">
            <label htmlFor="disable-code">Authenticator Code</label>
            <input
              id="disable-code"
              type="text"
              value={disableCode}
              onChange={e =>
                setDisableCode(e.target.value.replace(/\D/g, '').slice(0, 6))
              }
              placeholder="000000"
              maxLength={6}
              required
              disabled={loading}
              autoComplete="one-time-code"
            />
          </div>

          <div className="form-actions">
            <button type="submit" className="btn-danger" disabled={loading}>
              {loading ? 'Disabling...' : 'Disable MFA'}
            </button>

            <button
              type="button"
              className="btn-secondary"
              onClick={() => setShowDisableForm(false)}
              disabled={loading}
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    );
  }

  return (
    <div className="auth-form">
      <div className="mfa-header">
        <h2>Multi-Factor Authentication</h2>
        {onClose && (
          <button type="button" className="close-button" onClick={onClose}>
            ✕
          </button>
        )}
      </div>

      <div className="mfa-status">
        <div
          className={`status-indicator ${mfaStatus.enabled ? 'enabled' : 'disabled'}`}
        >
          <span className="status-dot"></span>
          <span className="status-text">
            {mfaStatus.enabled ? 'Enabled' : 'Disabled'}
          </span>
        </div>

        {mfaStatus.enabled && mfaStatus.setup_at && (
          <div className="setup-date">
            Enabled on {new Date(mfaStatus.setup_at).toLocaleDateString()}
          </div>
        )}
      </div>

      {error && <div className="error-message">{error}</div>}

      {mfaStatus.enabled ? (
        <div className="mfa-actions">
          <div className="info-section">
            <h3>✅ MFA is Active</h3>
            <p>
              Your account is protected with multi-factor authentication. You'll
              need both your password and authenticator code to sign in.
            </p>
          </div>

          <div className="action-buttons">
            <button
              type="button"
              className="btn-secondary"
              onClick={() => setShowDisableForm(true)}
              disabled={loading}
            >
              Disable MFA
            </button>

            <button
              type="button"
              className="btn-secondary"
              onClick={() => setShowBackupCodes(true)}
              disabled={loading}
            >
              Regenerate Backup Codes
            </button>
          </div>
        </div>
      ) : (
        <div className="mfa-actions">
          <div className="info-section">
            <h3>⚠️ MFA is Disabled</h3>
            <p>
              Your account is only protected by your password. Enable MFA for
              enhanced security.
            </p>
          </div>

          <div className="action-buttons">
            <button
              type="button"
              className="btn-primary"
              onClick={onClose}
              disabled={loading}
            >
              Enable MFA
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default MFAManagement;
