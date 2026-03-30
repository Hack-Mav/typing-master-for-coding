// Typing engine related types

export interface KeystrokeEvent {
  key: string;
  code: string;
  timestamp: number;
  action: 'keydown' | 'keyup';
  cursorPosition: number;
  modifiers: {
    ctrl: boolean;
    alt: boolean;
    shift: boolean;
    meta: boolean;
  };
}

export interface ValidationResult {
  isValid: boolean;
  expectedChar: string;
  actualChar: string;
  position: number;
  errorType: ValidationErrorType;
  suggestions?: string[];
}

export type ValidationErrorType =
  | 'correct'
  | 'wrong_character'
  | 'missing_character'
  | 'extra_character'
  | 'syntax_error'
  | 'whitespace_error'
  | 'delimiter_mismatch';

export interface SyntaxValidation {
  isValid: boolean;
  errors: SyntaxError[];
  warnings: SyntaxWarning[];
  delimitersBalanced: boolean;
  structuralIntegrity: boolean;
}

export interface SyntaxError {
  type:
    | 'missing_delimiter'
    | 'extra_delimiter'
    | 'mismatched_delimiter'
    | 'invalid_syntax';
  message: string;
  position: number;
  severity: 'error' | 'warning';
  expectedToken?: string;
  actualToken?: string;
}

export interface SyntaxWarning {
  type: 'whitespace_style' | 'formatting' | 'convention';
  message: string;
  position: number;
  suggestion: string;
}

export interface WhitespacePolicy {
  enforceIndentation: boolean;
  indentSize: number;
  useSpaces: boolean;
  trimTrailingWhitespace: boolean;
  enforceNewlineAtEOF: boolean;
  maxLineLength?: number;
}

export interface DelimiterBalance {
  isBalanced: boolean;
  openDelimiters: Array<{
    char: string;
    position: number;
  }>;
  errors: Array<{
    type: 'unclosed' | 'unexpected' | 'mismatched';
    position: number;
    expected?: string;
    actual: string;
  }>;
}

export interface TypingState {
  currentPosition: number;
  expectedText: string;
  actualText: string;
  isComplete: boolean;
  hasErrors: boolean;
  lastKeystroke: KeystrokeEvent | null;
}
