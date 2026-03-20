package services

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"sysmind/internal/models"
)

const (
	baselineSampleInterval = 30 * time.Second
	baselinePortRetention  = 30 * 24 * time.Hour
)

type BaselineService struct {
	filePath string
	profile  baselineProfile
	mu       sync.RWMutex
}

type baselineProfile struct {
	FirstSampleAt int64                             `json:"firstSampleAt"`
	LastSampleAt  int64                             `json:"lastSampleAt"`
	SampleCount   int                               `json:"sampleCount"`
	GlobalMetrics map[string]rollingStat            `json:"globalMetrics"`
	MetricSlots   map[string]map[int]rollingStat    `json:"metricSlots"`
	Processes     map[string]baselineProcessProfile `json:"processes"`
	Ports         map[string]baselinePortProfile    `json:"ports"`
}

type rollingStat struct {
	Count int     `json:"count"`
	Mean  float64 `json:"mean"`
	M2    float64 `json:"m2"`
	Max   float64 `json:"max"`
}

type baselineProcessProfile struct {
	Name        string      `json:"name"`
	FirstSeen   int64       `json:"firstSeen"`
	LastSeen    int64       `json:"lastSeen"`
	SampleCount int         `json:"sampleCount"`
	LastPID     int32       `json:"lastPid"`
	CPU         rollingStat `json:"cpu"`
	Memory      rollingStat `json:"memory"`
}

type baselinePortProfile struct {
	Key         string `json:"key"`
	Port        uint32 `json:"port"`
	Protocol    string `json:"protocol"`
	ProcessName string `json:"processName"`
	FirstSeen   int64  `json:"firstSeen"`
	LastSeen    int64  `json:"lastSeen"`
	SampleCount int    `json:"sampleCount"`
}

func NewBaselineService() (*BaselineService, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}

	dataDir := filepath.Join(configDir, "sysmind", "baseline")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	bs := &BaselineService{
		filePath: filepath.Join(dataDir, "profile.json"),
		profile: baselineProfile{
			GlobalMetrics: map[string]rollingStat{},
			MetricSlots:   map[string]map[int]rollingStat{},
			Processes:     map[string]baselineProcessProfile{},
			Ports:         map[string]baselinePortProfile{},
		},
	}
	bs.load()
	return bs, nil
}

func (bs *BaselineService) BootstrapFromTimeMachine(samples []models.TimeMachineSample) {
	if len(samples) == 0 {
		return
	}

	bs.mu.Lock()
	if bs.profile.SampleCount > 0 {
		bs.mu.Unlock()
		return
	}

	for _, sample := range samples {
		stats := &models.SystemStats{
			Timestamp:      sample.Timestamp,
			CPUPercent:     sample.CPUPercent,
			MemoryPercent:  sample.MemoryPercent,
			DiskPercent:    sample.DiskPercent,
			DiskUsedGB:     sample.DiskUsedGB,
			DiskTotalGB:    sample.DiskTotalGB,
			NetUploadSpeed: sample.NetUploadSpeed,
			NetDownSpeed:   sample.NetDownSpeed,
		}
		bs.updateLocked(stats, incidentProcessesToProcessInfo(sample.Processes), nil)
	}

	snapshot := bs.profile
	bs.mu.Unlock()
	_ = bs.save(snapshot)
}

func (bs *BaselineService) Save() {
	bs.mu.RLock()
	snapshot := bs.profile
	bs.mu.RUnlock()
	_ = bs.save(snapshot)
}

func (bs *BaselineService) ShouldCapture(timestamp int64) bool {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	if bs.profile.LastSampleAt == 0 {
		return true
	}
	return timestamp-bs.profile.LastSampleAt >= int64(baselineSampleInterval/time.Millisecond)
}

func (bs *BaselineService) Update(stats *models.SystemStats, processes []models.ProcessInfo, ports []models.PortInfo) {
	if stats == nil {
		return
	}

	bs.mu.Lock()
	if bs.profile.LastSampleAt != 0 && stats.Timestamp-bs.profile.LastSampleAt < int64(baselineSampleInterval/time.Millisecond) {
		bs.mu.Unlock()
		return
	}

	bs.updateLocked(stats, processes, ports)
	snapshot := bs.profile
	bs.mu.Unlock()

	_ = bs.save(snapshot)
}

