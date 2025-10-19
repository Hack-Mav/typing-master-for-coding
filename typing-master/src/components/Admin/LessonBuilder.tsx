import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useAuth } from '../../services/AuthContext';

interface Language {
  id: string;
  name: string;
  version: string;
  parser_id: string;
  grammar_config: any;
  whitespace_rules: any;
  created_at: string;
}

interface Lesson {
  id: string;
  language_id: string;
  title: string;
  difficulty: number;
  objectives: string[];
  prerequisites: string[];
  estimated_minutes: number;
  tokens_covered: string[];
  snippet_ids: string[];
  version: number;
  created_at: string;
}

interface TokenCoverage {
  token: string;
  covered: boolean;
  difficulty: number;
  description: string;
}

const LessonBuilder: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [lesson, setLesson] = useState<Lesson | null>(null);
  const [languages, setLanguages] = useState<Language[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [tokenCoverage, setTokenCoverage] = useState<TokenCoverage[]>([]);
  const { token, user } = useAuth();
  const navigate = useNavigate();

  // Check if user is admin
  useEffect(() => {
    if (user && user.role !== 'admin' && user.role !== 'moderator') {
      navigate('/');
      return;
    }
  }, [user, navigate]);

  // Fetch languages for dropdown
  useEffect(() => {
    const fetchLanguages = async () => {
      try {
        const response = await fetch('/api/v1/languages', {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        setLanguages(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch languages');
      }
    };

    if (token && user) {
      fetchLanguages();
    }
  }, [token, user]);

  // Fetch lesson data if editing
  useEffect(() => {
    if (id && id !== 'new') {
      const fetchLesson = async () => {
        try {
          const response = await fetch(`/api/v1/lessons/${id}`, {
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
          });

          if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
          }

          const data = await response.json();
          setLesson(data);
          setTokenCoverage(generateTokenCoverage(data.language_id));
        } catch (err) {
          setError(err instanceof Error ? err.message : 'Failed to fetch lesson');
        } finally {
          setLoading(false);
        }
      };

      fetchLesson();
    } else {
      // New lesson
      setLesson({
        id: '',
        language_id: '',
        title: '',
        difficulty: 1,
        objectives: [],
        prerequisites: [],
        estimated_minutes: 30,
        tokens_covered: [],
        snippet_ids: [],
        version: 1,
        created_at: new Date().toISOString(),
      });
      setLoading(false);
    }
  }, [id, token]);

  // Generate token coverage checklist based on language
  const generateTokenCoverage = (languageId: string): TokenCoverage[] => {
    // This would typically come from the language's grammar configuration
    // For now, using a sample set of tokens for different languages
    const tokenSets: { [key: string]: TokenCoverage[] } = {
      python: [
        { token: 'function', covered: false, difficulty: 1, description: 'Function definitions' },
        { token: 'class', covered: false, difficulty: 2, description: 'Class definitions' },
        { token: 'if', covered: false, difficulty: 1, description: 'Conditional statements' },
        { token: 'for', covered: false, difficulty: 1, description: 'For loops' },
        { token: 'while', covered: false, difficulty: 1, description: 'While loops' },
        { token: 'import', covered: false, difficulty: 1, description: 'Import statements' },
        { token: 'list', covered: false, difficulty: 2, description: 'List comprehensions' },
        { token: 'dict', covered: false, difficulty: 2, description: 'Dictionary operations' },
      ],
      javascript: [
        { token: 'function', covered: false, difficulty: 1, description: 'Function declarations' },
        { token: 'const', covered: false, difficulty: 1, description: 'Constant declarations' },
        { token: 'let', covered: false, difficulty: 1, description: 'Variable declarations' },
        { token: 'if', covered: false, difficulty: 1, description: 'Conditional statements' },
        { token: 'for', covered: false, difficulty: 1, description: 'For loops' },
        { token: 'arrow', covered: false, difficulty: 2, description: 'Arrow functions' },
        { token: 'async', covered: false, difficulty: 3, description: 'Async/await' },
        { token: 'class', covered: false, difficulty: 2, description: 'Class definitions' },
      ],
      yaml: [
        { token: 'key', covered: false, difficulty: 1, description: 'YAML keys' },
        { token: 'value', covered: false, difficulty: 1, description: 'YAML values' },
        { token: 'list', covered: false, difficulty: 1, description: 'YAML lists' },
        { token: 'object', covered: false, difficulty: 2, description: 'YAML objects' },
        { token: 'anchor', covered: false, difficulty: 3, description: 'YAML anchors' },
      ],
    };

    return tokenSets[languageId] || [];
  };

  const handleTokenToggle = (tokenIndex: number) => {
    if (!lesson) return;

    const updatedTokens = [...tokenCoverage];
    updatedTokens[tokenIndex].covered = !updatedTokens[tokenIndex].covered;

    setTokenCoverage(updatedTokens);

    // Update lesson tokens_covered
    const coveredTokens = updatedTokens
      .filter(t => t.covered)
      .map(t => t.token);

    setLesson({
      ...lesson,
      tokens_covered: coveredTokens,
    });
  };

  const handleSave = async () => {
    if (!lesson || !token) return;

    setSaving(true);
    try {
      const url = lesson.id ? `/api/v1/admin/lessons/${lesson.id}` : '/api/v1/admin/lessons';
      const method = lesson.id ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(lesson),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      navigate('/admin/lessons');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save lesson');
    } finally {
      setSaving(false);
    }
  };

  const handleInputChange = (field: keyof Lesson, value: any) => {
    if (!lesson) return;

    setLesson({
      ...lesson,
      [field]: value,
    });

    // Regenerate token coverage if language changed
    if (field === 'language_id') {
      setTokenCoverage(generateTokenCoverage(value));
    }
  };

  const addObjective = () => {
    if (!lesson) return;
    setLesson({
      ...lesson,
      objectives: [...lesson.objectives, ''],
    });
  };

  const updateObjective = (index: number, value: string) => {
    if (!lesson) return;
    const updatedObjectives = [...lesson.objectives];
    updatedObjectives[index] = value;
    setLesson({
      ...lesson,
      objectives: updatedObjectives,
    });
  };

  const removeObjective = (index: number) => {
    if (!lesson) return;
    const updatedObjectives = lesson.objectives.filter((_, i) => i !== index);
    setLesson({
      ...lesson,
      objectives: updatedObjectives,
    });
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading lesson builder...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 mb-4">Error: {error}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  if (!lesson) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-gray-600">Lesson not found</p>
      </div>
    );
  }

  const coveredTokens = tokenCoverage.filter(t => t.covered).length;
  const totalTokens = tokenCoverage.length;
  const coveragePercentage = totalTokens > 0 ? (coveredTokens / totalTokens) * 100 : 0;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Admin Navigation */}
      <nav className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <button
                onClick={() => navigate('/admin/dashboard')}
                className="text-blue-600 hover:text-blue-900 mr-4"
              >
                ← Back to Dashboard
              </button>
              <h1 className="text-xl font-semibold text-gray-900">
                {lesson.id ? 'Edit Lesson' : 'Create New Lesson'}
              </h1>
            </div>
            <div className="flex items-center space-x-4">
              <button
                onClick={handleSave}
                disabled={saving}
                className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
              >
                {saving ? 'Saving...' : 'Save Lesson'}
              </button>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Form */}
          <div className="lg:col-span-2 space-y-6">
            {/* Basic Information */}
            <div className="bg-white shadow rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Basic Information
                </h3>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label htmlFor="title" className="block text-sm font-medium text-gray-700">
                      Title
                    </label>
                    <input
                      type="text"
                      id="title"
                      value={lesson.title}
                      onChange={(e) => handleInputChange('title', e.target.value)}
                      className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                      placeholder="Lesson title"
                    />
                  </div>

                  <div>
                    <label htmlFor="language" className="block text-sm font-medium text-gray-700">
                      Language
                    </label>
                    <select
                      id="language"
                      value={lesson.language_id}
                      onChange={(e) => handleInputChange('language_id', e.target.value)}
                      className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                    >
                      <option value="">Select a language</option>
                      {languages.map((lang) => (
                        <option key={lang.id} value={lang.id}>
                          {lang.name}
                        </option>
                      ))}
                    </select>
                  </div>

                  <div>
                    <label htmlFor="difficulty" className="block text-sm font-medium text-gray-700">
                      Difficulty (1-5)
                    </label>
                    <select
                      id="difficulty"
                      value={lesson.difficulty}
                      onChange={(e) => handleInputChange('difficulty', parseInt(e.target.value))}
                      className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                    >
                      <option value={1}>1 - Beginner</option>
                      <option value={2}>2 - Easy</option>
                      <option value={3}>3 - Intermediate</option>
                      <option value={4}>4 - Advanced</option>
                      <option value={5}>5 - Expert</option>
                    </select>
                  </div>

                  <div>
                    <label htmlFor="estimated_minutes" className="block text-sm font-medium text-gray-700">
                      Estimated Time (minutes)
                    </label>
                    <input
                      type="number"
                      id="estimated_minutes"
                      value={lesson.estimated_minutes}
                      onChange={(e) => handleInputChange('estimated_minutes', parseInt(e.target.value))}
                      className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                      min="5"
                      max="300"
                    />
                  </div>
                </div>
              </div>
            </div>

            {/* Objectives */}
            <div className="bg-white shadow rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <div className="flex justify-between items-center mb-4">
                  <h3 className="text-lg leading-6 font-medium text-gray-900">
                    Learning Objectives
                  </h3>
                  <button
                    type="button"
                    onClick={addObjective}
                    className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                  >
                    Add Objective
                  </button>
                </div>

                <div className="space-y-3">
                  {lesson.objectives.map((objective, index) => (
                    <div key={index} className="flex items-center space-x-3">
                      <input
                        type="text"
                        value={objective}
                        onChange={(e) => updateObjective(index, e.target.value)}
                        className="flex-1 border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                        placeholder={`Objective ${index + 1}`}
                      />
                      <button
                        type="button"
                        onClick={() => removeObjective(index)}
                        className="text-red-600 hover:text-red-900"
                      >
                        Remove
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            {/* Prerequisites */}
            <div className="bg-white shadow rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Prerequisites
                </h3>
                <div className="text-sm text-gray-600 mb-4">
                  Leave empty if no prerequisites are required
                </div>
                <input
                  type="text"
                  value={lesson.prerequisites.join(', ')}
                  onChange={(e) => handleInputChange('prerequisites', e.target.value.split(',').map(s => s.trim()).filter(s => s))}
                  className="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                  placeholder="Enter prerequisite lesson IDs separated by commas"
                />
              </div>
            </div>
          </div>

          {/* Sidebar - Token Coverage Checklist */}
          <div className="lg:col-span-1">
            <div className="bg-white shadow rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Token Coverage Checklist
                </h3>

                {/* Coverage Progress */}
                <div className="mb-6">
                  <div className="flex justify-between text-sm text-gray-600 mb-1">
                    <span>Progress</span>
                    <span>{coveredTokens}/{totalTokens} tokens</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                      style={{ width: `${coveragePercentage}%` }}
                    ></div>
                  </div>
                  <p className="text-xs text-gray-500 mt-1">
                    {coveragePercentage.toFixed(0)}% complete
                  </p>
                </div>

                {/* Token Checklist */}
                <div className="space-y-3 max-h-96 overflow-y-auto">
                  {tokenCoverage.map((token, index) => (
                    <div key={index} className="flex items-start space-x-3">
                      <input
                        type="checkbox"
                        checked={token.covered}
                        onChange={() => handleTokenToggle(index)}
                        className="mt-1 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                      />
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center space-x-2">
                          <span className="text-sm font-medium text-gray-900">
                            {token.token}
                          </span>
                          <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                            token.difficulty <= 2 ? 'bg-green-100 text-green-800' :
                            token.difficulty <= 3 ? 'bg-yellow-100 text-yellow-800' :
                            'bg-red-100 text-red-800'
                          }`}>
                            Level {token.difficulty}
                          </span>
                        </div>
                        <p className="text-xs text-gray-500 mt-1">
                          {token.description}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>

                {totalTokens === 0 && (
                  <p className="text-sm text-gray-500 text-center py-4">
                    Select a language to see token coverage options
                  </p>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default LessonBuilder;
