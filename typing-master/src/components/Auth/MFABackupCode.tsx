import React, { useState } from 'react';
import { useAuth } from '../../services/AuthContext';

interface MFABackupCodeProps {
  email: string;
  password: string;
  onSuccess?: () => void;
  onCancel?: () => void;
  onBackToApp?: () => void;
}

export const MFABackupCode: React.FC<MFABackupCodeProps> = ({
  email,
  password,
  onSuccess,
  onCancel,
  onBackToApp,
}) => {
  const { loginWithMFA } = useAuth();
  const [backupCode, setBackupCode] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      // Clean the backup code (remove spaces, hyphens, convert to uppercase)
      const cleanCode = backupCode.replace(/[-\s]/g, '').toUpperCase();
      await loginWithMFA(email, password, cleanCode);
      onSuccess?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Invalid backup code');
    } finally {
      setLoading(false);
    }
  };

  const handleCodeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    let value = e.target.value.replace(/[^A-Za-z0-9-]/g, '').toUpperCase();
    // Add hyphen after 4 characters if it's a full backup code
    if (value.length === 8 && !value.includes('-')) {
      value = value.slice(0, 4) + '-' + value.slice(4);
    }
    setBackupCode(value.slice(0, 9)); // Max length: XXXX-XXXX
  };

  return (
    <div className="auth-form">
      <h2>Use Backup Code</h2>

      {error && <div className="error-message">{error}</div>}

      <div className="mfa-info">
        <p>
          Enter one of your backup codes to complete sign-in. Each backup code
          can only be used once.
        </p>

        <div className="backup-code-format">
          <strong>Format:</strong> XXXX-XXXX (8 characters total)
        </div>
      </div>

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="backup-code">Backup Code</label>
          <input
            id="backup-code"
            type="text"
            value={backupCode}
            onChange={handleCodeChange}
            placeholder="XXXX-XXXX"
            maxLength={9}
            required
            disabled={loading}
            autoComplete="off"
            className="backup-code-input"
          />
        </div>

        <button
          type="submit"
          className="btn-primary"
          disabled={loading || backupCode.replace(/[-\s]/g, '').length !== 8}
        >
          {loading ? 'Verifying...' : 'Verify & Sign In'}
        </button>
      </form>

      <div className="auth-footer">
        <button
          type="button"
          className="link-button"
          onClick={onBackToApp}
          disabled={loading}
        >
          Try Authenticator App Again
        </button>

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
          <strong>Lost your backup codes?</strong>
          <br />
          You'll need to contact support or reset your MFA through your account
          settings.
        </p>
      </div>
    </div>
  );
};

export default MFABackupCode;
