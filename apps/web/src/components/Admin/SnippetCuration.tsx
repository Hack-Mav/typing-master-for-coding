import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../services/AuthContext';

interface Snippet {
  id: string;
  language_id: string;
  title: string;
  source_code: string;
  tags: string[];
  difficulty: number;
  estimated_time: number;
  checksum: string;
  accessibility_tags: any;
  created_at: string;
}

interface Language {
  id: string;
  name: string;
  version: string;
  parser_id: string;
  grammar_config: any;
  whitespace_rules: any;
  created_at: string;
}

const SnippetCuration: React.FC = () => {
  const [snippets, setSnippets] = useState<Snippet[]>([]);
  const [languages, setLanguages] = useState<Language[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedSnippets, setSelectedSnippets] = useState<Set<string>>(
    new Set()
  );
  const [filters, setFilters] = useState({
    language: '',
    difficulty: '',
    tags: '',
    search: '',
  });
  const { token, user } = useAuth();
  const navigate = useNavigate();

  // Check if user is admin
  useEffect(() => {
    if (user && user.role !== 'admin' && user.role !== 'moderator') {
      navigate('/');
      return;
    }
  }, [user, navigate]);

  // Fetch snippets and languages
  useEffect(() => {
    const fetchData = async () => {
      try {
        const [snippetsResponse, languagesResponse] = await Promise.all([
          fetch('/api/v1/admin/snippets', {
            headers: {
              Authorization: `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
          }),
          fetch('/api/v1/languages', {
            headers: {
              Authorization: `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
          }),
        ]);

        if (!snippetsResponse.ok || !languagesResponse.ok) {
          throw new Error('Failed to fetch data');
        }

        const snippetsData = await snippetsResponse.json();
        const languagesData = await languagesResponse.json();

        setSnippets(snippetsData);
        setLanguages(languagesData);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch data');
      } finally {
        setLoading(false);
      }
    };

    if (token && user) {
      fetchData();
    }
  }, [token, user]);

  // Filter snippets based on current filters
  const filteredSnippets = snippets.filter(snippet => {
    const matchesLanguage =
      !filters.language || snippet.language_id === filters.language;
    const matchesDifficulty =
      !filters.difficulty ||
      snippet.difficulty.toString() === filters.difficulty;
    const matchesSearch =
      !filters.search ||
      snippet.title.toLowerCase().includes(filters.search.toLowerCase()) ||
      snippet.source_code
        .toLowerCase()
        .includes(filters.search.toLowerCase()) ||
      snippet.tags.some(tag =>
        tag.toLowerCase().includes(filters.search.toLowerCase())
      );
    const matchesTags =
      !filters.tags ||
      snippet.tags.some(tag =>
        tag.toLowerCase().includes(filters.tags.toLowerCase())
      );

    return matchesLanguage && matchesDifficulty && matchesSearch && matchesTags;
  });

  const handleFilterChange = (filterType: string, value: string) => {
    setFilters(prev => ({
      ...prev,
      [filterType]: value,
    }));
  };

  const handleSnippetSelect = (snippetId: string) => {
    const newSelection = new Set(selectedSnippets);
    if (newSelection.has(snippetId)) {
      newSelection.delete(snippetId);
    } else {
      newSelection.add(snippetId);
    }
    setSelectedSnippets(newSelection);
  };

  const handleBulkTag = async (tag: string) => {
    if (selectedSnippets.size === 0) {
      alert('Please select snippets to tag');
      return;
    }

    try {
      const updatePromises = Array.from(selectedSnippets).map(
        async snippetId => {
          const snippet = snippets.find(s => s.id === snippetId);
          if (!snippet) return;

          const updatedTags = Array.from(new Set([...snippet.tags, tag]));

          const response = await fetch(`/api/v1/admin/snippets/${snippetId}`, {
            method: 'PUT',
            headers: {
              Authorization: `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              ...snippet,
              tags: updatedTags,
            }),
          });

          if (!response.ok) {
            throw new Error(`Failed to update snippet ${snippetId}`);
          }
        }
      );

      await Promise.all(updatePromises);

      // Refresh snippets
      const response = await fetch('/api/v1/admin/snippets', {
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const updatedSnippets = await response.json();
        setSnippets(updatedSnippets);
      }

      setSelectedSnippets(new Set());
      alert(`Added tag "${tag}" to ${selectedSnippets.size} snippets`);
    } catch (err) {
      alert(
        `Failed to add tag: ${err instanceof Error ? err.message : 'Unknown error'}`
      );
    }
  };

  const handleBulkDelete = async () => {
    if (selectedSnippets.size === 0) {
      alert('Please select snippets to delete');
      return;
    }

    if (
      !window.confirm(
        `Are you sure you want to delete ${selectedSnippets.size} snippets? This action cannot be undone.`
      )
    ) {
      return;
    }

    try {
      const deletePromises = Array.from(selectedSnippets).map(
        async snippetId => {
          const response = await fetch(`/api/v1/admin/snippets/${snippetId}`, {
            method: 'DELETE',
            headers: {
              Authorization: `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
          });

          if (!response.ok) {
            throw new Error(`Failed to delete snippet ${snippetId}`);
          }
        }
      );

      await Promise.all(deletePromises);

      // Refresh snippets
      const response = await fetch('/api/v1/admin/snippets', {
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const updatedSnippets = await response.json();
        setSnippets(updatedSnippets);
      }

      setSelectedSnippets(new Set());
      alert(`Deleted ${selectedSnippets.size} snippets`);
    } catch (err) {
      alert(
        `Failed to delete snippets: ${err instanceof Error ? err.message : 'Unknown error'}`
      );
    }
  };

  // Get all unique tags for filter dropdown
  const allTags = Array.from(
    new Set(snippets.flatMap(snippet => snippet.tags))
  ).sort();

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading snippet curation...</p>
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
                Snippet Curation
              </h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-600">
                {filteredSnippets.length} of {snippets.length} snippets
              </span>
              <button
                onClick={() => navigate('/admin/snippets/new')}
                className="px-3 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700"
              >
                Create Snippet
              </button>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        {/* Filters */}
        <div className="bg-white shadow rounded-lg mb-6">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
              Filters & Search
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              <div>
                <label
                  htmlFor="search"
                  className="block text-sm font-medium text-gray-700"
                >
                  Search
                </label>
                <input
                  type="text"
                  id="search"
                  value={filters.search}
                  onChange={e => handleFilterChange('search', e.target.value)}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                  placeholder="Search snippets..."
                />
              </div>

              <div>
                <label
                  htmlFor="language"
                  className="block text-sm font-medium text-gray-700"
                >
                  Language
                </label>
                <select
                  id="language"
                  value={filters.language}
                  onChange={e => handleFilterChange('language', e.target.value)}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">All Languages</option>
                  {languages.map(lang => (
                    <option key={lang.id} value={lang.id}>
                      {lang.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label
                  htmlFor="difficulty"
                  className="block text-sm font-medium text-gray-700"
                >
                  Difficulty
                </label>
                <select
                  id="difficulty"
                  value={filters.difficulty}
                  onChange={e =>
                    handleFilterChange('difficulty', e.target.value)
                  }
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">All Difficulties</option>
                  <option value="1">1 - Beginner</option>
                  <option value="2">2 - Easy</option>
                  <option value="3">3 - Intermediate</option>
                  <option value="4">4 - Advanced</option>
                  <option value="5">5 - Expert</option>
                </select>
              </div>

              <div>
                <label
                  htmlFor="tags"
                  className="block text-sm font-medium text-gray-700"
                >
                  Tags
                </label>
                <select
                  id="tags"
                  value={filters.tags}
                  onChange={e => handleFilterChange('tags', e.target.value)}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">All Tags</option>
                  {allTags.map(tag => (
                    <option key={tag} value={tag}>
                      {tag}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>
        </div>

        {/* Bulk Actions */}
        {selectedSnippets.size > 0 && (
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
            <div className="flex items-center justify-between">
              <span className="text-sm text-blue-800">
                {selectedSnippets.size} snippet
                {selectedSnippets.size !== 1 ? 's' : ''} selected
              </span>
              <div className="flex space-x-3">
                <select
                  onChange={e => {
                    if (e.target.value) {
                      handleBulkTag(e.target.value);
                      e.target.value = '';
                    }
                  }}
                  className="text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                  defaultValue=""
                >
                  <option value="">Add Tag...</option>
                  {allTags.map(tag => (
                    <option key={tag} value={tag}>
                      {tag}
                    </option>
                  ))}
                </select>
                <button
                  onClick={handleBulkDelete}
                  className="px-3 py-2 bg-red-600 text-white text-sm rounded hover:bg-red-700"
                >
                  Delete Selected
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Snippets Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredSnippets.map(snippet => (
            <div
              key={snippet.id}
              className={`bg-white shadow rounded-lg p-6 border-2 transition-colors ${
                selectedSnippets.has(snippet.id)
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-gray-200 hover:border-gray-300'
              }`}
            >
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    checked={selectedSnippets.has(snippet.id)}
                    onChange={() => handleSnippetSelect(snippet.id)}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                  <div>
                    <h3 className="text-lg font-medium text-gray-900">
                      {snippet.title}
                    </h3>
                    <p className="text-sm text-gray-500">
                      {languages.find(l => l.id === snippet.language_id)
                        ?.name || snippet.language_id}
                    </p>
                  </div>
                </div>
                <span
                  className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                    snippet.difficulty <= 2
                      ? 'bg-green-100 text-green-800'
                      : snippet.difficulty <= 3
                        ? 'bg-yellow-100 text-yellow-800'
                        : 'bg-red-100 text-red-800'
                  }`}
                >
                  Level {snippet.difficulty}
                </span>
              </div>

              <div className="mb-4">
                <div className="bg-gray-50 rounded-md p-3 max-h-32 overflow-y-auto">
                  <pre className="text-xs text-gray-800 font-mono whitespace-pre-wrap">
                    {snippet.source_code.length > 200
                      ? snippet.source_code.substring(0, 200) + '...'
                      : snippet.source_code}
                  </pre>
                </div>
              </div>

              <div className="mb-4">
                <div className="flex flex-wrap gap-1">
                  {snippet.tags.map((tag, index) => (
                    <span
                      key={index}
                      className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800"
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              </div>

              <div className="flex items-center justify-between text-sm text-gray-500">
                <span>{snippet.estimated_time}min</span>
                <span>{new Date(snippet.created_at).toLocaleDateString()}</span>
              </div>

              <div className="mt-4 flex space-x-2">
                <button
                  onClick={() => navigate(`/admin/snippets/${snippet.id}/edit`)}
                  className="flex-1 px-3 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700"
                >
                  Edit
                </button>
                <button
                  onClick={() => {
                    if (
                      window.confirm(
                        'Are you sure you want to delete this snippet?'
                      )
                    ) {
                      // Handle delete
                    }
                  }}
                  className="px-3 py-2 bg-red-600 text-white text-sm rounded hover:bg-red-700"
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>

        {filteredSnippets.length === 0 && (
          <div className="text-center py-12">
            <svg
              className="mx-auto h-12 w-12 text-gray-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-gray-900">
              No snippets found
            </h3>
            <p className="mt-1 text-sm text-gray-500">
              {snippets.length === 0
                ? 'No snippets have been created yet.'
                : 'Try adjusting your filters to see more snippets.'}
            </p>
          </div>
        )}
      </div>
    </div>
  );
};

export default SnippetCuration;
