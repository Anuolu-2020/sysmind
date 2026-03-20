import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

// Helper to get computed CSS variable values
function getCSSVariable(name) {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
}

// Helper to convert hex to rgba
function hexToRgba(hex, alpha) {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  if (result) {
    return `rgba(${parseInt(result[1], 16)}, ${parseInt(result[2], 16)}, ${parseInt(result[3], 16)}, ${alpha})`;
  }
  return hex;
}

// Get theme colors from CSS variables
function getThemeColors() {
  return {
    bgPrimary: getCSSVariable('--bg-primary') || '#1a1b26',
    bgSecondary: getCSSVariable('--bg-secondary') || '#24283b',
    bgTertiary: getCSSVariable('--bg-tertiary') || '#414868',
    textPrimary: getCSSVariable('--text-primary') || '#c0caf5',
    textSecondary: getCSSVariable('--text-secondary') || '#a9b1d6',
    textMuted: getCSSVariable('--text-muted') || '#565f89',
    accentPrimary: getCSSVariable('--accent-primary') || '#7aa2f7',
    accentSuccess: getCSSVariable('--accent-success') || '#9ece6a',
    accentWarning: getCSSVariable('--accent-warning') || '#e0af68',
    accentDanger: getCSSVariable('--accent-danger') || '#f7768e',
    borderColor: getCSSVariable('--border-color') || '#414868',
  };
}

function ResourceTimeline({ compact = false }) {
  const [points, setPoints] = useState([]);
  const [minutes, setMinutes] = useState(30);
  const [theme, setTheme] = useState(null);
  const percentCanvasRef = useRef(null);
  const networkCanvasRef = useRef(null);
  const animationRef = useRef(null);

  // Watch for theme changes
  useEffect(() => {
    const observer = new MutationObserver(() => {
      setTheme(document.documentElement.getAttribute('data-theme') || 'default');
    });
    
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['data-theme']
    });

    // Set initial theme
    setTheme(document.documentElement.getAttribute('data-theme') || 'default');

    return () => observer.disconnect();
  }, []);

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
    return Math.max(maxValue, 1024); // At least 1KB/s scale
  }, [points]);

  // Draw percentage chart (CPU, Memory, Disk)
  useEffect(() => {
    const canvas = percentCanvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const colors = getThemeColors();
    const dpr = window.devicePixelRatio || 1;
    const rect = canvas.getBoundingClientRect();

    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;
    ctx.scale(dpr, dpr);

    const width = rect.width;
    const height = rect.height;
    const padding = { top: 20, right: 50, bottom: 35, left: 10 };
    const chartWidth = width - padding.left - padding.right;
    const chartHeight = height - padding.top - padding.bottom;

    // Clear and draw background
    ctx.clearRect(0, 0, width, height);
    
    // Draw chart area background with subtle gradient
    const bgGradient = ctx.createLinearGradient(0, padding.top, 0, height - padding.bottom);
    bgGradient.addColorStop(0, hexToRgba(colors.bgPrimary, 0.8));
    bgGradient.addColorStop(1, hexToRgba(colors.bgPrimary, 0.95));
    ctx.fillStyle = bgGradient;
    ctx.fillRect(padding.left, padding.top, chartWidth, chartHeight);

    // Draw grid
    drawGrid(ctx, padding, chartWidth, chartHeight, colors);

    if (points.length < 2) {
      drawNoData(ctx, width, height, colors);
      return;
    }

    // Draw charts with gradients - order matters for layering
    drawSmoothAreaChart(ctx, points, p => Number(p?.diskPercent || 0), padding, chartWidth, chartHeight, 100, colors.accentWarning, 0.15);
    drawSmoothAreaChart(ctx, points, p => Number(p?.memoryPercent || 0), padding, chartWidth, chartHeight, 100, colors.accentSuccess, 0.2);
    drawSmoothAreaChart(ctx, points, p => Number(p?.cpuPercent || 0), padding, chartWidth, chartHeight, 100, colors.accentPrimary, 0.25);

    // Draw Y-axis labels
    drawYLabels(ctx, padding, chartWidth, chartHeight, ['100%', '75%', '50%', '25%', '0%'], colors);

    // Draw time labels
    drawTimeLabels(ctx, padding, chartWidth, chartHeight, points, minutes, colors);

  }, [points, minutes, theme]);

  // Draw network chart
  useEffect(() => {
    const canvas = networkCanvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const colors = getThemeColors();
    const dpr = window.devicePixelRatio || 1;
    const rect = canvas.getBoundingClientRect();

    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;
    ctx.scale(dpr, dpr);

    const width = rect.width;
    const height = rect.height;
    const padding = { top: 20, right: 60, bottom: 35, left: 10 };
    const chartWidth = width - padding.left - padding.right;
    const chartHeight = height - padding.top - padding.bottom;

    // Clear and draw background
    ctx.clearRect(0, 0, width, height);

    // Draw chart area background
    const bgGradient = ctx.createLinearGradient(0, padding.top, 0, height - padding.bottom);
    bgGradient.addColorStop(0, hexToRgba(colors.bgPrimary, 0.8));
    bgGradient.addColorStop(1, hexToRgba(colors.bgPrimary, 0.95));
    ctx.fillStyle = bgGradient;
    ctx.fillRect(padding.left, padding.top, chartWidth, chartHeight);

    // Draw grid
    drawGrid(ctx, padding, chartWidth, chartHeight, colors);

    if (points.length < 2) {
      drawNoData(ctx, width, height, colors);
      return;
    }

    // Draw network charts
    drawSmoothAreaChart(ctx, points, p => Number(p?.netDownSpeed || 0), padding, chartWidth, chartHeight, netMax, colors.accentSuccess, 0.25);
    drawSmoothAreaChart(ctx, points, p => Number(p?.netUploadSpeed || 0), padding, chartWidth, chartHeight, netMax, colors.accentDanger, 0.2);

    // Draw Y-axis labels for network
    drawNetworkYLabels(ctx, padding, chartWidth, chartHeight, netMax, colors);

    // Draw time labels
    drawTimeLabels(ctx, padding, chartWidth, chartHeight, points, minutes, colors);

  }, [points, netMax, minutes, theme]);

  return (
    <section className={`resource-timeline-card ${compact ? 'compact' : ''}`}>
      <header className="resource-timeline-header">
        <div className="timeline-title-section">
          <h3>Resource Timeline</h3>
          <p className="resource-timeline-subtitle">
            {compact ? 'Live resource trends' : 'Real-time system performance metrics'}
          </p>
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
        {/* System Resources Chart */}
        <div className="timeline-chart-block">
          <div className="timeline-label-row">
            <div className="timeline-legend-group">
              <span className="timeline-legend-item cpu">
                <span className="legend-dot"></span>
                CPU
              </span>
              <span className="timeline-legend-item memory">
                <span className="legend-dot"></span>
                Memory
              </span>
              <span className="timeline-legend-item disk">
                <span className="legend-dot"></span>
                Disk
              </span>
            </div>
            <span className="axis-note">Utilization %</span>
          </div>
          <canvas ref={percentCanvasRef} className="timeline-canvas" />
        </div>

        {/* Network Chart */}
        <div className="timeline-chart-block">
          <div className="timeline-label-row">
            <div className="timeline-legend-group">
              <span className="timeline-legend-item net-down">
                <span className="legend-dot"></span>
                Download
              </span>
              <span className="timeline-legend-item net-up">
                <span className="legend-dot"></span>
                Upload
              </span>
            </div>
            <span className="axis-note">Network Speed</span>
          </div>
          <canvas ref={networkCanvasRef} className="timeline-canvas network" />
        </div>
      </div>
    </section>
  );
}

