import { useCallback, useEffect, useMemo, useState } from 'react';

function TimeMachine({ compact = false }) {
  const [hours, setHours] = useState(6);
  const [data, setData] = useState({
    samples: [],
    annotations: [],
    forecasts: [],
    summary: '',
    retentionHours: 0,
    samplingSeconds: 0,
  });
  const [loading, setLoading] = useState(true);
  const [selectedMomentId, setSelectedMomentId] = useState(null);

  const fetchTimeMachine = useCallback(async () => {
    try {
      setLoading(true);
      if (window.go?.main?.App?.GetTimeMachine) {
        const response = await window.go.main.App.GetTimeMachine(hours);
        setData(response || {
          samples: [],
          annotations: [],
          forecasts: [],
          summary: 'Time machine data is unavailable.',
          retentionHours: 0,
          samplingSeconds: 0,
        });
      }
    } catch (err) {
      console.error('Error fetching time machine:', err);
      setData({
        samples: [],
        annotations: [],
        forecasts: [],
        summary: 'Time machine data is unavailable.',
        retentionHours: 0,
        samplingSeconds: 0,
      });
    } finally {
      setLoading(false);
    }
  }, [hours]);

  useEffect(() => {
    fetchTimeMachine();
    const interval = setInterval(fetchTimeMachine, 30000);
    return () => clearInterval(interval);
  }, [fetchTimeMachine]);

  const timelineMoments = useMemo(() => {
    const moments = [];
    const firstSample = data.samples?.[0];
    const lastSample = data.samples?.[data.samples.length - 1];

    if (firstSample) {
      moments.push({
        id: 'window-start',
        kind: 'window-start',
        timestamp: firstSample.timestamp,
        title: 'Window start',
        summary: `Reviewing the last ${hours} hours of persisted telemetry.`,
      });
    }

    (data.annotations || []).forEach((annotation) => {
      moments.push({
        id: annotation.id,
        kind: 'annotation',
        timestamp: annotation.timestamp,
        title: annotation.title,
        summary: annotation.summary,
        severity: annotation.severity,
        metric: annotation.metric,
        processName: annotation.processName,
        source: annotation,
      });
    });

    if (lastSample) {
      moments.push({
        id: 'now',
        kind: 'present',
        timestamp: lastSample.timestamp,
        title: 'Now',
        summary: 'Current endpoint of the persisted telemetry window.',
      });
    }

    (data.forecasts || []).forEach((forecast) => {
      moments.push({
        id: forecast.id,
        kind: 'forecast',
        timestamp: forecast.predictedAt,
        title: forecast.title,
        summary: forecast.summary,
        severity: forecast.severity,
        source: forecast,
      });
    });

    return moments.sort((a, b) => a.timestamp - b.timestamp);
  }, [data.annotations, data.forecasts, data.samples, hours]);

  useEffect(() => {
    if (!timelineMoments.length) {
      setSelectedMomentId(null);
      return;
    }

    const preferred = timelineMoments.find((moment) => moment.kind === 'annotation')
      || timelineMoments.find((moment) => moment.kind === 'forecast')
      || timelineMoments[timelineMoments.length - 1];

    const exists = timelineMoments.some((moment) => moment.id === selectedMomentId);
    if (!exists) {
      setSelectedMomentId(preferred.id);
    }
  }, [timelineMoments, selectedMomentId]);

  const selectedMoment = useMemo(() => (
    timelineMoments.find((moment) => moment.id === selectedMomentId) || timelineMoments[0] || null
  ), [selectedMomentId, timelineMoments]);

  const selectedIndex = useMemo(() => (
    Math.max(0, timelineMoments.findIndex((moment) => moment.id === selectedMomentId))
  ), [selectedMomentId, timelineMoments]);

  const selectedSample = useMemo(() => (
    findNearestSample(data.samples || [], selectedMoment?.timestamp)
  ), [data.samples, selectedMoment]);

  const chartBounds = useMemo(() => {
    const firstSample = data.samples?.[0];
    const lastSample = data.samples?.[data.samples.length - 1];
    const lastForecast = data.forecasts?.[data.forecasts.length - 1];
    if (!firstSample || !lastSample) return null;
    return {
      start: firstSample.timestamp,
      end: Math.max(lastSample.timestamp, lastForecast?.predictedAt || lastSample.timestamp),
    };
  }, [data.forecasts, data.samples]);

  return (
    <section className={`time-machine-card ${compact ? 'compact' : ''}`}>
      <header className="time-machine-header">
        <div>
          <h3>Predictive Time Machine</h3>
          <p className="time-machine-subtitle">
            {compact ? 'Past highlights and future projections' : 'Scrub past incidents and future projections from persisted telemetry'}
          </p>
        </div>
        <div className="time-machine-controls">
          {[3, 6, 24].map((option) => (
            <button
              key={option}
              className={`timeline-range-btn ${hours === option ? 'active' : ''}`}
              onClick={() => setHours(option)}
            >
              {option}h
            </button>
          ))}
        </div>
      </header>

      <div className={`time-machine-summary ${compact ? 'compact' : ''}`}>
        <p>{data.summary || 'Collecting persisted history for the time machine.'}</p>
        <div className="time-machine-summary-meta">
          <span className="rewind-pill">Persisted {data.retentionHours || '?'}h</span>
          <span className="rewind-pill muted">Sampling {data.samplingSeconds || '?'}s</span>
        </div>
      </div>

      {loading && !data.samples?.length ? (
        <div className="time-machine-empty">Collecting persisted history...</div>
      ) : !timelineMoments.length ? (
        <div className="time-machine-empty">No timeline moments are available yet.</div>
      ) : (
        <>
          <div className="time-machine-track">
            <div className="time-machine-line"></div>
            {timelineMoments.map((moment) => (
              <button
                key={moment.id}
                className={`time-machine-marker kind-${moment.kind} severity-${moment.severity || 'info'} ${selectedMoment?.id === moment.id ? 'active' : ''}`}
                style={{ left: `${markerPosition(moment.timestamp, chartBounds)}%` }}
                onClick={() => setSelectedMomentId(moment.id)}
                title={`${moment.title} - ${moment.summary}`}
              />
            ))}
          </div>

          <div className="time-machine-slider-row">
            <span>{formatTimelineEdge(chartBounds?.start)}</span>
            <input
              type="range"
              min="0"
              max={Math.max(timelineMoments.length - 1, 0)}
              value={selectedIndex}
              onChange={(e) => {
                const nextMoment = timelineMoments[Number(e.target.value)];
                if (nextMoment) {
                  setSelectedMomentId(nextMoment.id);
                }
              }}
              className="time-machine-slider"
            />
            <span>{formatTimelineEdge(chartBounds?.end)}</span>
          </div>

          <div className="time-machine-grid">
            <div className="time-machine-detail-card">
              <div className="time-machine-detail-meta">
                <span className={`incident-severity severity-${selectedMoment?.severity || 'info'}`}>
                  {selectedMoment?.kind === 'forecast' ? 'forecast' : selectedMoment?.severity || 'info'}
                </span>
                <span className="time-machine-relative">{formatRelativeTime(selectedMoment?.timestamp)}</span>
              </div>
              <h4>{selectedMoment?.title}</h4>
              <p>{selectedMoment?.summary}</p>

              {selectedMoment?.kind === 'forecast' ? (
                <ForecastDetail forecast={selectedMoment?.source} />
              ) : (
                <SnapshotDetail sample={selectedSample} annotation={selectedMoment?.source} />
              )}
            </div>

            <div className="time-machine-side-column">
              <div className="time-machine-side-card">
                <h4>Past Highlights</h4>
                {(data.annotations || []).slice(0, 5).map((annotation) => (
                  <button
                    key={annotation.id}
                    className="time-machine-list-item"
                    onClick={() => setSelectedMomentId(annotation.id)}
                  >
                    <span>{annotation.title}</span>
                    <span>{formatRelativeTime(annotation.timestamp)}</span>
                  </button>
                ))}
                {!data.annotations?.length && <p className="incident-note">No historical annotations yet.</p>}
              </div>

              <div className="time-machine-side-card">
                <h4>Future Projections</h4>
                {(data.forecasts || []).slice(0, 4).map((forecast) => (
                  <button
                    key={forecast.id}
                    className="time-machine-list-item"
                    onClick={() => setSelectedMomentId(forecast.id)}
                  >
                    <span>{forecast.title}</span>
                    <span>{formatRelativeTime(forecast.predictedAt)}</span>
                  </button>
                ))}
                {!data.forecasts?.length && <p className="incident-note">No stable forecast yet.</p>}
              </div>
            </div>
          </div>
        </>
      )}
    </section>
  );
}

