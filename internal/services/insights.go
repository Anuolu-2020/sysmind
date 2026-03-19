package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"sysmind/internal/models"
)

// InsightService manages automatic system insights
type InsightService struct {
	dataDir   string
	insights  []models.AutoInsight
	triggers  map[string]*models.InsightTrigger
	mu        sync.RWMutex
	lastStats models.SystemStats

	// Baseline tracking for anomaly detection
	baselineStats *BaselineStats
	lastCleanup   int64
}

// BaselineStats tracks historical metrics for anomaly detection
type BaselineStats struct {
	AvgCPU          float64
	AvgMemory       float64
	MaxCPU          float64
	MaxMemory       float64
	CommonProcesses map[string]ProcessProfile
	LastUpdated     int64
	SampleCount     int
}

// ProcessProfile tracks historical behavior of a process
type ProcessProfile struct {
	Name      string
	CPUMin    float64
	CPUMax    float64
	MemoryMin float64
	MemoryMax float64
	LastSeen  int64
	Count     int
}

// NewInsightService creates a new insight service
func NewInsightService() (*InsightService, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}

	dataDir := filepath.Join(configDir, "sysmind", "insights")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	is := &InsightService{
		dataDir:  dataDir,
		insights: []models.AutoInsight{},
		triggers: make(map[string]*models.InsightTrigger),
		baselineStats: &BaselineStats{
			CommonProcesses: make(map[string]ProcessProfile),
			AvgCPU:          30, // Start with reasonable defaults
			MaxCPU:          80,
			AvgMemory:       50,
			MaxMemory:       85,
		},
	}

	// Load existing insights
	is.loadInsights()

	return is, nil
}

// generateID creates a unique insight ID
func generateInsightID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// nowMsInsights returns current time in milliseconds
func nowMsInsights() int64 {
	return time.Now().UnixMilli()
}

