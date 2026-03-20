import { useState, useEffect, useCallback } from 'react';
import { useErrorDialog } from '../contexts/ErrorDialogContext';
import SystemStats from './SystemStats';
import TimeMachine from './TimeMachine';
import ResourceTimeline from './ResourceTimeline';

function Dashboard() {
  const { showError } = useErrorDialog();
  const [processes, setProcesses] = useState([]);
  const [ports, setPorts] = useState([]);
  const [networkUsage, setNetworkUsage] = useState([]);
  const [processFilter, setProcessFilter] = useState('');
  const [portFilter, setPortFilter] = useState('');
  const [selectedProcess, setSelectedProcess] = useState(null);
  const [sortConfig, setSortConfig] = useState({ key: 'cpuPercent', direction: 'desc' });
  const [explainModal, setExplainModal] = useState({ show: false, loading: false, content: '', title: '', type: '' });

  const fetchData = useCallback(async () => {
    try {
      if (window.go?.main?.App) {
        const [procs, pts, net] = await Promise.all([
          window.go.main.App.GetProcesses(),
          window.go.main.App.GetPorts(),
          window.go.main.App.GetNetworkUsage(),
        ]);
        setProcesses(procs || []);
        setPorts(pts || []);
        setNetworkUsage(net || []);
      }
    } catch (err) {
      console.error('Error fetching data:', err);
    }
  }, []);

  // Explain This functionality
  const explainProcess = async (process) => {
    setExplainModal({ 
      show: true, 
      loading: true, 
      content: '', 
      title: `Explain Process: ${process.name}`,
      type: 'process'
    });

    try {
      const question = `Please explain this process in detail:

Process Name: ${process.name}
PID: ${process.pid}
CPU Usage: ${process.cpuPercent}%
Memory Usage: ${process.memoryMB} MB
Status: ${process.status}

Please provide:
1. What this process does and its purpose
2. Whether this CPU/memory usage is normal for this process
3. If there are any concerns or recommendations
4. Common reasons why this process might be using high resources
5. Whether this process is safe/legitimate

Be concise but informative.`;

      const response = await window.go.main.App.AskAI(question);
      
      if (response.success) {
        setExplainModal(prev => ({ 
          ...prev, 
          loading: false, 
          content: response.response 
        }));
      } else {
        setExplainModal(prev => ({ 
          ...prev, 
          loading: false, 
          content: `Error: ${response.error}` 
        }));
      }
    } catch (err) {
      setExplainModal(prev => ({ 
        ...prev, 
        loading: false, 
        content: `Error: ${err.message}` 
      }));
    }
  };

  const explainPort = async (port) => {
    setExplainModal({ 
      show: true, 
      loading: true, 
      content: '', 
      title: `Explain Port: ${port.port}/${port.protocol}`,
      type: 'port'
    });

    try {
      const question = `Please explain this network port in detail:

Port: ${port.port}
Protocol: ${port.protocol}
Process: ${port.processName} (PID: ${port.pid})
State: ${port.state}
Local Address: ${port.localAddr}

Please provide:
1. What this port is commonly used for
2. Which applications typically use this port
3. Whether having this port open is normal/safe
4. Any security considerations or risks
5. Whether this port should be exposed or blocked

Be concise but informative.`;

      const response = await window.go.main.App.AskAI(question);
      
      if (response.success) {
        setExplainModal(prev => ({ 
          ...prev, 
          loading: false, 
          content: response.response 
        }));
      } else {
        setExplainModal(prev => ({ 
          ...prev, 
          loading: false, 
          content: `Error: ${response.error}` 
        }));
      }
    } catch (err) {
      setExplainModal(prev => ({ 
        ...prev, 
        loading: false, 
        content: `Error: ${err.message}` 
      }));
    }
  };

  const explainNetworkActivity = async () => {
    setExplainModal({ 
      show: true, 
      loading: true, 
      content: '', 
      title: 'Explain Network Activity',
      type: 'network'
    });

    try {
      const question = `Please analyze the current network activity and explain:

1. Which processes are using the most bandwidth
2. Whether the network usage patterns are normal
3. Any suspicious or concerning connections
4. Which applications might be responsible for high data usage
5. Recommendations for optimizing network performance

Focus on practical insights and actionable recommendations.`;

      const response = await window.go.main.App.AskAI(question);
      
      if (response.success) {
        setExplainModal(prev => ({ 
          ...prev, 
          loading: false, 
          content: response.response 
        }));
      } else {
        setExplainModal(prev => ({ 
          ...prev, 
          loading: false, 
          content: `Error: ${response.error}` 
        }));
      }
    } catch (err) {
      setExplainModal(prev => ({ 
        ...prev, 
        loading: false, 
        content: `Error: ${err.message}` 
      }));
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 3000);
    return () => clearInterval(interval);
  }, [fetchData]);

  const handleSort = (key) => {
    setSortConfig(prev => ({
      key,
      direction: prev.key === key && prev.direction === 'desc' ? 'asc' : 'desc'
    }));
  };

  const sortedProcesses = [...processes]
    .filter(p => p.name?.toLowerCase().includes(processFilter.toLowerCase()))
    .sort((a, b) => {
      const aVal = a[sortConfig.key] || 0;
      const bVal = b[sortConfig.key] || 0;
      return sortConfig.direction === 'desc' ? bVal - aVal : aVal - bVal;
    })
    .slice(0, 50);

  const filteredPorts = ports
    .filter(p => 
      p.processName?.toLowerCase().includes(portFilter.toLowerCase()) ||
      String(p.port).includes(portFilter) ||
      p.protocol?.toLowerCase().includes(portFilter.toLowerCase())
    )
    .slice(0, 50);

  const topNetworkUsage = [...networkUsage]
    .sort((a, b) => (b.downloadSpeed + b.uploadSpeed) - (a.downloadSpeed + a.uploadSpeed))
    .slice(0, 20);

  const formatBytes = (bytes) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
    return `${(bytes / 1024 / 1024 / 1024).toFixed(1)} GB`;
  };

  const formatSpeed = (bytesPerSec) => {
    return formatBytes(bytesPerSec) + '/s';
  };

  const handleKillProcess = async (pid) => {
    if (!confirm(`Kill process ${pid}?`)) return;
    try {
      const result = await window.go.main.App.KillProcess(pid);
      if (result.success) {
        fetchData();
      } else {
        showError('Kill Process Failed', result.error || 'Failed to kill the process');
      }
    } catch (err) {
      showError('Kill Process Error', err.message || 'An error occurred while killing the process');
    }
    setSelectedProcess(null);
  };

  const handleSetPriority = async (pid, priority) => {
    try {
      const result = await window.go.main.App.SetProcessPriority(pid, priority);
      if (!result.success) {
        showError('Set Priority Failed', result.error || 'Failed to set process priority');
      }
    } catch (err) {
      showError('Set Priority Error', err.message || 'An error occurred while setting priority');
    }
  };

  return (
    <div className="panel">
      {/* System Stats at top */}
      <SystemStats />

      <ResourceTimeline />

      <TimeMachine compact />

      <div className="dashboard-grid">
        {/* Processes Table */}
        <div className="table-container">
          <div className="table-header">
            <span className="table-title">Processes ({processes.length})</span>
            <input
              type="text"
              className="table-filter"
              placeholder="Filter processes..."
              value={processFilter}
              onChange={(e) => setProcessFilter(e.target.value)}
            />
          </div>
          <div className="table-scroll">
            <table className="data-table sortable">
              <thead>
                <tr>
                  <th onClick={() => handleSort('pid')}>
                    PID {sortConfig.key === 'pid' && (sortConfig.direction === 'desc' ? '↓' : '↑')}
                  </th>
                  <th>Name</th>
                  <th onClick={() => handleSort('cpuPercent')}>
                    CPU % {sortConfig.key === 'cpuPercent' && (sortConfig.direction === 'desc' ? '↓' : '↑')}
                  </th>
                  <th onClick={() => handleSort('memoryMB')}>
                    Memory {sortConfig.key === 'memoryMB' && (sortConfig.direction === 'desc' ? '↓' : '↑')}
                  </th>
                  <th>Status</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {sortedProcesses.map((proc) => (
                  <tr 
                    key={proc.pid}
                    className={selectedProcess === proc.pid ? 'selected' : ''}
                    onClick={() => setSelectedProcess(proc.pid)}
                  >
                    <td className="text-mono">{proc.pid}</td>
                    <td title={proc.commandLine}>{proc.name || 'Unknown'}</td>
                    <td className={`number ${proc.cpuPercent > 50 ? 'text-danger' : proc.cpuPercent > 20 ? 'text-warning' : ''}`}>
                      {proc.cpuPercent?.toFixed(1)}%
                    </td>
                    <td className="number">{proc.memoryMB?.toFixed(1)} MB</td>
                    <td>
                      <span className="badge badge-other">{proc.status}</span>
                    </td>
                    <td className="actions-cell">
                      <button 
                        className="btn-explain" 
                        title="🧩 Explain This Process"
                        onClick={(e) => { e.stopPropagation(); explainProcess(proc); }}
                      >
                        🧩
                      </button>
                      <button 
                        className="btn-icon" 
                        title="Kill Process"
                        onClick={(e) => { e.stopPropagation(); handleKillProcess(proc.pid); }}
                      >
                        ✕
                      </button>
                      <select 
                        className="priority-select"
                        title="Set Priority"
                        onChange={(e) => { e.stopPropagation(); handleSetPriority(proc.pid, parseInt(e.target.value)); }}
                        onClick={(e) => e.stopPropagation()}
                        defaultValue="0"
                      >
                        <option value="-10">High</option>
                        <option value="0">Normal</option>
                        <option value="10">Low</option>
                        <option value="19">Idle</option>
                      </select>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        <div className="table-container">
          <div className="table-header">
            <span className="table-title">Open Ports ({ports.length})</span>
            <input
              type="text"
              className="table-filter"
              placeholder="Filter ports..."
              value={portFilter}
              onChange={(e) => setPortFilter(e.target.value)}
            />
          </div>
          <div className="table-scroll">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Port</th>
                  <th>Protocol</th>
                  <th>State</th>
                  <th>Process</th>
                  <th>PID</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredPorts.map((port, idx) => (
                  <tr key={`${port.port}-${port.protocol}-${idx}`}>
                    <td className="text-mono">{port.port}</td>
                    <td>{port.protocol}</td>
                    <td>
                      <span className={`badge ${
                        port.state === 'LISTENING' ? 'badge-listening' :
                        port.state === 'ESTABLISHED' ? 'badge-established' :
                        'badge-other'
                      }`}>
                        {port.state}
                      </span>
                    </td>
                    <td>{port.processName || 'Unknown'}</td>
                    <td className="text-mono">{port.pid}</td>
                    <td className="actions-cell">
                      <button 
                        className="btn-explain" 
                        title="🧩 Explain This Port"
                        onClick={(e) => { e.stopPropagation(); explainPort(port); }}
                      >
                        🧩
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        <div className="table-container" style={{ gridColumn: '1 / -1' }}>
          <div className="table-header">
            <span className="table-title">Network Usage</span>
            <button 
              className="btn-explain" 
              title="🧩 Explain Network Activity"
              onClick={explainNetworkActivity}
            >
              🧩 Explain Network Activity
            </button>
          </div>
          <div className="table-scroll">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Process</th>
                  <th>PID</th>
                  <th>Download Speed</th>
                  <th>Upload Speed</th>
                  <th>Total Received</th>
                  <th>Total Sent</th>
                </tr>
              </thead>
              <tbody>
                {topNetworkUsage.map((usage) => (
                  <tr key={usage.pid}>
                    <td>{usage.processName || 'Unknown'}</td>
                    <td className="text-mono">{usage.pid}</td>
                    <td className="number text-success">{formatSpeed(usage.downloadSpeed)}</td>
                    <td className="number text-warning">{formatSpeed(usage.uploadSpeed)}</td>
                    <td className="number">{formatBytes(usage.bytesRecv)}</td>
                    <td className="number">{formatBytes(usage.bytesSent)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Explanation Modal */}
      {explainModal.show && (
        <div className="modal-overlay" onClick={() => setExplainModal({ ...explainModal, show: false })}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>{explainModal.title}</h3>
              <button 
                className="modal-close" 
                onClick={() => setExplainModal({ ...explainModal, show: false })}
              >
                ✕
              </button>
            </div>
            <div className="modal-body">
              {explainModal.loading ? (
                <div className="loading-spinner">
                  <div className="spinner"></div>
                  <p>AI is analyzing and generating explanation...</p>
                </div>
              ) : (
                <div className="explanation-content">
                  {explainModal.content.split('\n').map((line, idx) => (
                    <p key={idx}>{line}</p>
                  ))}
                </div>
              )}
            </div>
            <div className="modal-footer">
              <button 
                className="btn-primary" 
                onClick={() => setExplainModal({ ...explainModal, show: false })}
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default Dashboard;
