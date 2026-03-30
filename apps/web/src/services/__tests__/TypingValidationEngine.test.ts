import { typingValidationEngine } from '../TypingValidationEngine';
import { KeystrokeEvent } from '../../types/typing';

describe('TypingValidationEngine', () => {
  beforeEach(() => {
    typingValidationEngine.reset();
  });

  describe('initialization', () => {
    it('should initialize with expected text and language', () => {
      const expectedText = 'def hello():';
      typingValidationEngine.initialize(expectedText, 'python');

      const state = typingValidationEngine.getTypingState();
      expect(state.expectedText).toBe(expectedText);
      expect(state.currentPosition).toBe(0);
      expect(state.actualText).toBe('');
    });
  });

  describe('keystroke processing', () => {
    beforeEach(() => {
      typingValidationEngine.initialize('def test():', 'python');
    });

    it('should validate correct character input', () => {
      const event: KeystrokeEvent = {
        key: 'd',
        code: 'KeyD',
        timestamp: Date.now(),
        action: 'keydown',
        cursorPosition: 0,
        modifiers: { ctrl: false, alt: false, shift: false, meta: false },
      };

      const result = typingValidationEngine.processKeystroke(event);

      expect(result.isValid).toBe(true);
      expect(result.errorType).toBe('correct');
      expect(result.expectedChar).toBe('d');
      expect(result.actualChar).toBe('d');
    });

    it('should detect wrong character input', () => {
      const event: KeystrokeEvent = {
        key: 'x',
        code: 'KeyX',
        timestamp: Date.now(),
        action: 'keydown',
        cursorPosition: 0,
        modifiers: { ctrl: false, alt: false, shift: false, meta: false },
      };

      const result = typingValidationEngine.processKeystroke(event);

      expect(result.isValid).toBe(false);
      expect(result.errorType).toBe('wrong_character');
      expect(result.expectedChar).toBe('d');
      expect(result.actualChar).toBe('x');
    });

    it('should handle backspace correctly', () => {
      // Type a character first
      const typeEvent: KeystrokeEvent = {
        key: 'd',
        code: 'KeyD',
        timestamp: Date.now(),
        action: 'keydown',
        cursorPosition: 0,
        modifiers: { ctrl: false, alt: false, shift: false, meta: false },
      };
      typingValidationEngine.processKeystroke(typeEvent);

      // Then backspace
      const backspaceEvent: KeystrokeEvent = {
        key: 'Backspace',
        code: 'Backspace',
        timestamp: Date.now(),
        action: 'keydown',
        cursorPosition: 1,
        modifiers: { ctrl: false, alt: false, shift: false, meta: false },
      };

      const result = typingValidationEngine.processKeystroke(backspaceEvent);
      const state = typingValidationEngine.getTypingState();

      expect(result.isValid).toBe(true);
      expect(state.currentPosition).toBe(0);
      expect(state.actualText).toBe('');
    });

    it('should handle tab key with proper indentation', () => {
      typingValidationEngine.initialize('    test', 'python'); // 4 spaces expected

      const tabEvent: KeystrokeEvent = {
        key: 'Tab',
        code: 'Tab',
        timestamp: Date.now(),
        action: 'keydown',
        cursorPosition: 0,
        modifiers: { ctrl: false, alt: false, shift: false, meta: false },
      };

      const result = typingValidationEngine.processKeystroke(tabEvent);

      expect(result.isValid).toBe(true);
      expect(result.errorType).toBe('correct');
    });

    it('should detect extra characters', () => {
      typingValidationEngine.initialize('def', 'python');

      // Type the expected text
      ['d', 'e', 'f'].forEach((char, index) => {
        const event: KeystrokeEvent = {
          key: char,
          code: `Key${char.toUpperCase()}`,
          timestamp: Date.now(),
          action: 'keydown',
          cursorPosition: index,
          modifiers: { ctrl: false, alt: false, shift: false, meta: false },
        };
        typingValidationEngine.processKeystroke(event);
      });

      // Type an extra character
      const extraEvent: KeystrokeEvent = {
        key: 'x',
        code: 'KeyX',
        timestamp: Date.now(),
        action: 'keydown',
        cursorPosition: 3,
        modifiers: { ctrl: false, alt: false, shift: false, meta: false },
      };

      const result = typingValidationEngine.processKeystroke(extraEvent);

      expect(result.isValid).toBe(false);
      expect(result.errorType).toBe('extra_character');
    });
  });

  describe('delimiter balance checking', () => {
    it('should detect balanced delimiters', () => {
      const balancedText = 'def test(): return [1, 2, 3]';
      const balance =
        typingValidationEngine.checkDelimiterBalance(balancedText);

      expect(balance.isBalanced).toBe(true);
      expect(balance.errors).toHaveLength(0);
    });

    it('should detect unbalanced parentheses', () => {
      const unbalancedText = 'def test(: return 1';
      const balance =
        typingValidationEngine.checkDelimiterBalance(unbalancedText);

      expect(balance.isBalanced).toBe(false);
      expect(balance.errors.length).toBeGreaterThan(0);
    });

    it('should detect mismatched delimiters', () => {
      const mismatchedText = 'def test(]: return 1';
      const balance =
        typingValidationEngine.checkDelimiterBalance(mismatchedText);

      expect(balance.isBalanced).toBe(false);
      expect(balance.errors.some(e => e.type === 'mismatched')).toBe(true);
    });

    it('should handle quote pairs correctly', () => {
      const quotedText = 'print("Hello, World!")';
      const balance = typingValidationEngine.checkDelimiterBalance(quotedText);

      expect(balance.isBalanced).toBe(true);
    });

    it('should detect unclosed quotes', () => {
      const unclosedQuote = 'print("Hello, World!';
      const balance =
        typingValidationEngine.checkDelimiterBalance(unclosedQuote);

      expect(balance.isBalanced).toBe(false);
      expect(balance.errors.some(e => e.type === 'unclosed')).toBe(true);
    });
  });

  describe('performance tracking', () => {
    it('should track validation performance', () => {
      typingValidationEngine.initialize('test', 'python');

      const event: KeystrokeEvent = {
        key: 't',
        code: 'KeyT',
        timestamp: Date.now(),
        action: 'keydown',
        cursorPosition: 0,
        modifiers: { ctrl: false, alt: false, shift: false, meta: false },
      };

      typingValidationEngine.processKeystroke(event);

      const metrics = typingValidationEngine.getPerformanceMetrics();

      expect(metrics.lastValidationTime).toBeGreaterThan(0);
      expect(metrics.averageValidationTime).toBeGreaterThan(0);
      expect(metrics.maxValidationTime).toBeGreaterThan(0);
      expect(typeof metrics.validationsExceedingTarget).toBe('number');
    });

    it('should maintain performance under 8ms for simple validations', () => {
      typingValidationEngine.initialize('a'.repeat(100), 'python');

      // Perform multiple validations
      for (let i = 0; i < 10; i++) {
        const event: KeystrokeEvent = {
          key: 'a',
          code: 'KeyA',
          timestamp: Date.now(),
          action: 'keydown',
          cursorPosition: i,
          modifiers: { ctrl: false, alt: false, shift: false, meta: false },
        };
        typingValidationEngine.processKeystroke(event);
      }

      const metrics = typingValidationEngine.getPerformanceMetrics();

      // Most validations should be under 8ms
      expect(metrics.averageValidationTime).toBeLessThan(8);
    });
  });

  describe('typing state management', () => {
    it('should track typing progress correctly', () => {
      const expectedText = 'def test():';
      typingValidationEngine.initialize(expectedText, 'python');

      // Type first character
      const event: KeystrokeEvent = {
        key: 'd',
        code: 'KeyD',
        timestamp: Date.now(),
        action: 'keydown',
        cursorPosition: 0,
        modifiers: { ctrl: false, alt: false, shift: false, meta: false },
      };

      typingValidationEngine.processKeystroke(event);

      const state = typingValidationEngine.getTypingState();

      expect(state.currentPosition).toBe(1);
      expect(state.actualText).toBe('d');
      expect(state.expectedText).toBe(expectedText);
      expect(state.isComplete).toBe(false);
      expect(state.hasErrors).toBe(false);
      expect(state.lastKeystroke).toEqual(event);
    });

    it('should detect completion', () => {
      const expectedText = 'hi';
      typingValidationEngine.initialize(expectedText, 'python');

      // Type complete text
      ['h', 'i'].forEach((char, index) => {
        const event: KeystrokeEvent = {
          key: char,
          code: `Key${char.toUpperCase()}`,
          timestamp: Date.now(),
          action: 'keydown',
          cursorPosition: index,
          modifiers: { ctrl: false, alt: false, shift: false, meta: false },
        };
        typingValidationEngine.processKeystroke(event);
      });

      const state = typingValidationEngine.getTypingState();

      expect(state.isComplete).toBe(true);
      expect(state.hasErrors).toBe(false);
    });

    it('should detect errors in typing', () => {
      typingValidationEngine.initialize('def', 'python');

      // Type wrong character
      const event: KeystrokeEvent = {
        key: 'x',
        code: 'KeyX',
        timestamp: Date.now(),
        action: 'keydown',
        cursorPosition: 0,
        modifiers: { ctrl: false, alt: false, shift: false, meta: false },
      };

      typingValidationEngine.processKeystroke(event);

      const state = typingValidationEngine.getTypingState();

      expect(state.hasErrors).toBe(true);
    });
  });

  describe('reset functionality', () => {
    it('should reset state correctly', () => {
      typingValidationEngine.initialize('test', 'python');

      // Type some characters
      const event: KeystrokeEvent = {
        key: 't',
        code: 'KeyT',
        timestamp: Date.now(),
        action: 'keydown',
        cursorPosition: 0,
        modifiers: { ctrl: false, alt: false, shift: false, meta: false },
      };

      typingValidationEngine.processKeystroke(event);

      // Reset
      typingValidationEngine.reset();

      const state = typingValidationEngine.getTypingState();

      expect(state.currentPosition).toBe(0);
      expect(state.actualText).toBe('');
      expect(state.lastKeystroke).toBeNull();
    });
  });
});
