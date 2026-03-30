// Content management related types

export interface LanguageEntity {
  id: string;
  name: string;
  version: number;
  parserId: string;
  grammarConfig: Record<string, any>;
  whitespaceRules: Record<string, any>;
  createdBy: string;
  createdAt: Date;
}

export interface Lesson {
  id: string;
  languageId: string;
  title: string;
  description: string;
  difficulty: number;
  objectives: string[];
  prerequisites: string[];
  estimatedMinutes: number;
  tokensCovered: string[];
  snippetIds: string[];
  version: number;
  createdAt: Date;
}

export interface Snippet {
  id: string;
  languageId: string;
  title: string;
  sourceCode: string;
  tags: string[];
  difficulty: number;
  estimatedTime: number;
  checksum: string;
  accessibilityTags: Record<string, any>;
  createdAt: Date;
}

export interface Playlist {
  id: string;
  name: string;
  description: string;
  languageId: string;
  snippetIds: string[];
  tags: string[];
  difficulty: number;
  estimatedTime: number;
  isPublic: boolean;
  createdBy: string;
  createdAt: Date;
}

export interface ContentVersion {
  id: string;
  contentType: 'lesson' | 'snippet' | 'playlist';
  contentId: string;
  version: number;
  changes: string[];
  checksum: string;
  createdBy: string;
  createdAt: Date;
}

export interface ContentValidation {
  id: string;
  contentType: 'lesson' | 'snippet' | 'playlist';
  contentId: string;
  isValid: boolean;
  errors: ValidationError[];
  warnings: ValidationWarning[];
  checkedAt: Date;
}

export interface ValidationError {
  type: string;
  message: string;
  line?: number;
  column?: number;
  severity: 'error' | 'warning';
}

export interface ValidationWarning {
  type: string;
  message: string;
  suggestion: string;
  line?: number;
  column?: number;
}
