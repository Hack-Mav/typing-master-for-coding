import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../services/AuthContext';

interface ContentValidation {
  content_type: string;
  content_id: string;
  validation_status: string;
  validation_errors: string[];
  validated_at: string;
}

const ContentValidationQA: React.FC = () => {
  const [validations, setValidations] = useState<ContentValidation[]>([]);
  const [loading, setLoading] = useState(true);
  const [validating, setValidating] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const { token, user } = useAuth();
  const navigate = useNavigate();

  // Check if user is admin
  useEffect(() => {
    if (user && user.role !== 'admin' && user.role !== 'moderator') {
      navigate('/');
      return;
    }
  }, [user, navigate]);

  // Fetch validation history
  useEffect(() => {
    const fetchValidations = async () => {
      try {
        // For now, we'll simulate validation history since we don't have a validation history endpoint
        // In production, you'd fetch from /api/v1/admin/validations
        setValidations([]);
      } catch (err) {
        setError(
          err instanceof Error ? err.message : 'Failed to fetch validations'
        );
      } finally {
        setLoading(false);
      }
    };

    if (token && user) {
      fetchValidations();
    }
  }, [token, user]);

  const handleValidateContent = async (
    contentType: string,
    contentId: string
  ) => {
    setValidating(`${contentType}_${contentId}`);

    try {
      const response = await fetch('/api/v1/admin/content/validate', {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          content_type: contentType,
          content_id: contentId,
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();

      // Refresh validations list (in production)
      alert(`Validation completed: ${result.valid ? 'PASSED' : 'FAILED'}`);

      if (!result.valid) {
        alert(`Errors found:\n${result.validation_errors.join('\n')}`);
      }
    } catch (err) {
      alert(
        `Validation failed: ${err instanceof Error ? err.message : 'Unknown error'}`
      );
    } finally {
      setValidating(null);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading validation tools...</p>
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
                Content Validation & QA
              </h1>
            </div>
            <div className="flex items-center space-x-4">
              <button
                onClick={() => navigate('/admin/yaml-validator')}
                className="px-3 py-2 text-sm text-green-600 hover:text-green-900"
              >
                YAML Validator
              </button>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Content Validation */}
          <div className="bg-white shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                Content Validation Tools
              </h3>

              <div className="space-y-4">
                {/* Language Validation */}
                <div className="border border-gray-200 rounded-lg p-4">
                  <h4 className="text-md font-medium text-gray-900 mb-3">
                    🌐 Languages
                  </h4>
                  <p className="text-sm text-gray-600 mb-3">
                    Validate language definitions, parser configurations, and
                    grammar rules.
                  </p>
                  <button
                    onClick={() => handleValidateContent('language', 'all')}
                    disabled={validating === 'language_all'}
                    className="w-full px-3 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-50"
                  >
                    {validating === 'language_all'
                      ? 'Validating...'
                      : 'Validate All Languages'}
                  </button>
                </div>

                {/* Lesson Validation */}
                <div className="border border-gray-200 rounded-lg p-4">
                  <h4 className="text-md font-medium text-gray-900 mb-3">
                    📚 Lessons
                  </h4>
                  <p className="text-sm text-gray-600 mb-3">
                    Check lesson structure, token coverage, prerequisites, and
                    learning objectives.
                  </p>
                  <button
                    onClick={() => handleValidateContent('lesson', 'all')}
                    disabled={validating === 'lesson_all'}
                    className="w-full px-3 py-2 bg-green-600 text-white text-sm rounded hover:bg-green-700 disabled:opacity-50"
                  >
                    {validating === 'lesson_all'
                      ? 'Validating...'
                      : 'Validate All Lessons'}
                  </button>
                </div>

                {/* Snippet Validation */}
                <div className="border border-gray-200 rounded-lg p-4">
                  <h4 className="text-md font-medium text-gray-900 mb-3">
                    📄 Snippets
                  </h4>
                  <p className="text-sm text-gray-600 mb-3">
                    Validate code snippets, check YAML syntax, verify
                    accessibility tags, and ensure proper formatting.
                  </p>
                  <button
                    onClick={() => handleValidateContent('snippet', 'all')}
                    disabled={validating === 'snippet_all'}
                    className="w-full px-3 py-2 bg-purple-600 text-white text-sm rounded hover:bg-purple-700 disabled:opacity-50"
                  >
                    {validating === 'snippet_all'
                      ? 'Validating...'
                      : 'Validate All Snippets'}
                  </button>
                </div>
              </div>
            </div>
          </div>

          {/* Validation Rules & Guidelines */}
          <div className="bg-white shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                Validation Rules & Guidelines
              </h3>

              <div className="space-y-6">
                {/* Language Validation Rules */}
                <div>
                  <h4 className="text-md font-medium text-gray-900 mb-3">
                    Language Validation Rules
                  </h4>
                  <ul className="text-sm text-gray-600 space-y-2">
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-blue-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>
                        Parser ID must be valid and correspond to an existing
                        Tree-sitter parser
                      </span>
                    </li>
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-blue-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>Grammar configuration must be valid JSON</span>
                    </li>
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-blue-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>
                        Whitespace rules must define valid token boundaries
                      </span>
                    </li>
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-blue-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>Language name must be unique and descriptive</span>
                    </li>
                  </ul>
                </div>

                {/* Lesson Validation Rules */}
                <div>
                  <h4 className="text-md font-medium text-gray-900 mb-3">
                    Lesson Validation Rules
                  </h4>
                  <ul className="text-sm text-gray-600 space-y-2">
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-green-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>Title must be descriptive and not empty</span>
                    </li>
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-green-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>
                        Language ID must reference an existing language
                      </span>
                    </li>
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-green-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>
                        Token coverage must include at least basic syntax tokens
                      </span>
                    </li>
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-green-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>Estimated time must be between 5-300 minutes</span>
                    </li>
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-green-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>
                        Prerequisites must reference existing lessons if
                        specified
                      </span>
                    </li>
                  </ul>
                </div>

                {/* Snippet Validation Rules */}
                <div>
                  <h4 className="text-md font-medium text-gray-900 mb-3">
                    Snippet Validation Rules
                  </h4>
                  <ul className="text-sm text-gray-600 space-y-2">
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-purple-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>Title must be descriptive and unique</span>
                    </li>
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-purple-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>
                        Source code must be valid for the specified language
                      </span>
                    </li>
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-purple-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>YAML snippets must have valid YAML syntax</span>
                    </li>
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-purple-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>
                        Content length must be appropriate (10-1500 characters)
                      </span>
                    </li>
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-purple-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>
                        Accessibility tags must be automatically generated
                      </span>
                    </li>
                    <li className="flex items-start space-x-2">
                      <span className="w-2 h-2 bg-purple-500 rounded-full mt-2 flex-shrink-0"></span>
                      <span>Checksum must match the source code content</span>
                    </li>
                  </ul>
                </div>

                {/* Quality Assurance Guidelines */}
                <div>
                  <h4 className="text-md font-medium text-gray-900 mb-3">
                    Quality Assurance Guidelines
                  </h4>
                  <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
                    <ul className="text-sm text-yellow-800 space-y-1">
                      <li>• All content must be reviewed before publication</li>
                      <li>• Code snippets should demonstrate best practices</li>
                      <li>• Lessons should have clear learning objectives</li>
                      <li>• Accessibility considerations must be included</li>
                      <li>
                        • Version control must be maintained for all changes
                      </li>
                      <li>• Regular validation checks should be performed</li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Validation History */}
        <div className="mt-8 bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
              Validation History
            </h3>

            {validations.length === 0 ? (
              <p className="text-gray-500 text-center py-8">
                No validation history available. Run some validations to see
                results here.
              </p>
            ) : (
              <div className="space-y-3">
                {validations.map((validation, index) => (
                  <div
                    key={index}
                    className="border border-gray-200 rounded-lg p-4"
                  >
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-sm font-medium text-gray-900">
                          {validation.content_type} • {validation.content_id}
                        </p>
                        <p className="text-xs text-gray-500">
                          {formatDate(validation.validated_at)}
                        </p>
                      </div>
                      <span
                        className={`px-2 py-1 text-xs font-medium rounded-full ${
                          validation.validation_status === 'valid'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-red-100 text-red-800'
                        }`}
                      >
                        {validation.validation_status}
                      </span>
                    </div>

                    {validation.validation_errors.length > 0 && (
                      <div className="mt-3">
                        <p className="text-sm text-red-600 font-medium">
                          Errors:
                        </p>
                        <ul className="text-sm text-red-600 list-disc list-inside">
                          {validation.validation_errors.map(
                            (error, errorIndex) => (
                              <li key={errorIndex}>{error}</li>
                            )
                          )}
                        </ul>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ContentValidationQA;
