import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import ZenMode from '../ZenMode';

// Mock the SessionManager
jest.mock('../../services/SessionManager', () => ({
  sessionManager: {
    createSession: jest.fn().mockResolvedValue({
      id: 'test-session-id',
      targetText: 'console.log("Hello, World!");',
      config: { mode: 'zen' },
      state: { status: 'created' },
      events: [],
      progress: {
        percentComplete: 0,
        charactersTyped: 0,
        currentSpeed: 0,
        currentAccuracy: 100,
      },
    }),
    startSession: jest.fn().mockResolvedValue(undefined),
    recordKeystroke: jest.fn().mockResolvedValue(undefined),
    finalizeSession: jest.fn().mockResolvedValue({
      sessionId: 'test-session-id',
      metrics: {
        speed: { twpm: 45 },
        accuracy: { overallAccuracy: 0.95 },
      },
      summary: {
        mode: 'zen',
        language: 'javascript',
        duration: 30000,
        completionRate: 100,
        finalSpeed: 45,
        finalAccuracy: 95,
        errorCount: 2,
        improvementAreas: [],
        strengths: ['high_accuracy'],
      },
    }),
    abandonSession: jest.fn().mockResolvedValue(undefined),
  },
}));

describe('ZenMode', () => {
  const defaultProps = {
    languageId: 'javascript',
    onComplete: jest.fn(),
    onExit: jest.fn(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders loading state initially', () => {
    render(<ZenMode {...defaultProps} />);

    expect(
      screen.getByText('Preparing your zen session...')
    ).toBeInTheDocument();
    expect(
      screen.getByText('Preparing your zen session...')
    ).toBeInTheDocument();
  });

  test('renders zen interface after session loads', async () => {
    render(<ZenMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByTestId('typing-target-text')
      ).toHaveTextContent('console.log("Hello, World!");');
    });

    expect(
      screen.getByText(/Type the code above.*to exit/i)
    ).toBeInTheDocument();
  });

  test('handles exit button click', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<ZenMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByTestId('typing-target-text')
      ).toHaveTextContent('console.log("Hello, World!");');
    });
    await waitFor(() => expect(sessionManager.startSession).toHaveBeenCalled());

    const exitButton = screen.getByTitle('Exit (Esc)');
    fireEvent.click(exitButton);

    await waitFor(() => {
      expect(defaultProps.onExit).toHaveBeenCalled();
    });
  });

  test('handles escape key press', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<ZenMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByTestId('typing-target-text')
      ).toHaveTextContent('console.log("Hello, World!");');
    });
    await waitFor(() => expect(sessionManager.startSession).toHaveBeenCalled());
    await sessionManager.startSession.mock.results[0].value;

    // Find the hidden textarea
    const textarea = screen.getByRole('textbox');
    expect(textarea).toBeInTheDocument();

    fireEvent.keyDown(textarea, { key: 'Escape', code: 'Escape' });

    await waitFor(() => {
      expect(defaultProps.onExit).toHaveBeenCalled();
    });
  });

  test('handles typing input', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<ZenMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByTestId('typing-target-text')
      ).toHaveTextContent('console.log("Hello, World!");');
    });
    await waitFor(() => expect(sessionManager.startSession).toHaveBeenCalled());
    await sessionManager.startSession.mock.results[0].value;

    const textarea = screen.getByRole('textbox');

    // Simulate typing 'c'
    fireEvent.keyDown(textarea, {
      key: 'c',
      code: 'KeyC',
      timestamp: Date.now(),
    });

    await waitFor(() => {
      expect(sessionManager.recordKeystroke).toHaveBeenCalledWith(
        'test-session-id',
        expect.objectContaining({
          key: 'c',
          code: 'KeyC',
          action: 'keydown',
        })
      );
    });
  });

  test('handles backspace input', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<ZenMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByTestId('typing-target-text')
      ).toHaveTextContent('console.log("Hello, World!");');
    });
    await waitFor(() => expect(sessionManager.startSession).toHaveBeenCalled());
    await sessionManager.startSession.mock.results[0].value;

    const textarea = screen.getByRole('textbox');

    // Simulate backspace
    fireEvent.keyDown(textarea, {
      key: 'Backspace',
      code: 'Backspace',
    });

    await waitFor(() => {
      expect(sessionManager.recordKeystroke).toHaveBeenCalledWith(
        'test-session-id',
        expect.objectContaining({
          key: 'Backspace',
          code: 'Backspace',
          action: 'keydown',
        })
      );
    });
  });

  test('shows summary when toggle button is clicked after completion', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    // Mock session completion
    sessionManager.finalizeSession.mockResolvedValueOnce({
      sessionId: 'test-session-id',
      metrics: {
        speed: { twpm: 45 },
        accuracy: { overallAccuracy: 0.95 },
      },
      summary: {
        mode: 'zen',
        language: 'javascript',
        duration: 30000,
        completionRate: 100,
        finalSpeed: 45,
        finalAccuracy: 95,
        errorCount: 2,
        improvementAreas: [],
        strengths: ['high_accuracy'],
      },
    });

    render(<ZenMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByTestId('typing-target-text')
      ).toHaveTextContent('console.log("Hello, World!");');
    });
    await waitFor(() => expect(sessionManager.startSession).toHaveBeenCalled());
    await sessionManager.startSession.mock.results[0].value;

    // Simulate completing the text by setting the input value
    const textarea = screen.getByRole('textbox');
    const targetText = 'console.log("Hello, World!");';
    fireEvent.change(textarea, { target: { value: targetText } });

    // Wait for completion and summary button to appear
    await waitFor(() => {
      expect(screen.getByTitle('Toggle Summary')).toBeInTheDocument();
    });

    const summaryButton = screen.getByTitle('Toggle Summary');
    fireEvent.click(summaryButton);
    expect(screen.getByText('Session Complete')).toBeInTheDocument();
  });

  test('supports different languages and lesson/snippet IDs', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(
      <ZenMode
        languageId="python"
        lessonId="python-basics-1"
        snippetId="snippet-123"
        onComplete={defaultProps.onComplete}
        onExit={defaultProps.onExit}
      />
    );

    await waitFor(() => {
      expect(sessionManager.createSession).toHaveBeenCalledWith(
        expect.objectContaining({
          mode: 'zen',
          languageId: 'python',
          lessonId: 'python-basics-1',
          snippetId: 'snippet-123',
        })
      );
    });
  });

  test('applies zen mode settings correctly', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<ZenMode {...defaultProps} />);

    await waitFor(() => {
      expect(sessionManager.createSession).toHaveBeenCalledWith(
        expect.objectContaining({
          settings: expect.objectContaining({
            showMetrics: false,
            showTimer: false,
            showProgress: false,
            enableSound: false,
          }),
        })
      );
    });
  });
});
