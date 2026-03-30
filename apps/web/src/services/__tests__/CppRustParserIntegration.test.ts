import { parserManager } from '../ParserManager';
import { contentService } from '../ContentService';

describe('C++ and Rust Parser Integration', () => {
  beforeAll(async () => {
    // Initialize parser manager and content service
    await parserManager.initialize();
    await contentService.initialize();
  });

  afterAll(() => {
    // Clean up resources
    parserManager.dispose();
  });

  describe('C++ Language Support', () => {
    it('should load C++ parser successfully', async () => {
      const parser = await parserManager.loadParser('cpp');
      expect(parser).toBeDefined();
    });

    it('should tokenize C++ code correctly', async () => {
      const cppCode = `#include <iostream>

int main() {
    std::cout << "Hello, World!" << std::endl;
    return 0;
}`;

      const tokens = await parserManager.tokenize(cppCode, 'cpp');
      expect(tokens).toBeDefined();
      expect(tokens.length).toBeGreaterThan(0);

      // Check for specific C++ tokens
      const tokenValues = tokens.map(t => t.value);
      expect(tokenValues).toContain('#include');
      expect(tokenValues).toContain('int');
      expect(tokenValues).toContain('main');
      expect(tokenValues).toContain('std');
      expect(tokenValues).toContain('cout');
    });

    it('should parse C++ AST correctly', async () => {
      const cppCode = `int factorial(int n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}`;

      const ast = await parserManager.parseAST(cppCode, 'cpp');
      expect(ast).toBeDefined();
      expect(ast.type).toBe('translation_unit');
      expect(ast.children.length).toBeGreaterThan(0);
    });

    it('should handle C++ syntax errors', async () => {
      const invalidCppCode = `int main() {
    std::cout << "Missing semicolon"
    return 0;
}`;

      const parseResult = await parserManager.parse(invalidCppCode, 'cpp');
      expect(parseResult.errors.length).toBeGreaterThan(0);
    });

    it('should get C++ language grammar', () => {
      const grammar = parserManager.getLanguageGrammar('cpp');
      expect(grammar).toBeDefined();
      expect(grammar.keywords).toContain('class');
      expect(grammar.keywords).toContain('namespace');
      expect(grammar.keywords).toContain('template');
      expect(grammar.operators).toContain('::');
      expect(grammar.operators).toContain('->');
      expect(grammar.whitespaceRules.indentSize).toBe(4);
    });

    it('should load C++ lessons', async () => {
      const lessons = await contentService.getLessonsForLanguage('cpp');
      expect(lessons).toBeDefined();
      expect(lessons.length).toBeGreaterThan(0);

      const introLesson = lessons.find(l => l.id === 'cpp-intro-1');
      expect(introLesson).toBeDefined();
      expect(introLesson?.title).toContain('Hello World');
    });

    it('should load C++ snippets', async () => {
      const snippets = await contentService.getSnippetsForLanguage('cpp');
      expect(snippets).toBeDefined();
      expect(snippets.length).toBeGreaterThan(0);

      const pointerSnippet = snippets.find(s => s.tags.includes('pointers'));
      expect(pointerSnippet).toBeDefined();
    });
  });

  describe('Rust Language Support', () => {
    it('should load Rust parser successfully', async () => {
      const parser = await parserManager.loadParser('rust');
      expect(parser).toBeDefined();
    });

    it('should tokenize Rust code correctly', async () => {
      const rustCode = `fn main() {
    println!("Hello, World!");
}`;

      const tokens = await parserManager.tokenize(rustCode, 'rust');
      expect(tokens).toBeDefined();
      expect(tokens.length).toBeGreaterThan(0);

      // Check for specific Rust tokens
      const tokenValues = tokens.map(t => t.value);
      expect(tokenValues).toContain('fn');
      expect(tokenValues).toContain('main');
      expect(tokenValues).toContain('println!');
    });

    it('should parse Rust AST correctly', async () => {
      const rustCode = `fn calculate_length(s: &String) -> usize {
    s.len()
}`;

      const ast = await parserManager.parseAST(rustCode, 'rust');
      expect(ast).toBeDefined();
      expect(ast.type).toBe('source_file');
      expect(ast.children.length).toBeGreaterThan(0);
    });

    it('should handle Rust syntax errors', async () => {
      const invalidRustCode = `fn main() {
    let x = 5
    println!("Missing semicolon: {}", x);
}`;

      const parseResult = await parserManager.parse(invalidRustCode, 'rust');
      expect(parseResult.errors.length).toBeGreaterThan(0);
    });

    it('should get Rust language grammar', () => {
      const grammar = parserManager.getLanguageGrammar('rust');
      expect(grammar).toBeDefined();
      expect(grammar.keywords).toContain('fn');
      expect(grammar.keywords).toContain('let');
      expect(grammar.keywords).toContain('mut');
      expect(grammar.keywords).toContain('impl');
      expect(grammar.operators).toContain('->');
      expect(grammar.operators).toContain('=>');
      expect(grammar.operators).toContain('::');
      expect(grammar.whitespaceRules.indentSize).toBe(4);
    });

    it('should load Rust lessons', async () => {
      const lessons = await contentService.getLessonsForLanguage('rust');
      expect(lessons).toBeDefined();
      expect(lessons.length).toBeGreaterThan(0);

      const introLesson = lessons.find(l => l.id === 'rust-intro-1');
      expect(introLesson).toBeDefined();
      expect(introLesson?.title).toContain('Hello World');
    });

    it('should load Rust snippets', async () => {
      const snippets = await contentService.getSnippetsForLanguage('rust');
      expect(snippets).toBeDefined();
      expect(snippets.length).toBeGreaterThan(0);

      const iteratorSnippet = snippets.find(s => s.tags.includes('iterators'));
      expect(iteratorSnippet).toBeDefined();
    });
  });

  describe('Advanced Language Features', () => {
    it('should handle C++ template syntax', async () => {
      const templateCode = `template<typename T>
void printVector(const std::vector<T>& vec) {
    for (const auto& item : vec) {
        std::cout << item << " ";
    }
}`;

      const tokens = await parserManager.tokenize(templateCode, 'cpp');
      const tokenValues = tokens.map(t => t.value);
      expect(tokenValues).toContain('template');
      expect(tokenValues).toContain('typename');
      expect(tokenValues).toContain('auto');
    });

    it('should handle Rust ownership syntax', async () => {
      const ownershipCode = `fn longest<'a>(x: &'a str, y: &'a str) -> &'a str {
    if x.len() > y.len() {
        x
    } else {
        y
    }
}`;

      const tokens = await parserManager.tokenize(ownershipCode, 'rust');
      const tokenValues = tokens.map(t => t.value);
      expect(tokenValues).toContain('fn');
      expect(tokenValues).toContain('longest');
      // Lifetime parameters should be tokenized
      expect(tokens.some(t => t.value.includes("'a"))).toBe(true);
    });

    it('should handle C++ preprocessor directives', async () => {
      const preprocessorCode = `#include <iostream>
#define MAX_SIZE 100
#ifdef DEBUG
    #define LOG(x) std::cout << x << std::endl
#endif`;

      const tokens = await parserManager.tokenize(preprocessorCode, 'cpp');
      const tokenValues = tokens.map(t => t.value);
      expect(tokenValues).toContain('#include');
      expect(tokenValues).toContain('#define');
      expect(tokenValues).toContain('#ifdef');
    });

    it('should handle Rust macro syntax', async () => {
      const macroCode = `macro_rules! say_hello {
    () => {
        println!("Hello!");
    };
}

fn main() {
    say_hello!();
}`;

      const tokens = await parserManager.tokenize(macroCode, 'rust');
      const tokenValues = tokens.map(t => t.value);
      expect(tokenValues).toContain('macro_rules!');
      expect(tokenValues).toContain('println!');
    });
  });

  describe('Language-specific Whitespace Rules', () => {
    it('should apply C++ whitespace rules', () => {
      const grammar = parserManager.getLanguageGrammar('cpp');
      expect(grammar.whitespaceRules.indentSize).toBe(4);
      expect(grammar.whitespaceRules.useSpaces).toBe(true);
      expect(grammar.whitespaceRules.trimTrailingWhitespace).toBe(true);
    });

    it('should apply Rust whitespace rules', () => {
      const grammar = parserManager.getLanguageGrammar('rust');
      expect(grammar.whitespaceRules.indentSize).toBe(4);
      expect(grammar.whitespaceRules.useSpaces).toBe(true);
      expect(grammar.whitespaceRules.trimTrailingWhitespace).toBe(true);
    });
  });

  describe('Error Recovery', () => {
    it('should handle partial C++ code gracefully', async () => {
      const partialCode = `class MyClass {
public:
    void method() {
        // Incomplete method`;

      const parseResult = await parserManager.parse(partialCode, 'cpp');
      expect(parseResult).toBeDefined();
      expect(parseResult.tokens.length).toBeGreaterThan(0);
    });

    it('should handle partial Rust code gracefully', async () => {
      const partialCode = `struct Person {
    name: String,
    age: u32,
    // Incomplete struct`;

      const parseResult = await parserManager.parse(partialCode, 'rust');
      expect(parseResult).toBeDefined();
      expect(parseResult.tokens.length).toBeGreaterThan(0);
    });
  });
});
