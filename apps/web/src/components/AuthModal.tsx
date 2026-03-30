import React, { useState } from 'react';
import { LoginForm } from './Auth/LoginForm';
import { RegisterForm } from './Auth/RegisterForm';

export type AuthMode = 'login' | 'register';

interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
  initialMode?: AuthMode;
  onAuthenticated: () => void;
}

export const AuthModal: React.FC<AuthModalProps> = ({
  isOpen,
  onClose,
  initialMode = 'login',
  onAuthenticated,
}) => {
  const [mode, setMode] = useState<AuthMode>(initialMode);

  if (!isOpen) return null;

  const handleLoginSuccess = () => {
    onAuthenticated();
    onClose();
  };

  const handleRegisterSuccess = () => {
    onAuthenticated();
    onClose();
  };

  const handleAnonymousSuccess = () => {
    onAuthenticated();
    onClose();
  };

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50 p-5"
      onClick={onClose}
    >
      <div
        className="bg-slate-900/95 backdrop-blur-sm border border-slate-700 rounded-lg shadow-xl w-full max-w-md h-auto max-h-[90vh] flex flex-col"
        onClick={e => e.stopPropagation()}
      >
        <div className="flex justify-end p-4">
          <button
            className="text-slate-400 hover:text-slate-200 text-2xl leading-none transition-colors"
            onClick={onClose}
          >
            ×
          </button>
        </div>

        <div className="px-8 pb-8 flex-grow overflow-y-auto">
          {mode === 'login' ? (
            <LoginForm
              onSuccess={handleLoginSuccess}
              onSwitchToRegister={() => setMode('register')}
              onAnonymousMode={handleAnonymousSuccess}
            />
          ) : (
            <RegisterForm
              onSuccess={handleRegisterSuccess}
              onSwitchToLogin={() => setMode('login')}
            />
          )}
        </div>
      </div>
    </div>
  );
};

export default AuthModal;
