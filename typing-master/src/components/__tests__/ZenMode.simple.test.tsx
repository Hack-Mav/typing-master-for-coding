import React from 'react';
import { render, screen } from '@testing-library/react';
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

describe('ZenMode Simple Tests', () => {
  const defaultProps = {
    languageId: 'javascript',
    onComplete: jest.fn(),
    onExit: jest.fn(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders without crashing', () => {
    render(<ZenMode {...defaultProps} />);
    expect(
      screen.getByText('Preparing your zen session...')
    ).toBeInTheDocument();
  });

  test('creates session with correct config', async () => {
    const { sessionManager } = require('../../services/SessionManager');

    render(<ZenMode {...defaultProps} />);

    // Wait a bit for useEffect to run
    await new Promise(resolve => setTimeout(resolve, 100));

    expect(sessionManager.createSession).toHaveBeenCalledWith(
      expect.objectContaining({
        mode: 'zen',
        languageId: 'javascript',
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
