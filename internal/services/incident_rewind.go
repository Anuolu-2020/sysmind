package services

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"sysmind/internal/models"
)

const incidentProcessLimit = 6

type metricConfig struct {
	key           string
	category      string
	title         string
	metric        string
	hardThreshold float64
	minIncrease   float64
	value         func(models.IncidentSample) float64
}

// BuildIncidentSample creates a synchronized incident snapshot from current stats.
func BuildIncidentSample(stats *models.SystemStats, processes []models.ProcessInfo) models.IncidentSample {
	sample := models.IncidentSample{
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
		if ranked[i].CPUPercent != ranked[j].CPUPercent {
			return ranked[i].CPUPercent > ranked[j].CPUPercent
		}
		if ranked[i].MemoryMB != ranked[j].MemoryMB {
			return ranked[i].MemoryMB > ranked[j].MemoryMB
		}
		return ranked[i].NumThreads > ranked[j].NumThreads
	})

	limit := incidentProcessLimit
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

// AnalyzeIncidentRewind detects notable anomalies within the selected rewind window.
func AnalyzeIncidentRewind(samples []models.IncidentSample, minutes int) models.IncidentRewind {
	rewind := models.IncidentRewind{
		WindowMinutes:     minutes,
		ResolutionSeconds: incidentResolutionSeconds(samples),
		Samples:           samples,
		Findings:          []models.IncidentFinding{},
		Summary:           "Collecting incident history.",
	}

	if len(samples) == 0 {
		return rewind
	}

	findings := detectMetricFindings(samples)
	sort.Slice(findings, func(i, j int) bool {
		if severityWeight(findings[i].Severity) != severityWeight(findings[j].Severity) {
			return severityWeight(findings[i].Severity) > severityWeight(findings[j].Severity)
		}
		if findings[i].Confidence != findings[j].Confidence {
			return findings[i].Confidence > findings[j].Confidence
		}
		return findings[i].StartedAt < findings[j].StartedAt
	})

	rewind.Findings = findings
	if len(findings) > 0 {
		rewind.HighlightedAt = findings[0].StartedAt
		rewind.Summary = fmt.Sprintf(
			"%d finding%s detected. Earliest onset: %s.",
			len(findings),
			pluralSuffix(len(findings)),
			formatIncidentTime(findings[0].StartedAt),
		)
		return rewind
	}

	rewind.Summary = fmt.Sprintf("No material anomaly detected in the last %d minutes.", minutes)
	return rewind
}

func detectMetricFindings(samples []models.IncidentSample) []models.IncidentFinding {
	configs := []metricConfig{
		{
			key:           "cpu",
			category:      "performance",
			title:         "CPU spike detected",
			metric:        "cpu",
			hardThreshold: 80,
			minIncrease:   18,
			value: func(sample models.IncidentSample) float64 {
				return sample.CPUPercent
			},
		},
		{
			key:           "memory",
			category:      "performance",
			title:         "Memory pressure detected",
			metric:        "memory",
			hardThreshold: 85,
			minIncrease:   12,
			value: func(sample models.IncidentSample) float64 {
				return sample.MemoryPercent
			},
		},
		{
			key:           "disk",
			category:      "performance",
			title:         "Disk pressure detected",
			metric:        "disk",
			hardThreshold: 80,
			minIncrease:   10,
			value: func(sample models.IncidentSample) float64 {
				return sample.DiskPercent
			},
		},
		{
			key:           "network",
			category:      "network",
			title:         "Network surge detected",
			metric:        "network",
			hardThreshold: 512 * 1024,
			minIncrease:   256 * 1024,
			value: func(sample models.IncidentSample) float64 {
				return sample.NetUploadSpeed + sample.NetDownSpeed
			},
		},
	}

	findings := make([]models.IncidentFinding, 0, len(configs))
	for _, config := range configs {
		finding, ok := detectMetricFinding(samples, config)
		if ok {
			findings = append(findings, finding)
		}
	}

	return findings
}

