import React, { useState, useEffect } from 'react';
import { sessionManager } from '../services/SessionManager';
import { authService } from '../services/AuthService';
import { SessionEvent } from '../types/session';

interface AuthenticationChoiceProps {
  onChoiceMade?: (choice: 'anonymous' | 'authenticate') => void;
}

export const AuthenticationChoice: React.FC<AuthenticationChoiceProps> = ({ onChoiceMade }) => {
  const [isVisible, setIsVisible] = useState(false);
  const [sessionId, setSessionId] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    const handleSessionEvent = (event: SessionEvent) => {
      if (event.type === 'AUTHENTICATION_CHOICE_REQUIRED') {
        setSessionId(event.payload.sessionId);
        setIsVisible(true);
      } else if (event.type === 'AUTHENTICATION_REQUIRED') {
        // Redirect to login or show login modal
        setIsVisible(false);
        // You could emit a custom event or use a callback prop
        window.dispatchEvent(new CustomEvent('showLogin', { 
          detail: { sessionId: event.payload.sessionId } 
        }));
      }
    };

    sessionManager.addEventListener(handleSessionEvent);

    return () => {
      sessionManager.removeEventListener(handleSessionEvent);
    };
  }, []);

  const handleChoice = async (choice: 'anonymous' | 'authenticate') => {
    setIsLoading(true);
    
    try {
      sessionManager.setAuthenticationChoice(choice);
      
      if (choice === 'anonymous') {
        // Resume session creation with anonymous mode
        await sessionManager.resumeSessionCreation(sessionId);
      } else {
        // Show login modal/dialog
        setIsVisible(false);
        window.dispatchEvent(new CustomEvent('showLogin', { 
          detail: { sessionId } 
        }));
        return; // Don't close yet - wait for authentication
      }
      
      setIsVisible(false);
      onChoiceMade?.(choice);
    } catch (error) {
      console.error('Failed to handle authentication choice:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleLoginSuccessAfterAuth = async () => {
    try {
      await sessionManager.resumeSessionCreation(sessionId);
      setIsVisible(false);
      onChoiceMade?.('authenticate');
    } catch (error) {
      console.error('Failed to resume session after login:', error);
    }
  };

  // Listen for successful login
  useEffect(() => {
    const handleLoginSuccessEvent = (event: CustomEvent) => {
      if (event.detail.sessionId === sessionId) {
        handleLoginSuccessAfterAuth();
      }
    };

    window.addEventListener('loginSuccess', handleLoginSuccessEvent as EventListener);
    
    return () => {
      window.removeEventListener('loginSuccess', handleLoginSuccessEvent as EventListener);
    };
  }, [sessionId]);

  if (!isVisible) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
        <h2 className="text-xl font-bold mb-4">Choose Your Experience</h2>
        
        <p className="text-gray-600 mb-6">
          Would you like to practice anonymously or create an account to track your progress?
        </p>

        <div className="space-y-3">
          <button
            onClick={() => handleChoice('anonymous')}
            disabled={isLoading}
            className="w-full py-3 px-4 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors disabled:opacity-50"
          >
            <div className="font-semibold">Practice Anonymously</div>
            <div className="text-sm text-gray-600">
              No registration required, practice right away
            </div>
          </button>

          <button
            onClick={() => handleChoice('authenticate')}
            disabled={isLoading}
            className="w-full py-3 px-4 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition-colors disabled:opacity-50"
          >
            <div className="font-semibold">Create Account / Sign In</div>
            <div className="text-sm text-blue-100">
              Save progress, compete on leaderboards, unlock features
            </div>
          </button>
        </div>

        {isLoading && (
          <div className="mt-4 text-center text-gray-600">
            Setting up your session...
          </div>
        )}
      </div>
    </div>
  );
};

export default AuthenticationChoice;
