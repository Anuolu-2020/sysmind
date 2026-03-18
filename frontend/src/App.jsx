import { useState, useEffect, useCallback } from 'react';
import { ThemeProvider } from './contexts/ThemeContext';
import Dashboard from './components/Dashboard';
import Chat from './components/Chat';
import Settings from './components/Settings';
import Security from './components/Security';
import Alerts from './components/Alerts';
import AutoInsights from './components/AutoInsights';

function AppContent() {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [systemStats, setSystemStats] = useState({ cpu: 0, memory: 0 });
  const [quickStats, setQuickStats] = useState({ cpu: 0, memory: 0, disk: 0, alerts: 0 });
  const [isConfigured, setIsConfigured] = useState(false);
  const [showShortcuts, setShowShortcuts] = useState(false);

  const playAlertBeep = useCallback((severity) => {
    try {
      const Ctx = window.AudioContext || window.webkitAudioContext;
      if (!Ctx) return;
      const audioContext = new Ctx();
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
      console.error('Alert sound error:', err);
    }
  }, []);

  const showDesktopAlert = useCallback(async (title, message, severity) => {
    if (!('Notification' in window)) return;
    let permission = Notification.permission;
    if (permission !== 'granted') {
      permission = await Notification.requestPermission();
    }
    if (permission !== 'granted') return;

    const icon = severity === 'critical' ? '🚨' : severity === 'warning' ? '⚠️' : 'ℹ️';
    const notification = new Notification(`${icon} ${title}`, {
      body: message,
      tag: `sysmind-${title}`,
      requireInteraction: severity === 'critical',
    });

    if (severity !== 'critical') {
      setTimeout(() => notification.close(), 5000);
    }
  }, []);

  const checkConfig = useCallback(async () => {
    try {
      if (window.go?.main?.App?.IsAIConfigured) {
        const configured = await window.go.main.App.IsAIConfigured();
        setIsConfigured(configured);
      }
    } catch (err) {
      console.error('Error checking config:', err);
    }
  }, []);

  const fetchStats = useCallback(async () => {
    try {
      if (window.go?.main?.App?.GetSystemStats) {
        const stats = await window.go.main.App.GetSystemStats();
        setSystemStats(stats);
      }
      if (window.go?.main?.App?.GetQuickStats) {
        const quick = await window.go.main.App.GetQuickStats();
        setQuickStats(quick);
      }
    } catch (err) {
      console.error('Error fetching stats:', err);
    }
  }, []);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e) => {
      // Ignore if typing in input
      if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;

      if (e.key === '?' && e.shiftKey) {
        e.preventDefault();
        setShowShortcuts(s => !s);
      } else if (e.key === '1' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setActiveTab('dashboard');
      } else if (e.key === '2' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setActiveTab('chat');
      } else if (e.key === '3' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setActiveTab('security');
      } else if (e.key === '4' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setActiveTab('insights');
      } else if (e.key === '5' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setActiveTab('alerts');
      } else if (e.key === '6' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setActiveTab('settings');
      } else if (e.key === 'Escape') {
        setShowShortcuts(false);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  useEffect(() => {
    checkConfig();
    fetchStats();
    const interval = setInterval(fetchStats, 3000);
    return () => clearInterval(interval);
  }, [checkConfig, fetchStats]);

  // Global alert listeners so notifications work on every tab
  useEffect(() => {
    if (!window.runtime?.EventsOn) return undefined;

    const onNewAlert = async (payload) => {
      const alert = payload?.alert;
      if (!alert || alert.dismissed) return;

      const notifiedKey = `notified-${alert.id}`;
      if (sessionStorage.getItem(notifiedKey)) return;

      if (payload?.enableDesktopNotf) {
        await showDesktopAlert(alert.title, alert.message, alert.severity);
      }
      if (payload?.enableSound) {
        playAlertBeep(alert.severity);
      }

      sessionStorage.setItem(notifiedKey, 'true');
    };

    const onAlertNotify = async (payload) => {
      if (payload?.title && payload?.message) {
        await showDesktopAlert(payload.title, payload.message, 'info');
      }
    };

    window.runtime.EventsOn('alerts:new', onNewAlert);
    window.runtime.EventsOn('alerts:notify', onAlertNotify);

    return () => {
      if (window.runtime?.EventsOff) {
        window.runtime.EventsOff('alerts:new', onNewAlert);
        window.runtime.EventsOff('alerts:notify', onAlertNotify);
      }
    };
  }, [playAlertBeep, showDesktopAlert]);

  return (
    <div className="app">
      <header className="header">
        <h1>SysMind</h1>
        <div className="header-stats">
          <div className="stat-item">
            <span className="stat-label">CPU:</span>
            <span className={`stat-value ${systemStats.cpu > 80 ? 'text-danger' : ''}`}>
              {systemStats.cpu?.toFixed(1)}%
            </span>
          </div>
          <div className="stat-item">
            <span className="stat-label">Memory:</span>
            <span className={`stat-value ${systemStats.memory > 85 ? 'text-danger' : ''}`}>
              {systemStats.memory?.toFixed(1)}%
            </span>
          </div>
          <div className="stat-item">
            <span className="stat-label">Disk:</span>
            <span className={`stat-value ${quickStats.disk > 90 ? 'text-danger' : quickStats.disk > 75 ? 'text-warning' : ''}`}>
              {quickStats.disk?.toFixed(1)}%
            </span>
          </div>
          {quickStats.alerts > 0 && (
            <div className="stat-item alert-indicator" onClick={() => setActiveTab('alerts')} title="View Alerts">
              <span className="alert-badge">{quickStats.alerts}</span>
              <span className="stat-label">Alert{quickStats.alerts > 1 ? 's' : ''}</span>
            </div>
          )}
          <div className="stat-item">
            <span className={`badge ${isConfigured ? 'badge-listening' : 'badge-other'}`}>
              {isConfigured ? 'AI Ready' : 'AI Not Configured'}
            </span>
          </div>
          <button 
            className="shortcuts-btn" 
            onClick={() => setShowShortcuts(true)}
            title="Keyboard Shortcuts (?)"
          >
            ?
          </button>
        </div>
      </header>

      <nav className="nav">
        <button
          className={`nav-btn ${activeTab === 'dashboard' ? 'active' : ''}`}
          onClick={() => setActiveTab('dashboard')}
        >
          Dashboard <span className="shortcut-hint">1</span>
        </button>
        <button
          className={`nav-btn ${activeTab === 'chat' ? 'active' : ''}`}
          onClick={() => setActiveTab('chat')}
        >
          AI Chat <span className="shortcut-hint">2</span>
        </button>
        <button
          className={`nav-btn ${activeTab === 'security' ? 'active' : ''}`}
          onClick={() => setActiveTab('security')}
        >
          Security <span className="shortcut-hint">3</span>
        </button>
        <button
          className={`nav-btn ${activeTab === 'insights' ? 'active' : ''}`}
          onClick={() => setActiveTab('insights')}
        >
          Auto Insights <span className="shortcut-hint">4</span>
        </button>
        <button
          className={`nav-btn ${activeTab === 'alerts' ? 'active' : ''}`}
          onClick={() => setActiveTab('alerts')}
        >
          Alerts <span className="shortcut-hint">5</span>
        </button>
        <button
          className={`nav-btn ${activeTab === 'settings' ? 'active' : ''}`}
          onClick={() => setActiveTab('settings')}
        >
          Settings <span className="shortcut-hint">6</span>
        </button>
      </nav>

      <main className="main-content">
        {activeTab === 'dashboard' && <Dashboard />}
        <div style={{ display: activeTab === 'chat' ? 'block' : 'none' }}>
          <Chat />
        </div>
        {activeTab === 'security' && <Security />}
        {activeTab === 'insights' && <AutoInsights />}
        {activeTab === 'alerts' && <Alerts />}
        {activeTab === 'settings' && <Settings onConfigChange={checkConfig} />}
      </main>

      {/* Keyboard Shortcuts Modal */}
      {showShortcuts && (
        <>
          <div className="modal-backdrop" onClick={() => setShowShortcuts(false)} />
          <div className="shortcuts-modal">
            <h2>Keyboard Shortcuts</h2>
            <div className="shortcut-item">
              <span>Dashboard</span>
              <span className="shortcut-key"><kbd>Ctrl</kbd><kbd>1</kbd></span>
            </div>
            <div className="shortcut-item">
              <span>AI Chat</span>
              <span className="shortcut-key"><kbd>Ctrl</kbd><kbd>2</kbd></span>
            </div>
            <div className="shortcut-item">
              <span>Security</span>
              <span className="shortcut-key"><kbd>Ctrl</kbd><kbd>3</kbd></span>
            </div>
            <div className="shortcut-item">
              <span>Auto Insights</span>
              <span className="shortcut-key"><kbd>Ctrl</kbd><kbd>4</kbd></span>
            </div>
            <div className="shortcut-item">
              <span>Alerts</span>
              <span className="shortcut-key"><kbd>Ctrl</kbd><kbd>5</kbd></span>
            </div>
            <div className="shortcut-item">
              <span>Settings</span>
              <span className="shortcut-key"><kbd>Ctrl</kbd><kbd>6</kbd></span>
            </div>
            <div className="shortcut-item">
              <span>Show Shortcuts</span>
              <span className="shortcut-key"><kbd>Shift</kbd><kbd>?</kbd></span>
            </div>
            <div className="shortcut-item">
              <span>Close Modal</span>
              <span className="shortcut-key"><kbd>Esc</kbd></span>
            </div>
            <button 
              className="refresh-btn" 
              style={{ marginTop: 16 }}
              onClick={() => setShowShortcuts(false)}
            >
              Close
            </button>
          </div>
        </>
      )}
    </div>
  );
}

function App() {
  return (
    <ThemeProvider>
      <AppContent />
    </ThemeProvider>
  );
}

export default App;
