import { parserManager } from '../ParserManager';
import { tokenizationService } from '../TokenizationService';

/**
 * Integration tests for parser functionality
 * These tests verify the API contracts without requiring WASM files
 */
describe('Parser Integration', () => {
  describe('ParserManager API', () => {
    it('should have correct singleton pattern', () => {
      const instance1 = parserManager;
      const instance2 = parserManager;
      expect(instance1).toBe(instance2);
    });

    it('should support MVP languages', () => {
      expect(parserManager.isLanguageSupported('python')).toBe(true);
      expect(parserManager.isLanguageSupported('javascript')).toBe(true);
      expect(parserManager.isLanguageSupported('yaml')).toBe(true);
      expect(parserManager.isLanguageSupported('unsupported')).toBe(false);
    });

    it('should return correct supported languages', () => {
      const languages = parserManager.getSupportedLanguages();
      expect(languages).toEqual([
        'python',
        'javascript',
        'yaml',
        'cpp',
        'rust',
      ]);
    });

    it('should provide language grammars', () => {
      const pythonGrammar = parserManager.getLanguageGrammar('python');
      expect(pythonGrammar).toHaveProperty('tokenTypes');
      expect(pythonGrammar).toHaveProperty('keywords');
      expect(pythonGrammar).toHaveProperty('operators');
      expect(pythonGrammar).toHaveProperty('delimiters');
      expect(pythonGrammar).toHaveProperty('whitespaceRules');

      expect(pythonGrammar.keywords).toContain('def');
      expect(pythonGrammar.keywords).toContain('class');
      expect(pythonGrammar.operators).toContain('+');
      expect(pythonGrammar.delimiters).toContain('(');
    });

    it('should have correct whitespace rules for each language', () => {
      const pythonRules =
        parserManager.getLanguageGrammar('python').whitespaceRules;
      expect(pythonRules.indentSize).toBe(4);
      expect(pythonRules.useSpaces).toBe(true);

      const jsRules =
        parserManager.getLanguageGrammar('javascript').whitespaceRules;
      expect(jsRules.indentSize).toBe(2);
      expect(jsRules.useSpaces).toBe(true);

      const yamlRules =
        parserManager.getLanguageGrammar('yaml').whitespaceRules;
      expect(yamlRules.indentSize).toBe(2);
      expect(yamlRules.useSpaces).toBe(true);
    });
  });

  describe('TokenizationService API', () => {
    it('should have correct singleton pattern', () => {
      const instance1 = tokenizationService;
      const instance2 = tokenizationService;
      expect(instance1).toBe(instance2);
    });

    it('should extract whitespace tokens correctly', () => {
      const code = '  def test():\n    return True\n  ';
      const whitespaceTokens =
        tokenizationService.extractWhitespaceTokens(code);

      expect(whitespaceTokens.length).toBeGreaterThan(0);
      whitespaceTokens.forEach(token => {
        expect(token.type).toBe('whitespace');
        expect(/^\s+$/.test(token.value)).toBe(true);
      });
    });

    it('should validate delimiter balance', () => {
      // Create mock tokens for testing
      const balancedTokens = [
        {
          type: 'delimiter',
          value: '(',
          category: 'delimiter',
          isDelimiter: true,
        },
        {
          type: 'identifier',
          value: 'test',
          category: 'identifier',
          isDelimiter: false,
        },
        {
          type: 'delimiter',
          value: ')',
          category: 'delimiter',
          isDelimiter: true,
        },
      ] as any[];

      const validation = tokenizationService.validateTokenSequence(
        balancedTokens,
        'python'
      );
      expect(validation.isValid).toBe(true);
      expect(validation.errors).toHaveLength(0);
    });

    it('should detect unbalanced delimiters', () => {
      const unbalancedTokens = [
        {
          type: 'delimiter',
          value: '(',
          category: 'delimiter',
          isDelimiter: true,
        },
        {
          type: 'identifier',
          value: 'test',
          category: 'identifier',
          isDelimiter: false,
        },
        // Missing closing parenthesis
      ] as any[];

      const validation = tokenizationService.validateTokenSequence(
        unbalancedTokens,
        'python'
      );
      expect(validation.isValid).toBe(false);
      expect(validation.errors.length).toBeGreaterThan(0);
      expect(validation.errors[0].message).toContain('Unclosed delimiter');
    });

    it('should compare token sequences correctly', () => {
      const tokens1 = [
        { type: 'keyword', value: 'def', category: 'keyword' },
        { type: 'identifier', value: 'test', category: 'identifier' },
      ] as any[];

      const tokens2 = [
        { type: 'keyword', value: 'def', category: 'keyword' },
        { type: 'identifier', value: 'test', category: 'identifier' },
      ] as any[];

      const comparison = tokenizationService.compareTokenSequences(
        tokens1,
        tokens2
      );
      expect(comparison.accuracy).toBe(1);
      expect(comparison.tokenAccuracy).toBe(1);
      expect(comparison.firstMismatchIndex).toBeNull();
    });

    it('should detect token mismatches', () => {
      const tokens1 = [
        { type: 'keyword', value: 'def', category: 'keyword' },
        { type: 'identifier', value: 'test', category: 'identifier' },
      ] as any[];

      const tokens2 = [
        { type: 'keyword', value: 'def', category: 'keyword' },
        { type: 'identifier', value: 'wrong', category: 'identifier' },
      ] as any[];

      const comparison = tokenizationService.compareTokenSequences(
        tokens1,
        tokens2
      );
      expect(comparison.accuracy).toBe(0.5);
      expect(comparison.tokenAccuracy).toBe(0.5);
      expect(comparison.firstMismatchIndex).toBe(1);
    });
  });

  describe('Language-specific configurations', () => {
    it('should have Python-specific configuration', () => {
      const grammar = parserManager.getLanguageGrammar('python');
      expect(grammar.keywords).toContain('def');
      expect(grammar.keywords).toContain('class');
      expect(grammar.keywords).toContain('import');
      expect(grammar.operators).toContain('and');
      expect(grammar.operators).toContain('or');
      expect(grammar.delimiters).toContain(':');
    });

    it('should have JavaScript-specific configuration', () => {
      const grammar = parserManager.getLanguageGrammar('javascript');
      expect(grammar.keywords).toContain('function');
      expect(grammar.keywords).toContain('const');
      expect(grammar.keywords).toContain('let');
      expect(grammar.operators).toContain('===');
      expect(grammar.operators).toContain('!==');
      expect(grammar.delimiters).toContain(';');
    });

    it('should have YAML-specific configuration', () => {
      const grammar = parserManager.getLanguageGrammar('yaml');
      expect(grammar.keywords).toContain('true');
      expect(grammar.keywords).toContain('false');
      expect(grammar.keywords).toContain('null');
      expect(grammar.operators).toContain(':');
      expect(grammar.operators).toContain('-');
    });
  });
});
