import { useState, useEffect, useCallback, useRef } from 'react';

function SystemStats() {
  const [stats, setStats] = useState(null);
  const [history, setHistory] = useState({
    cpu: [],
    memory: [],
    disk: [],
    netUp: [],
    netDown: [],
  });
  const maxHistoryPoints = 60; // 60 seconds of history

  const fetchStats = useCallback(async () => {
    try {
      if (window.go?.main?.App?.GetDetailedStats) {
        const data = await window.go.main.App.GetDetailedStats();
        setStats(data);
        
        setHistory(prev => ({
          cpu: [...prev.cpu.slice(-maxHistoryPoints + 1), data.cpuPercent || 0],
          memory: [...prev.memory.slice(-maxHistoryPoints + 1), data.memoryPercent || 0],
          disk: [...prev.disk.slice(-maxHistoryPoints + 1), data.diskPercent || 0],
          netUp: [...prev.netUp.slice(-maxHistoryPoints + 1), data.netUploadSpeed || 0],
          netDown: [...prev.netDown.slice(-maxHistoryPoints + 1), data.netDownSpeed || 0],
        }));
      }
    } catch (err) {
      console.error('Error fetching stats:', err);
    }
  }, []);

  useEffect(() => {
    fetchStats();
    const interval = setInterval(fetchStats, 1000);
    return () => clearInterval(interval);
  }, [fetchStats]);

  const formatBytes = (bytes) => {
    if (bytes == null || isNaN(bytes)) return '0 B/s';
    if (bytes < 1024) return `${bytes.toFixed(0)} B/s`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB/s`;
    return `${(bytes / 1024 / 1024).toFixed(1)} MB/s`;
  };

  const formatUptime = (seconds) => {
    if (seconds == null || isNaN(seconds)) return '0m';
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    if (days > 0) return `${days}d ${hours}h`;
    if (hours > 0) return `${hours}h ${mins}m`;
    return `${mins}m`;
  };

  if (!stats) {
    return (
      <div className="system-stats loading">
        <div className="loading-spinner" />
        <span>Loading system stats...</span>
      </div>
    );
  }

  return (
    <div className="system-stats">
      <div className="stats-grid">
        {/* CPU Card */}
        <div className="stat-card">
          <div className="stat-header">
            <span className="stat-icon">CPU</span>
            <span className={`stat-value ${(stats.cpuPercent || 0) > 80 ? 'critical' : (stats.cpuPercent || 0) > 60 ? 'warning' : ''}`}>
              {(stats.cpuPercent || 0).toFixed(1)}%
            </span>
          </div>
          <MiniGraph data={history.cpu} max={100} color="var(--accent-primary)" />
          {stats.cpuPerCore && (
            <div className="cpu-cores">
              {stats.cpuPerCore.map((core, i) => (
                <div key={i} className="core-bar">
                  <div 
                    className="core-fill" 
                    style={{ 
                      width: `${core}%`,
                      backgroundColor: core > 80 ? 'var(--accent-danger)' : 'var(--accent-primary)'
                    }} 
                  />
                </div>
              ))}
            </div>
          )}
          <div className="stat-details">
            <span>Load: {(stats.loadAvg1 || 0).toFixed(2)} / {(stats.loadAvg5 || 0).toFixed(2)} / {(stats.loadAvg15 || 0).toFixed(2)}</span>
          </div>
        </div>

        {/* Memory Card */}
        <div className="stat-card">
          <div className="stat-header">
            <span className="stat-icon">RAM</span>
            <span className={`stat-value ${(stats.memoryPercent || 0) > 85 ? 'critical' : (stats.memoryPercent || 0) > 70 ? 'warning' : ''}`}>
              {(stats.memoryPercent || 0).toFixed(1)}%
            </span>
          </div>
          <MiniGraph data={history.memory} max={100} color="var(--accent-success)" />
          <div className="stat-details">
            <span>{(stats.memoryUsedGB || 0).toFixed(1)} / {(stats.memoryTotalGB || 0).toFixed(1)} GB</span>
            {(stats.swapTotalGB || 0) > 0 && (
              <span className="swap-info">Swap: {(stats.swapUsedGB || 0).toFixed(1)} / {(stats.swapTotalGB || 0).toFixed(1)} GB</span>
            )}
          </div>
        </div>

        {/* Disk Card */}
        <div className="stat-card">
          <div className="stat-header">
            <span className="stat-icon">Disk</span>
            <span className={`stat-value ${(stats.diskPercent || 0) > 90 ? 'critical' : (stats.diskPercent || 0) > 75 ? 'warning' : ''}`}>
              {(stats.diskPercent || 0).toFixed(1)}%
            </span>
          </div>
          <div className="disk-bar">
            <div 
              className="disk-fill" 
              style={{ 
                width: `${stats.diskPercent || 0}%`,
                backgroundColor: (stats.diskPercent || 0) > 90 ? 'var(--accent-danger)' : 'var(--accent-warning)'
              }} 
            />
          </div>
          <div className="stat-details">
            <span>{(stats.diskUsedGB || 0).toFixed(0)} / {(stats.diskTotalGB || 0).toFixed(0)} GB</span>
            <span className="io-stats">
              R: {formatBytes(stats.diskReadSpeed || 0)} | W: {formatBytes(stats.diskWriteSpeed || 0)}
            </span>
          </div>
        </div>

        {/* Network Card */}
        <div className="stat-card">
          <div className="stat-header">
            <span className="stat-icon">Net</span>
            <span className="stat-value net-speed">
              <span className="download">{formatBytes(stats.netDownSpeed || 0)}</span>
              <span className="upload">{formatBytes(stats.netUploadSpeed || 0)}</span>
            </span>
          </div>
          <div className="dual-graph">
            <MiniGraph data={history.netDown} color="var(--accent-success)" label="Down" />
            <MiniGraph data={history.netUp} color="var(--accent-warning)" label="Up" />
          </div>
          <div className="stat-details">
            <span>Uptime: {formatUptime(stats.uptime || 0)}</span>
          </div>
        </div>
      </div>
    </div>
  );
}

function MiniGraph({ data, max, color, label }) {
  const canvasRef = useRef(null);

  useEffect(() => {
    try {
      const canvas = canvasRef.current;
      if (!canvas || !data || data.length < 2) return;

      const ctx = canvas.getContext('2d');
      if (!ctx) return;

      const width = canvas.width;
      const height = canvas.height;

      ctx.clearRect(0, 0, width, height);

      // Calculate auto-max for network graphs
      const actualMax = max || Math.max(...data.filter(v => v != null && !isNaN(v)), 1);

      // Draw gradient fill
      const gradient = ctx.createLinearGradient(0, 0, 0, height);
      gradient.addColorStop(0, color + '40');
      gradient.addColorStop(1, color + '00');

      ctx.beginPath();
      ctx.moveTo(0, height);

      data.forEach((value, i) => {
        const val = value || 0;
        const x = (i / (data.length - 1)) * width;
        const y = height - (val / actualMax) * height;
        ctx.lineTo(x, y);
      });

      ctx.lineTo(width, height);
      ctx.closePath();
      ctx.fillStyle = gradient;
      ctx.fill();

      // Draw line
      ctx.beginPath();
      data.forEach((value, i) => {
        const val = value || 0;
        const x = (i / (data.length - 1)) * width;
        const y = height - (val / actualMax) * height;
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
      });

      ctx.strokeStyle = color;
      ctx.lineWidth = 2;
      ctx.stroke();
    } catch (err) {
      console.error('MiniGraph render error:', err);
    }
  }, [data, max, color]);

  return (
    <div className="mini-graph">
      {label && <span className="graph-label">{label}</span>}
      <canvas ref={canvasRef} width={200} height={40} />
    </div>
  );
}

export default SystemStats;
