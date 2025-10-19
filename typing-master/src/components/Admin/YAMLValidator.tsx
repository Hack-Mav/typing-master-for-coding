import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../services/AuthContext';

interface ValidationResult {
  valid: boolean;
  validation_errors: string[];
  validation_id: string;
}

interface SnippetValidation {
  language_id: string;
  title: string;
  source_code: string;
  tags: string[];
  difficulty: number;
}

const YAMLValidator: React.FC = () => {
  const [yamlContent, setYamlContent] = useState('');
  const [validationResult, setValidationResult] = useState<ValidationResult | null>(null);
  const [validating, setValidating] = useState(false);
  const [snippetData, setSnippetData] = useState<SnippetValidation>({
    language_id: 'yaml',
    title: '',
    source_code: '',
    tags: [],
    difficulty: 1,
  });
  const { token, user } = useAuth();
  const navigate = useNavigate();

  // Check if user is admin
  React.useEffect(() => {
    if (user && user.role !== 'admin' && user.role !== 'moderator') {
      navigate('/');
      return;
    }
  }, [user, navigate]);

  const handleValidateYAML = async () => {
    if (!yamlContent.trim()) {
      setValidationResult({
        valid: false,
        validation_errors: ['YAML content cannot be empty'],
        validation_id: '',
      });
      return;
    }

    setValidating(true);
    try {
      const response = await fetch('/api/v1/admin/content/validate', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          content_type: 'snippet',
          content_id: 'temp_validation',
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      setValidationResult(result);
    } catch (err) {
      setValidationResult({
        valid: false,
        validation_errors: [err instanceof Error ? err.message : 'Validation failed'],
        validation_id: '',
      });
    } finally {
      setValidating(false);
    }
  };

  const handleCreateSnippet = async () => {
    if (!snippetData.title || !snippetData.source_code) {
      alert('Please provide title and source code');
      return;
    }

    try {
      const response = await fetch('/api/v1/admin/snippets', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(snippetData),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      alert('Snippet created successfully!');
      navigate('/admin/snippets');
    } catch (err) {
      alert(`Failed to create snippet: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  };

  const formatYAML = () => {
    try {
      // Basic YAML formatting (in production, use a proper YAML formatter)
      setYamlContent(yamlContent.trim());
    } catch (err) {
      alert('Failed to format YAML');
    }
  };

  const loadSampleYAML = () => {
    const sample = `name: John Doe
age: 30
city: New York

skills:
  - Python
  - JavaScript
  - TypeScript

experience:
  software_engineer:
    company: Tech Corp
    years: 5
  senior_developer:
    company: Dev Inc
    years: 3

preferences:
  theme: dark
  notifications: true
  language: en`;
    setYamlContent(sample);
    setSnippetData(prev => ({
      ...prev,
      source_code: sample,
      title: 'Sample YAML Configuration',
    }));
  };

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
                YAML Validator & Schema Tools
              </h1>
            </div>
            <div className="flex items-center space-x-4">
              <button
                onClick={loadSampleYAML}
                className="px-3 py-2 text-sm text-gray-600 hover:text-gray-900"
              >
                Load Sample
              </button>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* YAML Editor */}
          <div className="bg-white shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                YAML Content
              </h3>

              <div className="mb-4">
                <textarea
                  value={yamlContent}
                  onChange={(e) => {
                    setYamlContent(e.target.value);
                    setSnippetData(prev => ({ ...prev, source_code: e.target.value }));
                  }}
                  className="w-full h-96 border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                  placeholder="Enter your YAML content here..."
                />
              </div>

              <div className="flex space-x-3">
                <button
                  onClick={formatYAML}
                  className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  Format YAML
                </button>
                <button
                  onClick={handleValidateYAML}
                  disabled={validating || !yamlContent.trim()}
                  className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
                >
                  {validating ? 'Validating...' : 'Validate YAML'}
                </button>
              </div>
            </div>
          </div>

          {/* Validation Results & Snippet Creation */}
          <div className="space-y-6">
            {/* Validation Results */}
            <div className="bg-white shadow rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Validation Results
                </h3>

                {validationResult ? (
                  <div className={`p-4 rounded-md ${validationResult.valid ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'}`}>
                    <div className="flex items-center mb-2">
                      <div className={`flex-shrink-0 w-5 h-5 rounded-full flex items-center justify-center ${validationResult.valid ? 'bg-green-500' : 'bg-red-500'}`}>
                        <svg className="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          {validationResult.valid ? (
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7" />
                          ) : (
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                          )}
                        </svg>
                      </div>
                      <span className={`ml-2 text-sm font-medium ${validationResult.valid ? 'text-green-800' : 'text-red-800'}`}>
                        {validationResult.valid ? 'Valid YAML' : 'Invalid YAML'}
                      </span>
                    </div>

                    {validationResult.validation_errors.length > 0 && (
                      <div className="mt-2">
                        <p className={`text-sm ${validationResult.valid ? 'text-green-700' : 'text-red-700'}`}>
                          {validationResult.valid ? 'No errors found' : 'Errors found:'}
                        </p>
                        <ul className="mt-1 text-sm text-red-700 list-disc list-inside">
                          {validationResult.validation_errors.map((error, index) => (
                            <li key={index}>{error}</li>
                          ))}
                        </ul>
                      </div>
                    )}
                  </div>
                ) : (
                  <p className="text-sm text-gray-500">
                    Click "Validate YAML" to check your content
                  </p>
                )}
              </div>
            </div>

            {/* Snippet Creation Form */}
            <div className="bg-white shadow rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Create Snippet
                </h3>

                <div className="space-y-4">
                  <div>
                    <label htmlFor="snippet_title" className="block text-sm font-medium text-gray-700">
                      Title
                    </label>
                    <input
                      type="text"
                      id="snippet_title"
                      value={snippetData.title}
                      onChange={(e) => setSnippetData(prev => ({ ...prev, title: e.target.value }))}
                      className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                      placeholder="Snippet title"
                    />
                  </div>

                  <div>
                    <label htmlFor="snippet_language" className="block text-sm font-medium text-gray-700">
                      Language
                    </label>
                    <select
                      id="snippet_language"
                      value={snippetData.language_id}
                      onChange={(e) => setSnippetData(prev => ({ ...prev, language_id: e.target.value }))}
                      className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                    >
                      <option value="yaml">YAML</option>
                      <option value="python">Python</option>
                      <option value="javascript">JavaScript</option>
                    </select>
                  </div>

                  <div>
                    <label htmlFor="snippet_difficulty" className="block text-sm font-medium text-gray-700">
                      Difficulty
                    </label>
                    <select
                      id="snippet_difficulty"
                      value={snippetData.difficulty}
                      onChange={(e) => setSnippetData(prev => ({ ...prev, difficulty: parseInt(e.target.value) }))}
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
                    <label htmlFor="snippet_tags" className="block text-sm font-medium text-gray-700">
                      Tags (comma-separated)
                    </label>
                    <input
                      type="text"
                      id="snippet_tags"
                      value={snippetData.tags.join(', ')}
                      onChange={(e) => setSnippetData(prev => ({
                        ...prev,
                        tags: e.target.value.split(',').map(s => s.trim()).filter(s => s)
                      }))}
                      className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                      placeholder="yaml, configuration, example"
                    />
                  </div>

                  <button
                    onClick={handleCreateSnippet}
                    disabled={!snippetData.title || !snippetData.source_code || !validationResult?.valid}
                    className="w-full inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 disabled:opacity-50"
                  >
                    Create Snippet
                  </button>
                </div>
              </div>
            </div>

            {/* Schema Validation Rules */}
            <div className="bg-white shadow rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Schema Validation Rules
                </h3>

                <div className="space-y-3 text-sm">
                  <div className="flex items-start space-x-3">
                    <div className="flex-shrink-0 w-5 h-5 bg-blue-100 rounded-full flex items-center justify-center mt-0.5">
                      <span className="text-xs font-medium text-blue-800">✓</span>
                    </div>
                    <div>
                      <p className="text-gray-900 font-medium">YAML Structure</p>
                      <p className="text-gray-500">Valid YAML syntax with proper indentation</p>
                    </div>
                  </div>

                  <div className="flex items-start space-x-3">
                    <div className="flex-shrink-0 w-5 h-5 bg-blue-100 rounded-full flex items-center justify-center mt-0.5">
                      <span className="text-xs font-medium text-blue-800">✓</span>
                    </div>
                    <div>
                      <p className="text-gray-900 font-medium">Required Fields</p>
                      <p className="text-gray-500">Title and source code must be provided</p>
                    </div>
                  </div>

                  <div className="flex items-start space-x-3">
                    <div className="flex-shrink-0 w-5 h-5 bg-blue-100 rounded-full flex items-center justify-center mt-0.5">
                      <span className="text-xs font-medium text-blue-800">✓</span>
                    </div>
                    <div>
                      <p className="text-gray-900 font-medium">Content Length</p>
                      <p className="text-gray-500">Between 10 and 1500 characters</p>
                    </div>
                  </div>

                  <div className="flex items-start space-x-3">
                    <div className="flex-shrink-0 w-5 h-5 bg-blue-100 rounded-full flex items-center justify-center mt-0.5">
                      <span className="text-xs font-medium text-blue-800">✓</span>
                    </div>
                    <div>
                      <p className="text-gray-900 font-medium">Accessibility Tags</p>
                      <p className="text-gray-500">Automatically generated based on content complexity</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default YAMLValidator;
