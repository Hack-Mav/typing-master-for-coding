import { Parser, Language as TreeSitterLanguage } from 'web-tree-sitter';
import {
  Language,
  Token,
  ASTNode,
  ParseResult,
  ParseError,
  ParserConfig,
  LanguageGrammar,
} from '../types/parser';

/**
 * ParserManager handles Tree-sitter WASM parser integration with lazy loading
 * and language-specific grammar handling for syntax-aware typing practice.
 */
export class ParserManager {
  private static instance: ParserManager;
  private parsers: Map<Language, any> = new Map();
  private isInitialized = false;
  private initPromise: Promise<void> | null = null;

  // Language configurations for all supported languages
  private readonly languageConfigs: Record<Language, ParserConfig> = {
    python: {
      language: 'python',
      wasmPath: '/tree-sitter-python.wasm',
      grammarName: 'tree-sitter-python',
    },
    javascript: {
      language: 'javascript',
      wasmPath: '/tree-sitter-javascript.wasm',
      grammarName: 'tree-sitter-javascript',
    },
    yaml: {
      language: 'yaml',
      wasmPath: '/tree-sitter-yaml.wasm',
      grammarName: 'tree-sitter-yaml',
    },
    cpp: {
      language: 'cpp',
      wasmPath: '/tree-sitter-cpp.wasm',
      grammarName: 'tree-sitter-cpp',
    },
    rust: {
      language: 'rust',
      wasmPath: '/tree-sitter-rust.wasm',
      grammarName: 'tree-sitter-rust',
    },
  };

