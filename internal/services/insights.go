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

	// 1. High CPU Usage Detection (more conservative)
	if stats.CPUPercent > 90 {
		triggerKey := "cpu_high"
		trigger := is.getTrigger(triggerKey)

		if trigger == nil {
			// Start tracking high CPU
			is.triggers[triggerKey] = &models.InsightTrigger{
				Type:       triggerKey,
				StartTime:  nowMsInsights(),
				LastUpdate: nowMsInsights(),
				Count:      1,
				Data:       fmt.Sprintf("%.1f", stats.CPUPercent),
			}
		} else {
			// Update existing trigger
			trigger.LastUpdate = nowMsInsights()
			trigger.Count++
			trigger.Data = fmt.Sprintf("%.1f", stats.CPUPercent)

			// Generate insight if high for more than 10 minutes and not already triggered
			if !trigger.Triggered && (nowMsInsights()-trigger.StartTime) > 10*60*1000 {
				insight := models.AutoInsight{
					ID:        generateInsightID(),
					Title:     "Sustained High CPU Usage",
					Message:   fmt.Sprintf("CPU has been above 90%% for %d minutes (currently %.1f%%) - this may impact performance", (nowMsInsights()-trigger.StartTime)/(60*1000), stats.CPUPercent),
					Category:  "performance",
					Severity:  "warning",
					Timestamp: nowMsInsights(),
					IsRead:    false,
					Data:      trigger.Data,
					ActionItems: []string{
						"Check top CPU-consuming processes",
						"Look for runaway or stuck processes",
						"Consider restarting high-usage applications",
					},
				}
				newInsights = append(newInsights, insight)
				trigger.Triggered = true
			}
		}
	} else {
		// Clear CPU trigger if usage dropped below 85% (hysteresis)
		if trigger := is.getTrigger("cpu_high"); trigger != nil && stats.CPUPercent < 85 {
			delete(is.triggers, "cpu_high")
		}
	}

	// 2. High Memory Usage Detection (more conservative)
	if stats.MemoryPercent > 92 {
		triggerKey := "memory_high"
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

			if !trigger.Triggered && (nowMsInsights()-trigger.StartTime) > 5*60*1000 {
				insight := models.AutoInsight{
					ID:        generateInsightID(),
					Title:     "Critical Memory Usage",
					Message:   fmt.Sprintf("Memory usage critically high at %.1f%% for %d minutes - system may become unstable", stats.MemoryPercent, (nowMsInsights()-trigger.StartTime)/(60*1000)),
					Category:  "performance",
					Severity:  "critical",
					Timestamp: nowMsInsights(),
					IsRead:    false,
					Data:      trigger.Data,
					ActionItems: []string{
						"Close memory-intensive applications immediately",
						"Check for memory leaks in running processes",
						"Consider restarting the system if issues persist",
					},
				}
				newInsights = append(newInsights, insight)
				trigger.Triggered = true
			}
		}
	} else {
		// Clear memory trigger if usage dropped below 88% (hysteresis)
		if trigger := is.getTrigger("memory_high"); trigger != nil && stats.MemoryPercent < 88 {
			delete(is.triggers, "memory_high")
		}
	}

	// 3. New Process Detection
	if len(processes) > 0 {
		is.detectNewProcesses(processes, &newInsights)
	}

	// 4. Network Connection Surge
	if security != nil && len(security.UnknownConns) > 0 {
		is.detectConnectionSurge(security, &newInsights)
	}

	// 5. Suspicious Process Detection
	if security != nil && len(security.SuspiciousProcs) > 0 {
		is.detectSuspiciousActivity(security, &newInsights)
	}

	// Store new insights
	for _, insight := range newInsights {
		is.insights = append(is.insights, insight)
	}

	if len(newInsights) > 0 {
		is.saveInsights()
	}

	is.lastStats = stats
	return newInsights
}

// detectNewProcesses identifies when new processes start consuming excessive resources
func (is *InsightService) detectNewProcesses(processes []models.ProcessInfo, newInsights *[]models.AutoInsight) {
	// Only alert for truly concerning processes - much higher thresholds
	for _, proc := range processes {
		// More conservative: 40% CPU AND 500MB+ memory (or 60% CPU regardless of memory)
		isHighUsage := (proc.CPUPercent > 40 && proc.MemoryMB > 500) || proc.CPUPercent > 60

		if isHighUsage {
			triggerKey := fmt.Sprintf("high_usage_process_%d", proc.PID)
			if is.getTrigger(triggerKey) == nil {

				// Skip common system processes and browsers that are expected to use resources
				commonProcesses := []string{"chrome", "firefox", "safari", "code", "docker", "node", "python", "go", "java"}
				skip := false
				for _, common := range commonProcesses {
					if len(proc.Name) >= len(common) && proc.Name[:len(common)] == common {
						skip = true
						break
					}
				}

				if !skip {
					severity := "info"
					if proc.CPUPercent > 60 {
						severity = "warning"
					}

					insight := models.AutoInsight{
						ID:        generateInsightID(),
						Title:     "Unusually High Resource Process",
						Message:   fmt.Sprintf("Process '%s' (PID %d) is consuming %.1f%% CPU and %.1f MB memory", proc.Name, proc.PID, proc.CPUPercent, proc.MemoryMB),
						Category:  "process",
						Severity:  severity,
						Timestamp: nowMsInsights(),
						IsRead:    false,
						Data:      fmt.Sprintf(`{"name":"%s","pid":%d,"cpu":%.1f,"memory":%.1f}`, proc.Name, proc.PID, proc.CPUPercent, proc.MemoryMB),
						ActionItems: []string{
							fmt.Sprintf("Investigate why %s is using high resources", proc.Name),
							"Check if this is expected behavior for this application",
							"Consider terminating if it appears to be stuck or malfunctioning",
						},
					}
					*newInsights = append(*newInsights, insight)

					// Track this process
					is.triggers[triggerKey] = &models.InsightTrigger{
						Type:       "high_usage_process",
						StartTime:  nowMsInsights(),
						LastUpdate: nowMsInsights(),
						Count:      1,
						Data:       proc.Name,
						Triggered:  true,
					}
				}
			}
		}
	}
}

