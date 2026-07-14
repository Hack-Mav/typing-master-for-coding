import React, { useState } from 'react';
import { authService, LoginRequest } from '../../services/AuthService';
import { useAuth } from '../../services/AuthContext';
import { MFAVerification } from './MFAVerification';
import { MFABackupCode } from './MFABackupCode';
import './Auth.css';

interface LoginFormProps {
  onSuccess?: () => void;
  onSwitchToRegister?: () => void;
  onAnonymousMode?: () => void;
}

type AuthStep = 'login' | 'mfa' | 'backup-code';

export const LoginForm: React.FC<LoginFormProps> = ({
  onSuccess,
  onSwitchToRegister,
  onAnonymousMode,
}) => {
  const { anonymousLogin } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [authStep, setAuthStep] = useState<AuthStep>('login');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const request: LoginRequest = { email, password };
      const response = await authService.login(request);

      // Check if MFA is required
      if (response.requires_mfa) {
        // MFA user ID is handled by the MFA verification component
        setAuthStep('mfa');
        return;
      }

      // MFA not required - login successful
      onSuccess?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  const handleMFASuccess = () => {
    setAuthStep('login');
    onSuccess?.();
  };

  const handleMFACancel = () => {
    setAuthStep('login');
  };

  const handleAnonymousMode = async () => {
    setError('');
    setLoading(true);

    try {
      const deviceId = localStorage.getItem('device_id') || generateDeviceId();
      localStorage.setItem('device_id', deviceId);

      await anonymousLogin(deviceId, 'QWERTY', navigator.language);

      onAnonymousMode?.();
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to start anonymous session'
      );
    } finally {
      setLoading(false);
    }
  };

  // Show MFA verification step
  if (authStep === 'mfa') {
    return (
      <MFAVerification
        email={email}
        password={password}
        onSuccess={handleMFASuccess}
        onCancel={handleMFACancel}
        onUseBackupCode={() => setAuthStep('backup-code')}
      />
    );
  }

  // Show backup code step
  if (authStep === 'backup-code') {
    return (
      <MFABackupCode
        email={email}
        password={password}
        onSuccess={handleMFASuccess}
        onCancel={handleMFACancel}
        onBackToApp={() => setAuthStep('mfa')}
      />
    );
  }

  return (
    <div className="auth-form">
      <h2>Login</h2>

      {error && <div className="error-message">{error}</div>}

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="email">Email</label>
          <input
            id="email"
            type="email"
            value={email}
            onChange={e => setEmail(e.target.value)}
            required
            disabled={loading}
            autoComplete="email"
          />
        </div>

        <div className="form-group">
          <label htmlFor="password">Password</label>
          <input
            id="password"
            type="password"
            value={password}
            onChange={e => setPassword(e.target.value)}
            required
            disabled={loading}
            autoComplete="current-password"
          />
        </div>

        <button type="submit" className="btn-primary" disabled={loading}>
          {loading ? 'Logging in...' : 'Login'}
        </button>
      </form>

      <div className="auth-divider">
        <span>OR</span>
      </div>

      <button
        type="button"
        className="btn-secondary"
        onClick={handleAnonymousMode}
        disabled={loading}
      >
        Continue Anonymously
      </button>

      <div className="auth-footer">
        <p>
          Don't have an account?{' '}
          <button
            type="button"
            className="link-button"
            onClick={onSwitchToRegister}
            disabled={loading}
          >
            Register
          </button>
        </p>
      </div>

      <div className="privacy-notice">
        <p>
          <strong>Anonymous Mode:</strong> All your data stays on your device.
          No account required, no data sent to servers.
        </p>
      </div>
    </div>
  );
};

function generateDeviceId(): string {
  return `device_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

export default LoginForm;
