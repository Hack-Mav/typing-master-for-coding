import React, { useState, useEffect } from 'react';
import { parserManager } from '../services/ParserManager';
import { tokenizationService } from '../services/TokenizationService';
import { Language } from '../types/parser';
import { EnhancedToken } from '../services/TokenizationService';

/**
 * Example component demonstrating Tree-sitter parser integration
 * This shows how to use the ParserManager and TokenizationService
 */
export const ParserExample: React.FC = () => {
  const [selectedLanguage, setSelectedLanguage] = useState<Language>('python');
  const [code, setCode] = useState<string>(
    'def hello_world():\n    print("Hello, World!")'
  );
  const [tokens, setTokens] = useState<EnhancedToken[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [breakdown, setBreakdown] = useState<any>(null);

  useEffect(() => {
    // Sample code for different languages
    const sampleCode: Record<Language, string> = {
      python: 'def hello_world():\n    print("Hello, World!")\n    return True',
      javascript:
        'function helloWorld() {\n    console.log("Hello, World!");\n    return true;\n}',
      yaml: 'name: example\nversion: 1.0.0\ndependencies:\n  - react\n  - typescript',
    };

    setCode(sampleCode[selectedLanguage]);
  }, [selectedLanguage]);

  const parseCode = async () => {
    if (!code.trim()) return;

    setIsLoading(true);
    setError(null);

    try {
      // Initialize parser if needed
      await parserManager.initialize();

      // Get enhanced tokens
      const enhancedTokens = await tokenizationService.tokenizeWithMetadata(
        code,
        selectedLanguage
      );
      setTokens(enhancedTokens);

      // Get breakdown analysis
      const analysis = await tokenizationService.getTokenBreakdown(
        code,
        selectedLanguage
      );
      setBreakdown(analysis);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error occurred');
      console.error('Parser error:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const getCategoryColor = (category: string): string => {
    const colors: Record<string, string> = {
      keyword: 'text-blue-600 bg-blue-100',
      identifier: 'text-green-600 bg-green-100',
      string: 'text-red-600 bg-red-100',
      number: 'text-purple-600 bg-purple-100',
      operator: 'text-orange-600 bg-orange-100',
      delimiter: 'text-gray-600 bg-gray-100',
      comment: 'text-gray-500 bg-gray-50',
      whitespace: 'text-gray-300 bg-gray-50',
      unknown: 'text-black bg-yellow-100',
    };
    return colors[category] || 'text-black bg-gray-100';
  };

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <h1 className="text-3xl font-bold mb-6">
        Tree-sitter Parser Integration Demo
      </h1>

      {/* Language Selection */}
      <div className="mb-4">
        <label className="block text-sm font-medium mb-2">
          Select Language:
        </label>
        <select
          value={selectedLanguage}
          onChange={e => setSelectedLanguage(e.target.value as Language)}
          className="border border-gray-300 rounded px-3 py-2"
        >
          {parserManager.getSupportedLanguages().map(lang => (
            <option key={lang} value={lang}>
              {lang.charAt(0).toUpperCase() + lang.slice(1)}
            </option>
          ))}
        </select>
      </div>

      {/* Code Input */}
      <div className="mb-4">
        <label className="block text-sm font-medium mb-2">Code to Parse:</label>
        <textarea
          value={code}
          onChange={e => setCode(e.target.value)}
          className="w-full h-32 border border-gray-300 rounded px-3 py-2 font-mono text-sm"
          placeholder="Enter your code here..."
        />
      </div>

      {/* Parse Button */}
      <button
        onClick={parseCode}
        disabled={isLoading || !code.trim()}
        className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 disabled:bg-gray-400 mb-6"
      >
        {isLoading ? 'Parsing...' : 'Parse Code'}
      </button>

      {/* Error Display */}
      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          <strong>Error:</strong> {error}
        </div>
      )}

      {/* Results */}
      {tokens.length > 0 && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Token List */}
          <div>
            <h2 className="text-xl font-semibold mb-3">Parsed Tokens</h2>
            <div className="border border-gray-300 rounded p-4 max-h-96 overflow-y-auto">
              {tokens.map((token, index) => (
                <div
                  key={index}
                  className="mb-2 p-2 border-b border-gray-200 last:border-b-0"
                >
                  <div className="flex items-center gap-2 mb-1">
                    <span
                      className={`px-2 py-1 rounded text-xs font-medium ${getCategoryColor(token.category)}`}
                    >
                      {token.category}
                    </span>
                    <span className="text-xs text-gray-500">{token.type}</span>
                    {token.isKeyword && (
                      <span className="text-xs bg-blue-200 text-blue-800 px-1 rounded">
                        K
                      </span>
                    )}
                    {token.isOperator && (
                      <span className="text-xs bg-orange-200 text-orange-800 px-1 rounded">
                        O
                      </span>
                    )}
                    {token.isDelimiter && (
                      <span className="text-xs bg-gray-200 text-gray-800 px-1 rounded">
                        D
                      </span>
                    )}
                  </div>
                  <div className="font-mono text-sm bg-gray-50 px-2 py-1 rounded">
                    "{token.value}"
                  </div>
                  <div className="text-xs text-gray-500 mt-1">
                    Difficulty: {token.difficulty}/5 | Duration:{' '}
                    {token.expectedDuration}ms
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Analysis Breakdown */}
          <div>
            <h2 className="text-xl font-semibold mb-3">Analysis Breakdown</h2>
            {breakdown && (
              <div className="border border-gray-300 rounded p-4">
                <div className="grid grid-cols-2 gap-4 mb-4">
                  <div>
                    <div className="text-sm text-gray-600">Total Tokens</div>
                    <div className="text-lg font-semibold">
                      {breakdown.totalTokens}
                    </div>
                  </div>
                  <div>
                    <div className="text-sm text-gray-600">Avg Difficulty</div>
                    <div className="text-lg font-semibold">
                      {breakdown.averageDifficulty.toFixed(1)}/5
                    </div>
                  </div>
                  <div>
                    <div className="text-sm text-gray-600">Est. Duration</div>
                    <div className="text-lg font-semibold">
                      {(breakdown.estimatedDuration / 1000).toFixed(1)}s
                    </div>
                  </div>
                </div>

                <h3 className="font-medium mb-2">Tokens by Category</h3>
                <div className="space-y-1">
                  {Object.entries(breakdown.tokensByCategory).map(
                    ([category, count]) =>
                      (count as number) > 0 && (
                        <div
                          key={category}
                          className="flex justify-between items-center"
                        >
                          <span
                            className={`px-2 py-1 rounded text-xs ${getCategoryColor(category)}`}
                          >
                            {category}
                          </span>
                          <span className="font-medium">{count as number}</span>
                        </div>
                      )
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Usage Instructions */}
      <div className="mt-8 p-4 bg-blue-50 rounded">
        <h3 className="font-medium mb-2">How to Use:</h3>
        <ol className="list-decimal list-inside text-sm space-y-1">
          <li>Select a programming language from the dropdown</li>
          <li>Enter or modify the code in the textarea</li>
          <li>Click "Parse Code" to analyze the syntax</li>
          <li>View the tokenized results and analysis breakdown</li>
        </ol>
        <p className="text-xs text-gray-600 mt-2">
          This demonstrates the Tree-sitter parser integration that powers
          syntax-aware typing practice.
        </p>
      </div>
    </div>
  );
};

export default ParserExample;