function SnapshotDetail({ sample, annotation }) {
  if (!sample) {
    return <p className="incident-note">No matching persisted snapshot is available.</p>;
  }

  return (
    <div className="time-machine-detail-body">
      <div className="incident-kpi-row">
        <div className="incident-kpi">
          <span className="kpi-label">CPU</span>
          <span className="kpi-value">{sample.cpuPercent?.toFixed(1)}%</span>
        </div>
        <div className="incident-kpi">
          <span className="kpi-label">Memory</span>
          <span className="kpi-value">{sample.memoryPercent?.toFixed(1)}%</span>
        </div>
        <div className="incident-kpi">
          <span className="kpi-label">Disk</span>
          <span className="kpi-value">{sample.diskPercent?.toFixed(1)}%</span>
        </div>
      </div>

      <div className="time-machine-processes">
        {(sample.processes || []).slice(0, 4).map((proc) => (
          <div className="snapshot-process" key={`${sample.timestamp}-${proc.pid}`}>
            <span className="proc-name">{proc.name || 'Unknown'} (PID {proc.pid})</span>
            <span className="proc-metrics">
              {proc.cpuPercent?.toFixed(1)}% CPU, {proc.memoryMB?.toFixed(0)} MB, {proc.numThreads} thr
            </span>
          </div>
        ))}
      </div>

      {annotation?.processName && (
        <p className="incident-note">
          Annotation focus: {annotation.processName}{annotation.processPid ? ` (PID ${annotation.processPid})` : ''}.
        </p>
      )}
    </div>
  );
}

