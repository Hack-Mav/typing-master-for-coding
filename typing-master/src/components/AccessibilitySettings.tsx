import React, { useState, useEffect, createContext, useContext } from 'react';

export interface AccessibilityPreferences {
  reducedMotion: boolean;
  highContrast: boolean;
  fontSize: number;
  lineHeight: number;
  screenReaderOptimizations: boolean;
}

const defaultPreferences: AccessibilityPreferences = {
  reducedMotion: false,
  highContrast: false,
  fontSize: 16,
  lineHeight: 1.6,
  screenReaderOptimizations: true,
};

const AccessibilityContext = createContext<{
  preferences: AccessibilityPreferences;
  updatePreferences: (updates: Partial<AccessibilityPreferences>) => void;
}>({
  preferences: defaultPreferences,
  updatePreferences: () => {},
});

export const AccessibilityProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [preferences, setPreferences] = useState<AccessibilityPreferences>(
    () => {
      // Load from localStorage if available
      if (typeof window !== 'undefined') {
        const saved = localStorage.getItem('typing-master-accessibility');
        if (saved) {
          try {
            return { ...defaultPreferences, ...JSON.parse(saved) };
          } catch (e) {
            console.warn('Failed to parse accessibility preferences:', e);
          }
        }
      }
      return defaultPreferences;
    }
  );

  // Save to localStorage whenever preferences change
  useEffect(() => {
    if (typeof window !== 'undefined') {
      localStorage.setItem(
        'typing-master-accessibility',
        JSON.stringify(preferences)
      );
    }
  }, [preferences]);

  // Apply preferences to document
  useEffect(() => {
    if (typeof document !== 'undefined') {
      if (preferences.reducedMotion) {
        document.body.classList.add('reduce-motion');
      } else {
        document.body.classList.remove('reduce-motion');
      }

      if (preferences.highContrast) {
        document.body.classList.add('high-contrast');
      } else {
        document.body.classList.remove('high-contrast');
      }

      // Apply font size to root element
      document.documentElement.style.fontSize = `${preferences.fontSize}px`;

      // Apply line height to body
      document.body.style.lineHeight = preferences.lineHeight.toString();
    }
  }, [preferences]);

  const updatePreferences = (updates: Partial<AccessibilityPreferences>) => {
    setPreferences(prev => ({ ...prev, ...updates }));
  };

  return (
    <AccessibilityContext.Provider value={{ preferences, updatePreferences }}>
      {children}
    </AccessibilityContext.Provider>
  );
};

export const useAccessibility = () => {
  const context = useContext(AccessibilityContext);
  if (!context) {
    throw new Error(
      'useAccessibility must be used within an AccessibilityProvider'
    );
  }
  return context;
};

export const AccessibilitySettings: React.FC<{ className?: string }> = ({
  className = '',
}) => {
  const { preferences, updatePreferences } = useAccessibility();

  return (
    <div className={`accessibility-settings ${className}`}>
      <h3>Accessibility Settings</h3>

      <div className="setting-group">
        <label>
          <input
            type="checkbox"
            checked={preferences.reducedMotion}
            onChange={e =>
              updatePreferences({ reducedMotion: e.target.checked })
            }
          />
          Reduce Motion
        </label>
        <p className="setting-description">
          Reduces animations and transitions for better performance and comfort.
        </p>
      </div>

      <div className="setting-group">
        <label>
          <input
            type="checkbox"
            checked={preferences.highContrast}
            onChange={e =>
              updatePreferences({ highContrast: e.target.checked })
            }
          />
          High Contrast
        </label>
        <p className="setting-description">
          Increases contrast for better visibility and readability.
        </p>
      </div>

      <div className="setting-group">
        <label>
          Font Size: {preferences.fontSize}px
          <input
            type="range"
            min="12"
            max="24"
            value={preferences.fontSize}
            onChange={e =>
              updatePreferences({ fontSize: Number(e.target.value) })
            }
          />
        </label>
      </div>

      <div className="setting-group">
        <label>
          Line Height: {preferences.lineHeight}
          <input
            type="range"
            min="1.2"
            max="2.0"
            step="0.1"
            value={preferences.lineHeight}
            onChange={e =>
              updatePreferences({ lineHeight: Number(e.target.value) })
            }
          />
        </label>
      </div>

      <div className="setting-group">
        <label>
          <input
            type="checkbox"
            checked={preferences.screenReaderOptimizations}
            onChange={e =>
              updatePreferences({ screenReaderOptimizations: e.target.checked })
            }
          />
          Screen Reader Optimizations
        </label>
        <p className="setting-description">
          Enables additional announcements and optimizations for screen readers.
        </p>
      </div>
    </div>
  );
};

export default AccessibilityProvider;
