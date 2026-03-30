import { parserManager } from './ParserManager';
import { Language } from '../types/parser';
import {
  KeystrokeEvent,
  ValidationResult,
  ValidationErrorType,
  SyntaxValidation,
  SyntaxError,
  SyntaxWarning,
  WhitespacePolicy,
  DelimiterBalance,
  TypingState,
} from '../types/typing';

/**
 * Real-time typing validation engine that processes keystrokes and provides
 * immediate feedback for syntax-aware typing practice.
 */
export class TypingValidationEngine {
  private static instance: TypingValidationEngine;
  private language: Language = 'python';
  private expectedText: string = '';
  private actualText: string = '';
  private currentPosition: number = 0;
  private keystrokeHistory: KeystrokeEvent[] = [];
  private validationCache: Map<string, ValidationResult> = new Map();

  // Performance tracking for sub-8ms latency requirement
  private lastValidationTime: number = 0;
  private validationTimes: number[] = [];

  private constructor() {}

  public static getInstance(): TypingValidationEngine {
    if (!TypingValidationEngine.instance) {
      TypingValidationEngine.instance = new TypingValidationEngine();
    }
    return TypingValidationEngine.instance;
  }

  /**
   * Initialize validation engine with target text and language
   */
  public initialize(expectedText: string, language: Language): void {
    this.expectedText = expectedText;
    this.language = language;
    this.actualText = '';
    this.currentPosition = 0;
    this.keystrokeHistory = [];
    this.validationCache.clear();
  }

  /**
   * Process keystroke event and return validation result
   * Optimized for <8ms latency requirement
   */
  public processKeystroke(event: KeystrokeEvent): ValidationResult {
    const startTime = performance.now();

    try {
      // Only process keydown events for character input
      if (event.action !== 'keydown') {
        return this.createValidationResult(
          'correct',
          event.key,
          event.key,
          this.currentPosition
        );
      }

      // Handle special keys
      if (this.isSpecialKey(event.key)) {
        return this.handleSpecialKey(event);
      }

      // Record keystroke
      this.keystrokeHistory.push(event);

      // Update actual text and position
      const result = this.validateCharacterInput(event);

      // Update position and text based on validation result
      if (result.isValid) {
        this.actualText += event.key;
        this.currentPosition++;
      } else {
        // For wrong characters, still update actual text to track errors
        this.actualText += event.key;
        // Don't advance position for wrong characters
      }

      return result;
    } finally {
      // Track validation performance
      const validationTime = performance.now() - startTime;
      this.lastValidationTime = validationTime;
      this.validationTimes.push(validationTime);

      // Keep only last 100 measurements for performance tracking
      if (this.validationTimes.length > 100) {
        this.validationTimes.shift();
      }

      // Log warning if validation exceeds 8ms target
      if (validationTime > 8) {
        console.warn(
          `Validation latency exceeded target: ${validationTime.toFixed(2)}ms`
        );
      }
    }
  }

  /**
   * Validate character input against expected text
   */
  private validateCharacterInput(event: KeystrokeEvent): ValidationResult {
    const expectedChar = this.expectedText[this.currentPosition];
    const actualChar = event.key;

    // Check cache first for performance
    const cacheKey = `${this.currentPosition}:${actualChar}`;
    if (this.validationCache.has(cacheKey)) {
      return this.validationCache.get(cacheKey)!;
    }

    let result: ValidationResult;

    if (!expectedChar) {
      // Extra character beyond expected text
      result = this.createValidationResult(
        'extra_character',
        '',
        actualChar,
        this.currentPosition
      );
    } else if (expectedChar === actualChar) {
      // Correct character
      result = this.createValidationResult(
        'correct',
        expectedChar,
        actualChar,
        this.currentPosition
      );
    } else {
      // Wrong character - determine specific error type
      const errorType = this.determineErrorType(expectedChar, actualChar);
      result = this.createValidationResult(
        errorType,
        expectedChar,
        actualChar,
        this.currentPosition
      );
    }

    // Cache result for performance
    this.validationCache.set(cacheKey, result);
    return result;
  }

  /**
   * Handle special keys (backspace, delete, arrow keys, etc.)
   */
  private handleSpecialKey(event: KeystrokeEvent): ValidationResult {
    switch (event.key) {
      case 'Backspace':
        return this.handleBackspace();
      case 'Delete':
        return this.handleDelete();
      case 'ArrowLeft':
      case 'ArrowRight':
      case 'ArrowUp':
      case 'ArrowDown':
        return this.handleArrowKey(event.key);
      case 'Tab':
        return this.handleTab();
      case 'Enter':
        return this.handleEnter();
      default:
        return this.createValidationResult(
          'correct',
          event.key,
          event.key,
          this.currentPosition
        );
    }
  }

