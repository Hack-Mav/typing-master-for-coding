import React, { useState } from 'react';
import { authService, LoginRequest } from '../../services/AuthService';
import './Auth.css';

interface LoginFormProps {
  onSuccess?: () => void;
  onSwitchToRegister?: () => void;
  onAnonymousMode?: () => void;
}

export const LoginForm: React.FC<LoginFormProps> = ({
  onSuccess,
  onSwitchToRegister,
  onAnonymousMode,
}) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const request: LoginRequest = { email, password };
      await authService.login(request);
      onSuccess?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  const handleAnonymousMode = async () => {
    setError('');
    setLoading(true);

    try {
      const deviceId = localStorage.getItem('device_id') || generateDeviceId();
      localStorage.setItem('device_id', deviceId);

      await authService.createAnonymousSession({
        deviceId,
        keyboardLayout: 'QWERTY',
        locale: navigator.language,
      });

      onAnonymousMode?.();
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to start anonymous session'
      );
    } finally {
      setLoading(false);
    }
  };

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