function ForecastDetail({ forecast }) {
  if (!forecast) {
    return <p className="incident-note">No forecast details available.</p>;
  }

  return (
    <div className="time-machine-detail-body">
      <div className="incident-kpi-row">
        <div className="incident-kpi">
          <span className="kpi-label">Current</span>
          <span className="kpi-value">{formatForecastValue(forecast.currentValue, forecast.unit)}</span>
        </div>
        <div className="incident-kpi">
          <span className="kpi-label">Projected</span>
          <span className="kpi-value">{formatForecastValue(forecast.projectedValue, forecast.unit)}</span>
        </div>
        <div className="incident-kpi">
          <span className="kpi-label">Confidence</span>
          <span className="kpi-value">{Math.round((forecast.confidence || 0) * 100)}%</span>
        </div>
      </div>
      <p className="incident-note">Predicted time: {formatAbsoluteTime(forecast.predictedAt)}</p>
    </div>
  );
}

function findNearestSample(samples, timestamp) {
  if (!samples?.length || !timestamp) return null;

  let best = samples[0];
  let bestDistance = Math.abs(samples[0].timestamp - timestamp);
  for (const sample of samples) {
    const distance = Math.abs(sample.timestamp - timestamp);
    if (distance < bestDistance) {
      best = sample;
      bestDistance = distance;
    }
  }
  return best;
}

function markerPosition(timestamp, bounds) {
  if (!bounds || !timestamp || bounds.end <= bounds.start) return 0;
  return ((timestamp - bounds.start) / (bounds.end - bounds.start)) * 100;
}

function formatRelativeTime(timestamp) {
  if (!timestamp) return 'Unknown';
  const diffMs = timestamp - Date.now();
  const absMinutes = Math.round(Math.abs(diffMs) / 60000);
  if (absMinutes < 1) return 'now';
  if (absMinutes < 60) return diffMs < 0 ? `${absMinutes}m ago` : `in ${absMinutes}m`;
  const absHours = Math.round(absMinutes / 60);
  if (absHours < 48) return diffMs < 0 ? `${absHours}h ago` : `in ${absHours}h`;
  const absDays = Math.round(absHours / 24);
  return diffMs < 0 ? `${absDays}d ago` : `in ${absDays}d`;
}

function formatTimelineEdge(timestamp) {
  if (!timestamp) return '--';
  return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function formatAbsoluteTime(timestamp) {
  if (!timestamp) return 'Unknown';
  return new Date(timestamp).toLocaleString([], {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function formatForecastValue(value, unit) {
  if (unit === '%') return `${value.toFixed(1)}%`;
  return `${value.toFixed(1)} ${unit}`;
}

export default TimeMachine;
