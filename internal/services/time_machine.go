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
	timeMachineSampleInterval = 30 * time.Second
	timeMachineRetention      = 72 * time.Hour
	timeMachineProcessLimit   = 4
)

// TimeMachineService manages persisted lower-frequency telemetry for historical playback.
type TimeMachineService struct {
	filePath string
	samples  []models.TimeMachineSample
	mu       sync.RWMutex
}

// NewTimeMachineService creates a persisted time machine history store.
func NewTimeMachineService() (*TimeMachineService, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}

	dataDir := filepath.Join(configDir, "sysmind", "time-machine")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	tms := &TimeMachineService{
		filePath: filepath.Join(dataDir, "history.json"),
		samples:  []models.TimeMachineSample{},
	}
	tms.load()
	return tms, nil
}

// AppendSample stores a time machine sample if the sampling interval has elapsed.
func (tms *TimeMachineService) AppendSample(stats *models.SystemStats, processes []models.ProcessInfo) {
	if stats == nil {
		return
	}

	tms.mu.Lock()
	if len(tms.samples) > 0 {
		lastTs := tms.samples[len(tms.samples)-1].Timestamp
		if stats.Timestamp-lastTs < int64(timeMachineSampleInterval/time.Millisecond) {
			tms.mu.Unlock()
			return
		}
	}

	tms.samples = append(tms.samples, BuildTimeMachineSample(stats, processes))
	tms.pruneLocked()

	snapshot := make([]models.TimeMachineSample, len(tms.samples))
	copy(snapshot, tms.samples)
	tms.mu.Unlock()

	_ = tms.save(snapshot)
}

// Save flushes the current history to disk.
func (tms *TimeMachineService) Save() {
	tms.mu.RLock()
	snapshot := make([]models.TimeMachineSample, len(tms.samples))
	copy(snapshot, tms.samples)
	tms.mu.RUnlock()
	_ = tms.save(snapshot)
}

// SamplesSnapshot returns a copy of persisted time-machine samples for bootstrap use.
func (tms *TimeMachineService) SamplesSnapshot() []models.TimeMachineSample {
	tms.mu.RLock()
	defer tms.mu.RUnlock()

	snapshot := make([]models.TimeMachineSample, len(tms.samples))
	copy(snapshot, tms.samples)
	return snapshot
}

// GetView returns a filtered time machine view with annotations and forecasts.
func (tms *TimeMachineService) GetView(windowHours int) models.TimeMachineView {
	if windowHours <= 0 {
		windowHours = 6
	}
	if windowHours > int(timeMachineRetention.Hours()) {
		windowHours = int(timeMachineRetention.Hours())
	}

	cutoff := time.Now().Add(-time.Duration(windowHours) * time.Hour).UnixMilli()

	tms.mu.RLock()
	filtered := make([]models.TimeMachineSample, 0, len(tms.samples))
	for _, sample := range tms.samples {
		if sample.Timestamp >= cutoff {
			filtered = append(filtered, sample)
		}
	}
	tms.mu.RUnlock()

	annotations := detectTimeMachineAnnotations(filtered)
	forecasts := detectTimeMachineForecasts(filtered)

	view := models.TimeMachineView{
		WindowHours:        windowHours,
		RetentionHours:     int(timeMachineRetention.Hours()),
		SamplingSeconds:    int(timeMachineSampleInterval / time.Second),
		Samples:            filtered,
		Annotations:        annotations,
		Forecasts:          forecasts,
		Summary:            buildTimeMachineSummary(filtered, annotations, forecasts),
		PersistenceEnabled: true,
	}
	if len(filtered) > 0 {
		view.LastUpdated = filtered[len(filtered)-1].Timestamp
	}

	return view
}

