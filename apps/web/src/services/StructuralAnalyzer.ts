import {
  StructuralConformityScore,
  StructuralPenalty,
  StructuralPenaltyType,
  ASTConformityAnalysis,
  ASTNodeDifference,
} from '../types/assessment';
import { Language } from '../types/parser';
import { parserManager } from './ParserManager';

/**
 * Structural Analyzer - Performs AST-shape conformity checking and structural validation
 */
class StructuralAnalyzer {
  /**
   * Analyze structural conformity between expected and actual code
   */
  async analyzeStructuralConformity(
    expectedCode: string,
    actualCode: string,
    language: Language
  ): Promise<StructuralConformityScore> {
    try {
      // Parse both code snippets to AST
      const expectedAST = await parserManager.parseAST(expectedCode, language);
      const actualAST = await parserManager.parseAST(actualCode, language);

      // Perform AST conformity analysis
      const astAnalysis = this.compareASTStructures(expectedAST, actualAST);

      // Calculate token sequence accuracy
      const expectedTokens = await parserManager.tokenize(
        expectedCode,
        language
      );
      const actualTokens = await parserManager.tokenize(actualCode, language);
      const tokenAccuracy = this.calculateTokenSequenceAccuracy(
        expectedTokens,
        actualTokens
      );

      // Validate syntax correctness
      const syntaxScore = this.validateSyntaxCorrectness(actualCode, language);

      // Identify structural penalties
      const penalties = this.identifyStructuralPenalties(
        expectedCode,
        actualCode,
        astAnalysis
      );

      // Calculate overall conformity score
      const overallConformity = this.calculateOverallConformity(
        astAnalysis.structuralSimilarity,
        tokenAccuracy,
        syntaxScore,
        penalties
      );

      return {
        astSimilarity: astAnalysis.structuralSimilarity,
        tokenSequenceAccuracy: tokenAccuracy,
        syntaxValidationScore: syntaxScore,
        structuralPenalties: penalties,
        overallConformity,
      };
    } catch (error) {
      console.error('Error analyzing structural conformity:', error);
      // Return fallback analysis based on string comparison
      return this.fallbackAnalysis(expectedCode, actualCode);
    }
  }

  /**
   * Compare AST structures for similarity
   */
  private compareASTStructures(
    expectedAST: any,
    actualAST: any
  ): ASTConformityAnalysis {
    if (!expectedAST || !actualAST) {
      return {
        structuralSimilarity: 0,
        nodeTypeMatches: 0,
        nodeTypeMismatches: 0,
        depthSimilarity: 0,
        branchingSimilarity: 0,
        missingNodes: [],
        extraNodes: [],
        mismatchedNodes: [],
      };
    }

    // Flatten AST nodes for comparison
    const expectedNodes = this.flattenAST(expectedAST);
    const actualNodes = this.flattenAST(actualAST);

    // Calculate node type matches
    const nodeTypeMatches = this.countNodeTypeMatches(
      expectedNodes,
      actualNodes
    );
    const nodeTypeMismatches = Math.abs(
      expectedNodes.length - actualNodes.length
    );

    // Calculate structural similarity
    const structuralSimilarity =
      expectedNodes.length > 0
        ? nodeTypeMatches / Math.max(expectedNodes.length, actualNodes.length)
        : 0;

    // Calculate depth similarity
    const expectedDepth = this.calculateASTDepth(expectedAST);
    const actualDepth = this.calculateASTDepth(actualAST);
    const depthSimilarity =
      1 -
      Math.abs(expectedDepth - actualDepth) /
        Math.max(expectedDepth, actualDepth, 1);

    // Calculate branching similarity
    const expectedBranching = this.calculateBranchingFactor(expectedAST);
    const actualBranching = this.calculateBranchingFactor(actualAST);
    const branchingSimilarity =
      1 -
      Math.abs(expectedBranching - actualBranching) /
        Math.max(expectedBranching, actualBranching, 1);

    // Identify node differences
    const { missingNodes, extraNodes, mismatchedNodes } =
      this.identifyNodeDifferences(expectedNodes, actualNodes);

    return {
      structuralSimilarity,
      nodeTypeMatches,
      nodeTypeMismatches,
      depthSimilarity,
      branchingSimilarity,
      missingNodes,
      extraNodes,
      mismatchedNodes,
    };
  }

