import { useState, useEffect, useCallback } from 'react';
import ConnectionMap from './ConnectionMap';

function Security() {
  const [securityInfo, setSecurityInfo] = useState(null);
  const [geoSecurityInfo, setGeoSecurityInfo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [loadingGeo, setLoadingGeo] = useState(false);
  const [explainModal, setExplainModal] = useState({ show: false, loading: false, content: '', title: '', type: '' });

  const fetchSecurityInfo = useCallback(async () => {
    try {
      if (window.go?.main?.App?.GetSecurityInfo) {
        // First get basic security info quickly
        const info = await window.go.main.App.GetSecurityInfo();
        setSecurityInfo(info);
        setLoading(false);
        
        // Then get geo-enhanced info if there are external connections
        if (info.unknownConns && info.unknownConns.length > 0) {
          setLoadingGeo(true);
          try {
            const geoInfo = await window.go.main.App.GetSecurityInfoWithGeo();
            setGeoSecurityInfo(geoInfo);
          } catch (err) {
            console.error('Error fetching geo security info:', err);
            // Fallback to basic info for geo display too
            setGeoSecurityInfo(info);
          } finally {
            setLoadingGeo(false);
          }
        } else {
          setGeoSecurityInfo(info);
        }
      }
    } catch (err) {
      console.error('Error fetching security info:', err);
      setLoading(false);
    }
  }, []);

  // Explain This functionality
  const explainSuspiciousProcess = async (process) => {
    setExplainModal({ 
      show: true, 
      loading: true, 
      content: '', 
      title: `Explain Suspicious Process: ${process.name}`,
      type: 'suspicious-process'
    });

    try {
      const question = `Please analyze this suspicious process in detail:

Process Name: ${process.name}
PID: ${process.pid}
Risk Level: ${process.riskLevel}
Detected Issues: ${process.reasons?.join(', ') || 'None specified'}

Please provide:
1. What this process is and its normal purpose (if legitimate)
2. Analysis of why it was flagged as suspicious
3. Whether the detected risk level is accurate
4. Recommended actions (kill, monitor, investigate further)
5. How to determine if this is a false positive
6. Security implications if this is malicious

Be thorough and provide actionable security guidance.`;

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

  const explainNetworkConnections = async () => {
    setExplainModal({ 
      show: true, 
      loading: true, 
      content: '', 
      title: 'Explain Network Security Status',
      type: 'network-security'
    });

    try {
      const question = `Please analyze the current network security status:

Open Ports: ${securityInfo.openPorts || 0}
Listening Ports: ${securityInfo.listeningPorts || 0}  
External Connections: ${securityInfo.externalConns || 0}
Firewall Status: ${securityInfo.firewallEnabled ? 'Enabled' : 'Disabled'}

Please provide:
1. Analysis of whether these connection numbers are normal
2. Security risks associated with current open ports
3. Impact of firewall status on overall security
4. Recommendations for improving network security posture  
5. What external connections might indicate
6. Best practices for port and connection management

Focus on practical security advice and risk assessment.`;

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
    fetchSecurityInfo();
    const interval = setInterval(fetchSecurityInfo, 10000); // Refresh every 10s
    return () => clearInterval(interval);
  }, [fetchSecurityInfo]);

  if (loading) {
    return (
      <div className="panel security-panel loading">
        <div className="loading-spinner" />
        <span>Scanning system security...</span>
      </div>
    );
  }

  if (!securityInfo) {
    return (
      <div className="panel security-panel">
        <h2>Security</h2>
        <p className="text-muted">Unable to fetch security information</p>
      </div>
    );
  }

  const hasIssues = securityInfo.suspiciousProcs?.length > 0;
  const overallStatus = hasIssues ? 'warning' : 'secure';

  return (
    <div className="panel security-panel">
      <div className="security-header">
        <h2>Security Overview</h2>
        <span className={`security-badge ${overallStatus}`}>
          {overallStatus === 'secure' ? 'System Secure' : 'Issues Detected'}
        </span>
      </div>

      <div className="security-grid">
        {/* Firewall Status */}
        <div className={`security-card ${securityInfo.firewallEnabled ? 'good' : 'warning'}`}>
          <div className="security-card-icon">
            {securityInfo.firewallEnabled ? '🛡️' : '⚠️'}
          </div>
          <div className="security-card-content">
            <h3>Firewall</h3>
            <p>{securityInfo.firewallStatus || 'Unknown'}</p>
          </div>
        </div>

        {/* Connections */}
        <div className="security-card neutral">
          <div className="security-card-icon">🌐</div>
          <div className="security-card-content">
            <h3>Connections</h3>
            <p>{securityInfo.openPorts || 0} open ports</p>
            <p className="sub">{securityInfo.listeningPorts || 0} listening, {securityInfo.externalConns || 0} external</p>
            <button 
              className="btn-explain" 
              title="🧩 Explain Network Security"
              onClick={explainNetworkConnections}
            >
              🧩 Explain
            </button>
          </div>
        </div>

        {/* Suspicious Processes */}
        <div className={`security-card ${hasIssues ? 'danger' : 'good'}`}>
          <div className="security-card-icon">
            {hasIssues ? '🚨' : '✅'}
          </div>
          <div className="security-card-content">
            <h3>Processes</h3>
            <p>
              {hasIssues 
                ? `${securityInfo.suspiciousProcs.length} suspicious found`
                : 'No suspicious activity'}
            </p>
          </div>
        </div>
      </div>

      {/* Suspicious Processes List */}
      {hasIssues && (
        <div className="suspicious-list">
          <h3>Suspicious Processes</h3>
          <div className="table-scroll">
            <table className="data-table">
              <thead>
                <tr>
                  <th>PID</th>
                  <th>Name</th>
                  <th>Risk</th>
                  <th>Reasons</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {securityInfo.suspiciousProcs.map((proc) => (
                  <tr key={proc.pid} className={`risk-${proc.riskLevel}`}>
                    <td className="text-mono">{proc.pid}</td>
                    <td>{proc.name}</td>
                    <td>
                      <span className={`badge badge-${proc.riskLevel}`}>
                        {proc.riskLevel.toUpperCase()}
                      </span>
                    </td>
                    <td>
                      <ul className="reasons-list">
                        {proc.reasons?.map((reason, i) => (
                          <li key={i}>{reason}</li>
                        ))}
                      </ul>
                    </td>
                    <td>
                      <button 
                        className="btn-explain" 
                        title="🧩 Explain This Process"
                        onClick={(e) => { e.stopPropagation(); explainSuspiciousProcess(proc); }}
                        style={{ marginRight: '8px' }}
                      >
                        🧩
                      </button>
                      <button 
                        className="btn-small btn-danger"
                        onClick={() => handleKillProcess(proc.pid)}
                      >
                        Kill
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Connection Map */}
      {geoSecurityInfo?.unknownConns && geoSecurityInfo.unknownConns.length > 0 && (
        <div>
          {loadingGeo && (
            <div className="geo-loading">
              <span>Loading connection locations...</span>
            </div>
          )}
          <ConnectionMap connections={geoSecurityInfo.unknownConns} />
        </div>
      )}

      <button className="refresh-btn" onClick={fetchSecurityInfo}>
        Refresh Scan
      </button>

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
                  <p>AI is analyzing security information...</p>
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

  async function handleKillProcess(pid) {
    if (!confirm(`Are you sure you want to terminate process ${pid}?`)) {
      return;
    }
    
    try {
      const result = await window.go.main.App.KillProcess(pid);
      if (result.success) {
        alert(result.message);
        fetchSecurityInfo();
      } else {
        alert('Error: ' + result.error);
      }
    } catch (err) {
      alert('Error: ' + err.message);
    }
  }
}

export default Security;