// BuildTimeMachineSample creates a persisted sample from current telemetry.
func BuildTimeMachineSample(stats *models.SystemStats, processes []models.ProcessInfo) models.TimeMachineSample {
	sample := models.TimeMachineSample{
		Timestamp:      stats.Timestamp,
		CPUPercent:     stats.CPUPercent,
		MemoryPercent:  stats.MemoryPercent,
		DiskPercent:    stats.DiskPercent,
		DiskUsedGB:     stats.DiskUsedGB,
		DiskTotalGB:    stats.DiskTotalGB,
		NetUploadSpeed: stats.NetUploadSpeed,
		NetDownSpeed:   stats.NetDownSpeed,
		LoadAvg1:       stats.LoadAvg1,
		Processes:      []models.IncidentProcessSample{},
	}

	if len(processes) == 0 {
		return sample
	}

	ranked := make([]models.ProcessInfo, len(processes))
	copy(ranked, processes)
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].MemoryMB != ranked[j].MemoryMB {
			return ranked[i].MemoryMB > ranked[j].MemoryMB
		}
		if ranked[i].CPUPercent != ranked[j].CPUPercent {
			return ranked[i].CPUPercent > ranked[j].CPUPercent
		}
		return ranked[i].NumThreads > ranked[j].NumThreads
	})

	limit := timeMachineProcessLimit
	if len(ranked) < limit {
		limit = len(ranked)
	}

	for _, proc := range ranked[:limit] {
		sample.Processes = append(sample.Processes, models.IncidentProcessSample{
			PID:        proc.PID,
			Name:       proc.Name,
			CPUPercent: proc.CPUPercent,
			MemoryMB:   proc.MemoryMB,
			NumThreads: proc.NumThreads,
			Status:     proc.Status,
		})
	}

	return sample
}

func (tms *TimeMachineService) pruneLocked() {
	cutoff := time.Now().Add(-timeMachineRetention).UnixMilli()
	if len(tms.samples) == 0 {
		return
	}

	pruned := tms.samples[:0]
	for _, sample := range tms.samples {
		if sample.Timestamp >= cutoff {
			pruned = append(pruned, sample)
		}
	}
	tms.samples = pruned
}

func (tms *TimeMachineService) save(samples []models.TimeMachineSample) error {
	data, err := json.Marshal(samples)
	if err != nil {
		return err
	}
	return os.WriteFile(tms.filePath, data, 0644)
}

func (tms *TimeMachineService) load() {
	data, err := os.ReadFile(tms.filePath)
	if err != nil {
		return
	}

	var samples []models.TimeMachineSample
	if err := json.Unmarshal(data, &samples); err != nil {
		return
	}

	tms.samples = samples
	tms.pruneLocked()
}

func detectTimeMachineAnnotations(samples []models.TimeMachineSample) []models.TimeMachineAnnotation {
	annotations := []models.TimeMachineAnnotation{}
	annotations = append(annotations, detectMemoryLeakAnnotations(samples)...)
	annotations = append(annotations, detectSpikeAnnotations(samples)...)

	sort.Slice(annotations, func(i, j int) bool {
		if annotations[i].Timestamp != annotations[j].Timestamp {
			return annotations[i].Timestamp < annotations[j].Timestamp
		}
		return severityWeight(annotations[i].Severity) > severityWeight(annotations[j].Severity)
	})

	return dedupeAnnotations(annotations)
}

