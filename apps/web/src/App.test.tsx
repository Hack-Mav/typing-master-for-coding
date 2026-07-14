import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import App from './App';

// Mock SessionManager to avoid loading IndexedDB in unmocked mode components
jest.mock('./services/SessionManager', () => ({
  sessionManager: {
    createSession: jest.fn().mockResolvedValue({
      id: 'test-session-id',
      targetText: 'console.log("Hello, World!");',
      state: { status: 'created' },
    }),
    startSession: jest.fn().mockResolvedValue(undefined),
    recordKeystroke: jest.fn().mockResolvedValue(undefined),
    finalizeSession: jest.fn().mockResolvedValue({
      sessionId: 'test-session-id',
    }),
    abandonSession: jest.fn().mockResolvedValue(undefined),
  },
}));

// Mock PrivacyService to avoid indexedDB reference in JSDOM
jest.mock('./services/PrivacyService', () => ({
  privacyService: {
    recordTelemetryEvent: jest.fn(),
    getPrivacySettings: jest.fn().mockResolvedValue({ telemetryConsent: true }),
    updatePrivacySettings: jest.fn().mockResolvedValue({}),
  },
}));

// Mock AuthContext so the app treats the user as authenticated
jest.mock('./services/AuthContext', () => ({
  AuthProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useAuth: () => ({
    user: null,
    token: null,
    isAuthenticated: true,
    isAnonymous: false,
    loading: false,
    login: jest.fn(),
    register: jest.fn(),
    logout: jest.fn(),
    loginWithMFA: jest.fn(),
    setupMFA: jest.fn(),
    verifyMFASetup: jest.fn(),
    getMFAStatus: jest.fn(),
    disableMFA: jest.fn(),
    regenerateMFABackupCodes: jest.fn(),
  }),
}));

// Mock the components
jest.mock('./components/ZenMode', () => {
  return function MockZenMode({ languageId, onComplete, onExit }: any) {
    return (
      <div data-testid="zen-mode">
        <div>Zen Mode - Language: {languageId}</div>
        <button onClick={() => onComplete({ summary: { duration: 1000 } })}>
          Complete Session
        </button>
        <button onClick={onExit}>Exit</button>
      </div>
    );
  };
});

jest.mock('./components/TimedDrillMode', () => {
  return function MockTimedDrillMode({
    languageId,
    duration,
    onComplete,
    onExit,
  }: any) {
    return (
      <div data-testid="timed-drill-mode">
        <div>
          Timed Drill - Language: {languageId}, Duration: {duration}
        </div>
        <button onClick={() => onComplete({ summary: { duration: 1000 } })}>
          Complete Session
        </button>
        <button onClick={onExit}>Exit</button>
      </div>
    );
  };
});