// detectConnectionSurge identifies unusual network activity (less paranoid)
func (is *InsightService) detectConnectionSurge(security *models.SecurityInfo, newInsights *[]models.AutoInsight) {
	connCount := len(security.UnknownConns)
	triggerKey := "connection_surge"

	// Much higher threshold - 25+ connections, and only warn if they're to many different countries
	if connCount > 25 {
		trigger := is.getTrigger(triggerKey)
		if trigger == nil || !trigger.Triggered {
			// Count connections by country for insight
			countryMap := make(map[string]int)
			for _, conn := range security.UnknownConns {
				if conn.Country != "" && conn.Country != "Unknown" {
					countryMap[conn.Country]++
				}
			}

			// Only alert if connections are to many different countries (potential concern)
			if len(countryMap) > 5 {
				countries := []string{}
				for country, count := range countryMap {
					if count > 2 {
						countries = append(countries, fmt.Sprintf("%s (%d)", country, count))
					}
				}

				severity := "info"
				if len(countryMap) > 10 {
					severity = "warning"
				}

				insight := models.AutoInsight{
					ID:        generateInsightID(),
					Title:     "Extensive Network Activity",
					Message:   fmt.Sprintf("Your system has %d active external connections to %d different countries", connCount, len(countryMap)),
					Category:  "network",
					Severity:  severity,
					Timestamp: nowMsInsights(),
					IsRead:    false,
					Data:      fmt.Sprintf(`{"count":%d,"countries":%d,"country_list":%v}`, connCount, len(countryMap), countries),
					ActionItems: []string{
						"This is usually normal for modern applications",
						"Check if you're running torrents, cloud sync, or streaming apps",
						"Only investigate if you notice performance issues",
					},
				}
				*newInsights = append(*newInsights, insight)

				is.triggers[triggerKey] = &models.InsightTrigger{
					Type:       triggerKey,
					StartTime:  nowMsInsights(),
					LastUpdate: nowMsInsights(),
					Count:      1,
					Triggered:  true,
				}
			}
		}
	}
}

// detectSuspiciousActivity identifies potential security concerns (less paranoid)
func (is *InsightService) detectSuspiciousActivity(security *models.SecurityInfo, newInsights *[]models.AutoInsight) {
	for _, suspProc := range security.SuspiciousProcs {
		triggerKey := fmt.Sprintf("suspicious_%d", suspProc.PID)
		if is.getTrigger(triggerKey) == nil {

			// Only alert on high-risk processes, ignore medium/low risk to reduce false alarms
			if suspProc.RiskLevel == "high" || suspProc.RiskLevel == "critical" {
				severity := "warning"
				if suspProc.RiskLevel == "critical" {
					severity = "critical"
				}

				actionItems := []string{
					fmt.Sprintf("Investigate %s process - flagged as %s risk", suspProc.Name, suspProc.RiskLevel),
					"Verify this is a legitimate application you installed",
					"Check the process file location and digital signature",
				}

				if suspProc.RiskLevel == "critical" {
					actionItems = append(actionItems, "Consider immediately terminating this process")
				} else {
					actionItems = append(actionItems, "Monitor this process closely for unusual behavior")
				}

				insight := models.AutoInsight{
					ID:          generateInsightID(),
					Title:       fmt.Sprintf("%s Risk Process Detected", strings.Title(suspProc.RiskLevel)),
					Message:     fmt.Sprintf("Process '%s' has been flagged as %s risk by security analysis", suspProc.Name, suspProc.RiskLevel),
					Category:    "security",
					Severity:    severity,
					Timestamp:   nowMsInsights(),
					IsRead:      false,
					Data:        fmt.Sprintf(`{"name":"%s","pid":%d,"risk":"%s"}`, suspProc.Name, suspProc.PID, suspProc.RiskLevel),
					ActionItems: actionItems,
				}
				*newInsights = append(*newInsights, insight)

				is.triggers[triggerKey] = &models.InsightTrigger{
					Type:       "suspicious_process",
					StartTime:  nowMsInsights(),
					LastUpdate: nowMsInsights(),
					Count:      1,
					Triggered:  true,
				}
			}
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
