import React, { useState } from 'react';
import { useAuth } from '../../services/AuthContext';

interface MFASetupProps {
  onSuccess?: () => void;
  onCancel?: () => void;
}

export const MFASetup: React.FC<MFASetupProps> = ({ onSuccess, onCancel }) => {
  const { setupMFA, verifyMFASetup } = useAuth();
  const [step, setStep] = useState<'setup' | 'verify' | 'success'>('setup');
  const [setupData, setSetupData] = useState<any>(null);
  const [verificationCode, setVerificationCode] = useState('');
  const [backupCodes, setBackupCodes] = useState<string[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSetup = async () => {
    setLoading(true);
    setError('');

    try {
      const data = await setupMFA();
      setSetupData(data);
      setBackupCodes(data.backup_codes);
      setStep('verify');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to setup MFA');
    } finally {
      setLoading(false);
    }
  };

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      await verifyMFASetup(verificationCode);
      setStep('success');
      onSuccess?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to verify MFA');
    } finally {
      setLoading(false);
    }
  };

  const downloadBackupCodes = () => {
    const content = backupCodes.join('\n');
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

  if (step === 'setup') {
    return (
      <div className="auth-form">
        <h2>Setup Multi-Factor Authentication</h2>

        {error && <div className="error-message">{error}</div>}

        <div className="mfa-info">
          <p>
            Multi-factor authentication adds an extra layer of security to your
            account. You'll need both your password and a code from your
            authenticator app to sign in.
          </p>

          <div className="mfa-steps">
            <h3>Steps to enable MFA:</h3>
            <ol>
              <li>
                Install an authenticator app (Google Authenticator, Authy, etc.)
              </li>
              <li>Click "Setup MFA" below</li>
              <li>Scan the QR code or enter the secret key manually</li>
              <li>Enter the 6-digit code from your app to verify</li>
            </ol>
          </div>
        </div>

        <div className="form-actions">
          <button
            type="button"
            className="btn-primary"
            onClick={handleSetup}
            disabled={loading}
          >
            {loading ? 'Setting up...' : 'Setup MFA'}
          </button>

          <button
            type="button"
            className="btn-secondary"
            onClick={onCancel}
            disabled={loading}
          >
            Cancel
          </button>
        </div>
      </div>
    );
  }

  if (step === 'verify') {
    return (
      <div className="auth-form">
        <h2>Verify MFA Setup</h2>

        {error && <div className="error-message">{error}</div>}

        {setupData && (
          <div className="mfa-setup-data">
            <div className="qr-code-section">
              <h3>Scan QR Code</h3>
              <p>Scan this code with your authenticator app:</p>
              <img
                src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(setupData.qr_code_url)}`}
                alt="MFA QR Code"
                className="qr-code"
              />
            </div>

            <div className="manual-entry-section">
              <h3>Or Enter Manually</h3>
              <p>Enter this code in your authenticator app:</p>
              <div className="secret-key">
                <code>{setupData.secret}</code>
                <button
                  type="button"
                  className="copy-button"
                  onClick={() =>
                    navigator.clipboard.writeText(setupData.secret)
                  }
                >
                  Copy
                </button>
              </div>
            </div>
          </div>
        )}

        <form onSubmit={handleVerify}>
          <div className="form-group">
            <label htmlFor="mfa-code">6-Digit Code</label>
            <input
              id="mfa-code"
              type="text"
              value={verificationCode}
              onChange={e =>
                setVerificationCode(
                  e.target.value.replace(/\D/g, '').slice(0, 6)
                )
              }
              placeholder="000000"
              maxLength={6}
              required
              disabled={loading}
              autoComplete="one-time-code"
            />
          </div>

          <button
            type="submit"
            className="btn-primary"
            disabled={loading || verificationCode.length !== 6}
          >
            {loading ? 'Verifying...' : 'Verify & Enable MFA'}
          </button>
        </form>
      </div>
    );
  }

  if (step === 'success') {
    return (
      <div className="auth-form">
        <h2>MFA Setup Complete!</h2>

        <div className="success-message">
          <p>
            ✅ Multi-factor authentication has been successfully enabled for
            your account.
          </p>
        </div>

        <div className="backup-codes-section">
          <h3>⚠️ Save Your Backup Codes</h3>
          <p>
            These codes can be used to access your account if you lose access to
            your authenticator app.
            <strong>
              {' '}
              Save them in a safe place - you won't see them again!
            </strong>
          </p>

          <div className="backup-codes">
            {backupCodes.map((code, index) => (
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
                navigator.clipboard.writeText(backupCodes.join('\n'))
              }
            >
              Copy All Codes
            </button>
          </div>
        </div>

        <div className="form-actions">
          <button type="button" className="btn-primary" onClick={onSuccess}>
            Continue
          </button>
        </div>
      </div>
    );
  }

  return null;
};

export default MFASetup;