  // Language-specific grammar rules and token types
  private readonly languageGrammars: Record<Language, LanguageGrammar> = {
    python: {
      tokenTypes: [
        'identifier',
        'string',
        'number',
        'keyword',
        'operator',
        'delimiter',
        'comment',
      ],
      keywords: [
        'def',
        'class',
        'if',
        'else',
        'elif',
        'for',
        'while',
        'try',
        'except',
        'import',
        'from',
        'return',
      ],
      operators: [
        '+',
        '-',
        '*',
        '/',
        '//',
        '%',
        '**',
        '==',
        '!=',
        '<',
        '>',
        '<=',
        '>=',
        'and',
        'or',
        'not',
      ],
      delimiters: ['(', ')', '[', ']', '{', '}', ':', ',', '.'],
      whitespaceRules: {
        indentSize: 4,
        useSpaces: true,
        trimTrailingWhitespace: true,
      },
    },
    javascript: {
      tokenTypes: [
        'identifier',
        'string',
        'number',
        'keyword',
        'operator',
        'delimiter',
        'comment',
      ],
      keywords: [
        'function',
        'const',
        'let',
        'var',
        'if',
        'else',
        'for',
        'while',
        'return',
        'class',
        'import',
        'export',
      ],
      operators: [
        '+',
        '-',
        '*',
        '/',
        '%',
        '==',
        '===',
        '!=',
        '!==',
        '<',
        '>',
        '<=',
        '>=',
        '&&',
        '||',
        '!',
      ],
      delimiters: ['(', ')', '[', ']', '{', '}', ';', ',', '.'],
      whitespaceRules: {
        indentSize: 2,
        useSpaces: true,
        trimTrailingWhitespace: true,
      },
    },
    yaml: {
      tokenTypes: [
        'key',
        'value',
        'string',
        'number',
        'boolean',
        'null',
        'comment',
      ],
      keywords: ['true', 'false', 'null'],
      operators: [':', '-', '|', '>'],
      delimiters: ['[', ']', '{', '}', ','],
      whitespaceRules: {
        indentSize: 2,
        useSpaces: true,
        trimTrailingWhitespace: true,
      },
    },
    cpp: {
      tokenTypes: [
        'identifier',
        'string',
        'number',
        'keyword',
        'operator',
        'delimiter',
        'comment',
        'preprocessor',
        'type',
      ],
      keywords: [
        'auto',
        'bool',
        'break',
        'case',
        'catch',
        'char',
        'class',
        'const',
        'continue',
        'default',
        'delete',
        'do',
        'double',
        'else',
        'enum',
        'explicit',
        'extern',
        'false',
        'float',
        'for',
        'friend',
        'goto',
        'if',
        'inline',
        'int',
        'long',
        'namespace',
        'new',
        'nullptr',
        'operator',
        'private',
        'protected',
        'public',
        'return',
        'short',
        'signed',
        'sizeof',
        'static',
        'struct',
        'switch',
        'template',
        'this',
        'throw',
        'true',
        'try',
        'typedef',
        'typename',
        'union',
        'unsigned',
        'using',
        'virtual',
        'void',
        'volatile',
        'while',
      ],
      operators: [
        '+',
        '-',
        '*',
        '/',
        '%',
        '++',
        '--',
        '==',
        '!=',
        '<',
        '>',
        '<=',
        '>=',
        '&&',
        '||',
        '!',
        '&',
        '|',
        '^',
        '~',
        '<<',
        '>>',
        '=',
        '+=',
        '-=',
        '*=',
        '/=',
        '%=',
        '&=',
        '|=',
        '^=',
        '<<=',
        '>>=',
        '->',
        '::',
        '.*',
        '->*',
      ],
      delimiters: ['(', ')', '[', ']', '{', '}', ';', ',', '.', '<', '>'],
      whitespaceRules: {
        indentSize: 4,
        useSpaces: true,
        trimTrailingWhitespace: true,
      },
    },
    rust: {
      tokenTypes: [
        'identifier',
        'string',
        'number',
        'keyword',
        'operator',
        'delimiter',
        'comment',
        'attribute',
        'lifetime',
        'type',
      ],
      keywords: [
        'as',
        'async',
        'await',
        'break',
        'const',
        'continue',
        'crate',
        'dyn',
        'else',
        'enum',
        'extern',
        'false',
        'fn',
        'for',
        'if',
        'impl',
        'in',
        'let',
        'loop',
        'match',
        'mod',
        'move',
        'mut',
        'pub',
        'ref',
        'return',
        'self',
        'Self',
        'static',
        'struct',
        'super',
        'trait',
        'true',
        'type',
        'unsafe',
        'use',
        'where',
        'while',
        'abstract',
        'become',
        'box',
        'do',
        'final',
        'macro',
        'override',
        'priv',
        'typeof',
        'unsized',
        'virtual',
        'yield',
      ],
      operators: [
        '+',
        '-',
        '*',
        '/',
        '%',
        '==',
        '!=',
        '<',
        '>',
        '<=',
        '>=',
        '&&',
        '||',
        '!',
        '&',
        '|',
        '^',
        '<<',
        '>>',
        '=',
        '+=',
        '-=',
        '*=',
        '/=',
        '%=',
        '&=',
        '|=',
        '^=',
        '<<=',
        '>>=',
        '->',
        '=>',
        '::',
        '..',
        '..=',
        '?',
      ],
      delimiters: ['(', ')', '[', ']', '{', '}', ';', ',', '.', '<', '>'],
      whitespaceRules: {
        indentSize: 4,
        useSpaces: true,
        trimTrailingWhitespace: true,
      },
    },
  };

  private constructor() {}

  /**
   * Get singleton instance of ParserManager
   */
  public static getInstance(): ParserManager {
    if (!ParserManager.instance) {
      ParserManager.instance = new ParserManager();
    }
    return ParserManager.instance;
  }

  /**
   * Initialize Tree-sitter WASM runtime
   */
  public async initialize(): Promise<void> {
    if (this.isInitialized) {
      return;
    }

    if (this.initPromise) {
      return this.initPromise;
    }

    this.initPromise = this.doInitialize();
    return this.initPromise;
  }

