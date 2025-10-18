import React, { useState, useEffect, useCallback } from 'react';
import { SessionResult } from '../types/session';

interface ResultsAnalyticsInterfaceProps {
  sessionResults: SessionResult[];
  currentSession?: SessionResult;
  className?: string;
}

interface AnalyticsData {
  totalSessions: number;
  averageCPM: number;
  averageAccuracy: number;
  bestCPM: number;
  bestAccuracy: number;
  totalTimeSpent: number;
  languagesUsed: string[];
  progressData: Array<{
    date: string;
    cpm: number;
    accuracy: number;
    sessionCount: number;
  }>;
  errorPatterns: Array<{
    errorType: string;
    frequency: number;
    commonPositions: number[];
  }>;
}

const ResultsAnalyticsInterface: React.FC<ResultsAnalyticsInterfaceProps> = ({
  sessionResults,
  currentSession,
  className = '',
}) => {
  const [analytics, setAnalytics] = useState<AnalyticsData | null>(null);
  const [activeTab, setActiveTab] = useState<
    'overview' | 'progress' | 'errors' | 'export'
  >('overview');
  const [isExporting, setIsExporting] = useState(false);

  const calculateAnalytics = useCallback(() => {
    if (sessionResults.length === 0 && !currentSession) {
      setAnalytics(null);
      return;
    }

    const allResults = [...sessionResults];
    if (currentSession) {
      allResults.push(currentSession);
    }

    const totalSessions = allResults.length;
    const averageCPM =
      allResults.reduce(
        (sum, result) => sum + (result.metrics.speed.cpm || 0),
        0
      ) / totalSessions;
    const averageAccuracy =
      allResults.reduce(
        (sum, result) => sum + (result.metrics.accuracy.overallAccuracy || 0),
        0
      ) / totalSessions;
    const bestCPM = Math.max(
      ...allResults.map(result => result.metrics.speed.cpm || 0)
    );
    const bestAccuracy = Math.max(
      ...allResults.map(result => result.metrics.accuracy.overallAccuracy || 0)
    );

    const totalTimeSpent = allResults.reduce((sum, result) => {
      return sum + (result.summary.duration || 0);
    }, 0);

    const languagesUsed = Array.from(
      new Set(allResults.map(result => result.summary.language || 'unknown'))
    );

    // Calculate progress data (last 30 days)
    const progressData = calculateProgressData(allResults);

    // Analyze error patterns
    const errorPatterns = analyzeErrorPatterns(allResults);

    setAnalytics({
      totalSessions,
      averageCPM,
      averageAccuracy,
      bestCPM,
      bestAccuracy,
      totalTimeSpent,
      languagesUsed,
      progressData,
      errorPatterns,
    });
  }, [sessionResults, currentSession]);

  useEffect(() => {
    calculateAnalytics();
  }, [calculateAnalytics]);

  const calculateProgressData = (results: SessionResult[]) => {
    const last30Days = new Date();
    last30Days.setDate(last30Days.getDate() - 30);

    const dailyData: Record<
      string,
      { cpm: number; accuracy: number; count: number }
    > = {};

    results
      .filter(result => {
        const sessionDate = new Date(
          result.metrics.metadata.timestamp || Date.now()
        );
        return sessionDate >= last30Days;
      })
      .forEach(result => {
        const date = new Date(result.metrics.metadata.timestamp || Date.now())
          .toISOString()
          .split('T')[0];
        if (!dailyData[date]) {
          dailyData[date] = { cpm: 0, accuracy: 0, count: 0 };
        }
        dailyData[date].cpm += result.metrics.speed.cpm || 0;
        dailyData[date].accuracy +=
          result.metrics.accuracy.overallAccuracy || 0;
        dailyData[date].count += 1;
      });

    return Object.entries(dailyData).map(([date, data]) => ({
      date,
      cpm: data.cpm / data.count,
      accuracy: data.accuracy / data.count,
      sessionCount: data.count,
    }));
  };

  const analyzeErrorPatterns = (_results: SessionResult[]) => {
    // This would analyze the error patterns from session events
    // For now, return mock data structure
    return [
      {
        errorType: 'Typos',
        frequency: 45,
        commonPositions: [10, 25, 40],
      },
      {
        errorType: 'Missed Characters',
        frequency: 30,
        commonPositions: [5, 15, 35],
      },
      {
        errorType: 'Extra Characters',
        frequency: 25,
        commonPositions: [20, 30, 45],
      },
    ];
  };

  const exportToCSV = async () => {
    if (!analytics) return;

    setIsExporting(true);
    try {
      const csvData = sessionResults.map(result => ({
        Date: new Date(
          result.metrics.metadata.timestamp || Date.now()
        ).toLocaleDateString(),
        Language: result.summary.language || 'Unknown',
        CPM: result.metrics.speed.cpm || 0,
        Accuracy: `${((result.metrics.accuracy.overallAccuracy || 0) * 100).toFixed(1)}%`,
        'Token WPM': result.metrics.speed.twpm || 0,
        'Raw Accuracy': `${((result.metrics.accuracy.rawAccuracy || 0) * 100).toFixed(1)}%`,
        Duration: `${(result.summary.duration || 0) / 1000}s`,
      }));

      const headers = Object.keys(csvData[0] || {});
      const csvContent = [
        headers.join(','),
        ...csvData.map(row =>
          headers
            .map(header => `"${row[header as keyof typeof row]}"`)
            .join(',')
        ),
      ].join('\n');

      const blob = new Blob([csvContent], { type: 'text/csv' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `typing-results-${new Date().toISOString().split('T')[0]}.csv`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to export CSV:', error);
    } finally {
      setIsExporting(false);
    }
  };

  const exportToPDF = async () => {
    if (!analytics) return;

    setIsExporting(true);
    try {
      // For PDF export, we would need a library like jsPDF or html2canvas
      // For now, show a placeholder
      alert(
        'PDF export functionality would be implemented here with jsPDF library'
      );
    } catch (error) {
      console.error('Failed to export PDF:', error);
    } finally {
      setIsExporting(false);
    }
  };

  const shareResults = async () => {
    if (!analytics || !currentSession) return;

    const shareData = {
      title: 'My Typing Master Results',
      text: `Check out my typing results! CPM: ${currentSession.metrics.speed.cpm?.toFixed(1) || 'N/A'}, Accuracy: ${((currentSession.metrics.accuracy.overallAccuracy || 0) * 100).toFixed(1)}%`,
      url: window.location.href,
    };

    if (navigator.share && navigator.canShare(shareData)) {
      try {
        await navigator.share(shareData);
      } catch (error) {
        console.error('Failed to share:', error);
      }
    } else {
      // Fallback: copy to clipboard
      const text = `${shareData.text}\n${shareData.url}`;
      await navigator.clipboard.writeText(text);
      alert('Results copied to clipboard!');
    }
  };

  if (!analytics) {
    return (
      <div className={`results-analytics ${className}`}>
        <div className="no-data">No session data available</div>
      </div>
    );
  }

  return (
    <div className={`results-analytics ${className}`}>
      {/* Tab Navigation */}
      <div className="analytics-tabs" role="tablist">
        {[
          { key: 'overview', label: 'Overview', icon: '📊' },
          { key: 'progress', label: 'Progress', icon: '📈' },
          { key: 'errors', label: 'Error Analysis', icon: '🔍' },
          { key: 'export', label: 'Export', icon: '📤' },
        ].map(tab => (
          <button
            key={tab.key}
            className={`analytics-tab ${activeTab === tab.key ? 'active' : ''}`}
            onClick={() => setActiveTab(tab.key as typeof activeTab)}
            role="tab"
            aria-selected={activeTab === tab.key}
          >
            <span className="tab-icon">{tab.icon}</span>
            <span className="tab-label">{tab.label}</span>
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div className="analytics-content">
        {activeTab === 'overview' && (
          <div className="overview-panel">
            <div className="stats-grid">
              <div className="stat-card">
                <div className="stat-value">{analytics.totalSessions}</div>
                <div className="stat-label">Total Sessions</div>
              </div>
              <div className="stat-card">
                <div className="stat-value">
                  {analytics.averageCPM.toFixed(1)}
                </div>
                <div className="stat-label">Average CPM</div>
              </div>
              <div className="stat-card">
                <div className="stat-value">
                  {(analytics.averageAccuracy * 100).toFixed(1)}%
                </div>
                <div className="stat-label">Average Accuracy</div>
              </div>
              <div className="stat-card">
                <div className="stat-value">{analytics.bestCPM.toFixed(1)}</div>
                <div className="stat-label">Best CPM</div>
              </div>
              <div className="stat-card">
                <div className="stat-value">
                  {analytics.bestAccuracy.toFixed(1)}
                </div>
                <div className="stat-label">Best Accuracy</div>
              </div>
              <div className="stat-card">
                <div className="stat-value">
                  {Math.floor(analytics.totalTimeSpent / 60000)}m
                </div>
                <div className="stat-label">Time Practiced</div>
              </div>
            </div>

            <div className="languages-section">
              <h3>Languages Practiced</h3>
              <div className="language-badges">
                {analytics.languagesUsed.map(language => (
                  <span key={language} className="language-badge">
                    {language}
                  </span>
                ))}
              </div>
            </div>
          </div>
        )}

        {activeTab === 'progress' && (
          <div className="progress-panel">
            <h3>Progress Over Time</h3>
            {analytics.progressData.length > 0 ? (
              <div className="progress-chart">
                {/* Simple progress visualization */}
                <div className="chart-container">
                  <div className="chart-legend">
                    <div className="legend-item">
                      <div className="legend-color cpm"></div>
                      <span>CPM</span>
                    </div>
                    <div className="legend-item">
                      <div className="legend-color accuracy"></div>
                      <span>Accuracy</span>
                    </div>
                  </div>
                  <div className="chart-bars">
                    {analytics.progressData.slice(-10).map((day, _index) => (
                      <div key={day.date} className="chart-day">
                        <div className="day-bars">
                          <div
                            className="bar cpm"
                            style={{
                              height: `${(day.cpm / analytics.bestCPM) * 100}px`,
                            }}
                            title={`CPM: ${day.cpm.toFixed(1)}`}
                          ></div>
                          <div
                            className="bar accuracy"
                            style={{ height: `${day.accuracy * 100}px` }}
                            title={`Accuracy: ${(day.accuracy * 100).toFixed(1)}%`}
                          ></div>
                        </div>
                        <div className="day-label">
                          {new Date(day.date).toLocaleDateString('en-US', {
                            month: 'short',
                            day: 'numeric',
                          })}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            ) : (
              <div className="no-progress-data">
                No progress data available yet
              </div>
            )}
          </div>
        )}

        {activeTab === 'errors' && (
          <div className="errors-panel">
            <h3>Error Analysis</h3>
            <div className="error-heatmap">
              {analytics.errorPatterns.map((pattern, index) => (
                <div key={index} className="error-pattern">
                  <div className="pattern-header">
                    <h4>{pattern.errorType}</h4>
                    <span className="frequency">{pattern.frequency}%</span>
                  </div>
                  <div className="pattern-details">
                    <p>
                      Common positions: {pattern.commonPositions.join(', ')}
                    </p>
                    <div className="position-indicators">
                      {Array.from({ length: 50 }, (_, i) => (
                        <div
                          key={i}
                          className={`position-indicator ${
                            pattern.commonPositions.includes(i)
                              ? 'error-position'
                              : ''
                          }`}
                        ></div>
                      ))}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'export' && (
          <div className="export-panel">
            <h3>Export & Share</h3>
            <div className="export-options">
              <div className="export-section">
                <h4>Export Data</h4>
                <div className="export-buttons">
                  <button
                    className="export-button csv"
                    onClick={exportToCSV}
                    disabled={isExporting}
                  >
                    📄 Export CSV
                  </button>
                  <button
                    className="export-button pdf"
                    onClick={exportToPDF}
                    disabled={isExporting}
                  >
                    📋 Export PDF
                  </button>
                </div>
              </div>

              <div className="export-section">
                <h4>Share Results</h4>
                <button className="share-button" onClick={shareResults}>
                  🔗 Share Results
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ResultsAnalyticsInterface;