func (bs *BaselineService) GetDriftView(stats *models.SystemStats, processes []models.ProcessInfo, ports []models.PortInfo) models.BaselineDriftView {
	now := time.Now().UnixMilli()
	if stats != nil && stats.Timestamp > 0 {
		now = stats.Timestamp
	}

	bs.mu.RLock()
	profile := bs.profile
	bs.mu.RUnlock()

	view := models.BaselineDriftView{
		GeneratedAt:   now,
		CoverageHours: baselineCoverageHours(profile),
		SampleCount:   profile.SampleCount,
		Learning:      profile.SampleCount < 24,
		Findings:      []models.BaselineDriftFinding{},
	}

	if profile.SampleCount == 0 {
		view.Learning = true
		view.Summary = "Learning your machine's normal behavior. Findings will appear after baseline samples accumulate."
		return view
	}

	timestamp := now
	if stats != nil && stats.Timestamp > 0 {
		timestamp = stats.Timestamp
	}
	slot := baselineSlot(time.UnixMilli(timestamp))

	findings := []models.BaselineDriftFinding{}
	if stats != nil {
		findings = append(findings, bs.metricFindings(profile, slot, *stats, timestamp)...)
	}
	findings = append(findings, bs.processFindings(profile, processes, timestamp)...)
	findings = append(findings, bs.portFindings(profile, ports, timestamp)...)

	sort.Slice(findings, func(i, j int) bool {
		if severityWeight(findings[i].Severity) != severityWeight(findings[j].Severity) {
			return severityWeight(findings[i].Severity) > severityWeight(findings[j].Severity)
		}
		if findings[i].Confidence != findings[j].Confidence {
			return findings[i].Confidence > findings[j].Confidence
		}
		return findings[i].CurrentValue > findings[j].CurrentValue
	})
	if len(findings) > 6 {
		findings = findings[:6]
	}

	view.Findings = findings
	view.Summary = buildBaselineSummary(view)
	return view
}

func (bs *BaselineService) updateLocked(stats *models.SystemStats, processes []models.ProcessInfo, ports []models.PortInfo) {
	bs.ensureMapsLocked()

	if bs.profile.FirstSampleAt == 0 {
		bs.profile.FirstSampleAt = stats.Timestamp
	}
	bs.profile.LastSampleAt = stats.Timestamp
	bs.profile.SampleCount++

	slot := baselineSlot(time.UnixMilli(stats.Timestamp))
	bs.updateMetricLocked("cpu", slot, stats.CPUPercent)
	bs.updateMetricLocked("memory", slot, stats.MemoryPercent)
	bs.updateMetricLocked("disk", slot, stats.DiskPercent)
	bs.updateMetricLocked("network", slot, stats.NetUploadSpeed+stats.NetDownSpeed)

	for _, proc := range processes {
		if !shouldRecordProcess(proc) {
			continue
		}

		key := normalizeName(proc.Name)
		profile := bs.profile.Processes[key]
		if profile.Name == "" {
			profile = baselineProcessProfile{
				Name:      proc.Name,
				FirstSeen: stats.Timestamp,
			}
		}
		profile.LastSeen = stats.Timestamp
		profile.LastPID = proc.PID
		profile.SampleCount++
		profile.CPU.update(proc.CPUPercent)
		profile.Memory.update(proc.MemoryMB)
		bs.profile.Processes[key] = profile
	}

	for _, port := range ports {
		if !shouldTrackPort(port) {
			continue
		}

		key := baselinePortKey(port)
		profile := bs.profile.Ports[key]
		if profile.Key == "" {
			profile = baselinePortProfile{
				Key:         key,
				Port:        port.Port,
				Protocol:    port.Protocol,
				ProcessName: port.ProcessName,
				FirstSeen:   stats.Timestamp,
			}
		}
		profile.LastSeen = stats.Timestamp
		profile.SampleCount++
		bs.profile.Ports[key] = profile
	}

	bs.prunePortsLocked(stats.Timestamp)
}