func detectMetricFinding(samples []models.IncidentSample, config metricConfig) (models.IncidentFinding, bool) {
	values := make([]float64, 0, len(samples))
	for _, sample := range samples {
		values = append(values, config.value(sample))
	}

	baseline, stdDev := robustBaseline(values)
	threshold := math.Max(config.hardThreshold, baseline+2.5*math.Max(stdDev, 1))
	threshold = math.Max(threshold, baseline+config.minIncrease)

	startIndex := -1
	for idx := range samples {
		current := config.value(samples[idx])
		if current < threshold {
			continue
		}
		if idx+1 < len(samples) && config.value(samples[idx+1]) < threshold*0.9 {
			continue
		}
		startIndex = idx
		break
	}
	if startIndex == -1 {
		return models.IncidentFinding{}, false
	}

	peakIndex := startIndex
	peakValue := config.value(samples[startIndex])
	for idx := startIndex; idx < len(samples); idx++ {
		current := config.value(samples[idx])
		if current > peakValue {
			peakIndex = idx
			peakValue = current
		}
	}

	if peakValue < threshold {
		return models.IncidentFinding{}, false
	}

	startValue := config.value(samples[startIndex])
	culprit, confidence := identifyCulprit(samples, startIndex, peakIndex, config.key)
	severity := findingSeverity(config.key, peakValue, threshold)

	finding := models.IncidentFinding{
		ID:                  fmt.Sprintf("%s-%d", config.key, samples[startIndex].Timestamp),
		Category:            config.category,
		Severity:            severity,
		Title:               config.title,
		Summary:             findingSummary(config.key, baseline, startValue, peakValue, culprit),
		Metric:              config.metric,
		StartedAt:           samples[startIndex].Timestamp,
		PeakAt:              samples[peakIndex].Timestamp,
		StartValue:          startValue,
		PeakValue:           peakValue,
		CulpritPID:          culprit.PID,
		CulpritName:         culprit.Name,
		CulpritCPUPercent:   culprit.CPUPercent,
		CulpritMemoryMB:     culprit.MemoryMB,
		CulpritThreads:      culprit.NumThreads,
		CulpritStatus:       culprit.Status,
		ThreadHint:          buildThreadHint(config.key, culprit),
		SyscallHint:         buildSyscallHint(config.key, culprit),
		Confidence:          confidence,
		ExactTraceAvailable: false,
	}

	return finding, true
}

func robustBaseline(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}

	sortedValues := make([]float64, len(values))
	copy(sortedValues, values)
	sort.Float64s(sortedValues)

	limit := int(math.Ceil(float64(len(sortedValues)) * 0.6))
	if limit < 3 {
		limit = len(sortedValues)
	}

	baseSlice := sortedValues[:limit]
	mean := average(baseSlice)
	var variance float64
	for _, value := range baseSlice {
		diff := value - mean
		variance += diff * diff
	}

	return mean, math.Sqrt(variance / float64(len(baseSlice)))
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func identifyCulprit(samples []models.IncidentSample, startIndex, peakIndex int, metric string) (models.IncidentProcessSample, float64) {
	target := samples[peakIndex]
	previousIndex := startIndex - 1
	if previousIndex < 0 {
		previousIndex = startIndex
	}
	previousMap := make(map[int32]models.IncidentProcessSample)
	for _, proc := range samples[previousIndex].Processes {
		previousMap[proc.PID] = proc
	}

	bestScore := 0.0
	best := models.IncidentProcessSample{}
	for _, proc := range target.Processes {
		prev := previousMap[proc.PID]
		cpuDelta := math.Max(proc.CPUPercent-prev.CPUPercent, 0)
		memDelta := math.Max(proc.MemoryMB-prev.MemoryMB, 0)
		threadDelta := math.Max(float64(proc.NumThreads-prev.NumThreads), 0)

		score := proc.CPUPercent*1.2 + cpuDelta*1.8 + proc.MemoryMB/256 + memDelta/96 + threadDelta*1.5
		switch metric {
		case "memory":
			score += proc.MemoryMB/96 + memDelta/64
		case "disk":
			if isBlockedStatus(proc.Status) {
				score += 40
			}
		case "network":
			score += cpuDelta * 0.5
		}

		if isBlockedStatus(proc.Status) {
			score += 15
		}

		if score > bestScore {
			bestScore = score
			best = proc
		}
	}

	if bestScore < 15 {
		return models.IncidentProcessSample{}, 0.35
	}

	confidence := math.Min(0.98, 0.45+bestScore/220)
	return best, confidence
}