  /**
   * Handle backspace key
   */
  private handleBackspace(): ValidationResult {
    if (this.currentPosition > 0) {
      this.currentPosition--;
      this.actualText = this.actualText.slice(0, -1);

      // Clear cache for affected positions
      this.clearCacheFromPosition(this.currentPosition);
    }

    return this.createValidationResult(
      'correct',
      '',
      'Backspace',
      this.currentPosition
    );
  }

  /**
   * Handle delete key
   */
  private handleDelete(): ValidationResult {
    // Delete doesn't change current position, just removes character ahead
    return this.createValidationResult(
      'correct',
      '',
      'Delete',
      this.currentPosition
    );
  }

  /**
   * Handle arrow keys (cursor movement)
   */
  private handleArrowKey(key: string): ValidationResult {
    // For now, we don't allow cursor movement in typing practice
    // This could be enhanced later for more advanced editing modes
    return this.createValidationResult(
      'correct',
      '',
      key,
      this.currentPosition
    );
  }

  /**
   * Handle tab key with language-specific indentation
   */
  private handleTab(): ValidationResult {
    const grammar = parserManager.getLanguageGrammar(this.language);
    const indentString = grammar.whitespaceRules.useSpaces
      ? ' '.repeat(grammar.whitespaceRules.indentSize)
      : '\t';

    // Check if tab matches expected indentation
    const expectedSubstring = this.expectedText.substring(
      this.currentPosition,
      this.currentPosition + indentString.length
    );

    if (expectedSubstring === indentString) {
      this.actualText += indentString;
      this.currentPosition += indentString.length;
      return this.createValidationResult(
        'correct',
        indentString,
        'Tab',
        this.currentPosition - indentString.length
      );
    } else {
      return this.createValidationResult(
        'whitespace_error',
        expectedSubstring,
        'Tab',
        this.currentPosition
      );
    }
  }

  /**
   * Handle enter key (newline)
   */
  private handleEnter(): ValidationResult {
    const expectedChar = this.expectedText[this.currentPosition];

    if (expectedChar === '\n') {
      this.actualText += '\n';
      this.currentPosition++;
      return this.createValidationResult(
        'correct',
        '\n',
        'Enter',
        this.currentPosition - 1
      );
    } else {
      return this.createValidationResult(
        'wrong_character',
        expectedChar,
        '\n',
        this.currentPosition
      );
    }
  }

  /**
   * Perform comprehensive syntax validation
   */
  public async validateSyntax(
    input: string,
    expected: string
  ): Promise<SyntaxValidation> {
    try {
      // Parse both input and expected text
      const [inputResult, expectedResult] = await Promise.all([
        parserManager.parse(input, this.language),
        parserManager.parse(expected, this.language),
      ]);

      const errors: SyntaxError[] = [];
      const warnings: SyntaxWarning[] = [];

      // Check for parse errors in input
      inputResult.errors.forEach(error => {
        errors.push({
          type: 'invalid_syntax',
          message: error.message,
          position: error.startIndex,
          severity: 'error',
        });
      });

      // Check delimiter balance
      const delimiterBalance = this.checkDelimiterBalance(input);
      if (!delimiterBalance.isBalanced) {
        delimiterBalance.errors.forEach(error => {
          errors.push({
            type:
              error.type === 'unclosed'
                ? 'missing_delimiter'
                : error.type === 'unexpected'
                  ? 'extra_delimiter'
                  : 'mismatched_delimiter',
            message: `Delimiter error: ${error.actual}`,
            position: error.position,
            severity: 'error',
            expectedToken: error.expected,
            actualToken: error.actual,
          });
        });
      }

      // Check whitespace policy
      const whitespaceWarnings = this.validateWhitespacePolicy(input);
      warnings.push(...whitespaceWarnings);

      // Compare AST structure if both parse successfully
      const structuralIntegrity =
        inputResult.errors.length === 0 &&
        this.compareASTStructure(inputResult.ast, expectedResult.ast);

      return {
        isValid: errors.length === 0,
        errors,
        warnings,
        delimitersBalanced: delimiterBalance.isBalanced,
        structuralIntegrity,
      };
    } catch (error) {
      return {
        isValid: false,
        errors: [
          {
            type: 'invalid_syntax',
            message: 'Failed to parse syntax',
            position: 0,
            severity: 'error',
          },
        ],
        warnings: [],
        delimitersBalanced: false,
        structuralIntegrity: false,
      };
    }
  }

