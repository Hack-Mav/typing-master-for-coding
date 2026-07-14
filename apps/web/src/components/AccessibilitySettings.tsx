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

const FONT_SIZE_MIN = 10;
const FONT_SIZE_MAX = 100;
const LINE_HEIGHT_MIN = 1;
const LINE_HEIGHT_MAX = 3;

function validatePreferences(value: unknown): AccessibilityPreferences | null {
  if (typeof value !== 'object' || value === null) return null;
  const v = value as Record<string, unknown>;
  const parsed: Partial<AccessibilityPreferences> = {};

  if (typeof v.reducedMotion === 'boolean') {
    parsed.reducedMotion = v.reducedMotion;
  }
  if (typeof v.highContrast === 'boolean') {
    parsed.highContrast = v.highContrast;
  }
  if (typeof v.screenReaderOptimizations === 'boolean') {
    parsed.screenReaderOptimizations = v.screenReaderOptimizations;
  }
  if (
    typeof v.fontSize === 'number' &&
    v.fontSize >= FONT_SIZE_MIN &&
    v.fontSize <= FONT_SIZE_MAX
  ) {
    parsed.fontSize = v.fontSize;
  }
  if (
    typeof v.lineHeight === 'number' &&
    v.lineHeight >= LINE_HEIGHT_MIN &&
    v.lineHeight <= LINE_HEIGHT_MAX
  ) {
    parsed.lineHeight = v.lineHeight;
  }

  return { ...defaultPreferences, ...parsed };
}

function getSystemPreferences(): Partial<AccessibilityPreferences> {
  if (typeof window === 'undefined') return {};
  return {
    reducedMotion:
      window.matchMedia?.('(prefers-reduced-motion: reduce)')?.matches || false,
    highContrast:
      window.matchMedia?.('(prefers-contrast: more)')?.matches || false,
  };
}

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
      if (typeof window === 'undefined') return defaultPreferences;

      const saved = localStorage.getItem('typing-master-accessibility');
      if (saved) {
        try {
          const parsed = JSON.parse(saved);
          const validated = validatePreferences(parsed);
          if (validated) return validated;
          console.warn('Invalid accessibility preferences; using defaults');
          localStorage.removeItem('typing-master-accessibility');
        } catch (e) {
          console.warn('Failed to parse accessibility preferences:', e);
        }
      }

      return { ...defaultPreferences, ...getSystemPreferences() };
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
      document.body.classList.toggle(
        'reduced-motion',
        preferences.reducedMotion
      );
      document.body.classList.toggle('high-contrast', preferences.highContrast);
    }
  }, [preferences]);

  // Respect system accessibility preferences when no user override is saved
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const reduce = window.matchMedia('(prefers-reduced-motion: reduce)');
    const contrast = window.matchMedia('(prefers-contrast: more)');

    const update = () => {
      if (localStorage.getItem('typing-master-accessibility')) return;
      setPreferences(prev => ({ ...prev, ...getSystemPreferences() }));
    };

    update();
    reduce.addEventListener('change', update);
    contrast.addEventListener('change', update);
    return () => {
      reduce.removeEventListener('change', update);
      contrast.removeEventListener('change', update);
    };
  }, []);

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
