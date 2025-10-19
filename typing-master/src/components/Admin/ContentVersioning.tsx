import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../services/AuthContext';

interface ContentVersion {
  id: string;
  content_type: string;
  content_id: string;
  version: number;
  content: any;
  checksum: string;
  created_by: string;
  change_notes: string;
  is_active: boolean;
  created_at: string;
}

interface ContentItem {
  id: string;
  title: string;
  type: string;
  current_version: number;
  versions: ContentVersion[];
}

const ContentVersioning: React.FC = () => {
  const [contentItems, setContentItems] = useState<ContentItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedItem, setSelectedItem] = useState<ContentItem | null>(null);
  const [versions, setVersions] = useState<ContentVersion[]>([]);
  const [showVersions, setShowVersions] = useState(false);
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

  // Fetch content items with version info
  useEffect(() => {
    const fetchContentOverview = async () => {
      try {
        // Fetch all content types and aggregate version info
        const [languagesResponse, lessonsResponse, snippetsResponse] = await Promise.all([
          fetch('/api/v1/admin/languages', {
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
          }),
          fetch('/api/v1/admin/lessons', {
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
          }),
          fetch('/api/v1/admin/snippets', {
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
          }),
        ]);

        if (!languagesResponse.ok || !lessonsResponse.ok || !snippetsResponse.ok) {
          throw new Error('Failed to fetch content');
        }

        const [languages, lessons, snippets] = await Promise.all([
          languagesResponse.json(),
          lessonsResponse.json(),
          snippetsResponse.json(),
        ]);

        // Combine all content items
        const allItems: ContentItem[] = [
          ...languages.map((lang: any) => ({
            id: lang.id,
            title: lang.name,
            type: 'language',
            current_version: lang.version || 1,
            versions: [],
          })),
          ...lessons.map((lesson: any) => ({
            id: lesson.id,
            title: lesson.title,
            type: 'lesson',
            current_version: lesson.version || 1,
            versions: [],
          })),
          ...snippets.map((snippet: any) => ({
            id: snippet.id,
            title: snippet.title,
            type: 'snippet',
            current_version: 1, // Snippets don't have version field yet
            versions: [],
          })),
        ];

        setContentItems(allItems);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch content');
      } finally {
        setLoading(false);
      }
    };

    if (token && user) {
      fetchContentOverview();
    }
  }, [token, user]);

  const fetchVersions = async (contentType: string, contentId: string) => {
    try {
      const response = await fetch(`/api/v1/admin/content/versions/${contentType}/${contentId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const versionsData = await response.json();
      setVersions(versionsData);
      setShowVersions(true);
    } catch (err) {
      alert(`Failed to fetch versions: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  };

  const handleRestoreVersion = async (contentType: string, contentId: string, version: number) => {
    if (!confirm(`Are you sure you want to restore version ${version}? This will replace the current content.`)) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/admin/content/versions/${contentType}/${contentId}/restore/${version}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      alert('Version restored successfully!');

      // Refresh versions
      await fetchVersions(contentType, contentId);

      // Refresh content overview
      window.location.reload();
    } catch (err) {
      alert(`Failed to restore version: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'language': return '🌐';
      case 'lesson': return '📚';
      case 'snippet': return '📄';
      default: return '📄';
    }
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'language': return 'bg-blue-100 text-blue-800';
      case 'lesson': return 'bg-green-100 text-green-800';
      case 'snippet': return 'bg-purple-100 text-purple-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading content versioning...</p>
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
                Content Versioning & Migration
              </h1>
            </div>
            <div className="flex items-center space-x-4">
              <button
                onClick={() => navigate('/admin/content/validate')}
                className="px-3 py-2 text-sm text-blue-600 hover:text-blue-900"
              >
                Validate Content
              </button>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Content Items List */}
          <div className="lg:col-span-1">
            <div className="bg-white shadow rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Content Items
                </h3>

                <div className="space-y-3 max-h-96 overflow-y-auto">
                  {contentItems.map((item) => (
                    <div
                      key={item.id}
                      className={`p-3 border rounded-lg cursor-pointer transition-colors ${
                        selectedItem?.id === item.id
                          ? 'border-blue-500 bg-blue-50'
                          : 'border-gray-200 hover:border-gray-300'
                      }`}
                      onClick={() => {
                        setSelectedItem(item);
                        fetchVersions(item.type, item.id);
                      }}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-3">
                          <span className="text-lg">{getTypeIcon(item.type)}</span>
                          <div>
                            <p className="text-sm font-medium text-gray-900 truncate">
                              {item.title}
                            </p>
                            <p className="text-xs text-gray-500">
                              v{item.current_version} • {item.type}
                            </p>
                          </div>
                        </div>
                        <span className={`px-2 py-1 text-xs font-medium rounded-full ${getTypeColor(item.type)}`}>
                          {item.type}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Version History */}
          <div className="lg:col-span-2">
            {selectedItem ? (
              <div className="space-y-6">
                {/* Selected Item Info */}
                <div className="bg-white shadow rounded-lg">
                  <div className="px-4 py-5 sm:p-6">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <span className="text-lg">{getTypeIcon(selectedItem.type)}</span>
                        <div>
                          <h3 className="text-lg font-medium text-gray-900">
                            {selectedItem.title}
                          </h3>
                          <p className="text-sm text-gray-500">
                            Current Version: {selectedItem.current_version}
                          </p>
                        </div>
                      </div>
                      <button
                        onClick={() => navigate(`/admin/${selectedItem.type}s/${selectedItem.id}/edit`)}
                        className="px-3 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700"
                      >
                        Edit Current
                      </button>
                    </div>
                  </div>
                </div>

                {/* Version History */}
                <div className="bg-white shadow rounded-lg">
                  <div className="px-4 py-5 sm:p-6">
                    <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                      Version History
                    </h3>

                    {versions.length === 0 ? (
                      <p className="text-gray-500 text-center py-4">
                        No version history available
                      </p>
                    ) : (
                      <div className="space-y-4">
                        {versions.map((version) => (
                          <div key={version.id} className="border border-gray-200 rounded-lg p-4">
                            <div className="flex items-center justify-between mb-3">
                              <div className="flex items-center space-x-3">
                                <span className={`w-3 h-3 rounded-full ${version.is_active ? 'bg-green-500' : 'bg-gray-400'}`}></span>
                                <div>
                                  <p className="text-sm font-medium text-gray-900">
                                    Version {version.version}
                                    {version.is_active && <span className="ml-2 text-xs text-green-600">(Active)</span>}
                                  </p>
                                  <p className="text-xs text-gray-500">
                                    {formatDate(version.created_at)} • {version.created_by}
                                  </p>
                                </div>
                              </div>
                              {!version.is_active && (
                                <button
                                  onClick={() => handleRestoreVersion(selectedItem.type, selectedItem.id, version.version)}
                                  className="px-3 py-2 bg-green-600 text-white text-sm rounded hover:bg-green-700"
                                >
                                  Restore
                                </button>
                              )}
                            </div>

                            {version.change_notes && (
                              <div className="mb-3">
                                <p className="text-sm text-gray-600">
                                  <strong>Changes:</strong> {version.change_notes}
                                </p>
                              </div>
                            )}

                            <details className="mt-3">
                              <summary className="text-sm text-blue-600 cursor-pointer hover:text-blue-800">
                                View Content
                              </summary>
                              <div className="mt-3 p-3 bg-gray-50 rounded text-xs font-mono max-h-48 overflow-y-auto">
                                <pre>{JSON.stringify(version.content, null, 2)}</pre>
                              </div>
                            </details>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ) : (
              <div className="bg-white shadow rounded-lg">
                <div className="px-4 py-5 sm:p-6 text-center">
                  <p className="text-gray-500">
                    Select a content item to view its version history
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ContentVersioning;
