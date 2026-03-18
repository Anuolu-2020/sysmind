import { useState, useEffect, useCallback } from 'react';

function Alerts() {
  const [alerts, setAlerts] = useState([]);
  const [config, setConfig] = useState({
    cpuThreshold: 80,
    memoryThreshold: 85,
    diskThreshold: 90,
    enableAlerts: true,
    enableDesktopNotf: true,
    enableSound: false,
  });
  const [notificationPermission, setNotificationPermission] = useState('default');

  const requestNotificationPermission = useCallback(async () => {
    if (!('Notification' in window)) {
      return 'unsupported';
    }

    try {
      if (Notification.permission === 'granted') {
        setNotificationPermission('granted');
        return 'granted';
      }
      const permission = await Notification.requestPermission();
      setNotificationPermission(permission);
      return permission;
    } catch (err) {
      console.error('Error requesting notification permission:', err);
      return 'denied';
    }
  }, []);

  const fetchAlerts = useCallback(async () => {
    try {
      if (window.go?.main?.App?.GetAlerts) {
        const alertsList = await window.go.main.App.GetAlerts();
        setAlerts(alertsList || []);
        
        // Show desktop notifications for new undismissed alerts
        if (alertsList && config.enableDesktopNotf && notificationPermission === 'granted') {
          alertsList.forEach(alert => {
            if (!alert.dismissed && !sessionStorage.getItem(`notified-${alert.id}`)) {
              showNotification(alert.title, alert.message, alert.severity);
              sessionStorage.setItem(`notified-${alert.id}`, 'true');
            }
          });
        }
      }
    } catch (err) {
      console.error('Error fetching alerts:', err);
    }
  }, [config.enableDesktopNotf, notificationPermission]);

  const fetchConfig = useCallback(async () => {
    try {
      if (window.go?.main?.App?.GetAlertConfig) {
        const cfg = await window.go.main.App.GetAlertConfig();
        if (cfg) {
          setConfig(cfg);
        }
      }
    } catch (err) {
      console.error('Error fetching alert config:', err);
    }
  }, []);

  useEffect(() => {
    if ('Notification' in window) {
      setNotificationPermission(Notification.permission);
    }

    fetchConfig();
    fetchAlerts();
    const interval = setInterval(fetchAlerts, 5000);
    return () => clearInterval(interval);
  }, [fetchAlerts, fetchConfig]);

  // Listen for alert events from backend so notifications work across tabs
  useEffect(() => {
    if (!window.runtime?.EventsOn) return undefined;

    const handleAlertEvent = async (payload) => {
      const alert = payload?.alert;
      if (!alert || alert.dismissed) return;

      const desktopEnabled = payload?.enableDesktopNotf ?? config.enableDesktopNotf;
      if (!desktopEnabled) return;

      const notifiedKey = `notified-${alert.id}`;
      if (sessionStorage.getItem(notifiedKey)) return;

      let permission = notificationPermission;
      if (permission !== 'granted') {
        permission = await requestNotificationPermission();
      }

      if (permission === 'granted') {
        showNotification(alert.title, alert.message, alert.severity);
        sessionStorage.setItem(notifiedKey, 'true');
      }

      if (payload?.enableSound) {
        playAlertSound(alert.severity);
      }
    };

    const handleNotifyEvent = async (payload) => {
      if (!config.enableDesktopNotf) return;
      let permission = notificationPermission;
      if (permission !== 'granted') {
        permission = await requestNotificationPermission();
      }
      if (permission === 'granted' && payload?.title && payload?.message) {
        showNotification(payload.title, payload.message, 'info');
      }
    };

    window.runtime.EventsOn('alerts:new', handleAlertEvent);
    window.runtime.EventsOn('alerts:notify', handleNotifyEvent);

    return () => {
      if (window.runtime?.EventsOff) {
        window.runtime.EventsOff('alerts:new', handleAlertEvent);
        window.runtime.EventsOff('alerts:notify', handleNotifyEvent);
      }
    };
  }, [config.enableDesktopNotf, notificationPermission, requestNotificationPermission]);

  const showNotification = (title, body, severity) => {
    if ('Notification' in window && Notification.permission === 'granted') {
      try {
        const icon = severity === 'critical' ? '🚨' : severity === 'warning' ? '⚠️' : 'ℹ️';
        const notification = new Notification(`${icon} ${title}`, {
          body,
          icon: '/icon.png',
          badge: '/icon.png',
          tag: title,
          requireInteraction: severity === 'critical',
        });
        
        // Auto-close after 5 seconds for non-critical alerts
        if (severity !== 'critical') {
          setTimeout(() => notification.close(), 5000);
        }
      } catch (err) {
        console.error('Error creating notification:', err);
      }
    } else if ('Notification' in window && Notification.permission === 'denied') {
      console.warn('Notifications are denied. Please enable them in browser settings.');
    } else if ('Notification' in window && Notification.permission === 'default') {
      console.warn('Notification permission not granted yet');
    } else {
      console.warn('Notification API not available');
    }
  };

  const playAlertSound = (severity) => {
    if (!config.enableSound) return;
    try {
      const audioContext = new (window.AudioContext || window.webkitAudioContext)();
      const oscillator = audioContext.createOscillator();
      const gainNode = audioContext.createGain();

      oscillator.type = severity === 'critical' ? 'square' : 'sine';
      oscillator.frequency.value = severity === 'critical' ? 880 : 660;
      gainNode.gain.value = 0.02;

      oscillator.connect(gainNode);
      gainNode.connect(audioContext.destination);

      oscillator.start();
      setTimeout(() => {
        oscillator.stop();
        audioContext.close();
      }, severity === 'critical' ? 250 : 140);
    } catch (err) {
      console.error('Error playing alert sound:', err);
    }
  };

  const dismissAlert = async (alertId) => {
    try {
      if (window.go?.main?.App?.DismissAlert) {
        await window.go.main.App.DismissAlert(alertId);
        await fetchAlerts();
      }
    } catch (err) {
      console.error('Error dismissing alert:', err);
    }
  };

  const clearDismissed = async () => {
    try {
      if (window.go?.main?.App?.ClearAlerts) {
        await window.go.main.App.ClearAlerts();
        await fetchAlerts();
      }
    } catch (err) {
      console.error('Error clearing alerts:', err);
    }
  };

  const saveConfig = async () => {
    try {
      if (window.go?.main?.App?.SetAlertConfig) {
        await window.go.main.App.SetAlertConfig(config);
      }
    } catch (err) {
      console.error('Error saving config:', err);
    }
  };

  const activeAlerts = alerts.filter(a => !a.dismissed);
  const dismissedAlerts = alerts.filter(a => a.dismissed);

  return (
    <div className="panel alerts-panel">
      <h2>System Alerts</h2>

      {/* Notification Permission */}
      {notificationPermission !== 'granted' && (
        <div className="notification-banner">
          <span>🔔 Desktop notifications are disabled</span>
          <button 
            onClick={() => {
              Notification.requestPermission().then(permission => {
                setNotificationPermission(permission);
              });
            }}
          >
            Enable Notifications
          </button>
        </div>
      )}

      {/* Alert Configuration */}
      <div className="alert-config-section">
        <h3>Alert Thresholds</h3>
        <div className="form-row">
          <div className="form-group">
            <label>
              <input 
                type="checkbox" 
                checked={config.enableAlerts}
                onChange={(e) => setConfig({...config, enableAlerts: e.target.checked})}
              />
              Enable Alerts
            </label>
          </div>
          <div className="form-group">
            <label>
              <input 
                type="checkbox" 
                checked={config.enableDesktopNotf}
                onChange={(e) => setConfig({...config, enableDesktopNotf: e.target.checked})}
                disabled={!config.enableAlerts}
              />
              Desktop Notifications
            </label>
            {config.enableDesktopNotf && (
              <button 
                className="test-notification-btn" 
                onClick={() => showNotification('Test Alert', 'This is a test notification from SysMind', 'info')}
                style={{ marginLeft: '10px', padding: '4px 8px', fontSize: '12px' }}
              >
                Test
              </button>
            )}
          </div>
          <div className="form-group">
            <label>
              <input
                type="checkbox"
                checked={config.enableSound}
                onChange={(e) => setConfig({...config, enableSound: e.target.checked})}
                disabled={!config.enableAlerts}
              />
              Sound Notifications
            </label>
            {config.enableSound && (
              <button
                className="test-notification-btn"
                onClick={() => playAlertSound('warning')}
                style={{ marginLeft: '10px', padding: '4px 8px', fontSize: '12px' }}
              >
                Test Sound
              </button>
            )}
          </div>
        </div>

        <div className="threshold-inputs">
          <div className="form-group">
            <label>CPU Threshold (%)</label>
            <input 
              type="number" 
              value={config.cpuThreshold}
              onChange={(e) => setConfig({...config, cpuThreshold: parseFloat(e.target.value)})}
              min="0"
              max="100"
              step="5"
            />
          </div>
          <div className="form-group">
            <label>Memory Threshold (%)</label>
            <input 
              type="number" 
              value={config.memoryThreshold}
              onChange={(e) => setConfig({...config, memoryThreshold: parseFloat(e.target.value)})}
              min="0"
              max="100"
              step="5"
            />
          </div>
          <div className="form-group">
            <label>Disk Threshold (%)</label>
            <input 
              type="number" 
              value={config.diskThreshold}
              onChange={(e) => setConfig({...config, diskThreshold: parseFloat(e.target.value)})}
              min="0"
              max="100"
              step="5"
            />
          </div>
        </div>

        <button className="save-btn" onClick={saveConfig}>
          Save Configuration
        </button>
      </div>

      {/* Active Alerts */}
      {activeAlerts.length > 0 && (
        <div className="alerts-section">
          <h3>Active Alerts ({activeAlerts.length})</h3>
          <div className="alerts-list">
            {activeAlerts.map(alert => (
              <div key={alert.id} className={`alert-item alert-${alert.severity}`}>
                <div className="alert-icon">
                  {alert.severity === 'critical' ? '🚨' : alert.severity === 'warning' ? '⚠️' : 'ℹ️'}
                </div>
                <div className="alert-content">
                  <div className="alert-header">
                    <span className="alert-title">{alert.title}</span>
                    <span className="alert-time">
                      {new Date(alert.timestamp).toLocaleTimeString()}
                    </span>
                  </div>
                  <div className="alert-message">{alert.message}</div>
                </div>
                <button 
                  className="alert-dismiss"
                  onClick={() => dismissAlert(alert.id)}
                  title="Dismiss"
                >
                  ×
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* No Active Alerts */}
      {activeAlerts.length === 0 && (
        <div className="no-alerts">
          <span className="no-alerts-icon">✅</span>
          <p>No active alerts. Your system is running smoothly!</p>
        </div>
      )}

      {/* Dismissed Alerts */}
      {dismissedAlerts.length > 0 && (
        <div className="dismissed-section">
          <div className="dismissed-header">
            <h3>Dismissed Alerts ({dismissedAlerts.length})</h3>
            <button className="clear-btn" onClick={clearDismissed}>
              Clear All
            </button>
          </div>
          <div className="alerts-list dismissed">
            {dismissedAlerts.slice(0, 10).map(alert => (
              <div key={alert.id} className="alert-item dismissed">
                <div className="alert-icon">
                  {alert.severity === 'critical' ? '🚨' : alert.severity === 'warning' ? '⚠️' : 'ℹ️'}
                </div>
                <div className="alert-content">
                  <div className="alert-header">
                    <span className="alert-title">{alert.title}</span>
                    <span className="alert-time">
                      {new Date(alert.timestamp).toLocaleTimeString()}
                    </span>
                  </div>
                  <div className="alert-message">{alert.message}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

export default Alerts;
