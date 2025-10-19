import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../services/AuthContext';

interface ContentAnalytics {
  content_type: string;
  content_id: string;
  views: number;
  completions: number;
  avg_time: number;
  success_rate: number;
  difficulty_feedback: number;
  last_accessed: string;
  usage_trend: 'increasing' | 'decreasing' | 'stable';
}

interface UsageMetrics {
  total_sessions: number;
  total_users: number;
  avg_session_duration: number;
  popular_languages: { language: string; count: number }[];
  popular_content: { content_id: string; content_type: string; title: string; views: number }[];
  engagement_metrics: {
    daily_active_users: number;
    weekly_active_users: number;
    monthly_active_users: number;
    retention_rate: number;
  };
}

const ContentAnalytics: React.FC = () => {
  const [analytics, setAnalytics] = useState<ContentAnalytics[]>([]);
  const [usageMetrics, setUsageMetrics] = useState<UsageMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [timeRange, setTimeRange] = useState('7d');
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

  // Fetch analytics data
  useEffect(() => {
    const fetchAnalytics = async () => {
      try {
        // Simulate analytics data since we don't have real endpoints yet
        const mockAnalytics: ContentAnalytics[] = [
          {
            content_type: 'lesson',
            content_id: 'python_basics_1',
            views: 1250,
            completions: 890,
            avg_time: 25,
            success_rate: 0.85,
            difficulty_feedback: 2.3,
            last_accessed: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
            usage_trend: 'increasing',
          },
          {
            content_type: 'snippet',
            content_id: 'yaml_config_example',
            views: 450,
            completions: 380,
            avg_time: 8,
            success_rate: 0.92,
            difficulty_feedback: 1.8,
            last_accessed: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
            usage_trend: 'stable',
          },
          {
            content_type: 'lesson',
            content_id: 'javascript_functions',
            views: 890,
            completions: 560,
            avg_time: 35,
            success_rate: 0.78,
            difficulty_feedback: 3.1,
            last_accessed: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
            usage_trend: 'decreasing',
          },
        ];

        const mockUsageMetrics: UsageMetrics = {
          total_sessions: 15420,
          total_users: 3240,
          avg_session_duration: 18.5,
          popular_languages: [
            { language: 'python', count: 5420 },
            { language: 'javascript', count: 4680 },
            { language: 'yaml', count: 2890 },
          ],
          popular_content: [
            { content_id: 'python_basics_1', content_type: 'lesson', title: 'Python Basics', views: 1250 },
            { content_id: 'js_variables', content_type: 'lesson', title: 'JavaScript Variables', views: 980 },
            { content_id: 'yaml_config_example', content_type: 'snippet', title: 'YAML Configuration', views: 450 },
          ],
          engagement_metrics: {
            daily_active_users: 180,
            weekly_active_users: 890,
            monthly_active_users: 2340,
            retention_rate: 0.67,
          },
        };

        setAnalytics(mockAnalytics);
        setUsageMetrics(mockUsageMetrics);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch analytics');
      } finally {
        setLoading(false);
      }
    };

    if (token && user) {
      fetchAnalytics();
    }
  }, [token, user, timeRange]);

  const formatNumber = (num: number) => {
    return new Intl.NumberFormat().format(num);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'increasing': return <span className="text-green-600">↗</span>;
      case 'decreasing': return <span className="text-red-600">↘</span>;
      case 'stable': return <span className="text-gray-600">→</span>;
      default: return <span className="text-gray-600">→</span>;
    }
  };

  const getTrendColor = (trend: string) => {
    switch (trend) {
      case 'increasing': return 'text-green-600';
      case 'decreasing': return 'text-red-600';
      case 'stable': return 'text-gray-600';
      default: return 'text-gray-600';
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading analytics...</p>
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
                Content Analytics & Usage Tracking
              </h1>
            </div>
            <div className="flex items-center space-x-4">
              <select
                value={timeRange}
                onChange={(e) => setTimeRange(e.target.value)}
                className="text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="24h">Last 24 hours</option>
                <option value="7d">Last 7 days</option>
                <option value="30d">Last 30 days</option>
                <option value="90d">Last 90 days</option>
              </select>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        {/* Usage Overview Cards */}
        {usageMetrics && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-blue-500 rounded-md flex items-center justify-center">
                      <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
                      </svg>
                    </div>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        Total Sessions
                      </dt>
                      <dd className="text-lg font-medium text-gray-900">
                        {formatNumber(usageMetrics.total_sessions)}
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-green-500 rounded-md flex items-center justify-center">
                      <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                      </svg>
                    </div>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        Active Users
                      </dt>
                      <dd className="text-lg font-medium text-gray-900">
                        {formatNumber(usageMetrics.total_users)}
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-purple-500 rounded-md flex items-center justify-center">
                      <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        Avg Session Time
                      </dt>
                      <dd className="text-lg font-medium text-gray-900">
                        {usageMetrics.avg_session_duration}min
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-orange-500 rounded-md flex items-center justify-center">
                      <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                      </svg>
                    </div>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        Retention Rate
                      </dt>
                      <dd className="text-lg font-medium text-gray-900">
                        {(usageMetrics.engagement_metrics.retention_rate * 100).toFixed(1)}%
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Popular Languages */}
          <div className="bg-white shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                Popular Languages
              </h3>

              <div className="space-y-3">
                {usageMetrics?.popular_languages.map((lang, index) => (
                  <div key={lang.language} className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <span className="flex-shrink-0 w-8 h-8 bg-blue-100 text-blue-800 rounded-full flex items-center justify-center text-sm font-medium">
                        {index + 1}
                      </span>
                      <span className="text-sm font-medium text-gray-900 capitalize">
                        {lang.language}
                      </span>
                    </div>
                    <span className="text-sm text-gray-600">
                      {formatNumber(lang.count)} sessions
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Popular Content */}
          <div className="bg-white shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                Most Viewed Content
              </h3>

              <div className="space-y-3">
                {usageMetrics?.popular_content.map((content, index) => (
                  <div key={content.content_id} className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <span className="flex-shrink-0 w-8 h-8 bg-green-100 text-green-800 rounded-full flex items-center justify-center text-sm font-medium">
                        {index + 1}
                      </span>
                      <div>
                        <p className="text-sm font-medium text-gray-900">
                          {content.title}
                        </p>
                        <p className="text-xs text-gray-500 capitalize">
                          {content.content_type}
                        </p>
                      </div>
                    </div>
                    <span className="text-sm text-gray-600">
                      {formatNumber(content.views)} views
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Content Performance Table */}
        <div className="mt-8 bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
              Content Performance Details
            </h3>

            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Content
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Views
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Completions
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Success Rate
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Avg Time
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Trend
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Last Accessed
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {analytics.map((item) => (
                    <tr key={`${item.content_type}_${item.content_id}`}>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">
                          {item.content_id}
                        </div>
                        <div className="text-sm text-gray-500 capitalize">
                          {item.content_type}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {formatNumber(item.views)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {formatNumber(item.completions)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`text-sm font-medium ${
                          item.success_rate >= 0.8 ? 'text-green-600' :
                          item.success_rate >= 0.6 ? 'text-yellow-600' : 'text-red-600'
                        }`}>
                          {(item.success_rate * 100).toFixed(1)}%
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {item.avg_time}min
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className={`text-sm font-medium ${getTrendColor(item.usage_trend)}`}>
                          {getTrendIcon(item.usage_trend)} {item.usage_trend}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {formatDate(item.last_accessed)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        {/* Engagement Metrics */}
        {usageMetrics && (
          <div className="mt-8 bg-white shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                User Engagement Metrics
              </h3>

              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <div className="text-center">
                  <div className="text-2xl font-bold text-blue-600">
                    {formatNumber(usageMetrics.engagement_metrics.daily_active_users)}
                  </div>
                  <div className="text-sm text-gray-500">Daily Active Users</div>
                </div>

                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">
                    {formatNumber(usageMetrics.engagement_metrics.weekly_active_users)}
                  </div>
                  <div className="text-sm text-gray-500">Weekly Active Users</div>
                </div>

                <div className="text-center">
                  <div className="text-2xl font-bold text-purple-600">
                    {formatNumber(usageMetrics.engagement_metrics.monthly_active_users)}
                  </div>
                  <div className="text-sm text-gray-500">Monthly Active Users</div>
                </div>

                <div className="text-center">
                  <div className="text-2xl font-bold text-orange-600">
                    {(usageMetrics.engagement_metrics.retention_rate * 100).toFixed(1)}%
                  </div>
                  <div className="text-sm text-gray-500">Retention Rate</div>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Optimization Recommendations */}
        <div className="mt-8 bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
              Optimization Recommendations
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-3">
                <div className="p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
                  <h4 className="text-sm font-medium text-yellow-800 mb-2">
                    📈 Content to Promote
                  </h4>
                  <p className="text-sm text-yellow-700">
                    "python_basics_1" has high engagement but could benefit from more promotion to increase completion rates.
                  </p>
                </div>

                <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
                  <h4 className="text-sm font-medium text-red-800 mb-2">
                    ⚠️ Content to Review
                  </h4>
                  <p className="text-sm text-red-700">
                    "javascript_functions" shows decreasing usage trend and lower success rate. Consider updating content or difficulty assessment.
                  </p>
                </div>
              </div>

              <div className="space-y-3">
                <div className="p-3 bg-blue-50 border border-blue-200 rounded-lg">
                  <h4 className="text-sm font-medium text-blue-800 mb-2">
                    🎯 Improvement Opportunities
                  </h4>
                  <p className="text-sm text-blue-700">
                    YAML content has stable but relatively low engagement. Consider creating more interactive YAML lessons.
                  </p>
                </div>

                <div className="p-3 bg-green-50 border border-green-200 rounded-lg">
                  <h4 className="text-sm font-medium text-green-800 mb-2">
                    ✅ High Performers
                  </h4>
                  <p className="text-sm text-green-700">
                    Python and JavaScript content consistently perform well. Consider expanding these language offerings.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ContentAnalytics;