// Draw grid lines
function drawGrid(ctx, padding, chartWidth, chartHeight, colors) {
  ctx.strokeStyle = hexToRgba(colors.textMuted, 0.2);
  ctx.lineWidth = 1;

  // Horizontal grid lines
  for (let i = 0; i <= 4; i++) {
    const y = padding.top + (i / 4) * chartHeight;
    ctx.beginPath();
    ctx.setLineDash([4, 4]);
    ctx.moveTo(padding.left, y);
    ctx.lineTo(padding.left + chartWidth, y);
    ctx.stroke();
  }

  // Vertical grid lines
  ctx.setLineDash([]);
  for (let i = 1; i < 4; i++) {
    const x = padding.left + (i / 4) * chartWidth;
    ctx.beginPath();
    ctx.setLineDash([4, 4]);
    ctx.moveTo(x, padding.top);
    ctx.lineTo(x, padding.top + chartHeight);
    ctx.stroke();
  }
  ctx.setLineDash([]);
}

// Draw "no data" message
function drawNoData(ctx, width, height, colors) {
  ctx.fillStyle = colors.textMuted;
  ctx.font = '13px -apple-system, BlinkMacSystemFont, sans-serif';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText('Collecting data...', width / 2, height / 2);
}

// Draw smooth area chart with bezier curves
function drawSmoothAreaChart(ctx, points, getter, padding, chartWidth, chartHeight, maxValue, color, fillAlpha) {
  if (points.length < 2) return;
  
  const safeMax = Math.max(maxValue, 1);
  const values = points.map(p => getter(p));
  
  // Calculate points
  const chartPoints = values.map((value, index) => ({
    x: padding.left + (index / (points.length - 1)) * chartWidth,
    y: padding.top + chartHeight - (Math.min(Math.max(value, 0), safeMax) / safeMax) * chartHeight
  }));

  // Draw filled area with gradient
  const gradient = ctx.createLinearGradient(0, padding.top, 0, padding.top + chartHeight);
  gradient.addColorStop(0, hexToRgba(color, fillAlpha));
  gradient.addColorStop(1, hexToRgba(color, 0.02));

  ctx.beginPath();
  ctx.moveTo(chartPoints[0].x, padding.top + chartHeight);
  
  // Draw smooth curve for area
  chartPoints.forEach((point, i) => {
    if (i === 0) {
      ctx.lineTo(point.x, point.y);
    } else {
      const prev = chartPoints[i - 1];
      const cpX = (prev.x + point.x) / 2;
      ctx.bezierCurveTo(cpX, prev.y, cpX, point.y, point.x, point.y);
    }
  });
  
  ctx.lineTo(chartPoints[chartPoints.length - 1].x, padding.top + chartHeight);
  ctx.closePath();
  ctx.fillStyle = gradient;
  ctx.fill();

  // Draw line
  ctx.beginPath();
  chartPoints.forEach((point, i) => {
    if (i === 0) {
      ctx.moveTo(point.x, point.y);
    } else {
      const prev = chartPoints[i - 1];
      const cpX = (prev.x + point.x) / 2;
      ctx.bezierCurveTo(cpX, prev.y, cpX, point.y, point.x, point.y);
    }
  });
  ctx.strokeStyle = color;
  ctx.lineWidth = 2;
  ctx.lineCap = 'round';
  ctx.lineJoin = 'round';
  ctx.stroke();

  // Draw glow effect for the line
  ctx.shadowColor = color;
  ctx.shadowBlur = 6;
  ctx.stroke();
  ctx.shadowBlur = 0;
}