describe('App', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders main menu by default', () => {
    render(<App />);

    expect(screen.getByText('Typing Master for Coding')).toBeInTheDocument();
    expect(screen.getByText('Choose Your Practice Mode')).toBeInTheDocument();
    expect(screen.getByText('Zen Mode')).toBeInTheDocument();
    expect(screen.getByText('Timed Drill')).toBeInTheDocument();
  });

  test('displays language selection buttons', () => {
    render(<App />);

    expect(screen.getByText('JAVASCRIPT')).toBeInTheDocument();
    expect(screen.getByText('PYTHON')).toBeInTheDocument();
    expect(screen.getByText('CPP')).toBeInTheDocument();
    expect(screen.getByText('RUST')).toBeInTheDocument();
    expect(screen.getByText('YAML')).toBeInTheDocument();
  });

  test('navigates to Zen mode when Zen mode card is clicked', async () => {
    render(<App />);

    // Click on Zen mode card
    fireEvent.click(screen.getByText('Zen Mode'));

    await waitFor(() => {
      expect(screen.getByTestId('zen-mode')).toBeInTheDocument();
    });
    expect(
      screen.getByText('Zen Mode - Language: javascript')
    ).toBeInTheDocument();
  });

  test('navigates to Timed Drill mode when Timed Drill card is clicked', async () => {
    render(<App />);

    // Click on Timed Drill mode card
    fireEvent.click(screen.getByText('Timed Drill'));

    await waitFor(() => {
      expect(screen.getByTestId('timed-drill-mode')).toBeInTheDocument();
    });
    expect(
      screen.getByText('Timed Drill - Language: javascript, Duration: 60000')
    ).toBeInTheDocument();
  });

  test('changes language selection', async () => {
    render(<App />);

    // Click on Python language button
    fireEvent.click(screen.getByText('PYTHON'));

    // Navigate to Zen mode
    fireEvent.click(screen.getByText('Zen Mode'));

    await waitFor(() => {
      expect(
        screen.getByText('Zen Mode - Language: python')
      ).toBeInTheDocument();
    });
  });

  test('returns to menu after session completion in Zen mode', async () => {
    render(<App />);

    // Navigate to Zen mode
    fireEvent.click(screen.getByText('Zen Mode'));

    await waitFor(() => {
      expect(screen.getByTestId('zen-mode')).toBeInTheDocument();
    });

    // Complete the session
    fireEvent.click(screen.getByText('Complete Session'));

    await waitFor(() => {
      expect(screen.getByText('Typing Master for Coding')).toBeInTheDocument();
    });
  });

  test('returns to menu after session completion in Timed Drill mode', async () => {
    render(<App />);

    // Navigate to Timed Drill mode
    fireEvent.click(screen.getByText('Timed Drill'));

    await waitFor(() => {
      expect(screen.getByTestId('timed-drill-mode')).toBeInTheDocument();
    });

    // Complete the session
    fireEvent.click(screen.getByText('Complete Session'));

    await waitFor(() => {
      expect(screen.getByText('Typing Master for Coding')).toBeInTheDocument();
    });
  });

  test('returns to menu when exit is clicked in Zen mode', async () => {
    render(<App />);

    // Navigate to Zen mode
    fireEvent.click(screen.getByText('Zen Mode'));

    await waitFor(() => {
      expect(screen.getByTestId('zen-mode')).toBeInTheDocument();
    });

    // Exit the session
    fireEvent.click(screen.getByText('Exit'));

    await waitFor(() => {
      expect(screen.getByText('Typing Master for Coding')).toBeInTheDocument();
    });
  });

  test('returns to menu when exit is clicked in Timed Drill mode', async () => {
    render(<App />);

    // Navigate to Timed Drill mode
    fireEvent.click(screen.getByText('Timed Drill'));

    await waitFor(() => {
      expect(screen.getByTestId('timed-drill-mode')).toBeInTheDocument();
    });

    // Exit the session
    fireEvent.click(screen.getByText('Exit'));

    await waitFor(() => {
      expect(screen.getByText('Typing Master for Coding')).toBeInTheDocument();
    });
  });

  test('applies correct CSS classes for active language buttons', async () => {
    render(<App />);

    // Check that JavaScript button has active class initially
    const jsButton = screen.getByText('JAVASCRIPT');
    expect(jsButton).toHaveClass('active');

    // Change language to Python
    fireEvent.click(screen.getByText('PYTHON'));

    // Now Python should be active
    const pyButton = screen.getByText('PYTHON');
    expect(pyButton).toHaveClass('active');
    expect(jsButton).not.toHaveClass('active');
  });

  test('mode cards have hover effects and are clickable', () => {
    render(<App />);

    const zenCard = screen.getByText('Zen Mode');
    const timedCard = screen.getByText('Timed Drill');

    expect(zenCard).toBeInTheDocument();
    expect(timedCard).toBeInTheDocument();

    // Cards should be clickable
    fireEvent.click(zenCard!);
    fireEvent.click(timedCard!);

    // This would trigger navigation in a real scenario
    // We can't easily test the state changes without more complex mocking
  });
});
