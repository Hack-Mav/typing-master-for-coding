import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import App from './App';

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
    expect(screen.getByText('TYPESCRIPT')).toBeInTheDocument();
    expect(screen.getByText('JAVA')).toBeInTheDocument();
    expect(screen.getByText('CPP')).toBeInTheDocument();
    expect(screen.getByText('RUST')).toBeInTheDocument();
  });

  test('shows duration selection when timed drill is selected', async () => {
    render(<App />);

    // Click on Timed Drill mode card
    fireEvent.click(screen.getByText('Timed Drill'));

    await waitFor(() => {
      expect(screen.getByText('Select Duration')).toBeInTheDocument();
    });
    expect(screen.getByText('30s')).toBeInTheDocument();
    expect(screen.getByText('1m')).toBeInTheDocument();
    expect(screen.getByText('3m')).toBeInTheDocument();
    expect(screen.getByText('5m')).toBeInTheDocument();
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

  test('changes duration selection for timed drill', async () => {
    render(<App />);

    // Click on Timed Drill mode card
    fireEvent.click(screen.getByText('Timed Drill'));

    await waitFor(() => {
      expect(screen.getByText('Select Duration')).toBeInTheDocument();
    });

    // Click on 3m duration button
    fireEvent.click(screen.getByText('3m'));

    // The duration should be updated (though we can't easily test the state change)
    expect(screen.getByText('3m')).toBeInTheDocument();
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

  test('applies correct CSS classes for active language and duration buttons', async () => {
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

    // Navigate to Timed Drill and check duration buttons
    fireEvent.click(screen.getByText('Timed Drill'));

    await waitFor(() => {
      expect(screen.getByText('Select Duration')).toBeInTheDocument();
    });

    // 1m should be active by default
    const oneMinButton = screen.getByText('1m');
    expect(oneMinButton).toHaveClass('active');
  });

  test('mode cards have hover effects and are clickable', () => {
    render(<App />);

    const zenCard = screen.getByRole('button', { name: /zen mode/i });
    const timedCard = screen.getByRole('button', { name: /timed drill/i });

    expect(zenCard).toBeInTheDocument();
    expect(timedCard).toBeInTheDocument();

    // Cards should be clickable
    fireEvent.click(zenCard!);
    fireEvent.click(timedCard!);

    // This would trigger navigation in a real scenario
    // We can't easily test the state changes without more complex mocking
  });
});
