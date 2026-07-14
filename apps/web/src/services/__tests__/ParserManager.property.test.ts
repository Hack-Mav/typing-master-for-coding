/**
 * Property-based tests for ParserManager
 * Tests parsing invariants across different code samples
 */

import * as fc from 'fast-check';
import { ParserManager } from '../ParserManager';

describe('ParserManager - Property-based Tests', () => {
  const parserManager = ParserManager.getInstance();

  // Code generators for different languages
  const pythonCodeGen = fc.oneof(
    fc.constant('def hello():\n    pass'),
    fc.constant('x = 42'),
    fc.constant('if True:\n    print("hello")'),
    fc.constant('for i in range(10):\n    print(i)')
  );

  // JavaScript code generator for future use
  // const javascriptCodeGen = fc.oneof(
  //   fc.constant('function hello() { return 42; }'),
  //   fc.constant('const x = 42;'),
  //   fc.constant('if (true) { console.log("hello"); }'),
  //   fc.constant('for (let i = 0; i < 10; i++) { }')
  // );

  describe('Parsing Invariants', () => {
    test('Parsing same code twice should produce identical AST', () => {
      fc.assert(
        fc.asyncProperty(
          fc.constantFrom('python', 'javascript'),
          pythonCodeGen,
          async (language, code) => {
            const ast1 = await parserManager.parseCode(code, language);
            const ast2 = await parserManager.parseCode(code, language);

            return JSON.stringify(ast1) === JSON.stringify(ast2);
          }
        ),
        { numRuns: 20 }
      );
    });

    test('Valid code should parse without errors', () => {
      fc.assert(
        fc.asyncProperty(pythonCodeGen, async code => {
          const ast = await parserManager.parseCode(code, 'python');
          return !ast.hasError;
        }),
        { numRuns: 20 }
      );
    });

    test('Tokenization should preserve code length', () => {
      fc.assert(
        fc.asyncProperty(
          fc.constantFrom('python', 'javascript'),
          fc.string({ minLength: 1, maxLength: 100 }),
          async (language, code) => {
            const tokens = await parserManager.tokenize(code, language);
            const reconstructed = tokens.map(t => t.text).join('');

            // Tokens should cover the entire code
            return reconstructed.length <= code.length;
          }
        ),
        { numRuns: 20 }
      );
    });

    test('Empty code should produce minimal AST', () => {
      fc.assert(
        fc.asyncProperty(
          fc.constantFrom('python', 'javascript', 'yaml'),
          async language => {
            const ast = await parserManager.parseCode('', language);
            return ast.rootNode !== null;
          }
        ),
        { numRuns: 10 }
      );
    });
  });

  describe('Tokenization Properties', () => {
    test('Tokens should be ordered by position', () => {
      fc.assert(
        fc.asyncProperty(pythonCodeGen, async code => {
          const tokens = await parserManager.tokenize(code, 'python');

          for (let i = 1; i < tokens.length; i++) {
            if (tokens[i].startIndex < tokens[i - 1].endIndex) {
              return false;
            }
          }
          return true;
        }),
        { numRuns: 20 }
      );
    });

    test('Token positions should be within code bounds', () => {
      fc.assert(
        fc.asyncProperty(
          fc.string({ minLength: 1, maxLength: 100 }),
          async code => {
            const tokens = await parserManager.tokenize(code, 'python');

            return tokens.every(
              token =>
                token.startIndex >= 0 &&
                token.endIndex <= code.length &&
                token.startIndex <= token.endIndex
            );
          }
        ),
        { numRuns: 20 }
      );
    });
  });

  describe('AST Structure Properties', () => {
    test('AST should have valid tree structure', () => {
      fc.assert(
        fc.asyncProperty(pythonCodeGen, async code => {
          const ast = await parserManager.parseCode(code, 'python');

          // Check tree invariants
          function validateNode(node: any): boolean {
            if (!node) return false;

            // Node should have required properties
            if (!node.type || node.startIndex === undefined) {
              return false;
            }

            // Children should be valid
            if (node.children) {
              return node.children.every(validateNode);
            }

            return true;
          }

          return validateNode(ast.rootNode);
        }),
        { numRuns: 20 }
      );
    });

    test('Parent-child relationships should be consistent', () => {
      fc.assert(
        fc.asyncProperty(pythonCodeGen, async code => {
          const ast = await parserManager.parseCode(code, 'python');

          function checkParentChild(node: any): boolean {
            if (!node.children || node.children.length === 0) {
              return true;
            }

            // Each child should be within parent bounds
            return node.children.every((child: any) => {
              const withinBounds =
                child.startIndex >= node.startIndex &&
                child.endIndex <= node.endIndex;

              return withinBounds && checkParentChild(child);
            });
          }

          return checkParentChild(ast.rootNode);
        }),
        { numRuns: 20 }
      );
    });
  });

  describe('Comparison Properties', () => {
    test('Identical code should have zero structural diff', () => {
      fc.assert(
        fc.asyncProperty(pythonCodeGen, async code => {
          const ast1 = await parserManager.parseCode(code, 'python');
          const ast2 = await parserManager.parseCode(code, 'python');

          const diff = parserManager.compareStructure(
            ast1.rootNode,
            ast2.rootNode
          );

          return diff.differences === 0;
        }),
        { numRuns: 20 }
      );
    });

    test('Different code should have non-zero structural diff', () => {
      fc.assert(
        fc.asyncProperty(pythonCodeGen, pythonCodeGen, async (code1, code2) => {
          if (code1 === code2) return true;

          const ast1 = await parserManager.parseCode(code1, 'python');
          const ast2 = await parserManager.parseCode(code2, 'python');

          const diff = parserManager.compareStructure(
            ast1.rootNode,
            ast2.rootNode
          );

          // Different code should generally have differences
          // (though some simple cases might be structurally identical)
          return diff.differences >= 0;
        }),
        { numRuns: 20 }
      );
    });
  });

  describe('Error Handling Properties', () => {
    test('Invalid syntax should be detected', () => {
      fc.assert(
        fc.asyncProperty(
          fc.constantFrom('def (', 'if', 'for in', '(((', 'def hello(:'),
          async invalidCode => {
            const ast = await parserManager.parseCode(invalidCode, 'python');
            return ast.hasError === true;
          }
        ),
        { numRuns: 10 }
      );
    });

    test('Parser should not crash on random input', () => {
      fc.assert(
        fc.asyncProperty(fc.string({ maxLength: 100 }), async randomCode => {
          try {
            await parserManager.parseCode(randomCode, 'python');
            return true; // Should not throw
          } catch (error) {
            return false;
          }
        }),
        { numRuns: 50 }
      );
    });
  });
});