  /**
   * Check delimiter balance (parentheses, brackets, braces, quotes)
   */
  public checkDelimiterBalance(text: string): DelimiterBalance {
    const stack: Array<{ char: string; position: number }> = [];
    const errors: DelimiterBalance['errors'] = [];

    const pairs: Record<string, string> = {
      '(': ')',
      '[': ']',
      '{': '}',
      '"': '"',
      "'": "'",
      '`': '`',
    };

    const closingChars = new Set(Object.values(pairs));
    const openingChars = new Set(Object.keys(pairs));

    for (let i = 0; i < text.length; i++) {
      const char = text[i];

      if (openingChars.has(char)) {
        // Handle quote pairs specially (they open and close with same character)
        if (char === '"' || char === "'" || char === '`') {
          const lastSame = stack.findIndex(item => item.char === char);
          if (lastSame !== -1) {
            // Close existing quote
            stack.splice(lastSame, 1);
          } else {
            // Open new quote
            stack.push({ char, position: i });
          }
        } else {
          // Regular opening delimiter
          stack.push({ char, position: i });
        }
      } else if (closingChars.has(char)) {
        const expected = stack.pop();
        if (!expected) {
          errors.push({
            type: 'unexpected',
            position: i,
            actual: char,
          });
        } else if (pairs[expected.char] !== char) {
          errors.push({
            type: 'mismatched',
            position: i,
            expected: pairs[expected.char],
            actual: char,
          });
        }
      }
    }

    // Check for unclosed delimiters
    stack.forEach(unclosed => {
      errors.push({
        type: 'unclosed',
        position: unclosed.position,
        actual: unclosed.char,
        expected: pairs[unclosed.char],
      });
    });

    return {
      isBalanced: errors.length === 0,
      openDelimiters: stack,
      errors,
    };
  }

  /**
   * Validate whitespace policy enforcement
   */
  private validateWhitespacePolicy(text: string): SyntaxWarning[] {
    const warnings: SyntaxWarning[] = [];
    const policy = this.getWhitespacePolicy();
    const lines = text.split('\n');

    lines.forEach((line, lineIndex) => {
      // Check trailing whitespace
      if (policy.trimTrailingWhitespace && /\s+$/.test(line)) {
        warnings.push({
          type: 'whitespace_style',
          message: 'Trailing whitespace detected',
          position:
            this.getLineStartPosition(text, lineIndex) +
            line.length -
            line.trimEnd().length,
          suggestion: 'Remove trailing whitespace',
        });
      }

      // Check indentation style
      const leadingWhitespace = line.match(/^[\s\t]*/)?.[0] || '';
      if (leadingWhitespace) {
        if (policy.useSpaces && leadingWhitespace.includes('\t')) {
          warnings.push({
            type: 'whitespace_style',
            message: 'Tabs found, spaces expected',
            position: this.getLineStartPosition(text, lineIndex),
            suggestion: `Use ${policy.indentSize} spaces for indentation`,
          });
        } else if (!policy.useSpaces && leadingWhitespace.includes(' ')) {
          warnings.push({
            type: 'whitespace_style',
            message: 'Spaces found, tabs expected',
            position: this.getLineStartPosition(text, lineIndex),
            suggestion: 'Use tabs for indentation',
          });
        }

        // Check indentation size
        if (
          policy.useSpaces &&
          leadingWhitespace.length % policy.indentSize !== 0
        ) {
          warnings.push({
            type: 'whitespace_style',
            message: 'Incorrect indentation size',
            position: this.getLineStartPosition(text, lineIndex),
            suggestion: `Use multiples of ${policy.indentSize} spaces`,
          });
        }
      }

      // Check line length
      if (policy.maxLineLength && line.length > policy.maxLineLength) {
        warnings.push({
          type: 'formatting',
          message: `Line exceeds maximum length (${policy.maxLineLength})`,
          position:
            this.getLineStartPosition(text, lineIndex) + policy.maxLineLength,
          suggestion: 'Break line or refactor to reduce length',
        });
      }
    });

    // Check newline at EOF
    if (policy.enforceNewlineAtEOF && text.length > 0 && !text.endsWith('\n')) {
      warnings.push({
        type: 'formatting',
        message: 'Missing newline at end of file',
        position: text.length,
        suggestion: 'Add newline at end of file',
      });
    }

    return warnings;
  }

