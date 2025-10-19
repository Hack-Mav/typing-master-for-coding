import React, { useState } from 'react';
import { authService, RegisterRequest } from '../../services/AuthService';
import './Auth.css';

interface RegisterFormProps {
  onSuccess?: () => void;
  onSwitchToLogin?: () => void;
}

export const RegisterForm: React.FC<RegisterFormProps> = ({
  onSuccess,
  onSwitchToLogin,
}) => {
  const [handle, setHandle] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [telemetryConsent, setTelemetryConsent] = useState(false);
  const [dataProcessingConsent, setDataProcessingConsent] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // Validation
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }

    if (!dataProcessingConsent) {
      setError('You must consent to data processing to create an account');
      return;
    }

    setLoading(true);

    try {
      const request: RegisterRequest = {
        handle,
        email,
        password,
        locale: navigator.language,
        keyboardLayout: 'QWERTY',
        telemetryConsent,
        dataProcessingConsent,
      };

      await authService.register(request);
      onSuccess?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-form">
      <h2>Create Account</h2>
      
      {error && <div className="error-message">{error}</div>}
      
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="handle">Username</label>
          <input
            id="handle"
            type="text"
            value={handle}
            onChange={(e) => setHandle(e.target.value)}
            required
            minLength={3}
            maxLength={30}
            disabled={loading}
            autoComplete="username"
          />
          <small>3-30 characters</small>
        </div>

        <div className="form-group">
          <label htmlFor="email">Email</label>
          <input
            id="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
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
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
            disabled={loading}
            autoComplete="new-password"
          />
          <small>Minimum 8 characters</small>
        </div>

        <div className="form-group">
          <label htmlFor="confirmPassword">Confirm Password</label>
          <input
            id="confirmPassword"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
            disabled={loading}
            autoComplete="new-password"
          />
        </div>

        <div className="consent-section">
          <h3>Privacy & Consent</h3>
          
          <div className="checkbox-group">
            <input
              id="dataProcessing"
              type="checkbox"
              checked={dataProcessingConsent}
              onChange={(e) => setDataProcessingConsent(e.target.checked)}
              disabled={loading}
            />
            <label htmlFor="dataProcessing">
              <strong>I consent to data processing (Required)</strong>
              <small>
                We will store your account information, practice sessions, and results
                to provide the service. You can export or delete your data at any time.
              </small>
            </label>
          </div>

          <div className="checkbox-group">
            <input
              id="telemetry"
              type="checkbox"
              checked={telemetryConsent}
              onChange={(e) => setTelemetryConsent(e.target.checked)}
              disabled={loading}
            />
            <label htmlFor="telemetry">
              <strong>Help improve the app (Optional)</strong>
              <small>
                Send anonymized usage data to help us improve the application.
                You can change this setting at any time.
              </small>
            </label>
          </div>
        </div>

        <button type="submit" className="btn-primary" disabled={loading}>
          {loading ? 'Creating Account...' : 'Create Account'}
        </button>
      </form>

      <div className="auth-footer">
        <p>
          Already have an account?{' '}
          <button
            type="button"
            className="link-button"
            onClick={onSwitchToLogin}
            disabled={loading}
          >
            Login
          </button>
        </p>
      </div>

      <div className="privacy-notice">
        <p>
          By creating an account, you agree to our Terms of Service and Privacy Policy.
          Your data is protected according to GDPR regulations.
        </p>
      </div>
    </div>
  );
};

export default RegisterForm;