func (bs *BaselineService) updateMetricLocked(metric string, slot int, value float64) {
	global := bs.profile.GlobalMetrics[metric]
	global.update(value)
	bs.profile.GlobalMetrics[metric] = global

	slots := bs.profile.MetricSlots[metric]
	if slots == nil {
		slots = map[int]rollingStat{}
		bs.profile.MetricSlots[metric] = slots
	}

	slotStats := slots[slot]
	slotStats.update(value)
	slots[slot] = slotStats
}

func (bs *BaselineService) metricFindings(profile baselineProfile, slot int, stats models.SystemStats, timestamp int64) []models.BaselineDriftFinding {
	type metricInput struct {
		name    string
		label   string
		value   float64
		unit    string
		current string
	}

	inputs := []metricInput{
		{name: "cpu", label: "CPU", value: stats.CPUPercent, unit: "%", current: "CPU usage"},
		{name: "memory", label: "Memory", value: stats.MemoryPercent, unit: "%", current: "memory usage"},
		{name: "disk", label: "Disk", value: stats.DiskPercent, unit: "%", current: "disk usage"},
		{name: "network", label: "Network", value: stats.NetUploadSpeed + stats.NetDownSpeed, unit: "B/s", current: "network traffic"},
	}

	findings := []models.BaselineDriftFinding{}
	for _, input := range inputs {
		bucket, sourceCount := baselineMetricStats(profile, input.name, slot)
		if sourceCount < 8 {
			continue
		}

		stdDev := bucket.stdDev()
		expectedHigh := bucket.Mean + math.Max(2.8*stdDev, metricDriftFloor(input.name, bucket.Mean))
		if input.value <= expectedHigh || input.value <= metricAbsoluteFloor(input.name) {
			continue
		}

		severity := "info"
		if input.value > expectedHigh*1.35 || input.name == "disk" && input.value > 92 {
			severity = "warning"
		}
		if input.value > expectedHigh*1.6 || (input.name == "cpu" && input.value > 95) || (input.name == "memory" && input.value > 96) {
			severity = "critical"
		}

		deltaPercent := percentageAboveBaseline(input.value, bucket.Mean)
		title := fmt.Sprintf("Unusual %s", input.current)
		summary := fmt.Sprintf("%s is %s right now, versus a usual %s for this time slot.", input.label, formatBaselineValue(input.value, input.unit), formatBaselineValue(bucket.Mean, input.unit))
		if input.name == "network" {
			title = "Unusual network burst"
			summary = fmt.Sprintf("Network traffic is %s right now, above the usual %s for this time slot.", formatBaselineValue(input.value, input.unit), formatBaselineValue(bucket.Mean, input.unit))
		}

		findings = append(findings, models.BaselineDriftFinding{
			ID:            fmt.Sprintf("metric-%s", input.name),
			Kind:          "metric-drift",
			Category:      "performance",
			Severity:      severity,
			Title:         title,
			Summary:       summary,
			Metric:        input.name,
			Unit:          input.unit,
			CurrentValue:  input.value,
			BaselineValue: bucket.Mean,
			ExpectedHigh:  expectedHigh,
			DeltaPercent:  deltaPercent,
			LastSeenAt:    timestamp,
			SampleCount:   sourceCount,
			Confidence:    driftConfidence(sourceCount, input.value, expectedHigh),
		})
	}

	return findings
}

