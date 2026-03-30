# Tree-sitter Parser Integration

This directory contains the Tree-sitter WASM parser integration for syntax-aware typing practice.

## Overview

The parser integration provides:

- **ParserManager**: Core Tree-sitter WASM parser management with lazy loading
- **TokenizationService**: Enhanced tokenization with metadata for typing practice
- Language support for Python, JavaScript, and YAML (MVP languages)

## Architecture

### ParserManager

- Singleton pattern for efficient resource management
- Lazy loading of language parsers to minimize initial bundle size
- WASM file management with proper initialization
- Language-specific grammar configurations
- AST parsing and token extraction

### TokenizationService

- Enhanced token metadata (difficulty, expected duration, categories)
- Token sequence validation with delimiter balance checking
- Whitespace policy enforcement per language
- Token comparison for typing validation
- Error analysis and pattern detection

## Usage

### Basic Parsing

```typescript
import { parserManager } from './services/ParserManager';

// Initialize parser (only needed once)
await parserManager.initialize();

// Load language parser
const parser = await parserManager.loadParser('python');

// Tokenize code
const tokens = await parserManager.tokenize('def hello(): pass', 'python');

// Parse AST
const ast = await parserManager.parseAST('def hello(): pass', 'python');
```

### Enhanced Tokenization

```typescript
import { tokenizationService } from './services/TokenizationService';

// Get enhanced tokens with metadata
const tokens = await tokenizationService.tokenizeWithMetadata(code, 'python');

// Get comprehensive breakdown
const breakdown = await tokenizationService.getTokenBreakdown(code, 'python');

// Validate token sequence
const validation = tokenizationService.validateTokenSequence(tokens, 'python');

// Compare token sequences for typing validation
const comparison = tokenizationService.compareTokenSequences(expected, actual);
```

## Language Support

### Python

- Keywords: `def`, `class`, `if`, `else`, `for`, `while`, `import`, etc.
- Operators: `+`, `-`, `*`, `/`, `and`, `or`, `not`, etc.
- Delimiters: `(`, `)`, `[`, `]`, `{`, `}`, `:`, `,`, `.`
- Whitespace: 4 spaces, no tabs, trim trailing whitespace

### JavaScript

- Keywords: `function`, `const`, `let`, `var`, `if`, `else`, `for`, `while`, etc.
- Operators: `+`, `-`, `*`, `/`, `===`, `!==`, `&&`, `||`, etc.
- Delimiters: `(`, `)`, `[`, `]`, `{`, `}`, `;`, `,`, `.`
- Whitespace: 2 spaces, no tabs, trim trailing whitespace

### YAML

- Keywords: `true`, `false`, `null`
- Operators: `:`, `-`, `|`, `>`
- Delimiters: `[`, `]`, `{`, `}`, `,`
- Whitespace: 2 spaces, no tabs, trim trailing whitespace

## WASM Files

The following WASM files must be available in the public directory:

- `tree-sitter.wasm` - Core Tree-sitter runtime
- `tree-sitter-python.wasm` - Python language grammar
- `tree-sitter-javascript.wasm` - JavaScript language grammar
- `tree-sitter-yaml.wasm` - YAML language grammar

## Performance Considerations

- **Lazy Loading**: Parsers are loaded on-demand to minimize initial bundle size
- **Caching**: Loaded parsers are cached for reuse
- **Web Workers**: Consider moving parsing to Web Workers for large files
- **Memory Management**: Parsers are properly disposed when no longer needed

## Error Handling

- Graceful fallback for unsupported languages
- Parse error detection and reporting
- Network error handling for WASM file loading
- Memory management for large syntax trees

## Testing

Run the integration tests:

```bash
npm test -- --testPathPattern=ParserIntegration.test.ts
```

The tests verify:

- API contracts and singleton patterns
- Language support and grammar configurations
- Token validation and comparison logic
- Whitespace extraction and delimiter balance checking

## Future Enhancements

- Support for C++ and Rust languages
- Advanced error recovery and partial parsing
- Performance optimization with Web Workers
- Custom grammar rule configuration
- Real-time incremental parsing for large files