func detectMemoryLeakAnnotations(samples []models.TimeMachineSample) []models.TimeMachineAnnotation {
	type processPoint struct {
		timestamp int64
		proc      models.IncidentProcessSample
	}

	history := make(map[string][]processPoint)
	for _, sample := range samples {
		for _, proc := range sample.Processes {
			if proc.MemoryMB <= 0 || proc.Name == "" {
				continue
			}
			history[proc.Name] = append(history[proc.Name], processPoint{timestamp: sample.Timestamp, proc: proc})
		}
	}

	annotations := []models.TimeMachineAnnotation{}
	for name, points := range history {
		if len(points) < 6 {
			continue
		}

		first := points[0]
		last := points[len(points)-1]
		durationHours := float64(last.timestamp-first.timestamp) / float64(time.Hour/time.Millisecond)
		if durationHours < 0.5 {
			continue
		}

		growth := last.proc.MemoryMB - first.proc.MemoryMB
		if growth < 300 {
			continue
		}

		positiveSteps := 0
		for i := 1; i < len(points); i++ {
			if points[i].proc.MemoryMB >= points[i-1].proc.MemoryMB {
				positiveSteps++
			}
		}
		if float64(positiveSteps)/float64(len(points)-1) < 0.7 {
			continue
		}

		title := fmt.Sprintf("%s memory kept climbing", friendlyProcessLabel(name))
		if isNodeFamily(name) {
			title = "Node.js memory leak pattern"
		}

		annotations = append(annotations, models.TimeMachineAnnotation{
			ID:          fmt.Sprintf("memleak-%s-%d", sanitizeID(name), first.timestamp),
			Kind:        "memory-leak",
			Severity:    leakSeverity(growth, last.proc.MemoryMB),
			Timestamp:   first.timestamp,
			Title:       title,
			Summary:     fmt.Sprintf("%s grew from %.0f MB to %.0f MB over %s, which looks like leak-style growth.", friendlyProcessLabel(name), first.proc.MemoryMB, last.proc.MemoryMB, humanDurationHours(durationHours)),
			ProcessName: name,
			ProcessPID:  last.proc.PID,
			Metric:      "memory",
			Value:       last.proc.MemoryMB,
		})
	}

	sort.Slice(annotations, func(i, j int) bool {
		return annotations[i].Value > annotations[j].Value
	})
	if len(annotations) > 3 {
		annotations = annotations[:3]
	}
	return annotations
}

func detectSpikeAnnotations(samples []models.TimeMachineSample) []models.TimeMachineAnnotation {
	type metricPeak struct {
		kind      string
		metric    string
		title     string
		valueFunc func(models.TimeMachineSample) float64
	}

	metrics := []metricPeak{
		{
			kind:   "cpu-spike",
			metric: "cpu",
			title:  "CPU spike",
			valueFunc: func(sample models.TimeMachineSample) float64 {
				return sample.CPUPercent
			},
		},
		{
			kind:   "disk-spike",
			metric: "disk",
			title:  "Disk pressure spike",
			valueFunc: func(sample models.TimeMachineSample) float64 {
				return sample.DiskPercent
			},
		},
		{
			kind:   "network-spike",
			metric: "network",
			title:  "Network burst",
			valueFunc: func(sample models.TimeMachineSample) float64 {
				return sample.NetUploadSpeed + sample.NetDownSpeed
			},
		},
	}

	annotations := []models.TimeMachineAnnotation{}
	for _, metric := range metrics {
		values := make([]float64, 0, len(samples))
		for _, sample := range samples {
			values = append(values, metric.valueFunc(sample))
		}
		baseline, stdDev := robustBaseline(values)
		threshold := baseline + math.Max(2.5*stdDev, metricMinimumThreshold(metric.metric, baseline))

		lastAddedTs := int64(0)
		for i := 1; i < len(samples)-1; i++ {
			current := metric.valueFunc(samples[i])
			if current < threshold {
				continue
			}
			if current < metric.valueFunc(samples[i-1]) || current < metric.valueFunc(samples[i+1]) {
				continue
			}
			if lastAddedTs != 0 && samples[i].Timestamp-lastAddedTs < int64((20*time.Minute)/time.Millisecond) {
				continue
			}

			culprit := dominantProcess(samples[i], metric.metric)
			title, summary := spikeNarrative(metric.metric, culprit, current, baseline)
			annotations = append(annotations, models.TimeMachineAnnotation{
				ID:          fmt.Sprintf("%s-%d", metric.kind, samples[i].Timestamp),
				Kind:        metric.kind,
				Severity:    spikeSeverity(metric.metric, current),
				Timestamp:   samples[i].Timestamp,
				Title:       title,
				Summary:     summary,
				ProcessName: culprit.Name,
				ProcessPID:  culprit.PID,
				Metric:      metric.metric,
				Value:       current,
			})
			lastAddedTs = samples[i].Timestamp
		}
	}

	sort.Slice(annotations, func(i, j int) bool {
		if severityWeight(annotations[i].Severity) != severityWeight(annotations[j].Severity) {
			return severityWeight(annotations[i].Severity) > severityWeight(annotations[j].Severity)
		}
		return annotations[i].Timestamp < annotations[j].Timestamp
	})
	if len(annotations) > 6 {
		annotations = annotations[:6]
	}
	return annotations
}