// AnalyzeSystem examines current system state and generates insights
func (is *InsightService) AnalyzeSystem(stats models.SystemStats, processes []models.ProcessInfo, security *models.SecurityInfo) []models.AutoInsight {
	is.mu.Lock()
	defer is.mu.Unlock()

	newInsights := []models.AutoInsight{}

	// Update baseline stats for anomaly detection
	is.updateBaseline(stats, processes)

	// 1. Sustained High CPU Usage (20+ minutes, not sporadic spikes)
	if stats.CPUPercent > 85 {
		triggerKey := "cpu_sustained"
		trigger := is.getTrigger(triggerKey)

		if trigger == nil {
			is.triggers[triggerKey] = &models.InsightTrigger{
				Type:       triggerKey,
				StartTime:  nowMsInsights(),
				LastUpdate: nowMsInsights(),
				Count:      1,
				Data:       fmt.Sprintf("%.1f", stats.CPUPercent),
			}
		} else {
			trigger.LastUpdate = nowMsInsights()
			trigger.Count++
			trigger.Data = fmt.Sprintf("%.1f", stats.CPUPercent)

			// Only generate insight after 20 minutes of sustained high CPU
			if !trigger.Triggered && (nowMsInsights()-trigger.StartTime) > 20*60*1000 {
				insight := models.AutoInsight{
					ID:        generateInsightID(),
					Title:     "Sustained High CPU Usage",
					Message:   fmt.Sprintf("CPU has been elevated above 85%% for %d minutes (currently %.1f%%) - this may impact performance", (nowMsInsights()-trigger.StartTime)/(60*1000), stats.CPUPercent),
					Category:  "performance",
					Severity:  "warning",
					Timestamp: nowMsInsights(),
					IsRead:    false,
					Data:      trigger.Data,
					ActionItems: []string{
						"Check which processes are consuming the most CPU",
						"Close unnecessary applications to free up resources",
						"Look for any stuck or runaway processes in the process list",
					},
				}
				newInsights = append(newInsights, insight)
				trigger.Triggered = true
			}
		}
	} else {
		// Clear trigger if CPU is normal
		if trigger := is.getTrigger("cpu_sustained"); trigger != nil && stats.CPUPercent < 75 {
			delete(is.triggers, "cpu_sustained")
		}
	}

	// 2. Critical Memory Usage (sustained for 10+ minutes)
	if stats.MemoryPercent > 95 {
		triggerKey := "memory_critical"
		trigger := is.getTrigger(triggerKey)

		if trigger == nil {
			is.triggers[triggerKey] = &models.InsightTrigger{
				Type:       triggerKey,
				StartTime:  nowMsInsights(),
				LastUpdate: nowMsInsights(),
				Count:      1,
				Data:       fmt.Sprintf("%.1f", stats.MemoryPercent),
			}
		} else {
			trigger.LastUpdate = nowMsInsights()
			trigger.Count++
			trigger.Data = fmt.Sprintf("%.1f", stats.MemoryPercent)

			// Generate insight after 10 minutes of critical memory
			if !trigger.Triggered && (nowMsInsights()-trigger.StartTime) > 10*60*1000 {
				insight := models.AutoInsight{
					ID:        generateInsightID(),
					Title:     "Critical Memory Level",
					Message:   fmt.Sprintf("Memory is at critically high level (%.1f%%) and has been for %d minutes - system may become unresponsive", stats.MemoryPercent, (nowMsInsights()-trigger.StartTime)/(60*1000)),
					Category:  "performance",
					Severity:  "critical",
					Timestamp: nowMsInsights(),
					IsRead:    false,
					Data:      trigger.Data,
					ActionItems: []string{
						"Close memory-heavy applications immediately (browsers with many tabs, VMs, etc.)",
						"Save your work and consider restarting the system",
						"Monitor memory usage closely to prevent system freeze",
					},
				}
				newInsights = append(newInsights, insight)
				trigger.Triggered = true
			}
		}
	} else {
		// Clear memory trigger if usage is normal
		if trigger := is.getTrigger("memory_critical"); trigger != nil && stats.MemoryPercent < 90 {
			delete(is.triggers, "memory_critical")
		}
	}

	// 3. Anomalous Process Detection (uses baseline)
	if len(processes) > 0 {
		is.detectAnomalousProcesses(processes, &newInsights)
	}

	// 4. Significant Network Anomaly (sustained unusual activity)
	if security != nil && len(security.UnknownConns) > 0 {
		is.detectNetworkAnomaly(security, &newInsights)
	}

	// 5. Confirmed Suspicious Process (only high/critical risk)
	if security != nil && len(security.SuspiciousProcs) > 0 {
		is.detectConfirmedThreats(security, &newInsights)
	}

	// Store new insights
	for _, insight := range newInsights {
		is.insights = append(is.insights, insight)
	}

	if len(newInsights) > 0 {
		is.saveInsights()
	}

	// Periodic cleanup of old triggers
	now := nowMsInsights()
	if now-is.lastCleanup > 60*60*1000 { // Cleanup hourly
		is.cleanupStaleTrigers()
		is.lastCleanup = now
	}

	is.lastStats = stats
	return newInsights
}

