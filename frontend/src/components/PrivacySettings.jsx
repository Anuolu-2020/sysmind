import { useState, useEffect, useCallback } from 'react';

function PrivacySettings() {
  const [privacyConfig, setPrivacyConfig] = useState({
    shareProcessNames: true,
    shareProcessDetails: true,
    shareNetworkPorts: true,
    shareConnectionIPs: true,
    shareConnectionGeo: true,
    shareSecurityInfo: true,
    shareSystemStats: true,
    anonymizeProcesses: false,
    anonymizeConnections: false,
  });
  const [isSaving, setIsSaving] = useState(false);
  const [saveStatus, setSaveStatus] = useState('');

  const loadPrivacyConfig = useCallback(async () => {
    try {
      if (window.go?.main?.App?.GetPrivacyConfig) {
        const config = await window.go.main.App.GetPrivacyConfig();
        if (config) {
          setPrivacyConfig(config);
        }
      }
    } catch (err) {
      console.error('Error loading privacy config:', err);
    }
  }, []);

  useEffect(() => {
    loadPrivacyConfig();
  }, [loadPrivacyConfig]);

  const handleToggle = (key) => {
    setPrivacyConfig((prev) => ({
      ...prev,
      [key]: !prev[key],
    }));
  };

  const handleSave = async () => {
    setIsSaving(true);
    setSaveStatus('');

    try {
      if (window.go?.main?.App?.SetPrivacyConfig) {
        const result = await window.go.main.App.SetPrivacyConfig(privacyConfig);
        if (result.success) {
          setSaveStatus('Privacy settings saved successfully!');
        } else {
          setSaveStatus('Error saving privacy settings: ' + result.error);
        }
      }
    } catch (err) {
      setSaveStatus('Error saving privacy settings: ' + err.message);
    } finally {
      setIsSaving(false);
      setTimeout(() => setSaveStatus(''), 3000);
    }
  };

  const handleReset = () => {
    setPrivacyConfig({
      shareProcessNames: true,
      shareProcessDetails: true,
      shareNetworkPorts: true,
      shareConnectionIPs: true,
      shareConnectionGeo: true,
      shareSecurityInfo: true,
      shareSystemStats: true,
      anonymizeProcesses: false,
      anonymizeConnections: false,
    });
    setSaveStatus('Reset to default settings');
    setTimeout(() => setSaveStatus(''), 3000);
  };

  const privacySettings = [
    {
      key: 'shareProcessNames',
      title: 'Share Process Names',
      description: 'Allow AI to see what programs are running (e.g., Firefox, VS Code)',
      category: 'Process Information',
    },
    {
      key: 'shareProcessDetails',
      title: 'Share Process Details',
      description: 'Allow AI to see CPU and memory usage per process',
      category: 'Process Information',
    },
    {
      key: 'shareNetworkPorts',
      title: 'Share Network Ports',
      description: 'Allow AI to see open ports and listening services',
      category: 'Network Information',
    },
    {
      key: 'shareConnectionIPs',
      title: 'Share Connection IPs',
      description: 'Allow AI to see remote IP addresses your system connects to',
      category: 'Network Information',
    },
    {
      key: 'shareConnectionGeo',
      title: 'Share Connection Locations',
      description: 'Allow AI to see geographic location of network connections',
      category: 'Network Information',
    },
    {
      key: 'shareSecurityInfo',
      title: 'Share Security Alerts',
      description: 'Allow AI to see security warnings and suspicious process information',
      category: 'Security Information',
    },
    {
      key: 'shareSystemStats',
      title: 'Share System Statistics',
      description: 'Allow AI to see overall CPU, memory, and disk usage percentages',
      category: 'System Statistics',
    },
    {
      key: 'anonymizeProcesses',
      title: 'Anonymize Process Names',
      description: 'Replace process names with categories (e.g., [Browser], [Dev Tool])',
      category: 'Anonymization',
    },
    {
      key: 'anonymizeConnections',
      title: 'Anonymize Connection IPs',
      description: 'Replace IP addresses with service categories (e.g., [Google Services], [AWS Cloud])',
      category: 'Anonymization',
    },
  ];

  const groupedSettings = {};
  privacySettings.forEach((setting) => {
    if (!groupedSettings[setting.category]) {
      groupedSettings[setting.category] = [];
    }
    groupedSettings[setting.category].push(setting);
  });

  return (
    <div className="settings-section">
      <h3>Privacy & Data Sharing</h3>
      
      <div className="privacy-intro">
        <p>
          Control what system information is shared with your AI provider. By default, all data is shared for better AI insights.
          Enable anonymization to replace specific information with categories while still getting helpful analysis.
        </p>
      </div>

      {Object.entries(groupedSettings).map(([category, settings]) => (
        <div key={category} className="privacy-category">
          <h4>{category}</h4>
          <div className="privacy-settings">
            {settings.map((setting) => (
              <div key={setting.key} className="privacy-toggle-group">
                <div className="privacy-toggle-content">
                  <label className="privacy-toggle-label">
                    <input
                      type="checkbox"
                      checked={privacyConfig[setting.key]}
                      onChange={() => handleToggle(setting.key)}
                      className="privacy-checkbox"
                    />
                    <span className="toggle-title">{setting.title}</span>
                  </label>
                  <p className="toggle-description">{setting.description}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      ))}

      <div className="privacy-actions">
        <button 
          className="save-btn" 
          onClick={handleSave} 
          disabled={isSaving}
          style={{ marginRight: '12px' }}
        >
          {isSaving ? 'Saving...' : 'Save Privacy Settings'}
        </button>
        
        <button 
          className="reset-btn" 
          onClick={handleReset}
          style={{
            padding: '10px 20px',
            border: '1px solid var(--border-color)',
            borderRadius: '6px',
            background: 'transparent',
            color: 'var(--text-secondary)',
            cursor: 'pointer',
            fontSize: '14px',
          }}
        >
          Reset to Default
        </button>
      </div>

      {saveStatus && (
        <div
          className={`status-indicator ${saveStatus.includes('Error') ? 'status-not-configured' : 'status-configured'}`}
          style={{ marginTop: '12px' }}
        >
          {saveStatus}
        </div>
      )}

      <div className="privacy-info-box" style={{ marginTop: '20px', padding: '12px', background: 'var(--bg-secondary)', borderRadius: '6px', fontSize: '12px' }}>
        <p style={{ margin: '0 0 8px 0', fontWeight: '600' }}>Data Privacy Assurance:</p>
        <ul style={{ margin: '0', paddingLeft: '20px', lineHeight: '1.6' }}>
          <li>Command line arguments are never shared</li>
          <li>Usernames are never shared</li>
          <li>File contents, passwords, and environment variables are never collected</li>
          <li>Your privacy preferences are stored locally on your device</li>
          <li>You can change these settings at any time</li>
        </ul>
      </div>
    </div>
  );
}

export default PrivacySettings;
