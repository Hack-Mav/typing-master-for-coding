import { parserManager } from '../ParserManager';

describe('C++ and Rust Simple Test', () => {
  it('should support C++ and Rust languages', () => {
    const languages = parserManager.getSupportedLanguages();
    expect(languages).toContain('cpp');
    expect(languages).toContain('rust');
  });

  it('should have C++ grammar configuration', () => {
    const grammar = parserManager.getLanguageGrammar('cpp');
    expect(grammar).toBeDefined();
    expect(grammar.keywords).toContain('class');
    expect(grammar.keywords).toContain('namespace');
  });

  it('should have Rust grammar configuration', () => {
    const grammar = parserManager.getLanguageGrammar('rust');
    expect(grammar).toBeDefined();
    expect(grammar.keywords).toContain('fn');
    expect(grammar.keywords).toContain('let');
  });
});