// updateBaseline continuously builds a profile of normal system behavior
func (is *InsightService) updateBaseline(stats models.SystemStats, processes []models.ProcessInfo) {
	baseline := is.baselineStats

	// Exponential moving average for baseline
	alpha := 0.2 // Weight for new samples
	if baseline.SampleCount < 100 {
		alpha = 0.5 // Start with higher weight for quicker convergence
	}

	baseline.AvgCPU = alpha*stats.CPUPercent + (1-alpha)*baseline.AvgCPU
	baseline.AvgMemory = alpha*stats.MemoryPercent + (1-alpha)*baseline.AvgMemory

	if stats.CPUPercent > baseline.MaxCPU {
		baseline.MaxCPU = stats.CPUPercent
	}
	if stats.MemoryPercent > baseline.MaxMemory {
		baseline.MaxMemory = stats.MemoryPercent
	}

	baseline.SampleCount++
	baseline.LastUpdated = nowMsInsights()

	// Track process profiles
	for _, proc := range processes {
		profile, exists := baseline.CommonProcesses[proc.Name]
		if exists {
			profile.Count++
			profile.LastSeen = nowMsInsights()
			if proc.CPUPercent > profile.CPUMax {
				profile.CPUMax = proc.CPUPercent
			}
			if proc.CPUPercent < profile.CPUMin || profile.Count == 1 {
				profile.CPUMin = proc.CPUPercent
			}
			if proc.MemoryMB > profile.MemoryMax {
				profile.MemoryMax = proc.MemoryMB
			}
			if proc.MemoryMB < profile.MemoryMin || profile.Count == 1 {
				profile.MemoryMin = proc.MemoryMB
			}
			baseline.CommonProcesses[proc.Name] = profile
		} else {
			// New process - start tracking
			baseline.CommonProcesses[proc.Name] = ProcessProfile{
				Name:      proc.Name,
				CPUMin:    proc.CPUPercent,
				CPUMax:    proc.CPUPercent,
				MemoryMin: proc.MemoryMB,
				MemoryMax: proc.MemoryMB,
				LastSeen:  nowMsInsights(),
				Count:     1,
			}
		}
	}
}

// detectAnomalousProcesses identifies processes with unusual behavior
func (is *InsightService) detectAnomalousProcesses(processes []models.ProcessInfo, newInsights *[]models.AutoInsight) {
	baseline := is.baselineStats

	for _, proc := range processes {
		triggerKey := fmt.Sprintf("process_%d_%s", proc.PID, proc.Name)

		// Skip if already triggered on this process
		if is.getTrigger(triggerKey) != nil {
			continue
		}

		profile, exists := baseline.CommonProcesses[proc.Name]

		// Only alert on truly anomalous behavior
		var isAnomalous bool
		var reason string

		if exists && profile.Count > 10 {
			// Process we know well - check if current usage significantly exceeds historical max
			cpuExcessPercentage := ((proc.CPUPercent - profile.CPUMax) / profile.CPUMax) * 100
			memExcessPercentage := ((proc.MemoryMB - profile.MemoryMax) / profile.MemoryMax) * 100

			if cpuExcessPercentage > 50 && proc.CPUPercent > 40 {
				isAnomalous = true
				reason = fmt.Sprintf("CPU usage (%.1f%%) is 50%% above historical max (%.1f%%) for this process", proc.CPUPercent, profile.CPUMax)
			} else if memExcessPercentage > 50 && proc.MemoryMB > 500 {
				isAnomalous = true
				reason = fmt.Sprintf("Memory usage (%.1fMB) is 50%% above historical max (%.1fMB)", proc.MemoryMB, profile.MemoryMax)
			}
		} else if !exists {
			// Unknown process - only alert if it's doing something extreme
			if proc.CPUPercent > 70 && proc.MemoryMB > 2000 {
				isAnomalous = true
				reason = fmt.Sprintf("New process using extreme resources: %.1f%% CPU and %.1fMB memory", proc.CPUPercent, proc.MemoryMB)
			} else if proc.CPUPercent > 80 {
				isAnomalous = true
				reason = fmt.Sprintf("New process using extremely high CPU: %.1f%%", proc.CPUPercent)
			}
		}

		if isAnomalous {
			severity := "info"
			if proc.CPUPercent > 70 || proc.MemoryMB > 3000 {
				severity = "warning"
			}

			insight := models.AutoInsight{
				ID:        generateInsightID(),
				Title:     fmt.Sprintf("Unusual Activity: %s", proc.Name),
				Message:   reason,
				Category:  "process",
				Severity:  severity,
				Timestamp: nowMsInsights(),
				IsRead:    false,
				Data:      fmt.Sprintf(`{"name":"%s","pid":%d,"cpu":%.1f,"memory":%.1f}`, proc.Name, proc.PID, proc.CPUPercent, proc.MemoryMB),
				ActionItems: []string{
					fmt.Sprintf("Review what %s is doing (check open files, network connections)", proc.Name),
					"Determine if this is expected for your current work",
					"Use the Process list to monitor or terminate if needed",
				},
			}
			*newInsights = append(*newInsights, insight)

			is.triggers[triggerKey] = &models.InsightTrigger{
				Type:       "anomalous_process",
				StartTime:  nowMsInsights(),
				LastUpdate: nowMsInsights(),
				Count:      1,
				Data:       proc.Name,
				Triggered:  true,
			}
		}
	}
}

