import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

function ResourceTimeline() {
  const [points, setPoints] = useState([]);
  const [minutes, setMinutes] = useState(30);
  const percentCanvasRef = useRef(null);
  const networkCanvasRef = useRef(null);

  const fetchTimeline = useCallback(async () => {
    try {
      if (window.go?.main?.App?.GetResourceTimeline) {
        const data = await window.go.main.App.GetResourceTimeline(minutes);
        setPoints(Array.isArray(data) ? data : []);
      }
    } catch (err) {
      console.error('Error fetching resource timeline:', err);
      setPoints([]);
    }
  }, [minutes]);

  useEffect(() => {
    fetchTimeline();
    const interval = setInterval(fetchTimeline, 5000);
    return () => clearInterval(interval);
  }, [fetchTimeline]);

  const netMax = useMemo(() => {
    const maxValue = points.reduce((max, p) => {
      const down = Number(p?.netDownSpeed || 0);
      const up = Number(p?.netUploadSpeed || 0);
      return Math.max(max, down, up);
    }, 0);
    return Math.max(maxValue, 1);
  }, [points]);

  const summary = useMemo(() => {
    if (!points.length) {
      return {
        cpuPeak: 0,
        memPeak: 0,
        diskPeak: 0,
        netPeak: 0,
      };
    }

    return {
      cpuPeak: Math.max(...points.map((p) => Number(p?.cpuPercent || 0))),
      memPeak: Math.max(...points.map((p) => Number(p?.memoryPercent || 0))),
      diskPeak: Math.max(...points.map((p) => Number(p?.diskPercent || 0))),
      netPeak: Math.max(
        ...points.map((p) => Math.max(Number(p?.netDownSpeed || 0), Number(p?.netUploadSpeed || 0)))
      ),
    };
  }, [points]);

  useEffect(() => {
    const canvas = percentCanvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const dpr = window.devicePixelRatio || 1;
    const rect = canvas.getBoundingClientRect();
    
    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;
    ctx.scale(dpr, dpr);
    
    const width = rect.width;
    const height = rect.height;
    
    ctx.clearRect(0, 0, width, height);
    drawBackground(ctx, width, height);
    drawGrid(ctx, width, height, 4, 4);

    if (points.length < 2) {
      drawNoData(ctx, width, height);
      return;
    }

    drawAreaAndLine(ctx, points, (p) => Number(p?.cpuPercent || 0), width, height, 100, 'rgba(122, 162, 247, 0.2)', '#7aa2f7');
    drawAreaAndLine(ctx, points, (p) => Number(p?.memoryPercent || 0), width, height, 100, 'rgba(158, 206, 106, 0.15)', '#9ece6a');
    drawAreaAndLine(ctx, points, (p) => Number(p?.diskPercent || 0), width, height, 100, 'rgba(224, 175, 104, 0.12)', '#e0af68');

    drawYLabels(ctx, width, height, ['100%', '75%', '50%', '25%', '0%']);
    drawTimeLabels(ctx, width, height, points, minutes);
  }, [points, minutes]);

  useEffect(() => {
    const canvas = networkCanvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const dpr = window.devicePixelRatio || 1;
    const rect = canvas.getBoundingClientRect();
    
    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;
    ctx.scale(dpr, dpr);
    
    const width = rect.width;
    const height = rect.height;
    
    ctx.clearRect(0, 0, width, height);
    drawBackground(ctx, width, height);
    drawGrid(ctx, width, height, 4, 4);

    if (points.length < 2) {
      drawNoData(ctx, width, height);
      return;
    }

    drawAreaAndLine(ctx, points, (p) => Number(p?.netDownSpeed || 0), width, height, netMax, 'rgba(158, 206, 106, 0.18)', '#9ece6a');
    drawAreaAndLine(ctx, points, (p) => Number(p?.netUploadSpeed || 0), width, height, netMax, 'rgba(247, 118, 142, 0.14)', '#f7768e');

    drawNetworkYLabels(ctx, width, height, netMax);
    drawTimeLabels(ctx, width, height, points, minutes);
  }, [points, netMax, minutes]);

  const latest = points.length ? points[points.length - 1] : null;

  return (
    <section className="resource-timeline-card">
      <header className="resource-timeline-header">
        <div>
          <h3>Resource Timeline</h3>
          <p className="resource-timeline-subtitle">System metrics and network activity</p>
        </div>
        <div className="timeline-controls">
          {[15, 30, 60].map((option) => (
            <button
              key={option}
              className={`timeline-range-btn ${minutes === option ? 'active' : ''}`}
              onClick={() => setMinutes(option)}
            >
              {option}m
            </button>
          ))}
        </div>
      </header>

      <div className="timeline-grid">
        <div className="timeline-chart-block">
          <div className="timeline-label-row">
            <span className="timeline-legend-item cpu">CPU</span>
            <span className="timeline-legend-item memory">Memory</span>
            <span className="timeline-legend-item disk">Disk</span>
            <span className="axis-note">Utilization (%)</span>
          </div>
          <canvas ref={percentCanvasRef} className="timeline-canvas" style={{width: '100%', height: '240px'}} />
        </div>

        <div className="timeline-chart-block">
          <div className="timeline-label-row">
            <span className="timeline-legend-item net-down">Download</span>
            <span className="timeline-legend-item net-up">Upload</span>
            <span className="axis-note">Network (bytes/s)</span>
          </div>
          <canvas ref={networkCanvasRef} className="timeline-canvas" style={{width: '100%', height: '220px'}} />
        </div>
      </div>

      <footer className="timeline-snapshot">
        <span>Samples: {points.length}</span>
        <span>CPU: {Number(latest?.cpuPercent || 0).toFixed(1)}%</span>
        <span>RAM: {Number(latest?.memoryPercent || 0).toFixed(1)}%</span>
        <span>Disk: {Number(latest?.diskPercent || 0).toFixed(1)}%</span>
        <span>Peak CPU: {summary.cpuPeak.toFixed(1)}%</span>
        <span>Peak RAM: {summary.memPeak.toFixed(1)}%</span>
        <span>Peak Net: {formatSpeed(summary.netPeak)}</span>
      </footer>
    </section>
  );
}