  private getWasmPath(wasmPath: string): string {
    // In Node-based test environments (Jest/jsdom) process.cwd is available and
    // points to the project root. Point to the public/ WASM files so the tests
    // can find them instead of resolving absolute paths to the drive root.
    if (typeof process !== 'undefined' && typeof process.cwd === 'function') {
      const cwd = process.cwd().replace(/\\/g, '/');
      const scriptName = wasmPath.replace(/^\//, '');
      return `${cwd}/public/${scriptName}`;
    }
    return wasmPath.startsWith('/') ? wasmPath : `/${wasmPath}`;
  }

  private async doInitialize(): Promise<void> {
    try {
      await Parser.init({
        locateFile: (scriptName: string, _scriptDirectory: string) =>
          this.getWasmPath(scriptName),
      });
      this.isInitialized = true;
    } catch (error) {
      console.error('Failed to initialize Tree-sitter:', error);
      throw new Error('Tree-sitter initialization failed');
    }
  }

  /**
   * Lazy load parser for specific language
   */
  public async loadParser(language: Language): Promise<any> {
    await this.initialize();

    // Return cached parser if already loaded
    if (this.parsers.has(language)) {
      return this.parsers.get(language)!;
    }

    const config = this.languageConfigs[language];
    if (!config) {
      throw new Error(`Unsupported language: ${language}`);
    }

    try {
      const parser = new Parser();

      // Load language grammar from WASM file
      let wasmInput: string | Uint8Array = this.getWasmPath(config.wasmPath);

      // In Node-based test environments, read the WASM file directly and pass
      // it as a Uint8Array. This avoids web-tree-sitter's dynamic import of
      // fs/promises, which Jest cannot execute without --experimental-vm-modules.
      if (
        typeof process !== 'undefined' &&
        process.versions &&
        process.versions.node
      ) {
        const fs = require('fs');
        // Convert to a plain Uint8Array; in some test contexts (jsdom/vm) a
        // Node Buffer is not an instance of the realm's Uint8Array, which
        // causes web-tree-sitter's Language.load to fall back to dynamic import.
        wasmInput = new Uint8Array(fs.readFileSync(wasmInput));
      }

      const languageGrammar = await TreeSitterLanguage.load(wasmInput);
      parser.setLanguage(languageGrammar);

      // Cache the parser for reuse
      this.parsers.set(language, parser);

      return parser;
    } catch (error) {
      console.error(`Failed to load parser for ${language}:`, error);
      throw new Error(`Failed to load ${language} parser`);
    }
  }

  /**
   * Tokenize code using language-specific grammar
   */
  public async tokenize(code: string, language: Language): Promise<Token[]> {
    const parser = await this.loadParser(language);
    const tree = parser.parse(code);

    const tokens: Token[] = [];
    this.extractTokens(tree.rootNode, tokens);

    // Sort tokens by position
    tokens.sort((a, b) => a.startIndex - b.startIndex);

    return tokens;
  }

  private extractTokens(node: any, tokens: Token[]): void {
    // Only include leaf nodes as tokens (nodes without children)
    if (node.childCount === 0 && node.text.trim()) {
      tokens.push({
        type: node.type,
        value: node.text,
        startIndex: node.startIndex,
        endIndex: node.endIndex,
        startPosition: node.startPosition,
        endPosition: node.endPosition,
      });
      return;
    }

    for (let i = 0; i < node.childCount; i++) {
      const child = node.child(i);
      const next = node.child(i + 1);

      if (this.shouldMergeTokens(child, next)) {
        tokens.push({
          type: child.type,
          value: child.text + next.text,
          startIndex: child.startIndex,
          endIndex: next.endIndex,
          startPosition: child.startPosition,
          endPosition: next.endPosition,
        });
        i++; // skip the merged token
      } else {
        this.extractTokens(child, tokens);
      }
    }
  }

  private shouldMergeTokens(child: any, next: any): boolean {
    if (!child || !next) return false;

    if (child.type === 'identifier' && next.text === '!') {
      return true;
    }

    if (
      child.text === "'" &&
      child.parent?.type === 'lifetime' &&
      next.childCount === 0 &&
      next.text.trim()
    ) {
      return true;
    }

    return false;
  }

  /**
   * Parse code and return AST
   */
  public async parseAST(code: string, language: Language): Promise<ASTNode> {
    const parser = await this.loadParser(language);
    const tree = parser.parse(code);

    return this.convertTreeToAST(tree.rootNode);
  }

  /**
   * Parse code and return comprehensive result with tokens, AST, and errors
   */
  public async parse(code: string, language: Language): Promise<ParseResult> {
    const parser = await this.loadParser(language);
    const tree = parser.parse(code);

    const tokens = await this.tokenize(code, language);
    const ast = this.convertTreeToAST(tree.rootNode);
    const errors = this.extractParseErrors(tree.rootNode);

    return {
      tokens,
      ast,
      errors,
    };
  }

  /**
   * Parse code and return AST result with error flag.
   * Convenience method used by property-based tests.
   */
  public async parseCode(
    code: string,
    language: Language
  ): Promise<{ rootNode: ASTNode; hasError: boolean }> {
    const result = await this.parse(code, language);
    return {
      rootNode: result.ast,
      hasError: result.errors.length > 0,
    };
  }

  /**
   * Compare two AST structures and return the number of differences.
   */
  public compareStructure(
    node1: ASTNode,
    node2: ASTNode
  ): { differences: number } {
    let differences = 0;

    const compare = (a: ASTNode, b: ASTNode): void => {
      if (a.type !== b.type) differences++;
      if (a.children.length !== b.children.length) differences++;

      const minChildren = Math.min(a.children.length, b.children.length);
      for (let i = 0; i < minChildren; i++) {
        compare(a.children[i], b.children[i]);
      }
    };

    compare(node1, node2);
    return { differences };
  }

  /**
   * Get language grammar configuration
   */
  public getLanguageGrammar(language: Language): LanguageGrammar {
    return this.languageGrammars[language];
  }

  /**
   * Check if language is supported
   */
  public isLanguageSupported(language: string): language is Language {
    return language in this.languageConfigs;
  }

  /**
   * Get list of supported languages
   */
  public getSupportedLanguages(): Language[] {
    return Object.keys(this.languageConfigs) as Language[];
  }

  /**
   * Convert Tree-sitter node to our AST format
   */
  private convertTreeToAST(node: any): ASTNode {
    const children: ASTNode[] = [];

    for (let i = 0; i < node.childCount; i++) {
      children.push(this.convertTreeToAST(node.child(i)));
    }

    return {
      type: node.type,
      startIndex: node.startIndex,
      endIndex: node.endIndex,
      startPosition: node.startPosition,
      endPosition: node.endPosition,
      children,
      text: node.text,
    };
  }

  /**
   * Extract parse errors from syntax tree
   */
  private extractParseErrors(node: any): ParseError[] {
    const errors: ParseError[] = [];

    const visitNode = (currentNode: any) => {
      // Check if node has errors or is missing
      if (currentNode.hasError || currentNode.isMissing) {
        errors.push({
          message: currentNode.isMissing ? 'Missing node' : 'Syntax error',
          startIndex: currentNode.startIndex,
          endIndex: currentNode.endIndex,
          startPosition: currentNode.startPosition,
          endPosition: currentNode.endPosition,
        });
      }

      // Recursively check children
      for (let i = 0; i < currentNode.childCount; i++) {
        visitNode(currentNode.child(i));
      }
    };

    visitNode(node);
    return errors;
  }

  /**
   * Clean up resources
   */
  public dispose(): void {
    this.parsers.clear();
    this.isInitialized = false;
    this.initPromise = null;
  }
}

// Export singleton instance
export const parserManager = ParserManager.getInstance();