// detectNetworkAnomaly identifies sustained unusual network patterns
func (is *InsightService) detectNetworkAnomaly(security *models.SecurityInfo, newInsights *[]models.AutoInsight) {
	connCount := len(security.UnknownConns)
	triggerKey := "network_anomaly"

	// Only alert on extreme sustained activity
	if connCount > 50 {
		trigger := is.getTrigger(triggerKey)
		if trigger == nil {
			is.triggers[triggerKey] = &models.InsightTrigger{
				Type:       triggerKey,
				StartTime:  nowMsInsights(),
				LastUpdate: nowMsInsights(),
				Count:      1,
				Triggered:  false,
			}
		} else {
			trigger.LastUpdate = nowMsInsights()
			trigger.Count++

			// Only alert after sustained high activity for 5+ minutes
			if !trigger.Triggered && (nowMsInsights()-trigger.StartTime) > 5*60*1000 {
				countryMap := make(map[string]int)
				for _, conn := range security.UnknownConns {
					if conn.Country != "" && conn.Country != "Unknown" {
						countryMap[conn.Country]++
					}
				}

				severity := "info"
				message := fmt.Sprintf("Sustained high network activity: %d connections detected", connCount)

				if len(countryMap) > 15 {
					severity = "warning"
					message = fmt.Sprintf("Your system has %d active connections to %d different countries", connCount, len(countryMap))
				}

				insight := models.AutoInsight{
					ID:        generateInsightID(),
					Title:     "Sustained Network Activity",
					Message:   message,
					Category:  "network",
					Severity:  severity,
					Timestamp: nowMsInsights(),
					IsRead:    false,
					Data:      fmt.Sprintf(`{"count":%d,"countries":%d}`, connCount, len(countryMap)),
					ActionItems: []string{
						"Check if you're running torrents, P2P apps, or cloud sync services",
						"Verify network activity matches your current applications",
						"Monitor performance to see if this is impacting speed",
					},
				}
				*newInsights = append(*newInsights, insight)
				trigger.Triggered = true
			}
		}
	} else {
		// Clear network trigger if activity returns to normal
		if trigger := is.getTrigger(triggerKey); trigger != nil && connCount < 30 {
			delete(is.triggers, triggerKey)
		}
	}
}