func (bs *BaselineService) processFindings(profile baselineProfile, processes []models.ProcessInfo, timestamp int64) []models.BaselineDriftFinding {
	findings := []models.BaselineDriftFinding{}
	for _, proc := range rankedProcessesForDrift(processes) {
		key := normalizeName(proc.Name)
		baseline, exists := profile.Processes[key]

		if !exists && (proc.CPUPercent >= 35 || proc.MemoryMB >= 1200) {
			severity := "info"
			if proc.CPUPercent >= 60 || proc.MemoryMB >= 2400 {
				severity = "warning"
			}
			findings = append(findings, models.BaselineDriftFinding{
				ID:            fmt.Sprintf("new-process-%s", sanitizeID(key)),
				Kind:          "new-process",
				Category:      "process",
				Severity:      severity,
				Title:         fmt.Sprintf("New heavy process: %s", friendlyProcessLabel(proc.Name)),
				Summary:       fmt.Sprintf("%s is active with %.1f%% CPU and %.0f MB memory, but it is not part of the learned baseline yet.", friendlyProcessLabel(proc.Name), proc.CPUPercent, proc.MemoryMB),
				ProcessName:   proc.Name,
				ProcessPID:    proc.PID,
				CurrentValue:  math.Max(proc.CPUPercent, proc.MemoryMB),
				BaselineValue: 0,
				DeltaPercent:  100,
				FirstSeenAt:   timestamp,
				LastSeenAt:    timestamp,
				SampleCount:   0,
				Confidence:    0.64,
				IsNew:         true,
			})
			continue
		}

		if !exists || baseline.SampleCount < 8 {
			continue
		}

		cpuExpectedHigh := baseline.CPU.Mean + math.Max(3*baseline.CPU.stdDev(), 20)
		if proc.CPUPercent > cpuExpectedHigh && proc.CPUPercent > 25 {
			findings = append(findings, models.BaselineDriftFinding{
				ID:            fmt.Sprintf("process-cpu-%s", sanitizeID(key)),
				Kind:          "process-cpu-drift",
				Category:      "process",
				Severity:      processSeverity("cpu", proc.CPUPercent, cpuExpectedHigh),
				Title:         fmt.Sprintf("%s is using unusually high CPU", friendlyProcessLabel(proc.Name)),
				Summary:       fmt.Sprintf("%s is at %.1f%% CPU, above its usual %.1f%% peak range.", friendlyProcessLabel(proc.Name), proc.CPUPercent, cpuExpectedHigh),
				Metric:        "cpu",
				Unit:          "%",
				ProcessName:   proc.Name,
				ProcessPID:    proc.PID,
				CurrentValue:  proc.CPUPercent,
				BaselineValue: baseline.CPU.Mean,
				ExpectedHigh:  cpuExpectedHigh,
				DeltaPercent:  percentageAboveBaseline(proc.CPUPercent, baseline.CPU.Mean),
				FirstSeenAt:   baseline.FirstSeen,
				LastSeenAt:    timestamp,
				SampleCount:   baseline.SampleCount,
				Confidence:    driftConfidence(baseline.SampleCount, proc.CPUPercent, cpuExpectedHigh),
			})
		}

		memExpectedHigh := baseline.Memory.Mean + math.Max(3*baseline.Memory.stdDev(), 400)
		if proc.MemoryMB > memExpectedHigh && proc.MemoryMB > 600 {
			findings = append(findings, models.BaselineDriftFinding{
				ID:            fmt.Sprintf("process-memory-%s", sanitizeID(key)),
				Kind:          "process-memory-drift",
				Category:      "process",
				Severity:      processSeverity("memory", proc.MemoryMB, memExpectedHigh),
				Title:         fmt.Sprintf("%s is using unusually high memory", friendlyProcessLabel(proc.Name)),
				Summary:       fmt.Sprintf("%s is at %.0f MB memory, above its usual %.0f MB peak range.", friendlyProcessLabel(proc.Name), proc.MemoryMB, memExpectedHigh),
				Metric:        "memory",
				Unit:          "MB",
				ProcessName:   proc.Name,
				ProcessPID:    proc.PID,
				CurrentValue:  proc.MemoryMB,
				BaselineValue: baseline.Memory.Mean,
				ExpectedHigh:  memExpectedHigh,
				DeltaPercent:  percentageAboveBaseline(proc.MemoryMB, baseline.Memory.Mean),
				FirstSeenAt:   baseline.FirstSeen,
				LastSeenAt:    timestamp,
				SampleCount:   baseline.SampleCount,
				Confidence:    driftConfidence(baseline.SampleCount, proc.MemoryMB, memExpectedHigh),
			})
		}
	}
	return findings
}

