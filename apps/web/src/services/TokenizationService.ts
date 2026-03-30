import { parserManager } from './ParserManager';
import { Language, Token, LanguageGrammar } from '../types/parser';

/**
 * Enhanced token with additional metadata for typing practice
 */
export interface EnhancedToken extends Token {
  category: TokenCategory;
  isKeyword: boolean;
  isOperator: boolean;
  isDelimiter: boolean;
  difficulty: number; // 1-5 scale
  expectedDuration: number; // milliseconds
}

export type TokenCategory =
  | 'keyword'
  | 'identifier'
  | 'string'
  | 'number'
  | 'operator'
  | 'delimiter'
  | 'comment'
  | 'whitespace'
  | 'unknown';

/**
 * TokenizationService provides enhanced tokenization with language-specific
 * grammar handling and metadata for typing practice optimization.
 */
export class TokenizationService {
  private static instance: TokenizationService;

  // Token difficulty weights based on complexity
  private readonly tokenDifficultyWeights: Record<TokenCategory, number> = {
    keyword: 2,
    identifier: 3,
    string: 4,
    number: 2,
    operator: 1,
    delimiter: 1,
    comment: 3,
    whitespace: 1,
    unknown: 3,
  };

  // Average typing time per character (milliseconds) by category
  private readonly avgTypingTimePerChar: Record<TokenCategory, number> = {
    keyword: 80, // Keywords are familiar, typed faster
    identifier: 100, // Variable names, moderate speed
    string: 120, // Strings may have special chars
    number: 90, // Numbers are straightforward
    operator: 70, // Operators are short and common
    delimiter: 60, // Delimiters are single chars
    comment: 110, // Comments are natural language
    whitespace: 50, // Whitespace is fast
    unknown: 130, // Unknown tokens are slowest
  };

  private constructor() {}

  public static getInstance(): TokenizationService {
    if (!TokenizationService.instance) {
      TokenizationService.instance = new TokenizationService();
    }
    return TokenizationService.instance;
  }

  /**
   * Tokenize code with enhanced metadata for typing practice
   */
  public async tokenizeWithMetadata(
    code: string,
    language: Language
  ): Promise<EnhancedToken[]> {
    const tokens = await parserManager.tokenize(code, language);
    const grammar = parserManager.getLanguageGrammar(language);

    return tokens.map(token => this.enhanceToken(token, grammar));
  }

  /**
   * Get token-level breakdown of code for practice sessions
   */
  public async getTokenBreakdown(
    code: string,
    language: Language
  ): Promise<{
    tokens: EnhancedToken[];
    totalTokens: number;
    tokensByCategory: Record<TokenCategory, number>;
    estimatedDuration: number;
    averageDifficulty: number;
  }> {
    const tokens = await this.tokenizeWithMetadata(code, language);

    const tokensByCategory: Record<TokenCategory, number> = {
      keyword: 0,
      identifier: 0,
      string: 0,
      number: 0,
      operator: 0,
      delimiter: 0,
      comment: 0,
      whitespace: 0,
      unknown: 0,
    };

    let totalDifficulty = 0;
    let estimatedDuration = 0;

    tokens.forEach(token => {
      tokensByCategory[token.category]++;
      totalDifficulty += token.difficulty;
      estimatedDuration += token.expectedDuration;
    });

    return {
      tokens,
      totalTokens: tokens.length,
      tokensByCategory,
      estimatedDuration,
      averageDifficulty:
        tokens.length > 0 ? totalDifficulty / tokens.length : 0,
    };
  }

  /**
   * Extract whitespace-only tokens for whitespace policy enforcement
   */
  public extractWhitespaceTokens(code: string): Token[] {
    const whitespaceTokens: Token[] = [];
    const lines = code.split('\n');
    let currentIndex = 0;

    lines.forEach((line, lineIndex) => {
      // Leading whitespace
      const leadingWhitespace = line.match(/^[\s\t]*/)?.[0] || '';
      if (leadingWhitespace) {
        whitespaceTokens.push({
          type: 'whitespace',
          value: leadingWhitespace,
          startIndex: currentIndex,
          endIndex: currentIndex + leadingWhitespace.length,
          startPosition: { row: lineIndex, column: 0 },
          endPosition: { row: lineIndex, column: leadingWhitespace.length },
        });
      }

      // Trailing whitespace
      const trailingWhitespace = line.match(/[\s\t]*$/)?.[0] || '';
      if (trailingWhitespace && trailingWhitespace !== leadingWhitespace) {
        const startCol = line.length - trailingWhitespace.length;
        whitespaceTokens.push({
          type: 'whitespace',
          value: trailingWhitespace,
          startIndex: currentIndex + startCol,
          endIndex: currentIndex + line.length,
          startPosition: { row: lineIndex, column: startCol },
          endPosition: { row: lineIndex, column: line.length },
        });
      }

      currentIndex += line.length + 1; // +1 for newline
    });

    return whitespaceTokens;
  }