  /**
   * Calculate token sequence accuracy
   */
  private calculateTokenSequenceAccuracy(
    expectedTokens: any[],
    actualTokens: any[]
  ): number {
    if (expectedTokens.length === 0) return actualTokens.length === 0 ? 1 : 0;

    let matches = 0;
    const minLength = Math.min(expectedTokens.length, actualTokens.length);

    for (let i = 0; i < minLength; i++) {
      if (this.tokensMatch(expectedTokens[i], actualTokens[i])) {
        matches++;
      }
    }

    // Penalize for length differences
    const lengthPenalty =
      Math.abs(expectedTokens.length - actualTokens.length) /
      Math.max(expectedTokens.length, actualTokens.length);
    const baseAccuracy = matches / expectedTokens.length;

    return Math.max(0, baseAccuracy - lengthPenalty * 0.5);
  }

  /**
   * Validate syntax correctness
   */
  private validateSyntaxCorrectness(code: string, language: Language): number {
    try {
      // Check for basic syntax issues
      const syntaxIssues = this.detectSyntaxIssues(code, language);

      // Calculate syntax score based on issues found
      const maxIssues = Math.max(10, code.length / 50); // Scale with code length
      const syntaxScore = Math.max(0, 1 - syntaxIssues.length / maxIssues);

      return syntaxScore;
    } catch (error) {
      console.error('Error validating syntax:', error);
      return 0.5; // Neutral score on error
    }
  }

  /**
   * Identify structural penalties
   */
  private identifyStructuralPenalties(
    expectedCode: string,
    actualCode: string,
    astAnalysis: ASTConformityAnalysis
  ): StructuralPenalty[] {
    const penalties: StructuralPenalty[] = [];

    // Add penalties for missing nodes
    astAnalysis.missingNodes.forEach((node, _index) => {
      penalties.push({
        type: 'missing_token',
        severity: node.severity,
        position: node.position,
        description: `Missing ${node.expectedType} at position ${node.position}`,
        penaltyPoints: this.calculatePenaltyPoints(
          'missing_token',
          node.severity
        ),
      });
    });

    // Add penalties for extra nodes
    astAnalysis.extraNodes.forEach((node, _index) => {
      penalties.push({
        type: 'extra_token',
        severity: node.severity,
        position: node.position,
        description: `Extra ${node.actualType} at position ${node.position}`,
        penaltyPoints: this.calculatePenaltyPoints(
          'extra_token',
          node.severity
        ),
      });
    });

    // Add penalties for mismatched nodes
    astAnalysis.mismatchedNodes.forEach((node, _index) => {
      penalties.push({
        type: 'wrong_token_type',
        severity: node.severity,
        position: node.position,
        description: `Expected ${node.expectedType}, got ${node.actualType} at position ${node.position}`,
        penaltyPoints: this.calculatePenaltyPoints(
          'wrong_token_type',
          node.severity
        ),
      });
    });

    // Check for delimiter issues
    const delimiterPenalties = this.checkDelimiterBalance(
      expectedCode,
      actualCode
    );
    penalties.push(...delimiterPenalties);

    // Check for indentation issues
    const indentationPenalties = this.checkIndentationConformity(
      expectedCode,
      actualCode
    );
    penalties.push(...indentationPenalties);

    return penalties;
  }