func (bs *BaselineService) portFindings(profile baselineProfile, ports []models.PortInfo, timestamp int64) []models.BaselineDriftFinding {
	findings := []models.BaselineDriftFinding{}
	for _, port := range ports {
		if !shouldTrackPort(port) {
			continue
		}

		key := baselinePortKey(port)
		if _, exists := profile.Ports[key]; exists {
			continue
		}

		severity := "info"
		if port.Port < 1024 || strings.EqualFold(port.ProcessName, "python") || strings.EqualFold(port.ProcessName, "node") {
			severity = "warning"
		}

		findings = append(findings, models.BaselineDriftFinding{
			ID:          fmt.Sprintf("new-port-%s", sanitizeID(key)),
			Kind:        "new-port",
			Category:    "network",
			Severity:    severity,
			Title:       fmt.Sprintf("New listening port: %d/%s", port.Port, port.Protocol),
			Summary:     fmt.Sprintf("%s opened %d/%s and this listener is new to the baseline.", friendlyProcessLabel(port.ProcessName), port.Port, strings.ToUpper(port.Protocol)),
			ProcessName: port.ProcessName,
			ProcessPID:  port.PID,
			Port:        port.Port,
			Protocol:    port.Protocol,
			FirstSeenAt: timestamp,
			LastSeenAt:  timestamp,
			SampleCount: 0,
			Confidence:  0.7,
			IsNew:       true,
		})
	}
	return findings
}

func buildBaselineSummary(view models.BaselineDriftView) string {
	if view.SampleCount == 0 {
		return "Learning your machine's normal behavior."
	}
	if len(view.Findings) == 0 {
		if view.Learning {
			return "Baseline is still learning, but nothing currently stands out as new or unusual."
		}
		return "Nothing currently stands out against this machine's learned baseline."
	}

	newCount := 0
	for _, finding := range view.Findings {
		if finding.IsNew {
			newCount++
		}
	}
	parts := []string{fmt.Sprintf("%d unusual finding%s", len(view.Findings), pluralSuffix(len(view.Findings)))}
	if newCount > 0 {
		parts = append(parts, fmt.Sprintf("%d new", newCount))
	}
	return strings.Join(parts, ", ") + "."
}

func baselineMetricStats(profile baselineProfile, metric string, slot int) (rollingStat, int) {
	if slots, ok := profile.MetricSlots[metric]; ok {
		if slotStats, ok := slots[slot]; ok && slotStats.Count >= 8 {
			return slotStats, slotStats.Count
		}
	}
	global := profile.GlobalMetrics[metric]
	return global, global.Count
}

func baselineCoverageHours(profile baselineProfile) int {
	if profile.FirstSampleAt == 0 || profile.LastSampleAt <= profile.FirstSampleAt {
		return 0
	}
	return int(math.Round(float64(profile.LastSampleAt-profile.FirstSampleAt) / float64(time.Hour/time.Millisecond)))
}

func baselineSlot(t time.Time) int {
	return int(t.Weekday())*24 + t.Hour()
}

func metricDriftFloor(metric string, mean float64) float64 {
	switch metric {
	case "cpu":
		return math.Max(18, mean*0.5)
	case "memory":
		return math.Max(12, mean*0.2)
	case "disk":
		return math.Max(10, mean*0.15)
	case "network":
		return math.Max(512*1024, mean*0.75)
	default:
		return math.Max(10, mean*0.25)
	}
}

func metricAbsoluteFloor(metric string) float64 {
	switch metric {
	case "cpu":
		return 35
	case "memory":
		return 55
	case "disk":
		return 70
	case "network":
		return 1024 * 1024
	default:
		return 0
	}
}

func processSeverity(metric string, current, expected float64) string {
	if current > expected*1.7 {
		return "critical"
	}
	if current > expected*1.35 {
		return "warning"
	}
	if metric == "memory" && current > 2500 {
		return "warning"
	}
	return "info"
}

func formatBaselineValue(value float64, unit string) string {
	switch unit {
	case "%":
		return fmt.Sprintf("%.1f%%", value)
	case "MB":
		return fmt.Sprintf("%.0f MB", value)
	case "B/s":
		if value < 1024 {
			return fmt.Sprintf("%.0f B/s", value)
		}
		if value < 1024*1024 {
			return fmt.Sprintf("%.1f KB/s", value/1024)
		}
		return fmt.Sprintf("%.1f MB/s", value/1024/1024)
	default:
		return fmt.Sprintf("%.1f %s", value, unit)
	}
}

