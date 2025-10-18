import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import TimedDrillMode from '../TimedDrillMode';

// Mock the SessionManager
jest.mock('../../services/SessionManager', () => ({
  sessionManager: {
    createSession: jest.fn().mockResolvedValue({
      id: 'test-drill-session-id',
      targetText: 'function hello() { return "world"; }',
      config: { mode: 'timed-drill' },
      state: { status: 'created' },
      events: [],
      progress: {
        percentComplete: 0,
        charactersTyped: 0,
        currentSpeed: 0,
        currentAccuracy: 100,
        errorsCount: 0,
      },
    }),
    startSession: jest.fn().mockResolvedValue(undefined),
    recordKeystroke: jest.fn().mockResolvedValue(undefined),
    getSession: jest.fn().mockResolvedValue({
      id: 'test-drill-session-id',
      progress: {
        currentSpeed: 45,
        currentAccuracy: 95,
        errorsCount: 2,
      },
    }),
    pauseSession: jest.fn().mockResolvedValue(undefined),
    resumeSession: jest.fn().mockResolvedValue(undefined),
    finalizeSession: jest.fn().mockResolvedValue({
      sessionId: 'test-drill-session-id',
      metrics: {
        speed: { twpm: 45, cpm: 225 },
        accuracy: { overallAccuracy: 0.95 },
      },
      summary: {
        mode: 'timed-drill',
        language: 'javascript',
        duration: 60000,
        completionRate: 85,
        finalSpeed: 45,
        finalAccuracy: 95,
        errorCount: 2,
        improvementAreas: [],
        strengths: ['good_speed'],
      },
    }),
    abandonSession: jest.fn().mockResolvedValue(undefined),
  },
}));

// Mock timers
jest.useFakeTimers();

describe('TimedDrillMode', () => {
  const defaultProps = {
    languageId: 'javascript',
    duration: 60000, // 1 minute
    onComplete: jest.fn(),
    onExit: jest.fn(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
    jest.clearAllTimers();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
    jest.useFakeTimers();
  });

  test('renders loading state initially', () => {
    render(<TimedDrillMode {...defaultProps} />);

    expect(
      screen.getByText('Preparing your timed drill...')
    ).toBeInTheDocument();
  });

  test('creates session with correct timed drill config', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<TimedDrillMode {...defaultProps} />);

    await waitFor(() => {
      expect(sessionManager.createSession).toHaveBeenCalledWith(
        expect.objectContaining({
          mode: 'timed-drill',
          languageId: 'javascript',
          duration: 60000,
          settings: expect.objectContaining({
            showMetrics: true,
            showTimer: true,
            showProgress: true,
          }),
        })
      );
    });
  });

  test('displays timer and HUD elements', async () => {
    render(<TimedDrillMode {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('1:00')).toBeInTheDocument(); // Timer display
    });
    expect(screen.getByText('Time Remaining')).toBeInTheDocument();
    expect(screen.getByText('CPM')).toBeInTheDocument();
    expect(screen.getByText('tWPM')).toBeInTheDocument();
    expect(screen.getByText('Accuracy')).toBeInTheDocument();
    expect(screen.getByText('Errors')).toBeInTheDocument();
  });

  test('handles pause and resume functionality', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<TimedDrillMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByText('function hello() { return "world"; }')
      ).toBeInTheDocument();
    });

    // Find pause button
    const pauseButton = screen.getByTitle(/Pause/);
    fireEvent.click(pauseButton);

    expect(sessionManager.pauseSession).toHaveBeenCalledWith(
      'test-drill-session-id'
    );

    // Should show pause overlay
    await waitFor(() => {
      expect(screen.getByText('Paused')).toBeInTheDocument();
    });

    // Resume
    const resumeButton = screen.getByText('Resume');
    fireEvent.click(resumeButton);

    expect(sessionManager.resumeSession).toHaveBeenCalledWith(
      'test-drill-session-id'
    );
  });

  test('handles restart functionality', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<TimedDrillMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByText('function hello() { return "world"; }')
      ).toBeInTheDocument();
    });

    const restartButton = screen.getByTitle('Restart');
    fireEvent.click(restartButton);

    expect(sessionManager.abandonSession).toHaveBeenCalledWith(
      'test-drill-session-id',
      'user_restart'
    );
  });

  test('handles exit functionality', async () => {
    render(<TimedDrillMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByText('function hello() { return "world"; }')
      ).toBeInTheDocument();
    });

    const exitButton = screen.getByTitle('Exit');
    fireEvent.click(exitButton);

    expect(defaultProps.onExit).toHaveBeenCalled();
  });

  test('handles typing input and updates metrics', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<TimedDrillMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByText('function hello() { return "world"; }')
      ).toBeInTheDocument();
    });

    const textarea = screen.getByRole('textbox');
    expect(textarea).toBeInTheDocument();

    // Simulate typing 'f'
    fireEvent.keyDown(textarea, {
      key: 'f',
      code: 'KeyF',
    });

    expect(sessionManager.recordKeystroke).toHaveBeenCalledWith(
      'test-drill-session-id',
      expect.objectContaining({
        key: 'f',
        code: 'KeyF',
        action: 'keydown',
      })
    );

    expect(sessionManager.getSession).toHaveBeenCalledWith(
      'test-drill-session-id'
    );
  });

  test('handles escape key for pause', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<TimedDrillMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByText('function hello() { return "world"; }')
      ).toBeInTheDocument();
    });

    const textarea = screen.getByRole('textbox');

    fireEvent.keyDown(textarea, {
      key: 'Escape',
      code: 'Escape',
    });

    expect(sessionManager.pauseSession).toHaveBeenCalledWith(
      'test-drill-session-id'
    );
  });

  test('shows results screen when time is up', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<TimedDrillMode {...defaultProps} />);

    // Wait for component to initialize
    await waitFor(() => {
      expect(sessionManager.createSession).toHaveBeenCalled();
    });

    // Fast forward time to trigger completion (60 seconds)
    jest.advanceTimersByTime(60000);

    await waitFor(() => {
      expect(sessionManager.finalizeSession).toHaveBeenCalledWith(
        'test-drill-session-id'
      );
    });

    await waitFor(() => {
      expect(screen.getByText('Drill Complete!')).toBeInTheDocument();
    });
    expect(screen.getByText('45')).toBeInTheDocument(); // tWPM
    expect(screen.getByText('95%')).toBeInTheDocument(); // Accuracy
    expect(screen.getByText('Try Again')).toBeInTheDocument();
  });

  test('supports different durations', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<TimedDrillMode {...defaultProps} duration={300000} />); // 5 minutes

    await waitFor(() => {
      expect(sessionManager.createSession).toHaveBeenCalledWith(
        expect.objectContaining({
          duration: 300000,
        })
      );
    });

    await waitFor(() => {
      expect(screen.getByText('5:00')).toBeInTheDocument(); // 5 minute timer
    });
  });

  test('supports lesson and snippet IDs', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(
      <TimedDrillMode
        {...defaultProps}
        lessonId="lesson-123"
        snippetId="snippet-456"
      />
    );

    await waitFor(() => {
      expect(sessionManager.createSession).toHaveBeenCalledWith(
        expect.objectContaining({
          lessonId: 'lesson-123',
          snippetId: 'snippet-456',
        })
      );
    });
  });
});