// Draw Y-axis labels
function drawYLabels(ctx, padding, chartWidth, chartHeight, labels, colors) {
  ctx.fillStyle = colors.textSecondary;
  ctx.font = '11px -apple-system, BlinkMacSystemFont, sans-serif';
  ctx.textAlign = 'left';
  ctx.textBaseline = 'middle';
  
  labels.forEach((label, index) => {
    const y = padding.top + (index / (labels.length - 1)) * chartHeight;
    ctx.fillText(label, padding.left + chartWidth + 8, y);
  });
}

// Draw network Y-axis labels
function drawNetworkYLabels(ctx, padding, chartWidth, chartHeight, maxValue, colors) {
  const values = [maxValue, maxValue * 0.75, maxValue * 0.5, maxValue * 0.25, 0];
  ctx.fillStyle = colors.textSecondary;
  ctx.font = '11px -apple-system, BlinkMacSystemFont, sans-serif';
  ctx.textAlign = 'left';
  ctx.textBaseline = 'middle';
  
  values.forEach((value, index) => {
    const y = padding.top + (index / (values.length - 1)) * chartHeight;
    ctx.fillText(formatSpeed(value), padding.left + chartWidth + 8, y);
  });
}

// Draw time labels
function drawTimeLabels(ctx, padding, chartWidth, chartHeight, points, minutes, colors) {
  if (!points.length) return;
  
  const startTs = points[0].timestamp;
  const endTs = points[points.length - 1].timestamp;
  const midTs = startTs + (endTs - startTs) / 2;
  const y = padding.top + chartHeight + 20;

  ctx.fillStyle = colors.textMuted;
  ctx.font = '11px -apple-system, BlinkMacSystemFont, sans-serif';
  ctx.textBaseline = 'middle';

  // Start time
  ctx.textAlign = 'left';
  ctx.fillText(formatTime(startTs), padding.left, y);

  // Middle time
  ctx.textAlign = 'center';
  ctx.fillText(formatTime(midTs), padding.left + chartWidth / 2, y);

  // End time with duration
  ctx.textAlign = 'right';
  ctx.fillText(`${formatTime(endTs)} (${minutes}m window)`, padding.left + chartWidth, y);
}

// Format timestamp to time string
function formatTime(ts) {
  const date = new Date(ts);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
}

// Format speed with appropriate units
function formatSpeed(value) {
  if (value < 1024) return `${value.toFixed(0)} B/s`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB/s`;
  if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB/s`;
  return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB/s`;
}

export default ResourceTimeline;