func buildThreadHint(metric string, proc models.IncidentProcessSample) string {
	if proc.PID == 0 {
		return "No single culprit process stood out in the stored snapshots."
	}
	if isBlockedStatus(proc.Status) {
		return fmt.Sprintf("%s was in blocked I/O wait state around the incident; its threads may have been stalled on disk or device work.", proc.Name)
	}
	if proc.NumThreads >= 64 {
		return fmt.Sprintf("%s had %d threads in the captured snapshot, which points to thread fan-out or contention rather than a single idle worker.", proc.Name, proc.NumThreads)
	}
	if metric == "cpu" {
		return fmt.Sprintf("%s became CPU-bound around the onset; the stall is more likely contention or runaway work than a sleeping thread.", proc.Name)
	}
	if metric == "memory" {
		return fmt.Sprintf("%s was the clearest memory hotspot in the captured snapshot.", proc.Name)
	}
	return fmt.Sprintf("%s was the strongest process-level correlate in the captured snapshot.", proc.Name)
}

func buildSyscallHint(metric string, proc models.IncidentProcessSample) string {
	base := "Exact syscall attribution is unavailable in the current build without OS tracing such as eBPF, ETW, DTrace, or strace."
	if proc.PID == 0 {
		return base
	}
	if metric == "disk" || isBlockedStatus(proc.Status) {
		return base + " The strongest hint here is blocking file or device I/O."
	}
	if metric == "network" {
		return base + " The strongest hint here is socket or network I/O."
	}
	if metric == "memory" {
		return base + " The strongest hint here is memory allocation or paging pressure."
	}
	return base
}

func findingSummary(metric string, baseline, startValue, peakValue float64, culprit models.IncidentProcessSample) string {
	switch metric {
	case "network":
		summary := fmt.Sprintf("Combined network throughput climbed from a %.1f KB/s baseline to %.1f KB/s, peaking at %.1f KB/s.", baseline/1024, startValue/1024, peakValue/1024)
		if culprit.PID != 0 {
			summary += fmt.Sprintf(" %s was the closest process-level correlate.", culprit.Name)
		}
		return summary
	case "memory":
		summary := fmt.Sprintf("Memory usage moved from a %.1f%% baseline to %.1f%% and peaked at %.1f%%.", baseline, startValue, peakValue)
		if culprit.PID != 0 {
			summary += fmt.Sprintf(" %s was holding %.1f MB in the captured snapshot.", culprit.Name, culprit.MemoryMB)
		}
		return summary
	default:
		summary := fmt.Sprintf("%s rose from a %.1f baseline to %.1f and peaked at %.1f.", strings.ToUpper(metric[:1])+metric[1:], baseline, startValue, peakValue)
		if culprit.PID != 0 {
			summary += fmt.Sprintf(" %s was the strongest process-level correlate.", culprit.Name)
		}
		return summary
	}
}

func findingSeverity(metric string, peakValue, threshold float64) string {
	if metric == "memory" && peakValue >= 95 {
		return "critical"
	}
	if metric == "cpu" && peakValue >= 95 {
		return "critical"
	}
	if peakValue >= threshold*1.25 {
		return "warning"
	}
	return "info"
}

func isBlockedStatus(status string) bool {
	return strings.Contains(strings.ToUpper(status), "D")
}

func incidentResolutionSeconds(samples []models.IncidentSample) int {
	if len(samples) < 2 {
		return 0
	}
	deltaMs := samples[1].Timestamp - samples[0].Timestamp
	if deltaMs <= 0 {
		return 0
	}
	return int(deltaMs / 1000)
}

func severityWeight(severity string) int {
	switch severity {
	case "critical":
		return 3
	case "warning":
		return 2
	default:
		return 1
	}
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func formatIncidentTime(ts int64) string {
	if ts == 0 {
		return "unknown"
	}
	return time.UnixMilli(ts).Format("15:04:05")
}
