import { useEffect, useRef, useMemo } from 'react';
import 'leaflet/dist/leaflet.css';

function ConnectionMap({ connections }) {
  const mapRef = useRef(null);
  const mapInstanceRef = useRef(null);
  const markersRef = useRef([]);

  // Debug logging
  console.log('ConnectionMap received connections:', connections);

  // Group connections by country
  const connectionsByCountry = useMemo(() => {
    const grouped = {};
    
    if (!connections || connections.length === 0) {
      return grouped;
    }
    
    connections.forEach(conn => {
      console.log('Processing connection:', conn);
      
      // Skip local connections
      if (conn.countryCode === 'LAN' || !conn.latitude || !conn.longitude) {
        console.log('Skipping local/invalid connection:', conn.countryCode, conn.latitude, conn.longitude);
        return;
      }
      
      const key = conn.countryCode || conn.country;
      if (!grouped[key]) {
        grouped[key] = {
          country: conn.country,
          countryCode: conn.countryCode,
          latitude: conn.latitude,
          longitude: conn.longitude,
          count: 0,
          connections: []
        };
      }
      grouped[key].count++;
      grouped[key].connections.push(conn);
    });
    
    console.log('Grouped connections by country:', grouped);
    return grouped;
  }, [connections]);

  useEffect(() => {
    // Import Leaflet dynamically to avoid SSR issues
    import('leaflet').then(L => {
      // Fix default markers in Leaflet
      delete L.Icon.Default.prototype._getIconUrl;
      L.Icon.Default.mergeOptions({
        iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon-2x.png',
        iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png',
        shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
      });

      if (!mapInstanceRef.current && mapRef.current) {
        // Initialize the map
        mapInstanceRef.current = L.map(mapRef.current, {
          center: [20, 0], // Center on world
          zoom: 2,
          maxZoom: 10,
          minZoom: 1,
          worldCopyJump: true
        });

        // Add OpenStreetMap tiles (free, no API key needed)
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
          attribution: '© OpenStreetMap contributors',
          maxZoom: 18,
        }).addTo(mapInstanceRef.current);

        // Alternative tile layers (you can switch between these)
        // 
        // Dark theme:
        // L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
        //   attribution: '© OpenStreetMap © CartoDB',
        //   maxZoom: 18,
        // }).addTo(mapInstanceRef.current);
        //
        // Satellite view:
        // L.tileLayer('https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}', {
        //   attribution: '© Esri',
        //   maxZoom: 18,
        // }).addTo(mapInstanceRef.current);
      }

      // Clear existing markers
      if (mapInstanceRef.current) {
        markersRef.current.forEach(marker => {
          mapInstanceRef.current.removeLayer(marker);
        });
        markersRef.current = [];

        // Add connection markers
        Object.values(connectionsByCountry).forEach((group) => {
          const { latitude, longitude, country, countryCode, count, connections } = group;

          // Create custom icon based on connection count
          const size = Math.min(20 + count * 8, 60);
          const customIcon = L.divIcon({
            className: 'custom-connection-marker',
            html: `
              <div class="marker-container">
                <div class="marker-pulse"></div>
                <div class="marker-main" style="width: ${size}px; height: ${size}px;">
                  <span class="marker-count">${count}</span>
                </div>
              </div>
            `,
            iconSize: [size, size],
            iconAnchor: [size/2, size/2],
          });

          // Create popup content
          const popupContent = `
            <div class="connection-popup">
              <h4>🌍 ${country} (${countryCode})</h4>
              <p><strong>${count}</strong> connection${count > 1 ? 's' : ''}</p>
              <p>📍 ${latitude.toFixed(2)}°, ${longitude.toFixed(2)}°</p>
              <div class="popup-connections">
                <strong>Connections:</strong>
                ${connections.slice(0, 5).map(conn => `
                  <div class="popup-connection">
                    🔗 ${conn.remoteAddr}${conn.processName ? ` (${conn.processName})` : ''}
                  </div>
                `).join('')}
                ${connections.length > 5 ? `<div class="popup-more">+${connections.length - 5} more</div>` : ''}
              </div>
            </div>
          `;

          // Create marker
          const marker = L.marker([latitude, longitude], { icon: customIcon })
            .bindPopup(popupContent, {
              maxWidth: 300,
              className: 'custom-popup'
            });

          // Add to map and track
          marker.addTo(mapInstanceRef.current);
          markersRef.current.push(marker);

          console.log(`Added marker for ${country} at ${latitude}, ${longitude} with ${count} connections`);
        });

        // If we have connections, fit the map to show all markers
        if (markersRef.current.length > 0) {
          const group = new L.featureGroup(markersRef.current);
          mapInstanceRef.current.fitBounds(group.getBounds().pad(0.1));
        }
      }
    }).catch(error => {
      console.error('Error loading Leaflet:', error);
    });
  }, [connectionsByCountry]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (mapInstanceRef.current) {
        mapInstanceRef.current.remove();
        mapInstanceRef.current = null;
      }
    };
  }, []);

  const hasValidConnections = Object.keys(connectionsByCountry).length > 0;

  if (!connections || connections.length === 0) {
    return (
      <div className="connection-map-empty">
        <p>No external connections to display</p>
      </div>
    );
  }

  if (!hasValidConnections) {
    return (
      <div className="connection-map-container">
        <h3>Connection Map</h3>
        <div className="connection-map-empty">
          <p>Found {connections.length} connections but no geo-located external connections</p>
          <details>
            <summary>Debug Info (click to expand)</summary>
            <pre style={{ fontSize: '10px', maxHeight: '200px', overflow: 'auto' }}>
              {JSON.stringify(connections, null, 2)}
            </pre>
          </details>
        </div>
      </div>
    );
  }

  return (
    <div className="connection-map-container">
      <h3>🌍 Network Connection Map ({Object.keys(connectionsByCountry).length} locations)</h3>
      
      {/* Real interactive map */}
      <div className="leaflet-map-wrapper">
        <div ref={mapRef} className="leaflet-map" style={{ height: '400px', width: '100%' }}></div>
      </div>
      
      {/* Map controls info */}
      <div className="map-controls-info">
        <p>
          🖱️ <strong>Interactive Map:</strong> Click and drag to pan • Scroll to zoom • Click markers for details
        </p>
      </div>
      
      {/* Connection statistics */}
      <div className="connection-stats">
        <div className="stats-grid">
          <div className="stat-item">
            <span className="stat-number">{Object.keys(connectionsByCountry).length}</span>
            <span className="stat-label">Countries</span>
          </div>
          <div className="stat-item">
            <span className="stat-number">{connections.length}</span>
            <span className="stat-label">Total Connections</span>
          </div>
          <div className="stat-item">
            <span className="stat-number">
              {Math.max(...Object.values(connectionsByCountry).map(g => g.count))}
            </span>
            <span className="stat-label">Max per Country</span>
          </div>
        </div>
      </div>
      
      {/* Connection details list */}
      <div className="connection-legend">
        <h4>🔗 Connection Details</h4>
        <div className="legend-list">
          {Object.values(connectionsByCountry)
            .sort((a, b) => b.count - a.count)
            .map((group, i) => (
              <div key={group.countryCode || i} className="legend-item">
                <div className="legend-marker-leaflet" />
                <div className="legend-info">
                  <span className="legend-country">
                    🌍 {group.country} ({group.countryCode})
                  </span>
                  <span className="legend-count">
                    {group.count} connection{group.count > 1 ? 's' : ''}
                  </span>
                </div>
                <div className="legend-coords">
                  📍 {group.latitude.toFixed(2)}°, {group.longitude.toFixed(2)}°
                </div>
                <div className="legend-details">
                  {group.connections.slice(0, 3).map((conn, j) => (
                    <div key={j} className="connection-detail">
                      <span className="text-mono">🔗 {conn.remoteAddr}</span>
                      {conn.processName && (
                        <span className="process-name">({conn.processName})</span>
                      )}
                    </div>
                  ))}
                  {group.connections.length > 3 && (
                    <div className="connection-detail text-muted">
                      +{group.connections.length - 3} more connections
                    </div>
                  )}
                </div>
              </div>
            ))}
        </div>
      </div>
    </div>
  );
}

export default ConnectionMap;