func detectTimeMachineForecasts(samples []models.TimeMachineSample) []models.TimeMachineForecast {
	forecasts := []models.TimeMachineForecast{}
	if forecast, ok := forecastDiskFull(samples); ok {
		forecasts = append(forecasts, forecast)
	}
	if forecast, ok := forecastMemoryPressure(samples); ok {
		forecasts = append(forecasts, forecast)
	}

	sort.Slice(forecasts, func(i, j int) bool {
		return forecasts[i].PredictedAt < forecasts[j].PredictedAt
	})
	return forecasts
}

func forecastDiskFull(samples []models.TimeMachineSample) (models.TimeMachineForecast, bool) {
	recent := trailingSamplesByHours(samples, 6)
	if len(recent) < 6 {
		return models.TimeMachineForecast{}, false
	}
	totalGB := recent[len(recent)-1].DiskTotalGB
	currentGB := recent[len(recent)-1].DiskUsedGB
	if totalGB <= 0 || currentGB <= 0 || currentGB >= totalGB {
		return models.TimeMachineForecast{}, false
	}

	slopePerHour, confidence := regressionSlopeDiskGBPerHour(recent)
	if slopePerHour <= 0.05 {
		return models.TimeMachineForecast{}, false
	}

	hoursToFull := (totalGB - currentGB) / slopePerHour
	if hoursToFull <= 0 || hoursToFull > 24*14 {
		return models.TimeMachineForecast{}, false
	}

	predictedAt := time.UnixMilli(recent[len(recent)-1].Timestamp).Add(time.Duration(hoursToFull * float64(time.Hour))).UnixMilli()
	return models.TimeMachineForecast{
		ID:             "disk-full",
		Kind:           "disk-full",
		Severity:       forecastSeverity(hoursToFull),
		Title:          "Disk exhaustion forecast",
		Summary:        fmt.Sprintf("At the current disk growth rate, your primary disk will fill in about %s.", humanDurationHours(hoursToFull)),
		PredictedAt:    predictedAt,
		CurrentValue:   currentGB,
		ProjectedValue: totalGB,
		Confidence:     confidence,
		Unit:           "GB",
	}, true
}

func forecastMemoryPressure(samples []models.TimeMachineSample) (models.TimeMachineForecast, bool) {
	recent := trailingSamplesByHours(samples, 3)
	if len(recent) < 6 {
		return models.TimeMachineForecast{}, false
	}

	current := recent[len(recent)-1].MemoryPercent
	if current < 70 {
		return models.TimeMachineForecast{}, false
	}

	slopePerHour, confidence := regressionSlopePercentPerHour(recent, func(sample models.TimeMachineSample) float64 {
		return sample.MemoryPercent
	})
	if slopePerHour <= 1.5 {
		return models.TimeMachineForecast{}, false
	}

	hoursTo90 := (90 - current) / slopePerHour
	if hoursTo90 <= 0 || hoursTo90 > 24 {
		return models.TimeMachineForecast{}, false
	}

	predictedAt := time.UnixMilli(recent[len(recent)-1].Timestamp).Add(time.Duration(hoursTo90 * float64(time.Hour))).UnixMilli()
	return models.TimeMachineForecast{
		ID:             "memory-pressure",
		Kind:           "memory-pressure",
		Severity:       forecastSeverity(hoursTo90),
		Title:          "Memory pressure forecast",
		Summary:        fmt.Sprintf("If the current memory trend holds, RAM usage could cross 90%% in about %s.", humanDurationHours(hoursTo90)),
		PredictedAt:    predictedAt,
		CurrentValue:   current,
		ProjectedValue: 90,
		Confidence:     confidence,
		Unit:           "%",
	}, true
}