  /**
   * Get current typing state
   */
  public getTypingState(): TypingState {
    return {
      currentPosition: this.currentPosition,
      expectedText: this.expectedText,
      actualText: this.actualText,
      isComplete: this.currentPosition >= this.expectedText.length,
      hasErrors:
        this.actualText !==
        this.expectedText.substring(0, this.actualText.length),
      lastKeystroke:
        this.keystrokeHistory[this.keystrokeHistory.length - 1] || null,
    };
  }

  /**
   * Get performance metrics for validation latency
   */
  public getPerformanceMetrics(): {
    lastValidationTime: number;
    averageValidationTime: number;
    maxValidationTime: number;
    validationsExceedingTarget: number;
  } {
    const avg =
      this.validationTimes.length > 0
        ? this.validationTimes.reduce((a, b) => a + b, 0) /
          this.validationTimes.length
        : 0;

    const max =
      this.validationTimes.length > 0 ? Math.max(...this.validationTimes) : 0;

    const exceedingTarget = this.validationTimes.filter(
      time => time > 8
    ).length;

    return {
      lastValidationTime: this.lastValidationTime,
      averageValidationTime: avg,
      maxValidationTime: max,
      validationsExceedingTarget: exceedingTarget,
    };
  }

  /**
   * Reset validation engine state
   */
  public reset(): void {
    this.actualText = '';
    this.currentPosition = 0;
    this.keystrokeHistory = [];
    this.validationCache.clear();
    this.validationTimes = [];
  }

  // Helper methods

  private isSpecialKey(key: string): boolean {
    return [
      'Backspace',
      'Delete',
      'Tab',
      'Enter',
      'ArrowLeft',
      'ArrowRight',
      'ArrowUp',
      'ArrowDown',
    ].includes(key);
  }

  private determineErrorType(
    expected: string,
    actual: string
  ): ValidationErrorType {
    // Check for common syntax elements
    const delimiters = ['(', ')', '[', ']', '{', '}', '"', "'", '`'];

    if (delimiters.includes(expected) || delimiters.includes(actual)) {
      return 'delimiter_mismatch';
    }

    if (/\s/.test(expected) || /\s/.test(actual)) {
      return 'whitespace_error';
    }

    return 'wrong_character';
  }

  private createValidationResult(
    errorType: ValidationErrorType,
    expectedChar: string,
    actualChar: string,
    position: number
  ): ValidationResult {
    return {
      isValid: errorType === 'correct',
      expectedChar,
      actualChar,
      position,
      errorType,
      suggestions: this.generateSuggestions(
        errorType,
        expectedChar,
        actualChar
      ),
    };
  }

  private generateSuggestions(
    errorType: ValidationErrorType,
    expected: string,
    actual: string
  ): string[] {
    const suggestions: string[] = [];

    switch (errorType) {
      case 'wrong_character':
        suggestions.push(`Expected '${expected}', got '${actual}'`);
        break;
      case 'delimiter_mismatch':
        suggestions.push(`Check delimiter pairing: expected '${expected}'`);
        break;
      case 'whitespace_error':
        suggestions.push('Check indentation and spacing');
        break;
      case 'extra_character':
        suggestions.push('Remove extra character');
        break;
      case 'missing_character':
        suggestions.push(`Add missing character: '${expected}'`);
        break;
    }

    return suggestions;
  }

  private compareASTStructure(actual: any, expected: any): boolean {
    // Simplified AST comparison - could be enhanced for more detailed structural analysis
    if (!actual || !expected) return false;

    return (
      actual.type === expected.type &&
      actual.children?.length === expected.children?.length
    );
  }

  private getWhitespacePolicy(): WhitespacePolicy {
    const grammar = parserManager.getLanguageGrammar(this.language);

    return {
      enforceIndentation: true,
      indentSize: grammar.whitespaceRules.indentSize,
      useSpaces: grammar.whitespaceRules.useSpaces,
      trimTrailingWhitespace: grammar.whitespaceRules.trimTrailingWhitespace,
      enforceNewlineAtEOF: true,
      maxLineLength: 100, // Default line length limit
    };
  }

  private getLineStartPosition(text: string, lineIndex: number): number {
    const lines = text.split('\n');
    let position = 0;

    for (let i = 0; i < lineIndex; i++) {
      position += lines[i].length + 1; // +1 for newline
    }

    return position;
  }

  private clearCacheFromPosition(position: number): void {
    const keysToDelete: string[] = [];

    for (const key of Array.from(this.validationCache.keys())) {
      const keyPosition = parseInt(key.split(':')[0]);
      if (keyPosition >= position) {
        keysToDelete.push(key);
      }
    }

    keysToDelete.forEach(key => this.validationCache.delete(key));
  }
}

// Export singleton instance
export const typingValidationEngine = TypingValidationEngine.getInstance();