  /**
   * Validate token sequence for structural correctness
   */
  public validateTokenSequence(
    tokens: EnhancedToken[],
    _language: Language
  ): {
    isValid: boolean;
    errors: Array<{
      message: string;
      tokenIndex: number;
      severity: 'error' | 'warning';
    }>;
  } {
    const errors: Array<{
      message: string;
      tokenIndex: number;
      severity: 'error' | 'warning';
    }> = [];

    // Check delimiter balance
    const delimiterStack: Array<{ token: string; index: number }> = [];
    const delimiterPairs: Record<string, string> = {
      '(': ')',
      '[': ']',
      '{': '}',
      '"': '"',
      "'": "'",
      '`': '`',
    };

    tokens.forEach((token, index) => {
      if (token.isDelimiter) {
        const value = token.value;

        // Opening delimiter
        if (value in delimiterPairs) {
          delimiterStack.push({ token: value, index });
        }
        // Closing delimiter
        else if (Object.values(delimiterPairs).includes(value)) {
          const expected = delimiterStack.pop();
          if (!expected) {
            errors.push({
              message: `Unexpected closing delimiter '${value}'`,
              tokenIndex: index,
              severity: 'error',
            });
          } else if (delimiterPairs[expected.token] !== value) {
            errors.push({
              message: `Mismatched delimiter: expected '${delimiterPairs[expected.token]}', got '${value}'`,
              tokenIndex: index,
              severity: 'error',
            });
          }
        }
      }
    });

    // Check for unclosed delimiters
    delimiterStack.forEach(unclosed => {
      errors.push({
        message: `Unclosed delimiter '${unclosed.token}'`,
        tokenIndex: unclosed.index,
        severity: 'error',
      });
    });

    return {
      isValid: errors.filter(e => e.severity === 'error').length === 0,
      errors,
    };
  }

  /**
   * Compare expected vs actual token sequences for typing validation
   */
  public compareTokenSequences(
    expected: EnhancedToken[],
    actual: EnhancedToken[]
  ): {
    matches: boolean[];
    accuracy: number;
    tokenAccuracy: number;
    firstMismatchIndex: number | null;
  } {
    const matches: boolean[] = [];
    let correctTokens = 0;
    let firstMismatchIndex: number | null = null;

    const maxLength = Math.max(expected.length, actual.length);

    for (let i = 0; i < maxLength; i++) {
      const expectedToken = expected[i];
      const actualToken = actual[i];

      if (!expectedToken || !actualToken) {
        matches.push(false);
        if (firstMismatchIndex === null) {
          firstMismatchIndex = i;
        }
      } else {
        const isMatch =
          expectedToken.value === actualToken.value &&
          expectedToken.type === actualToken.type;
        matches.push(isMatch);

        if (isMatch) {
          correctTokens++;
        } else if (firstMismatchIndex === null) {
          firstMismatchIndex = i;
        }
      }
    }

    const tokenAccuracy =
      expected.length > 0 ? correctTokens / expected.length : 0;
    const overallAccuracy =
      maxLength > 0 ? matches.filter(m => m).length / maxLength : 0;

    return {
      matches,
      accuracy: overallAccuracy,
      tokenAccuracy,
      firstMismatchIndex,
    };
  }

  /**
   * Enhance basic token with metadata for typing practice
   */
  private enhanceToken(token: Token, grammar: LanguageGrammar): EnhancedToken {
    const category = this.categorizeToken(token, grammar);
    const isKeyword = grammar.keywords.includes(token.value);
    const isOperator = grammar.operators.includes(token.value);
    const isDelimiter = grammar.delimiters.includes(token.value);

    const difficulty = this.calculateTokenDifficulty(
      token,
      category,
      isKeyword
    );
    const expectedDuration = this.calculateExpectedDuration(token, category);

    return {
      ...token,
      category,
      isKeyword,
      isOperator,
      isDelimiter,
      difficulty,
      expectedDuration,
    };
  }

  /**
   * Categorize token based on type and grammar rules
   */
  private categorizeToken(
    token: Token,
    grammar: LanguageGrammar
  ): TokenCategory {
    const { type, value } = token;

    // Check explicit categories first
    if (grammar.keywords.includes(value)) return 'keyword';
    if (grammar.operators.includes(value)) return 'operator';
    if (grammar.delimiters.includes(value)) return 'delimiter';

    // Check by token type
    if (type.includes('string') || type.includes('template')) return 'string';
    if (
      type.includes('number') ||
      type.includes('integer') ||
      type.includes('float')
    )
      return 'number';
    if (type.includes('comment')) return 'comment';
    if (type.includes('identifier') || type.includes('name'))
      return 'identifier';
    if (/^\s+$/.test(value)) return 'whitespace';

    return 'unknown';
  }

  /**
   * Calculate difficulty score for token (1-5 scale)
   */
  private calculateTokenDifficulty(
    token: Token,
    category: TokenCategory,
    isKeyword: boolean
  ): number {
    let baseDifficulty = this.tokenDifficultyWeights[category];

    // Adjust for token length
    if (token.value.length > 10) baseDifficulty += 1;
    if (token.value.length > 20) baseDifficulty += 1;

    // Adjust for special characters
    const specialCharCount = (token.value.match(/[^a-zA-Z0-9\s]/g) || [])
      .length;
    baseDifficulty += Math.min(specialCharCount * 0.5, 2);

    // Keywords are easier due to familiarity
    if (isKeyword) baseDifficulty -= 0.5;

    // Clamp to 1-5 range
    return Math.max(1, Math.min(5, Math.round(baseDifficulty)));
  }

  /**
   * Calculate expected typing duration for token
   */
  private calculateExpectedDuration(
    token: Token,
    category: TokenCategory
  ): number {
    const baseTimePerChar = this.avgTypingTimePerChar[category];
    const charCount = token.value.length;

    // Base duration
    let duration = charCount * baseTimePerChar;

    // Add overhead for token switching
    duration += 50; // 50ms overhead per token

    // Adjust for complexity
    const complexityMultiplier =
      token.value.match(/[^a-zA-Z0-9\s]/g)?.length || 0;
    duration += complexityMultiplier * 30;

    return Math.round(duration);
  }
}

// Export singleton instance
export const tokenizationService = TokenizationService.getInstance();
