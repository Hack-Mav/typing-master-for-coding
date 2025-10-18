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
        screen.getByText('console.log("Hello, World!");')
      ).toBeInTheDocument();
    });

    expect(screen.getByText('Type the code above. Press')).toBeInTheDocument();
    expect(screen.getByText('to exit.')).toBeInTheDocument();
  });

  test('handles exit button click', async () => {
    render(<ZenMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByText('console.log("Hello, World!");')
      ).toBeInTheDocument();
    });

    const exitButton = screen.getByTitle('Exit (Esc)');
    fireEvent.click(exitButton);

    expect(defaultProps.onExit).toHaveBeenCalled();
  });

  test('handles escape key press', async () => {
    render(<ZenMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByText('console.log("Hello, World!");')
      ).toBeInTheDocument();
    });

    // Find the hidden textarea
    const textarea = screen.getByRole('textbox');
    expect(textarea).toBeInTheDocument();

    fireEvent.keyDown(textarea, { key: 'Escape', code: 'Escape' });

    expect(defaultProps.onExit).toHaveBeenCalled();
  });

  test('handles typing input', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<ZenMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByText('console.log("Hello, World!");')
      ).toBeInTheDocument();
    });

    const textarea = screen.getByRole('textbox');

    // Simulate typing 'c'
    fireEvent.keyDown(textarea, {
      key: 'c',
      code: 'KeyC',
      timestamp: Date.now(),
    });

    expect(sessionManager.recordKeystroke).toHaveBeenCalledWith(
      'test-session-id',
      expect.objectContaining({
        key: 'c',
        code: 'KeyC',
        action: 'keydown',
      })
    );
  });

  test('handles backspace input', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<ZenMode {...defaultProps} />);

    await waitFor(() => {
      expect(
        screen.getByText('console.log("Hello, World!");')
      ).toBeInTheDocument();
    });

    const textarea = screen.getByRole('textbox');

    // Simulate backspace
    fireEvent.keyDown(textarea, {
      key: 'Backspace',
      code: 'Backspace',
    });

    expect(sessionManager.recordKeystroke).toHaveBeenCalledWith(
      'test-session-id',
      expect.objectContaining({
        key: 'Backspace',
        code: 'Backspace',
        action: 'keydown',
      })
    );
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
        screen.getByText('console.log("Hello, World!");')
      ).toBeInTheDocument();
    });

    // Simulate completing the text by typing the entire target text
    const textarea = screen.getByRole('textbox');
    const targetText = 'console.log("Hello, World!");';

    // Mock the state change that would happen during typing
    for (let i = 0; i < targetText.length; i++) {
      fireEvent.keyDown(textarea, {
        key: targetText[i],
        code: `Key${targetText[i].toUpperCase()}`,
      });
    }

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