  /**
   * Calculate overall conformity score
   */
  private calculateOverallConformity(
    astSimilarity: number,
    tokenAccuracy: number,
    syntaxScore: number,
    penalties: StructuralPenalty[]
  ): number {
    // Base score from structural components
    const baseScore =
      astSimilarity * 0.4 + tokenAccuracy * 0.4 + syntaxScore * 0.2;

    // Calculate penalty deduction
    const totalPenaltyPoints = penalties.reduce(
      (sum, penalty) => sum + penalty.penaltyPoints,
      0
    );
    const penaltyDeduction = Math.min(0.5, totalPenaltyPoints / 100); // Cap at 50% deduction

    return Math.max(0, baseScore - penaltyDeduction);
  }

  /**
   * Fallback analysis for when AST parsing fails
   */
  private fallbackAnalysis(
    expectedCode: string,
    actualCode: string
  ): StructuralConformityScore {
    // Simple string-based analysis
    const similarity = this.calculateStringSimilarity(expectedCode, actualCode);
    const penalties = this.checkDelimiterBalance(expectedCode, actualCode);

    return {
      astSimilarity: similarity,
      tokenSequenceAccuracy: similarity,
      syntaxValidationScore: similarity,
      structuralPenalties: penalties,
      overallConformity: Math.max(0, similarity - penalties.length * 0.1),
    };
  }

  // Helper methods

  private flattenAST(ast: any): any[] {
    if (!ast || !ast.children) return [ast].filter(Boolean);

    const nodes = [ast];
    for (const child of ast.children) {
      nodes.push(...this.flattenAST(child));
    }
    return nodes;
  }

  private countNodeTypeMatches(
    expectedNodes: any[],
    actualNodes: any[]
  ): number {
    let matches = 0;
    const minLength = Math.min(expectedNodes.length, actualNodes.length);

    for (let i = 0; i < minLength; i++) {
      if (expectedNodes[i]?.type === actualNodes[i]?.type) {
        matches++;
      }
    }

    return matches;
  }

  private calculateASTDepth(ast: any): number {
    if (!ast || !ast.children || ast.children.length === 0) return 1;

    let maxChildDepth = 0;
    for (const child of ast.children) {
      maxChildDepth = Math.max(maxChildDepth, this.calculateASTDepth(child));
    }

    return 1 + maxChildDepth;
  }

  private calculateBranchingFactor(ast: any): number {
    if (!ast || !ast.children) return 0;

    let totalBranches = ast.children.length;
    let nodeCount = 1;

    for (const child of ast.children) {
      const childStats = this.calculateBranchingFactor(child);
      totalBranches += childStats;
      nodeCount++;
    }

    return nodeCount > 0 ? totalBranches / nodeCount : 0;
  }

  private identifyNodeDifferences(
    expectedNodes: any[],
    actualNodes: any[]
  ): {
    missingNodes: ASTNodeDifference[];
    extraNodes: ASTNodeDifference[];
    mismatchedNodes: ASTNodeDifference[];
  } {
    const missingNodes: ASTNodeDifference[] = [];
    const extraNodes: ASTNodeDifference[] = [];
    const mismatchedNodes: ASTNodeDifference[] = [];

    const maxLength = Math.max(expectedNodes.length, actualNodes.length);

    for (let i = 0; i < maxLength; i++) {
      const expected = expectedNodes[i];
      const actual = actualNodes[i];

      if (expected && !actual) {
        missingNodes.push({
          expectedType: expected.type || 'unknown',
          position: i,
          severity: 'medium',
          impact: 'Missing required syntax element',
        });
      } else if (!expected && actual) {
        extraNodes.push({
          expectedType: 'none',
          actualType: actual.type || 'unknown',
          position: i,
          severity: 'low',
          impact: 'Extra syntax element',
        });
      } else if (expected && actual && expected.type !== actual.type) {
        mismatchedNodes.push({
          expectedType: expected.type || 'unknown',
          actualType: actual.type || 'unknown',
          position: i,
          severity: 'high',
          impact: 'Incorrect syntax element type',
        });
      }
    }

    return { missingNodes, extraNodes, mismatchedNodes };
  }