func trailingSamplesByHours(samples []models.TimeMachineSample, hours int) []models.TimeMachineSample {
	if len(samples) == 0 {
		return nil
	}

	cutoff := samples[len(samples)-1].Timestamp - int64((time.Duration(hours) * time.Hour / time.Millisecond))
	filtered := make([]models.TimeMachineSample, 0, len(samples))
	for _, sample := range samples {
		if sample.Timestamp >= cutoff {
			filtered = append(filtered, sample)
		}
	}
	return filtered
}

func regressionSlopeDiskGBPerHour(samples []models.TimeMachineSample) (float64, float64) {
	return regressionSlopePercentPerHour(samples, func(sample models.TimeMachineSample) float64 {
		return sample.DiskUsedGB
	})
}

func regressionSlopePercentPerHour(samples []models.TimeMachineSample, value func(models.TimeMachineSample) float64) (float64, float64) {
	if len(samples) < 2 {
		return 0, 0
	}

	var sumX, sumY, sumXY, sumXX float64
	start := float64(samples[0].Timestamp)
	for _, sample := range samples {
		x := (float64(sample.Timestamp) - start) / float64(time.Hour/time.Millisecond)
		y := value(sample)
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	n := float64(len(samples))
	denominator := n*sumXX - sumX*sumX
	if denominator == 0 {
		return 0, 0
	}

	slope := (n*sumXY - sumX*sumY) / denominator
	meanY := sumY / n
	var ssTot, ssRes float64
	intercept := (sumY - slope*sumX) / n
	for _, sample := range samples {
		x := (float64(sample.Timestamp) - start) / float64(time.Hour/time.Millisecond)
		y := value(sample)
		predicted := slope*x + intercept
		ssTot += math.Pow(y-meanY, 2)
		ssRes += math.Pow(y-predicted, 2)
	}

	confidence := 0.4
	if ssTot > 0 {
		confidence = math.Max(0.2, math.Min(0.98, 1-(ssRes/ssTot)))
	}
	return slope, confidence
}

func dominantProcess(sample models.TimeMachineSample, metric string) models.IncidentProcessSample {
	if len(sample.Processes) == 0 {
		return models.IncidentProcessSample{}
	}

	best := sample.Processes[0]
	bestScore := -1.0
	for _, proc := range sample.Processes {
		score := proc.CPUPercent + proc.MemoryMB/256
		if metric == "memory" {
			score = proc.MemoryMB + proc.CPUPercent*10
		}
		if metric == "disk" && isBlockedStatus(proc.Status) {
			score += 50
		}
		if score > bestScore {
			bestScore = score
			best = proc
		}
	}
	return best
}

func spikeNarrative(metric string, proc models.IncidentProcessSample, current, baseline float64) (string, string) {
	activity := describeProcessActivity(proc.Name)
	if proc.PID == 0 {
		switch metric {
		case "network":
			return "Unattributed network burst", fmt.Sprintf("Network traffic jumped well above its %.1f baseline.", baseline/1024)
		case "disk":
			return "Disk pressure spike", fmt.Sprintf("Disk usage surged well above its %.1f%% baseline.", baseline)
		default:
			return "CPU spike", fmt.Sprintf("CPU usage jumped well above its %.1f%% baseline.", baseline)
		}
	}

	switch metric {
	case "network":
		return fmt.Sprintf("Spike here was likely %s", activity), fmt.Sprintf("%s lined up with a network burst reaching %.1f KB/s.", friendlyProcessLabel(proc.Name), current/1024)
	case "disk":
		return fmt.Sprintf("Spike here was likely %s", activity), fmt.Sprintf("%s lined up with a disk-heavy period at %.1f%% disk usage.", friendlyProcessLabel(proc.Name), current)
	default:
		return fmt.Sprintf("Spike here was likely %s", activity), fmt.Sprintf("%s lined up with a CPU spike reaching %.1f%% versus a %.1f%% baseline.", friendlyProcessLabel(proc.Name), current, baseline)
	}
}

func buildTimeMachineSummary(samples []models.TimeMachineSample, annotations []models.TimeMachineAnnotation, forecasts []models.TimeMachineForecast) string {
	if len(samples) == 0 {
		return "Collecting historical samples for the time machine."
	}
	if len(annotations) == 0 && len(forecasts) == 0 {
		return "No major historical annotation or forecast stood out in the selected window."
	}

	parts := []string{}
	if len(annotations) > 0 {
		parts = append(parts, fmt.Sprintf("%d annotation%s in the selected history", len(annotations), pluralSuffix(len(annotations))))
	}
	if len(forecasts) > 0 {
		parts = append(parts, fmt.Sprintf("%d forecast%s available", len(forecasts), pluralSuffix(len(forecasts))))
	}
	return strings.Join(parts, " and ") + "."
}

func dedupeAnnotations(annotations []models.TimeMachineAnnotation) []models.TimeMachineAnnotation {
	result := []models.TimeMachineAnnotation{}
	seen := make(map[string]bool)
	for _, annotation := range annotations {
		key := fmt.Sprintf("%s-%d", annotation.Kind, annotation.Timestamp/(10*60*1000))
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, annotation)
	}
	return result
}

