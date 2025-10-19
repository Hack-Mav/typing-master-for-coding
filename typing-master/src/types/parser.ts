// Parser-related type definitions

export type Language = 'python' | 'javascript' | 'yaml' | 'cpp' | 'rust';

export interface Token {
  type: string;
  value: string;
  startIndex: number;
  endIndex: number;
  startPosition: { row: number; column: number };
  endPosition: { row: number; column: number };
}

export interface ASTNode {
  type: string;
  startIndex: number;
  endIndex: number;
  startPosition: { row: number; column: number };
  endPosition: { row: number; column: number };
  children: ASTNode[];
  text: string;
}

export interface ParseResult {
  tokens: Token[];
  ast: ASTNode;
  errors: ParseError[];
}

export interface ParseError {
  message: string;
  startIndex: number;
  endIndex: number;
  startPosition: { row: number; column: number };
  endPosition: { row: number; column: number };
}

export interface ParserConfig {
  language: Language;
  wasmPath: string;
  grammarName: string;
}

export interface LanguageGrammar {
  tokenTypes: string[];
  keywords: string[];
  operators: string[];
  delimiters: string[];
  whitespaceRules: {
    indentSize: number;
    useSpaces: boolean;
    trimTrailingWhitespace: boolean;
  };
}
