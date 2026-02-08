import React, { useState } from 'react';
import { useAuth } from '../../services/AuthContext';

interface MFAVerificationProps {
  email: string;
  password: string;
  onSuccess?: () => void;
  onCancel?: () => void;
  onUseBackupCode?: () => void;
}

export const MFAVerification: React.FC<MFAVerificationProps> = ({
  email,
  password,
  onSuccess,
  onCancel,
  onUseBackupCode,
}) => {
  const { loginWithMFA } = useAuth();
  const [code, setCode] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [attempts, setAttempts] = useState(0);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await loginWithMFA(email, password, code);
      onSuccess?.();
    } catch (err) {
      setAttempts(prev => prev + 1);
      setError(err instanceof Error ? err.message : 'Invalid code');

      // Show backup code option after 3 failed attempts
      if (attempts >= 2) {
        onUseBackupCode?.();
      }
    } finally {
      setLoading(false);
    }
  };

  const handleCodeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value.replace(/\D/g, '');
    if (value.length <= 6) {
      setCode(value);
    }
  };

  return (
    <div className="auth-form">
      <h2>Two-Factor Authentication</h2>

      {error && <div className="error-message">{error}</div>}

      <div className="mfa-info">
        <p>
          Enter the 6-digit code from your authenticator app to complete
          sign-in.
        </p>

        {attempts > 0 && (
          <div className="warning-message">
            {attempts >= 3 ? (
              <p>
                Too many failed attempts. Please use a backup code or try again
                later.
              </p>
            ) : (
              <p>Attempt {attempts + 1} of 3</p>
            )}
          </div>
        )}
      </div>

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="mfa-code">Authenticator Code</label>
          <input
            id="mfa-code"
            type="text"
            value={code}
            onChange={handleCodeChange}
            placeholder="000000"
            maxLength={6}
            required
            disabled={loading}
            autoComplete="one-time-code"
            className="mfa-input"
          />
        </div>

        <button
          type="submit"
          className="btn-primary"
          disabled={loading || code.length !== 6}
        >
          {loading ? 'Verifying...' : 'Verify & Sign In'}
        </button>
      </form>

      <div className="auth-divider">
        <span>OR</span>
      </div>

      {attempts >= 1 && (
        <button
          type="button"
          className="btn-secondary backup-code-btn"
          onClick={onUseBackupCode}
          disabled={loading}
        >
          Use Backup Code
        </button>
      )}

      <div className="auth-footer">
        <button
          type="button"
          className="link-button"
          onClick={onCancel}
          disabled={loading}
        >
          Back to Login
        </button>
      </div>

      <div className="mfa-help">
        <p>
          <strong>Don't have your authenticator app?</strong>
          <br />
          Make sure your device time is set correctly, or use a backup code.
        </p>
      </div>
    </div>
  );
};

export default MFAVerification;
