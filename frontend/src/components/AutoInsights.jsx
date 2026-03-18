import { useState, useEffect, useCallback } from 'react';

function AutoInsights() {
  const [insights, setInsights] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filterUnread, setFilterUnread] = useState(false);
  const [expandedInsight, setExpandedInsight] = useState(null);

  const fetchInsights = useCallback(async () => {
    try {
      if (window.go?.main?.App?.GetAutoInsights) {
        const insightsList = await window.go.main.App.GetAutoInsights(filterUnread);
        setInsights(insightsList || []);
      }
    } catch (err) {
      console.error('Error fetching insights:', err);
    } finally {
      setLoading(false);
    }
  }, [filterUnread]);

  const markAsRead = async (insightId) => {
    try {
      if (window.go?.main?.App?.MarkInsightAsRead) {
        await window.go.main.App.MarkInsightAsRead(insightId);
        fetchInsights(); // Refresh the list
      }
    } catch (err) {
      console.error('Error marking insight as read:', err);
    }
  };

  const clearAllInsights = async () => {
    try {
      if (window.go?.main?.App?.ClearAllInsights) {
        await window.go.main.App.ClearAllInsights();
        fetchInsights(); // Refresh the list
      }
    } catch (err) {
      console.error('Error clearing all insights:', err);
    }
  };

  const clearOldInsights = async () => {
    try {
      if (window.go?.main?.App?.ClearOldInsights) {
        await window.go.main.App.ClearOldInsights();
        fetchInsights(); // Refresh the list
      }
    } catch (err) {
      console.error('Error clearing old insights:', err);
    }
  };

  const getSeverityClass = (severity) => {
    switch (severity) {
      case 'critical':
        return 'insight-critical';
      case 'warning':
        return 'insight-warning';
      case 'info':
      default:
        return 'insight-info';
    }
  };

  const getCategoryIcon = (category) => {
    switch (category) {
      case 'performance':
        return '⚡';
      case 'security':
        return '🛡️';
      case 'network':
        return '🌐';
      case 'process':
        return '⚙️';
      default:
        return '💡';
    }
  };

  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now - date;
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);

    if (minutes < 1) return 'Just now';
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    return date.toLocaleDateString();
  };

  // Listen for new insights from backend
  useEffect(() => {
    const handleNewInsight = (data) => {
      if (data && data.insight) {
        // Add new insight to the list and show notification
        setInsights(prev => [data.insight, ...prev]);
      }
    };

    // Listen to events from Go backend
    if (window.runtime?.EventsOn) {
      window.runtime.EventsOn('insights:new', handleNewInsight);
      return () => {
        if (window.runtime?.EventsOff) {
          window.runtime.EventsOff('insights:new', handleNewInsight);
        }
      };
    }
  }, []);

  useEffect(() => {
    fetchInsights();
    const interval = setInterval(fetchInsights, 15000); // Check every 15 seconds
    return () => clearInterval(interval);
  }, [fetchInsights]);

  const unreadCount = insights.filter(insight => !insight.isRead).length;

  return (
    <div className="insights-container">
      <div className="insights-header">
        <div className="insights-title-section">
          <h2>Auto Insights</h2>
          <p className="insights-subtitle">
            AI-powered automatic system monitoring insights
          </p>
        </div>
        <div className="insights-controls">
          <label className="checkbox-container">
            <input
              type="checkbox"
              checked={filterUnread}
              onChange={(e) => setFilterUnread(e.target.checked)}
            />
            <span className="checkmark"></span>
            Show only unread ({unreadCount})
          </label>
          <button 
            className="refresh-btn" 
            onClick={fetchInsights}
            disabled={loading}
            title="Refresh insights"
          >
            {loading ? '⟳' : '↻'}
          </button>
          <button 
            className="clear-btn" 
            onClick={clearAllInsights}
            title="Clear all insights"
          >
            🗑️
          </button>
        </div>
      </div>

      {loading && insights.length === 0 ? (
        <div className="insights-loading">
          <div className="loading-spinner"></div>
          <p>Loading insights...</p>
        </div>
      ) : insights.length === 0 ? (
        <div className="insights-empty">
          <div className="empty-icon">💡</div>
          <h3>No insights available</h3>
          <p>
            {filterUnread 
              ? 'No unread insights. Check back later or view all insights.'
              : 'Auto Insights is monitoring your system. New insights will appear here automatically.'
            }
          </p>
        </div>
      ) : (
        <div className="insights-list">
          {insights.map((insight) => (
            <div 
              key={insight.id} 
              className={`insight-card ${getSeverityClass(insight.severity)} ${!insight.isRead ? 'insight-unread' : ''}`}
            >
              <div className="insight-main" onClick={() => setExpandedInsight(expandedInsight === insight.id ? null : insight.id)}>
                <div className="insight-header-row">
                  <div className="insight-meta">
                    <span className="insight-category">
                      {getCategoryIcon(insight.category)} {insight.category}
                    </span>
                    <span className="insight-severity">{insight.severity}</span>
                    <span className="insight-time">{formatTimestamp(insight.timestamp)}</span>
                  </div>
                  {!insight.isRead && (
                    <button
                      className="mark-read-btn"
                      onClick={(e) => {
                        e.stopPropagation();
                        markAsRead(insight.id);
                      }}
                      title="Mark as read"
                    >
                      ✓
                    </button>
                  )}
                </div>
                <h3 className="insight-title">{insight.title}</h3>
                <p className="insight-message">{insight.message}</p>
              </div>

              {expandedInsight === insight.id && (
                <div className="insight-details">
                  {insight.actionItems && insight.actionItems.length > 0 && (
                    <div className="action-items">
                      <h4>Recommended Actions:</h4>
                      <ul>
                        {insight.actionItems.map((action, index) => (
                          <li key={index}>{action}</li>
                        ))}
                      </ul>
                    </div>
                  )}
                  {insight.data && (
                    <div className="insight-data">
                      <h4>Technical Details:</h4>
                      <code className="data-block">{insight.data}</code>
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      <div className="insights-info">
        <p>
          <strong>Auto Insights</strong> continuously monitors your system and automatically generates 
          intelligent notifications when it detects patterns that need attention. This proactive 
          monitoring helps you stay ahead of potential issues.
        </p>
      </div>
    </div>
  );
}

export default AutoInsights;