function drawBackground(ctx, width, height) {
  ctx.fillStyle = 'rgba(30, 32, 48, 0.95)';
  ctx.fillRect(0, 0, width, height);
}

function drawNoData(ctx, width, height) {
  ctx.fillStyle = '#565f89';
  ctx.font = '13px sans-serif';
  ctx.textAlign = 'center';
  ctx.fillText('Collecting timeline data...', width / 2, height / 2);
}

function drawGrid(ctx, width, height, rows, cols) {
  ctx.strokeStyle = 'rgba(86, 95, 137, 0.3)';
  ctx.lineWidth = 1;
  
  for (let i = 1; i < rows; i += 1) {
    const y = (i / rows) * height;
    ctx.beginPath();
    ctx.moveTo(0, y);
    ctx.lineTo(width, y);
    ctx.stroke();
  }
  
  for (let i = 1; i < cols; i += 1) {
    const x = (i / cols) * width;
    ctx.beginPath();
    ctx.moveTo(x, 0);
    ctx.lineTo(x, height);
    ctx.stroke();
  }
}

function drawAreaAndLine(ctx, points, getter, width, height, maxValue, fillColor, lineColor) {
  if (points.length < 2) return;
  const safeMax = Math.max(maxValue, 1);

  // Draw area
  ctx.beginPath();
  ctx.moveTo(0, height);
  points.forEach((point, index) => {
    const x = (index / (points.length - 1)) * width;
    const value = getter(point);
    const y = height - (Math.min(Math.max(value, 0), safeMax) / safeMax) * height;
    ctx.lineTo(x, y);
  });
  ctx.lineTo(width, height);
  ctx.closePath();
  ctx.fillStyle = fillColor;
  ctx.fill();

  // Draw line
  ctx.beginPath();
  points.forEach((point, index) => {
    const x = (index / (points.length - 1)) * width;
    const value = getter(point);
    const y = height - (Math.min(Math.max(value, 0), safeMax) / safeMax) * height;
    if (index === 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
  });
  ctx.strokeStyle = lineColor;
  ctx.lineWidth = 2;
  ctx.stroke();
}

function drawYLabels(ctx, width, height, labels) {
  ctx.fillStyle = 'rgba(169, 177, 214, 0.8)';
  ctx.font = '11px sans-serif';
  ctx.textAlign = 'right';
  labels.forEach((label, index) => {
    const y = (index / (labels.length - 1)) * height + 4;
    ctx.fillText(label, width - 8, y);
  });
}

function drawNetworkYLabels(ctx, width, height, maxValue) {
  const values = [maxValue, maxValue * 0.75, maxValue * 0.5, maxValue * 0.25, 0];
  ctx.fillStyle = 'rgba(169, 177, 214, 0.8)';
  ctx.font = '11px sans-serif';
  ctx.textAlign = 'right';
  values.forEach((value, index) => {
    const y = (index / (values.length - 1)) * height + 4;
    ctx.fillText(formatSpeed(value), width - 8, y);
  });
}

function drawTimeLabels(ctx, width, height, points, minutes) {
  if (!points.length) return;
  const startTs = points[0].timestamp;
  const endTs = points[points.length - 1].timestamp;
  const midTs = startTs + (endTs - startTs) / 2;

  ctx.fillStyle = 'rgba(169, 177, 214, 0.7)';
  ctx.font = '11px sans-serif';
  
  ctx.textAlign = 'left';
  ctx.fillText(formatTime(startTs), 8, height - 8);
  
  ctx.textAlign = 'center';
  ctx.fillText(formatTime(midTs), width / 2, height - 8);
  
  ctx.textAlign = 'right';
  ctx.fillText(`${formatTime(endTs)} (${minutes}m)`, width - 8, height - 8);
}

function formatTime(ts) {
  const date = new Date(ts);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function formatSpeed(value) {
  if (value < 1024) return `${value.toFixed(0)} B/s`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB/s`;
  if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB/s`;
  return `${(value / (1024 * 1024 * 1024)).toFixed(1)} GB/s`;
}

export default ResourceTimeline;