  private tokensMatch(expected: any, actual: any): boolean {
    if (!expected || !actual) return false;
    return expected.type === actual.type && expected.value === actual.value;
  }

  private detectSyntaxIssues(code: string, language: Language): string[] {
    const issues: string[] = [];

    // Check delimiter balance
    const delimiters = this.getLanguageDelimiters(language);
    for (const [open, close] of delimiters) {
      if (!this.isDelimiterBalanced(code, open, close)) {
        issues.push(`Unbalanced ${open}${close} delimiters`);
      }
    }

    // Check for common syntax patterns
    const syntaxPatterns = this.getLanguageSyntaxPatterns(language);
    for (const [pattern, description] of syntaxPatterns) {
      if (pattern.test(code)) {
        issues.push(description);
      }
    }

    return issues;
  }

  private calculatePenaltyPoints(
    type: StructuralPenaltyType,
    severity: 'low' | 'medium' | 'high'
  ): number {
    const basePoints = {
      missing_delimiter: 15,
      extra_delimiter: 10,
      mismatched_delimiter: 20,
      incorrect_indentation: 5,
      missing_token: 12,
      extra_token: 8,
      wrong_token_type: 18,
      structural_mismatch: 25,
    };

    const severityMultiplier = {
      low: 0.5,
      medium: 1.0,
      high: 1.5,
    };

    return (basePoints[type] || 10) * severityMultiplier[severity];
  }

  private checkDelimiterBalance(
    expectedCode: string,
    actualCode: string
  ): StructuralPenalty[] {
    const penalties: StructuralPenalty[] = [];
    const delimiters: [string, string][] = [
      ['(', ')'],
      ['{', '}'],
      ['[', ']'],
      ['"', '"'],
      ["'", "'"],
    ];

    for (const [open, close] of delimiters) {
      const expectedBalance = this.countDelimiters(expectedCode, open, close);
      const actualBalance = this.countDelimiters(actualCode, open, close);

      if (expectedBalance !== actualBalance) {
        const diff = Math.abs(expectedBalance - actualBalance);
        penalties.push({
          type:
            expectedBalance > actualBalance
              ? 'missing_delimiter'
              : 'extra_delimiter',
          severity: diff > 2 ? 'high' : diff > 1 ? 'medium' : 'low',
          position: -1, // Would need more sophisticated position tracking
          description: `Delimiter imbalance: ${open}${close}`,
          penaltyPoints: this.calculatePenaltyPoints(
            expectedBalance > actualBalance
              ? 'missing_delimiter'
              : 'extra_delimiter',
            diff > 2 ? 'high' : diff > 1 ? 'medium' : 'low'
          ),
        });
      }
    }

    return penalties;
  }

  private checkIndentationConformity(
    expectedCode: string,
    actualCode: string
  ): StructuralPenalty[] {
    const penalties: StructuralPenalty[] = [];
    const expectedLines = expectedCode.split('\n');
    const actualLines = actualCode.split('\n');

    const minLines = Math.min(expectedLines.length, actualLines.length);

    for (let i = 0; i < minLines; i++) {
      const expectedIndent = this.getIndentationLevel(expectedLines[i]);
      const actualIndent = this.getIndentationLevel(actualLines[i]);

      if (expectedIndent !== actualIndent) {
        penalties.push({
          type: 'incorrect_indentation',
          severity:
            Math.abs(expectedIndent - actualIndent) > 4 ? 'high' : 'medium',
          position: i,
          description: `Incorrect indentation on line ${i + 1}`,
          penaltyPoints: this.calculatePenaltyPoints(
            'incorrect_indentation',
            Math.abs(expectedIndent - actualIndent) > 4 ? 'high' : 'medium'
          ),
        });
      }
    }

    return penalties;
  }