func percentageAboveBaseline(current, baseline float64) float64 {
	if baseline <= 0 {
		return 100
	}
	return math.Max(0, ((current-baseline)/baseline)*100)
}

func driftConfidence(sampleCount int, current, expected float64) float64 {
	if expected <= 0 {
		return 0.5
	}
	ratio := current / expected
	confidence := 0.35 + math.Min(0.35, float64(sampleCount)/40) + math.Min(0.25, (ratio-1)*0.3)
	return math.Max(0.25, math.Min(0.97, confidence))
}

func rankedProcessesForDrift(processes []models.ProcessInfo) []models.ProcessInfo {
	ranked := make([]models.ProcessInfo, 0, len(processes))
	for _, proc := range processes {
		if proc.Name == "" {
			continue
		}
		ranked = append(ranked, proc)
	}

	sort.Slice(ranked, func(i, j int) bool {
		iScore := ranked[i].CPUPercent*10 + ranked[i].MemoryMB/128
		jScore := ranked[j].CPUPercent*10 + ranked[j].MemoryMB/128
		return iScore > jScore
	})
	if len(ranked) > 20 {
		ranked = ranked[:20]
	}
	return ranked
}

func shouldRecordProcess(proc models.ProcessInfo) bool {
	if proc.Name == "" {
		return false
	}
	return proc.CPUPercent >= 1 || proc.MemoryMB >= 64 || proc.NumThreads >= 8
}

func shouldTrackPort(port models.PortInfo) bool {
	state := strings.ToUpper(port.State)
	return state == "LISTENING" || (port.Protocol == "udp" && port.Port != 0)
}

func baselinePortKey(port models.PortInfo) string {
	return fmt.Sprintf("%s:%d:%s", strings.ToLower(port.Protocol), port.Port, normalizeName(port.ProcessName))
}

func normalizeName(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func incidentProcessesToProcessInfo(processes []models.IncidentProcessSample) []models.ProcessInfo {
	result := make([]models.ProcessInfo, 0, len(processes))
	for _, proc := range processes {
		result = append(result, models.ProcessInfo{
			PID:        proc.PID,
			Name:       proc.Name,
			CPUPercent: proc.CPUPercent,
			MemoryMB:   proc.MemoryMB,
			NumThreads: proc.NumThreads,
			Status:     proc.Status,
		})
	}
	return result
}

func (bs *BaselineService) prunePortsLocked(now int64) {
	cutoff := now - int64(baselinePortRetention/time.Millisecond)
	for key, port := range bs.profile.Ports {
		if port.LastSeen < cutoff {
			delete(bs.profile.Ports, key)
		}
	}
}

func (bs *BaselineService) ensureMapsLocked() {
	if bs.profile.GlobalMetrics == nil {
		bs.profile.GlobalMetrics = map[string]rollingStat{}
	}
	if bs.profile.MetricSlots == nil {
		bs.profile.MetricSlots = map[string]map[int]rollingStat{}
	}
	if bs.profile.Processes == nil {
		bs.profile.Processes = map[string]baselineProcessProfile{}
	}
	if bs.profile.Ports == nil {
		bs.profile.Ports = map[string]baselinePortProfile{}
	}
}

func (rs *rollingStat) update(value float64) {
	rs.Count++
	if rs.Count == 1 {
		rs.Mean = value
		rs.Max = value
		rs.M2 = 0
		return
	}
	delta := value - rs.Mean
	rs.Mean += delta / float64(rs.Count)
	rs.M2 += delta * (value - rs.Mean)
	if value > rs.Max {
		rs.Max = value
	}
}

func (rs rollingStat) stdDev() float64 {
	if rs.Count < 2 {
		return 0
	}
	return math.Sqrt(rs.M2 / float64(rs.Count-1))
}

func (bs *BaselineService) save(profile baselineProfile) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	return os.WriteFile(bs.filePath, data, 0644)
}

func (bs *BaselineService) load() {
	data, err := os.ReadFile(bs.filePath)
	if err != nil {
		return
	}

	if err := json.Unmarshal(data, &bs.profile); err != nil {
		return
	}
	bs.ensureMapsLocked()
}