// detectConfirmedThreats only alerts on processes with proven high/critical risk
func (is *InsightService) detectConfirmedThreats(security *models.SecurityInfo, newInsights *[]models.AutoInsight) {
	for _, suspProc := range security.SuspiciousProcs {
		triggerKey := fmt.Sprintf("threat_%d", suspProc.PID)
		if is.getTrigger(triggerKey) != nil {
			continue // Already alerted on this
		}

		// Only alert on confirmed high-risk or critical threats
		if suspProc.RiskLevel == "high" || suspProc.RiskLevel == "critical" {
			severity := "warning"
			if suspProc.RiskLevel == "critical" {
				severity = "critical"
			}

			actionItems := []string{
				fmt.Sprintf("Investigate: %s (Risk: %s)", suspProc.Name, suspProc.RiskLevel),
				"Check if this is a legitimate application you installed",
				"Use your antivirus or security software to scan it",
			}

			if suspProc.RiskLevel == "critical" {
				actionItems = append(actionItems, "Consider immediately terminating and removing this process")
			} else {
				actionItems = append(actionItems, "Monitor this process for suspicious behavior")
			}

			insight := models.AutoInsight{
				ID:          generateInsightID(),
				Title:       fmt.Sprintf("%s Risk Process Detected", strings.Title(suspProc.RiskLevel)),
				Message:     fmt.Sprintf("'%s' has been flagged as %s risk", suspProc.Name, suspProc.RiskLevel),
				Category:    "security",
				Severity:    severity,
				Timestamp:   nowMsInsights(),
				IsRead:      false,
				Data:        fmt.Sprintf(`{"name":"%s","pid":%d,"risk":"%s"}`, suspProc.Name, suspProc.PID, suspProc.RiskLevel),
				ActionItems: actionItems,
			}
			*newInsights = append(*newInsights, insight)

			is.triggers[triggerKey] = &models.InsightTrigger{
				Type:       "confirmed_threat",
				StartTime:  nowMsInsights(),
				LastUpdate: nowMsInsights(),
				Count:      1,
				Triggered:  true,
			}
		}
	}
}

// cleanupStaleTrigers removes triggers that haven't been active in a while
func (is *InsightService) cleanupStaleTrigers() {
	now := nowMsInsights()
	staleThreshold := int64(24 * 60 * 60 * 1000) // 24 hours

	for key, trigger := range is.triggers {
		if now-trigger.LastUpdate > staleThreshold {
			delete(is.triggers, key)
		}
	}
}

// getTrigger retrieves an existing trigger
func (is *InsightService) getTrigger(key string) *models.InsightTrigger {
	return is.triggers[key]
}

// GetInsights returns all insights, optionally filtered
func (is *InsightService) GetInsights(onlyUnread bool) []models.AutoInsight {
	is.mu.RLock()
	defer is.mu.RUnlock()

	if !onlyUnread {
		return is.insights
	}

	unread := []models.AutoInsight{}
	for _, insight := range is.insights {
		if !insight.IsRead {
			unread = append(unread, insight)
		}
	}
	return unread
}

// MarkAsRead marks an insight as read
func (is *InsightService) MarkAsRead(insightID string) error {
	is.mu.Lock()
	defer is.mu.Unlock()

	for i := range is.insights {
		if is.insights[i].ID == insightID {
			is.insights[i].IsRead = true
			return is.saveInsights()
		}
	}
	return fmt.Errorf("insight not found")
}

// ClearInsights removes old insights (older than 24 hours)
func (is *InsightService) ClearInsights() error {
	is.mu.Lock()
	defer is.mu.Unlock()

	cutoff := nowMsInsights() - (24 * 60 * 60 * 1000) // 24 hours ago
	kept := []models.AutoInsight{}

	for _, insight := range is.insights {
		if insight.Timestamp > cutoff {
			kept = append(kept, insight)
		}
	}

	is.insights = kept
	return is.saveInsights()
}

// ClearAllInsights removes all insights
func (is *InsightService) ClearAllInsights() error {
	is.mu.Lock()
	defer is.mu.Unlock()

	is.insights = []models.AutoInsight{}
	return is.saveInsights()
}

// saveInsights persists insights to disk
func (is *InsightService) saveInsights() error {
	data, err := json.MarshalIndent(is.insights, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(is.dataDir, "insights.json")
	return os.WriteFile(filename, data, 0644)
}

// loadInsights loads insights from disk
func (is *InsightService) loadInsights() {
	filename := filepath.Join(is.dataDir, "insights.json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return // File doesn't exist yet
	}

	var insights []models.AutoInsight
	if err := json.Unmarshal(data, &insights); err != nil {
		return // Invalid format
	}

	is.insights = insights
}