  private calculateStringSimilarity(str1: string, str2: string): number {
    const longer = str1.length > str2.length ? str1 : str2;
    const shorter = str1.length > str2.length ? str2 : str1;

    if (longer.length === 0) return 1.0;

    const editDistance = this.levenshteinDistance(longer, shorter);
    return (longer.length - editDistance) / longer.length;
  }

  private levenshteinDistance(str1: string, str2: string): number {
    const matrix = Array(str2.length + 1)
      .fill(null)
      .map(() => Array(str1.length + 1).fill(null));

    for (let i = 0; i <= str1.length; i++) matrix[0][i] = i;
    for (let j = 0; j <= str2.length; j++) matrix[j][0] = j;

    for (let j = 1; j <= str2.length; j++) {
      for (let i = 1; i <= str1.length; i++) {
        const indicator = str1[i - 1] === str2[j - 1] ? 0 : 1;
        matrix[j][i] = Math.min(
          matrix[j][i - 1] + 1,
          matrix[j - 1][i] + 1,
          matrix[j - 1][i - 1] + indicator
        );
      }
    }

    return matrix[str2.length][str1.length];
  }

  private getLanguageDelimiters(language: Language): [string, string][] {
    const common: [string, string][] = [
      ['(', ')'],
      ['{', '}'],
      ['[', ']'],
    ];

    switch (language) {
      case 'javascript':
      case 'cpp':
      case 'rust':
        return [...common, ['"', '"'], ["'", "'"], ['`', '`']];
      case 'python':
        return [
          ...common,
          ['"', '"'],
          ["'", "'"],
          ['"""', '"""'],
          ["'''", "'''"],
        ];
      case 'yaml':
        return [
          ['"', '"'],
          ["'", "'"],
        ];
      default:
        return common;
    }
  }

  private getLanguageSyntaxPatterns(language: Language): [RegExp, string][] {
    switch (language) {
      case 'javascript':
        return [
          [/\bfunction\s+\w+\s*\([^)]*\)\s*(?!{)/, 'Missing function body'],
          [/\bif\s*\([^)]*\)\s*(?!{)(?![^;]*;)/, 'Missing if statement body'],
        ];
      case 'python':
        return [
          [/:\s*$(?!\n\s+)/, 'Missing indented block after colon'],
          [/^\s*def\s+\w+\([^)]*\):\s*$(?!\n\s+)/, 'Missing function body'],
        ];
      case 'cpp':
        return [
          [/\bclass\s+\w+\s*(?!{)/, 'Missing class body'],
          [/;\s*$(?=\s*})/, 'Unnecessary semicolon before closing brace'],
        ];
      case 'rust':
        return [
          [/\bfn\s+\w+\([^)]*\)\s*(?!{)/, 'Missing function body'],
          [/\bmatch\s+[^{]*(?!{)/, 'Missing match body'],
        ];
      default:
        return [];
    }
  }

  private isDelimiterBalanced(
    code: string,
    open: string,
    close: string
  ): boolean {
    return this.countDelimiters(code, open, close) === 0;
  }

  private countDelimiters(code: string, open: string, close: string): number {
    let count = 0;
    let inString = false;
    let stringChar = '';

    for (let i = 0; i < code.length; i++) {
      const char = code[i];
      const nextChars = code.substr(i, Math.max(open.length, close.length));

      // Handle string literals
      if ((char === '"' || char === "'") && !inString) {
        inString = true;
        stringChar = char;
        continue;
      } else if (char === stringChar && inString) {
        inString = false;
        stringChar = '';
        continue;
      }

      if (!inString) {
        if (nextChars.startsWith(open)) {
          count++;
          i += open.length - 1;
        } else if (nextChars.startsWith(close)) {
          count--;
          i += close.length - 1;
        }
      }
    }

    return count;
  }

  private getIndentationLevel(line: string): number {
    const match = line.match(/^(\s*)/);
    return match ? match[1].length : 0;
  }
}

export { StructuralAnalyzer };
export const structuralAnalyzer = new StructuralAnalyzer();