func friendlyProcessLabel(name string) string {
	switch {
	case isNodeFamily(name):
		return "Node.js"
	case strings.Contains(strings.ToLower(name), "docker"):
		return "Docker"
	default:
		return name
	}
}

func describeProcessActivity(name string) string {
	name = strings.ToLower(name)
	switch {
	case strings.Contains(name, "docker") || strings.Contains(name, "buildkit") || strings.Contains(name, "containerd"):
		return "Docker build activity"
	case isNodeFamily(name):
		return "Node.js workload"
	case strings.Contains(name, "go") || strings.Contains(name, "cargo") || strings.Contains(name, "rustc") || strings.Contains(name, "gcc") || strings.Contains(name, "clang"):
		return "a local build or compile"
	case strings.Contains(name, "chrome") || strings.Contains(name, "firefox") || strings.Contains(name, "brave"):
		return "browser activity"
	case strings.Contains(name, "postgres") || strings.Contains(name, "mysql") || strings.Contains(name, "redis"):
		return "database activity"
	case strings.Contains(name, "git"):
		return "Git activity"
	default:
		return fmt.Sprintf("%s activity", name)
	}
}

func isNodeFamily(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "node") || strings.Contains(name, "npm") || strings.Contains(name, "pnpm") || strings.Contains(name, "yarn") || strings.Contains(name, "bun") || strings.Contains(name, "deno")
}

func leakSeverity(growth, finalMB float64) string {
	if growth > 1500 || finalMB > 4000 {
		return "critical"
	}
	if growth > 800 || finalMB > 2000 {
		return "warning"
	}
	return "info"
}

func metricMinimumThreshold(metric string, baseline float64) float64 {
	switch metric {
	case "network":
		return math.Max(256*1024, baseline*0.7)
	case "disk":
		return math.Max(10, baseline*0.15)
	default:
		return math.Max(15, baseline*0.2)
	}
}

func spikeSeverity(metric string, value float64) string {
	switch metric {
	case "disk":
		if value > 92 {
			return "critical"
		}
	case "cpu":
		if value > 95 {
			return "critical"
		}
	case "network":
		if value > 8*1024*1024 {
			return "warning"
		}
	}
	return "info"
}

func forecastSeverity(hours float64) string {
	if hours <= 24 {
		return "critical"
	}
	if hours <= 72 {
		return "warning"
	}
	return "info"
}

func humanDurationHours(hours float64) string {
	if hours < 1 {
		minutes := int(math.Round(hours * 60))
		if minutes < 1 {
			minutes = 1
		}
		return fmt.Sprintf("%dm", minutes)
	}
	if hours < 48 {
		return fmt.Sprintf("%.1fh", hours)
	}
	return fmt.Sprintf("%.1fd", hours/24)
}

func sanitizeID(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.ReplaceAll(value, "/", "-")
	return value
}
