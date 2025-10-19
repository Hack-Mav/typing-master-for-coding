/**
 * Web Worker for parsing operations
 * Offloads Tree-sitter parsing to prevent UI blocking
 */

// @ts-ignore - web-tree-sitter has complex types that don't work well with --isolatedModules
import Parser from 'web-tree-sitter';

// Make this file a module for TypeScript's --isolatedModules
export {};

interface ParsingMessage {
  type: 'init' | 'parse' | 'tokenize';
  language?: string;
  code?: string;
  id?: string;
}

interface ParsingResponse {
  type: 'ready' | 'parsed' | 'tokenized' | 'error';
  id?: string;
  result?: any;
  error?: string;
}

let parser: any = null;
const loadedLanguages: Map<string, any> = new Map();

// Initialize Tree-sitter
async function initializeParser() {
  try {
    // @ts-ignore - Parser.init exists at runtime
    await Parser.init({
      locateFile(scriptName: string) {
        return `/${scriptName}`;
      },
    });
    // @ts-ignore - Parser is constructable at runtime
    parser = new Parser();
    postMessage({ type: 'ready' } as ParsingResponse);
  } catch (error) {
    postMessage({
      type: 'error',
      error: `Failed to initialize parser: ${error}`,
    } as ParsingResponse);
  }
}

// Load language grammar
async function loadLanguage(language: string): Promise<any> {
  if (loadedLanguages.has(language)) {
    return loadedLanguages.get(language)!;
  }

  const languageMap: Record<string, string> = {
    python: 'tree-sitter-python.wasm',
    javascript: 'tree-sitter-javascript.wasm',
    typescript: 'tree-sitter-typescript.wasm',
    cpp: 'tree-sitter-cpp.wasm',
    rust: 'tree-sitter-rust.wasm',
    yaml: 'tree-sitter-yaml.wasm',
  };

  const wasmFile = languageMap[language.toLowerCase()];
  if (!wasmFile) {
    throw new Error(`Unsupported language: ${language}`);
  }

  try {
    // @ts-ignore - Parser.Language.load exists at runtime
    const lang = await Parser.Language.load(`/${wasmFile}`);
    loadedLanguages.set(language, lang);
    return lang;
  } catch (error) {
    throw new Error(`Failed to load language ${language}: ${error}`);
  }
}

// Parse code and return AST
async function parseCode(language: string, code: string): Promise<any> {
  if (!parser) {
    throw new Error('Parser not initialized');
  }

  const lang = await loadLanguage(language);
  parser.setLanguage(lang);

  const tree = parser.parse(code);
  const rootNode = tree.rootNode;

  // Convert tree to serializable format
  function nodeToObject(node: any): any {
    return {
      type: node.type,
      startPosition: node.startPosition,
      endPosition: node.endPosition,
      startIndex: node.startIndex,
      endIndex: node.endIndex,
      text: node.text,
      isNamed: node.isNamed,
      isMissing: node.isMissing,
      hasError: node.hasError,
      children: node.children.map(nodeToObject),
    };
  }

  return {
    rootNode: nodeToObject(rootNode),
    hasError: rootNode.hasError,
  };
}

// Tokenize code
async function tokenizeCode(language: string, code: string): Promise<any[]> {
  if (!parser) {
    throw new Error('Parser not initialized');
  }

  const lang = await loadLanguage(language);
  parser.setLanguage(lang);

  const tree = parser.parse(code);
  const tokens: any[] = [];

  function extractTokens(node: any) {
    if (node.childCount === 0) {
      // Leaf node - this is a token
      tokens.push({
        type: node.type,
        text: node.text,
        startPosition: node.startPosition,
        endPosition: node.endPosition,
        startIndex: node.startIndex,
        endIndex: node.endIndex,
      });
    } else {
      // Recurse into children
      for (const child of node.children) {
        extractTokens(child);
      }
    }
  }

  extractTokens(tree.rootNode);
  return tokens;
}

// Message handler
globalThis.onmessage = async (event: MessageEvent<ParsingMessage>) => {
  const { type, language, code, id } = event.data;

  try {
    switch (type) {
      case 'init':
        await initializeParser();
        break;

      case 'parse':
        if (!language || !code) {
          throw new Error('Language and code required for parsing');
        }
        const parseResult = await parseCode(language, code);
        postMessage({
          type: 'parsed',
          id,
          result: parseResult,
        } as ParsingResponse);
        break;

      case 'tokenize':
        if (!language || !code) {
          throw new Error('Language and code required for tokenization');
        }
        const tokens = await tokenizeCode(language, code);
        postMessage({
          type: 'tokenized',
          id,
          result: tokens,
        } as ParsingResponse);
        break;

      default:
        throw new Error(`Unknown message type: ${type}`);
    }
  } catch (error) {
    postMessage({
      type: 'error',
      id,
      error: error instanceof Error ? error.message : String(error),
    } as ParsingResponse);
  }
};

// Auto-initialize on load
initializeParser();
