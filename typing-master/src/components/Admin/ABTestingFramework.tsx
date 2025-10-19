import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../services/AuthContext';

interface ABTest {
  id: string;
  name: string;
  description: string;
  test_type: string;
  variants: any[];
  status: string;
  start_date: string;
  end_date: string;
  user_percentage: number;
  participant_ids: string[];
  results: any;
  created_by: string;
  created_at: string;
  updated_at: string;
}

interface TestResults {
  test_id: string;
  test_name: string;
  status: string;
  total_participants: number;
  conversion_rates: { [key: string]: number };
}

const ABTestingFramework: React.FC = () => {
  const [tests, setTests] = useState<ABTest[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedTest, setSelectedTest] = useState<ABTest | null>(null);
  const [testResults, setTestResults] = useState<TestResults | null>(null);
  const { token, user } = useAuth();
  const navigate = useNavigate();

  // Check if user is admin
  useEffect(() => {
    if (user && user.role !== 'admin' && user.role !== 'moderator') {
      navigate('/');
      return;
    }
  }, [user, navigate]);

  // Fetch A/B tests
  useEffect(() => {
    const fetchTests = async () => {
      try {
        const response = await fetch('/api/v1/admin/ab-tests', {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        setTests(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch A/B tests');
      } finally {
        setLoading(false);
      }
    };

    if (token && user) {
      fetchTests();
    }
  }, [token, user]);

  const handleCreateTest = async (testData: any) => {
    setCreating(true);
    try {
      const response = await fetch('/api/v1/admin/ab-tests', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(testData),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const newTest = await response.json();

      // Refresh tests
      const testsResponse = await fetch('/api/v1/admin/ab-tests', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (testsResponse.ok) {
        const updatedTests = await testsResponse.json();
        setTests(updatedTests);
      }

      setShowCreateForm(false);
      alert('A/B test created successfully!');
    } catch (err) {
      alert(`Failed to create test: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setCreating(false);
    }
  };

  const handleUpdateTest = async (testId: string, updates: any) => {
    try {
      const response = await fetch(`/api/v1/admin/ab-tests/${testId}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updates),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      // Refresh tests
      const testsResponse = await fetch('/api/v1/admin/ab-tests', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (testsResponse.ok) {
        const updatedTests = await testsResponse.json();
        setTests(updatedTests);
      }

      alert('Test updated successfully!');
    } catch (err) {
      alert(`Failed to update test: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  };

  const handleDeleteTest = async (testId: string) => {
    if (!confirm('Are you sure you want to delete this A/B test? This action cannot be undone.')) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/admin/ab-tests/${testId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      // Refresh tests
      const testsResponse = await fetch('/api/v1/admin/ab-tests', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (testsResponse.ok) {
        const updatedTests = await testsResponse.json();
        setTests(updatedTests);
      }

      alert('Test deleted successfully!');
    } catch (err) {
      alert(`Failed to delete test: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  };

  const handleViewResults = async (testId: string) => {
    try {
      const response = await fetch(`/api/v1/admin/ab-tests/${testId}/results`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const results = await response.json();
      setTestResults(results);
      setSelectedTest(tests.find(t => t.id === testId) || null);
    } catch (err) {
      alert(`Failed to fetch results: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-green-100 text-green-800';
      case 'paused': return 'bg-yellow-100 text-yellow-800';
      case 'completed': return 'bg-gray-100 text-gray-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading A/B tests...</p>
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
                A/B Testing Framework
              </h1>
            </div>
            <div className="flex items-center space-x-4">
              <button
                onClick={() => setShowCreateForm(true)}
                className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
              >
                Create Test
              </button>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        {/* Test Results Modal */}
        {testResults && selectedTest && (
          <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
            <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
              <div className="mt-3">
                <h3 className="text-lg font-medium text-gray-900 mb-4">
                  Test Results: {selectedTest.name}
                </h3>

                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Status:</span>
                    <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(testResults.status)}`}>
                      {testResults.status}
                    </span>
                  </div>

                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Participants:</span>
                    <span className="text-sm font-medium">{testResults.total_participants}</span>
                  </div>

                  <div>
                    <span className="text-sm text-gray-600 block mb-2">Conversion Rates:</span>
                    <div className="space-y-1">
                      {Object.entries(testResults.conversion_rates).map(([variant, rate]) => (
                        <div key={variant} className="flex justify-between text-sm">
                          <span>{variant}:</span>
                          <span className="font-medium">{(rate * 100).toFixed(1)}%</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>

                <div className="flex justify-end mt-6 space-x-3">
                  <button
                    onClick={() => {
                      setTestResults(null);
                      setSelectedTest(null);
                    }}
                    className="px-4 py-2 bg-gray-300 text-gray-700 rounded hover:bg-gray-400"
                  >
                    Close
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Create Test Form Modal */}
        {showCreateForm && (
          <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
            <div className="relative top-20 mx-auto p-5 border w-full max-w-2xl shadow-lg rounded-md bg-white">
              <CreateTestForm
                onSubmit={handleCreateTest}
                onCancel={() => setShowCreateForm(false)}
                loading={creating}
              />
            </div>
          </div>
        )}

        {/* Tests List */}
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
              Active A/B Tests
            </h3>

            {tests.length === 0 ? (
              <p className="text-gray-500 text-center py-8">
                No A/B tests have been created yet.
              </p>
            ) : (
              <div className="space-y-4">
                {tests.map((test) => (
                  <div key={test.id} className="border border-gray-200 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-3">
                      <div>
                        <h4 className="text-lg font-medium text-gray-900">
                          {test.name}
                        </h4>
                        <p className="text-sm text-gray-600">
                          {test.description}
                        </p>
                      </div>
                      <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(test.status)}`}>
                        {test.status}
                      </span>
                    </div>

                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                      <div>
                        <span className="text-gray-500">Type:</span>
                        <span className="ml-2 font-medium">{test.test_type}</span>
                      </div>
                      <div>
                        <span className="text-gray-500">Variants:</span>
                        <span className="ml-2 font-medium">{test.variants.length}</span>
                      </div>
                      <div>
                        <span className="text-gray-500">Participants:</span>
                        <span className="ml-2 font-medium">{test.user_percentage}%</span>
                      </div>
                      <div>
                        <span className="text-gray-500">End Date:</span>
                        <span className="ml-2 font-medium">{formatDate(test.end_date)}</span>
                      </div>
                    </div>

                    <div className="flex justify-end mt-4 space-x-3">
                      <button
                        onClick={() => handleViewResults(test.id)}
                        className="px-3 py-2 text-sm text-blue-600 hover:text-blue-900"
                      >
                        View Results
                      </button>
                      <select
                        value={test.status}
                        onChange={(e) => handleUpdateTest(test.id, { status: e.target.value })}
                        className="text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                      >
                        <option value="active">Active</option>
                        <option value="paused">Paused</option>
                        <option value="completed">Completed</option>
                      </select>
                      <button
                        onClick={() => handleDeleteTest(test.id)}
                        className="px-3 py-2 text-sm text-red-600 hover:text-red-900"
                      >
                        Delete
                      </button>
                    </div>
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

// Create Test Form Component
interface CreateTestFormProps {
  onSubmit: (data: any) => void;
  onCancel: () => void;
  loading: boolean;
}

const CreateTestForm: React.FC<CreateTestFormProps> = ({ onSubmit, onCancel, loading }) => {
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    test_type: 'scoring_weights',
    variants: [
      { name: 'Control', twpm_weight: 0.4, raw_accuracy_weight: 0.3, syntax_accuracy_weight: 0.3, backspace_penalty: 0.1, idle_time_penalty: 0.05 },
      { name: 'Variant A', twpm_weight: 0.5, raw_accuracy_weight: 0.25, syntax_accuracy_weight: 0.25, backspace_penalty: 0.15, idle_time_penalty: 0.05 },
    ],
    duration_days: 7,
    rollout_percentage: 10,
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
  };

  const addVariant = () => {
    const newVariant = {
      name: `Variant ${String.fromCharCode(65 + formData.variants.length)}`,
      twpm_weight: 0.4,
      raw_accuracy_weight: 0.3,
      syntax_accuracy_weight: 0.3,
      backspace_penalty: 0.1,
      idle_time_penalty: 0.05,
    };
    setFormData(prev => ({
      ...prev,
      variants: [...prev.variants, newVariant],
    }));
  };

  const updateVariant = (index: number, field: string, value: number) => {
    const updatedVariants = [...formData.variants];
    updatedVariants[index] = { ...updatedVariants[index], [field]: value };
    setFormData(prev => ({ ...prev, variants: updatedVariants }));
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div>
        <h3 className="text-lg font-medium text-gray-900 mb-4">
          Create A/B Test
        </h3>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <label htmlFor="test_name" className="block text-sm font-medium text-gray-700">
            Test Name
          </label>
          <input
            type="text"
            id="test_name"
            value={formData.name}
            onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
            className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
            required
          />
        </div>

        <div>
          <label htmlFor="test_type" className="block text-sm font-medium text-gray-700">
            Test Type
          </label>
          <select
            id="test_type"
            value={formData.test_type}
            onChange={(e) => setFormData(prev => ({ ...prev, test_type: e.target.value }))}
            className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="scoring_weights">Scoring Weights</option>
            <option value="ui_variant">UI Variant</option>
          </select>
        </div>
      </div>

      <div>
        <label htmlFor="test_description" className="block text-sm font-medium text-gray-700">
          Description
        </label>
        <textarea
          id="test_description"
          value={formData.description}
          onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
          rows={3}
          className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
          placeholder="Describe what this test is measuring..."
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <label htmlFor="duration" className="block text-sm font-medium text-gray-700">
            Duration (days)
          </label>
          <input
            type="number"
            id="duration"
            value={formData.duration_days}
            onChange={(e) => setFormData(prev => ({ ...prev, duration_days: parseInt(e.target.value) }))}
            min="1"
            max="30"
            className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
          />
        </div>

        <div>
          <label htmlFor="rollout" className="block text-sm font-medium text-gray-700">
            Rollout Percentage
          </label>
          <input
            type="number"
            id="rollout"
            value={formData.rollout_percentage}
            onChange={(e) => setFormData(prev => ({ ...prev, rollout_percentage: parseFloat(e.target.value) }))}
            min="1"
            max="100"
            step="0.1"
            className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
          />
        </div>
      </div>

      <div>
        <div className="flex justify-between items-center mb-3">
          <label className="block text-sm font-medium text-gray-700">
            Test Variants
          </label>
          <button
            type="button"
            onClick={addVariant}
            className="px-3 py-2 text-sm bg-gray-100 text-gray-700 rounded hover:bg-gray-200"
          >
            Add Variant
          </button>
        </div>

        <div className="space-y-4">
          {formData.variants.map((variant, index) => (
            <div key={index} className="border border-gray-200 rounded-lg p-4">
              <h4 className="font-medium text-gray-900 mb-3">{variant.name}</h4>

              {formData.test_type === 'scoring_weights' && (
                <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
                  <div>
                    <label className="block text-xs font-medium text-gray-500">TWPM Weight</label>
                    <input
                      type="number"
                      value={variant.twpm_weight}
                      onChange={(e) => updateVariant(index, 'twpm_weight', parseFloat(e.target.value))}
                      step="0.01"
                      min="0"
                      max="1"
                      className="mt-1 block w-full text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-500">Raw Accuracy Weight</label>
                    <input
                      type="number"
                      value={variant.raw_accuracy_weight}
                      onChange={(e) => updateVariant(index, 'raw_accuracy_weight', parseFloat(e.target.value))}
                      step="0.01"
                      min="0"
                      max="1"
                      className="mt-1 block w-full text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-500">Syntax Accuracy Weight</label>
                    <input
                      type="number"
                      value={variant.syntax_accuracy_weight}
                      onChange={(e) => updateVariant(index, 'syntax_accuracy_weight', parseFloat(e.target.value))}
                      step="0.01"
                      min="0"
                      max="1"
                      className="mt-1 block w-full text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-500">Backspace Penalty</label>
                    <input
                      type="number"
                      value={variant.backspace_penalty}
                      onChange={(e) => updateVariant(index, 'backspace_penalty', parseFloat(e.target.value))}
                      step="0.01"
                      min="0"
                      max="1"
                      className="mt-1 block w-full text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-500">Idle Time Penalty</label>
                    <input
                      type="number"
                      value={variant.idle_time_penalty}
                      onChange={(e) => updateVariant(index, 'idle_time_penalty', parseFloat(e.target.value))}
                      step="0.01"
                      min="0"
                      max="1"
                      className="mt-1 block w-full text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      <div className="flex justify-end space-x-3 pt-6 border-t">
        <button
          type="button"
          onClick={onCancel}
          className="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={loading}
          className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
        >
          {loading ? 'Creating...' : 'Create Test'}
        </button>
      </div>
    </form>
  );
};

export default ABTestingFramework;
