import { useState, useEffect, useCallback } from 'react';
import { useErrorDialog } from '../contexts/ErrorDialogContext';

function DevEnvironments() {
  const { showError } = useErrorDialog();
  const [devInfo, setDevInfo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [expandedContainer, setExpandedContainer] = useState(null);
  const [containerActions, setContainerActions] = useState({});

  const normalizeDevInfo = (info) => ({
    containers: Array.isArray(info?.containers) ? info.containers : [],
    environments: Array.isArray(info?.environments) ? info.environments : [],
    devPorts: Array.isArray(info?.devPorts) ? info.devPorts : [],
    dockerRunning: Boolean(info?.dockerRunning),
  });

  const fetchDevInfo = useCallback(async () => {
    try {
      if (window.go?.main?.App?.GetDevEnvironmentInfo) {
        const info = await window.go.main.App.GetDevEnvironmentInfo();
        setDevInfo(normalizeDevInfo(info));
      }
    } catch (err) {
      console.error('Error fetching dev environment info:', err);
      setDevInfo(normalizeDevInfo(null));
    } finally {
      setLoading(false);
    }
  }, []);

  const handleContainerAction = async (containerId, action) => {
    setContainerActions(prev => ({ ...prev, [containerId]: action }));
    
    try {
      let result;
      switch (action) {
        case 'start':
          result = await window.go.main.App.StartContainer(containerId);
          break;
        case 'stop':
          result = await window.go.main.App.StopContainer(containerId);
          break;
        case 'restart':
          result = await window.go.main.App.RestartContainer(containerId);
          break;
        case 'remove':
          if (confirm('Are you sure you want to remove this container? This action cannot be undone.')) {
            result = await window.go.main.App.RemoveContainer(containerId);
          }
          break;
        default:
          return;
      }
      
      // Refresh the data after action
      setTimeout(() => {
        fetchDevInfo();
        setContainerActions(prev => {
          const updated = { ...prev };
          delete updated[containerId];
          return updated;
        });
      }, 1000);
    } catch (err) {
      console.error(`Error ${action}ing container:`, err);
      showError(`${action.charAt(0).toUpperCase() + action.slice(1)} Container Failed`, 
        `Failed to ${action} container: ${err.message || err}`);
      setContainerActions(prev => {
        const updated = { ...prev };
        delete updated[containerId];
        return updated;
      });
    }
  };

  useEffect(() => {
    fetchDevInfo();
    const interval = setInterval(fetchDevInfo, 5000); // Refresh every 5 seconds
    return () => clearInterval(interval);
  }, [fetchDevInfo]);

  const openURL = (url) => {
    if (window.runtime?.BrowserOpenURL) {
      window.runtime.BrowserOpenURL(url);
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'running':
        return '🟢';
      case 'exited':
        return '🔴';
      case 'paused':
        return '🟡';
      default:
        return '⚫';
    }
  };

  const getTypeIcon = (type) => {
    switch (type) {
      case 'web':
        return '🌐';
      case 'database':
        return '💾';
      case 'api':
        return '🔌';
      case 'proxy':
        return '🔀';
      case 'build':
        return '🔨';
      case 'container':
        return '🐳';
      default:
        return '⚙️';
    }
  };

  // Guard: Don't render until we have initial data
  if (loading || devInfo === null) {
    return (
      <div className="dev-environments-container">
        <div className="loading-spinner">
          <div className="spinner"></div>
          <p>Scanning development environments...</p>
        </div>
      </div>
    );
  }

  // Ensure devInfo is never null
  const safeDevInfo = normalizeDevInfo(devInfo);

  if (safeDevInfo.environments?.length === 0 && safeDevInfo.containers?.length === 0) {
    return (
      <div className="dev-environments-container">
        <div className="dev-environments-header">
          <h3>🚀 Development Environments</h3>
          <div className="docker-status">
            {safeDevInfo.dockerRunning ? (
              <span className="status-badge docker-running">🐳 Docker Available</span>
            ) : (
              <span className="status-badge docker-not-running">🐳 Docker Not Available</span>
            )}
          </div>
        </div>
        <div className="dev-environments-empty">
          <div className="empty-icon">🔍</div>
          <h4>No Development Environments Detected</h4>
          <p>
            SysMind automatically detects development servers, databases, and containerized applications.
            Start a development server or Docker container to see it here!
          </p>
          <div className="detection-examples">
            <p><strong>Detects:</strong></p>
            <ul>
              <li>⚛️ Next.js, React, Vue.js apps</li>
              <li>🐘 PostgreSQL, MySQL, Redis</li>
              <li>🐳 Docker containers</li>
              <li>🐍 Python Flask/Django apps</li>
              <li>🟢 Node.js servers</li>
            </ul>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="dev-environments-container">
      <div className="dev-environments-header">
        <h3>🚀 Development Environments ({safeDevInfo.environments.length})</h3>
        <div className="header-controls">
          <div className="docker-status">
            {safeDevInfo.dockerRunning ? (
              <span className="status-badge docker-running">🐳 Docker Active</span>
            ) : (
              <span className="status-badge docker-not-running">🐳 Docker Inactive</span>
            )}
          </div>
          <button 
            className="refresh-btn" 
            onClick={fetchDevInfo}
            title="Refresh environments"
          >
            ↻
          </button>
        </div>
      </div>

      {/* Development Environments Grid */}
      <div className="dev-environments-grid">
        {safeDevInfo.environments.map((env) => (
          <div 
            key={env.id} 
            className={`dev-env-card ${env.type} ${env.status}`}
          >
            <div className="dev-env-header">
              <div className="dev-env-icon-title">
                <span className="dev-env-icon">{env.icon}</span>
                <div className="dev-env-title-section">
                  <h4 className="dev-env-name">{env.name}</h4>
                  <p className="dev-env-description">{env.description}</p>
                </div>
              </div>
              <div className="dev-env-status">
                <span className="status-indicator" title={`Status: ${env.status}`}>
                  {getStatusIcon(env.status)}
                </span>
                <span className="type-indicator" title={`Type: ${env.type}`}>
                  {getTypeIcon(env.type)}
                </span>
              </div>
            </div>

            <div className="dev-env-details">
              {env.port && (
                <div className="dev-env-detail">
                  <span className="detail-label">Port:</span>
                  <span className="detail-value">:{env.port}</span>
                </div>
              )}
              {env.technology && (
                <div className="dev-env-detail">
                  <span className="detail-label">Tech:</span>
                  <span className="detail-value">{env.technology}</span>
                </div>
              )}
              {env.processName && (
                <div className="dev-env-detail">
                  <span className="detail-label">Process:</span>
                  <span className="detail-value">{env.processName}</span>
                </div>
              )}
              {env.containerID && (
                <div className="dev-env-detail">
                  <span className="detail-label">Container:</span>
                  <span className="detail-value">{env.containerID.substring(0, 12)}</span>
                </div>
              )}
            </div>

            {/* Action buttons */}
            <div className="dev-env-actions">
              {env.urls && env.urls.length > 0 && (
                <button 
                  className="btn-primary btn-small"
                  onClick={() => openURL(env.urls[0])}
                  title={`Open ${env.urls[0]}`}
                >
                  🌐 Open
                </button>
              )}
              {env.containerID && (
                <button 
                  className="btn-secondary btn-small"
                  onClick={() => setExpandedContainer(
                    expandedContainer === env.containerID ? null : env.containerID
                  )}
                  title="Container details"
                >
                  📋 Details
                </button>
              )}
            </div>

            {/* Expanded container details */}
            {expandedContainer === env.containerID && (
              <div className="container-details">
                {safeDevInfo.containers
                  .filter(c => c.id === env.containerID)
                  .map(container => (
                    <div key={container.id} className="container-info">
                      <div className="container-stats">
                        <div className="stat">
                          <span className="stat-label">CPU:</span>
                          <span className="stat-value">{container.cpuPercent.toFixed(1)}%</span>
                        </div>
                        <div className="stat">
                          <span className="stat-label">Memory:</span>
                          <span className="stat-value">{container.memoryMB.toFixed(0)} MB</span>
                        </div>
                        <div className="stat">
                          <span className="stat-label">Image:</span>
                          <span className="stat-value">{container.image}</span>
                        </div>
                      </div>
                      <div className="container-ports">
                        <span className="ports-label">Ports:</span>
                        {container.ports.map((port, idx) => (
                          <span key={idx} className="port-mapping">
                            {port.publicPort ? `${port.publicPort}:` : ''}{port.privatePort}/{port.type}
                          </span>
                        ))}
                      </div>
                      
                      {/* Container Management Actions */}
                      <div className="container-actions">
                        <div className="container-actions-header">
                          <span className="actions-label">Container Actions:</span>
                        </div>
                        <div className="container-actions-buttons">
                          {container.state === 'running' ? (
                            <>
                              <button 
                                className="btn-container-action btn-stop"
                                onClick={() => handleContainerAction(container.id, 'stop')}
                                disabled={containerActions[container.id] === 'stop'}
                                title="Stop container"
                              >
                                {containerActions[container.id] === 'stop' ? '⏸️ Stopping...' : '⏸️ Stop'}
                              </button>
                              <button 
                                className="btn-container-action btn-restart"
                                onClick={() => handleContainerAction(container.id, 'restart')}
                                disabled={containerActions[container.id] === 'restart'}
                                title="Restart container"
                              >
                                {containerActions[container.id] === 'restart' ? '🔄 Restarting...' : '🔄 Restart'}
                              </button>
                            </>
                          ) : (
                            <button 
                              className="btn-container-action btn-start"
                              onClick={() => handleContainerAction(container.id, 'start')}
                              disabled={containerActions[container.id] === 'start'}
                              title="Start container"
                            >
                              {containerActions[container.id] === 'start' ? '▶️ Starting...' : '▶️ Start'}
                            </button>
                          )}
                          
                          {container.state === 'exited' && (
                            <button 
                              className="btn-container-action btn-remove"
                              onClick={() => handleContainerAction(container.id, 'remove')}
                              disabled={containerActions[container.id] === 'remove'}
                              title="Remove container"
                            >
                              {containerActions[container.id] === 'remove' ? '🗑️ Removing...' : '🗑️ Remove'}
                            </button>
                          )}
                          
                          <button 
                            className="btn-container-action btn-info"
                            onClick={() => alert(`Container ID: ${container.id}\nName: ${container.name}\nImage: ${container.image}\nStatus: ${container.status}`)}
                            title="Container info"
                          >
                            ℹ️ Info
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Quick stats */}
      <div className="dev-environments-summary">
        <div className="summary-stat">
          <span className="summary-number">{safeDevInfo.containers.length}</span>
          <span className="summary-label">Containers</span>
        </div>
        <div className="summary-stat">
          <span className="summary-number">
            {safeDevInfo.environments.filter(e => e.status === 'running').length}
          </span>
          <span className="summary-label">Running</span>
        </div>
        <div className="summary-stat">
          <span className="summary-number">{safeDevInfo.devPorts.length}</span>
          <span className="summary-label">Dev Ports</span>
        </div>
        <div className="summary-stat">
          <span className="summary-number">
            {new Set(safeDevInfo.environments.map(e => e.technology)).size}
          </span>
          <span className="summary-label">Technologies</span>
        </div>
      </div>
    </div>
  );
}

export default DevEnvironments;
